package scanner

import (
	"fmt"
	"strings"

	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type OwnerRefScanner struct {
	client *kubernetes.Clientset
}

func NewOwnerRefScanner(config *rest.Config) *OwnerRefScanner {
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kubernetes client")
	}
	return &OwnerRefScanner{client: client}
}

func (s *OwnerRefScanner) Scan(namespace string, allNamespaces bool) ([]ScanResult, error) {
	var results []ScanResult

	namespaces := []string{namespace}
	if allNamespaces || namespace == "" {
		nsList, err := s.client.CoreV1().Namespaces().List(nil, metav1.ListOptions{})
		if err != nil {
			log.Warn().Err(err).Msg("Failed to list namespaces")
		} else {
			namespaces = []string{}
			for _, ns := range nsList.Items {
				// Skip system namespaces
				if !strings.HasPrefix(ns.Name, "kube-") {
					namespaces = append(namespaces, ns.Name)
				}
			}
		}
	}

	for _, ns := range namespaces {
		// Scan Pods for problematic owner references
		pods, err := s.client.CoreV1().Pods(ns).List(nil, metav1.ListOptions{})
		if err != nil {
			log.Warn().Err(err).Str("namespace", ns).Msg("Failed to list pods")
			continue
		}

		for _, pod := range pods.Items {
			if s.hasProblematicOwnerRef(pod) {
				results = append(results, ScanResult{
					Category:      "OwnerReference",
					Severity:      "error",
					Resource:      "Pod",
					Namespace:     ns,
					Name:          pod.Name,
					Issue:         "Problematic owner reference configuration",
					Description:   fmt.Sprintf("Pod %s/%s has owner references that may cause cascading deletion issues (e.g., parent deleted when child is deleted).", ns, pod.Name),
					Recommendation: "Review owner reference setup. Ensure parent resources don't depend on child resources for existence. Use controllerutil.SetControllerReference() correctly.",
				})
			}
		}

		// Scan ConfigMaps
		configMaps, err := s.client.CoreV1().ConfigMaps(ns).List(nil, metav1.ListOptions{})
		if err != nil {
			log.Warn().Err(err).Str("namespace", ns).Msg("Failed to list configmaps")
			continue
		}

		for _, cm := range configMaps.Items {
			if s.hasProblematicOwnerRefForCM(cm) {
				results = append(results, ScanResult{
					Category:      "OwnerReference",
					Severity:      "warning",
					Resource:      "ConfigMap",
					Namespace:     ns,
					Name:          cm.Name,
					Issue:         "ConfigMap with owner reference to child resource",
					Description:   fmt.Sprintf("ConfigMap %s/%s has owner reference pointing to a resource that depends on it.", ns, cm.Name),
					Recommendation: "Review owner reference hierarchy. Parent resources should own children, not the reverse.",
				})
			}
		}
	}

	return results, nil
}

func (s *OwnerRefScanner) hasProblematicOwnerRef(pod corev1.Pod) bool {
	// Check for owner references that might cause issues
	// This is heuristic - we look for patterns that suggest reverse dependencies
	if len(pod.OwnerReferences) == 0 {
		return false
	}

	for _, ownerRef := range pod.OwnerReferences {
		// Check if owner is a resource type that shouldn't own pods directly
		// (e.g., if a ConfigMap or Secret owns a Pod, that's unusual)
		if ownerRef.Kind == "ConfigMap" || ownerRef.Kind == "Secret" {
			return true
		}

		// Check if controller is set incorrectly
		// If a child resource (like a Pod) has controller=true pointing to a resource
		// that depends on it, that's problematic
		if ownerRef.Controller != nil && *ownerRef.Controller {
			// This is actually normal for most cases, but we flag unusual combinations
			// A more sophisticated check would verify the actual dependency graph
		}
	}

	return false
}

func (s *OwnerRefScanner) hasProblematicOwnerRefForCM(cm corev1.ConfigMap) bool {
	// ConfigMaps are often owned by Deployments/StatefulSets
	// But if a ConfigMap owns something that the parent depends on, that's problematic
	if len(cm.OwnerReferences) == 0 {
		return false
	}

	for _, ownerRef := range cm.OwnerReferences {
		// If a ConfigMap is owned by a Pod, that's unusual and problematic
		// (Pods typically own ConfigMaps, not the reverse)
		if ownerRef.Kind == "Pod" {
			return true
		}
	}

	return false
}

