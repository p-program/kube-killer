package scanner

import (
	"fmt"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type ControllerScanner struct {
	client *kubernetes.Clientset
}

func NewControllerScanner(config *rest.Config) *ControllerScanner {
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kubernetes client")
	}
	return &ControllerScanner{client: client}
}

func (s *ControllerScanner) Scan(namespace string, allNamespaces bool) ([]ScanResult, error) {
	var results []ScanResult

	// Scan Deployments (common place for operators/controllers)
	namespaces := []string{namespace}
	if allNamespaces || namespace == "" {
		nsList, err := s.client.CoreV1().Namespaces().List(nil, metav1.ListOptions{})
		if err != nil {
			log.Warn().Err(err).Msg("Failed to list namespaces")
		} else {
			namespaces = []string{}
			for _, ns := range nsList.Items {
				// Skip system namespaces
				if !strings.HasPrefix(ns.Name, "kube-") && ns.Name != "default" {
					namespaces = append(namespaces, ns.Name)
				}
			}
		}
	}

	for _, ns := range namespaces {
		// Scan Deployments
		deployments, err := s.client.AppsV1().Deployments(ns).List(nil, metav1.ListOptions{})
		if err != nil {
			log.Warn().Err(err).Str("namespace", ns).Msg("Failed to list deployments")
			continue
		}

		for _, deploy := range deployments.Items {
			// Check for potential reconciliation loop patterns
			// This is heuristic - we look for common patterns in annotations/labels
			if s.hasPotentialReconcileLoop(deploy) {
				results = append(results, ScanResult{
					Category:      "Controller",
					Severity:      "warning",
					Resource:      "Deployment",
					Namespace:     ns,
					Name:          deploy.Name,
					Issue:         "Potential reconciliation loop pattern",
					Description:   fmt.Sprintf("Deployment %s/%s may have a reconciliation loop (updating itself in Reconcile).", ns, deploy.Name),
					Recommendation: "Review controller code: avoid calling Update() on the same resource that triggered Reconcile. Use Patch() with proper comparison.",
				})
			}

			// Check for excessive event generation
			if s.hasExcessiveEventGeneration(deploy) {
				results = append(results, ScanResult{
					Category:      "Controller",
					Severity:      "warning",
					Resource:      "Deployment",
					Namespace:     ns,
					Name:          deploy.Name,
					Issue:         "Potential excessive event generation",
					Description:   fmt.Sprintf("Deployment %s/%s may generate events on every reconcile without checking for actual changes.", ns, deploy.Name),
					Recommendation: "Only emit events when status actually changes. Use reflect.DeepEqual() to compare old and new status before emitting events.",
				})
			}
		}

		// Scan Jobs for long requeue times
		jobs, err := s.client.BatchV1().Jobs(ns).List(nil, metav1.ListOptions{})
		if err != nil {
			log.Warn().Err(err).Str("namespace", ns).Msg("Failed to list jobs")
			continue
		}

		for _, job := range jobs.Items {
			if s.hasLongRequeueTime(job) {
				results = append(results, ScanResult{
					Category:      "Controller",
					Severity:      "error",
					Resource:      "Job",
					Namespace:     ns,
					Name:          job.Name,
					Issue:         "Job with extremely long requeue time",
					Description:   fmt.Sprintf("Job %s/%s may have RequeueAfter set to an extremely long duration (e.g., 1000000000 years).", ns, job.Name),
					Recommendation: "Review controller code and set reasonable RequeueAfter values (seconds, minutes, or hours, not years).",
				})
			}
		}
	}

	return results, nil
}

func (s *ControllerScanner) hasPotentialReconcileLoop(deploy appsv1.Deployment) bool {
	// Heuristic: Look for labels/annotations that suggest self-updating
	// This is a simplified check - real detection would require code analysis
	if deploy.Labels != nil {
		// Check for labels that are updated frequently
		if val, ok := deploy.Labels["lastSync"]; ok && val != "" {
			return true
		}
		if val, ok := deploy.Labels["lastReconcile"]; ok && val != "" {
			return true
		}
	}
	return false
}

func (s *ControllerScanner) hasExcessiveEventGeneration(deploy appsv1.Deployment) bool {
	// This is hard to detect without code analysis
	// We can check for annotations that suggest event generation
	if deploy.Annotations != nil {
		// If there are many event-related annotations, it might indicate excessive events
		eventCount := 0
		for key := range deploy.Annotations {
			if strings.Contains(key, "event") || strings.Contains(key, "Event") {
				eventCount++
			}
		}
		return eventCount > 5
	}
	return false
}

func (s *ControllerScanner) hasLongRequeueTime(job batchv1.Job) bool {
	// Check job annotations for requeue information
	// This is heuristic - real detection requires code analysis
	if job.Annotations != nil {
		if val, ok := job.Annotations["requeueAfter"]; ok {
			// Try to parse duration
			if d, err := time.ParseDuration(val); err == nil {
				// Check if it's unreasonably long (more than 1 year)
				if d > 365*24*time.Hour {
					return true
				}
			}
		}
	}
	return false
}

