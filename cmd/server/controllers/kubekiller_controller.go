package controllers

import (
	"context"
	"fmt"
	"time"

	"github.com/p-program/kube-killer/cmd/killer"
	"github.com/p-program/kube-killer/cmd/server/api/v1alpha1"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// KubeKillerReconciler reconciles a KubeKiller object
type KubeKillerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=kubekiller.p-program.github.io,resources=kubekillers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=kubekiller.p-program.github.io,resources=kubekillers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=kubekiller.p-program.github.io,resources=kubekillers/finalizers,verbs=update
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;delete
//+kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;delete
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch;delete
//+kubebuilder:rbac:groups="",resources=secrets,verbs=get;list;watch;delete
//+kubebuilder:rbac:groups="",resources=persistentvolumeclaims,verbs=get;list;watch;delete
//+kubebuilder:rbac:groups="",resources=persistentvolumes,verbs=get;list;watch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;delete
//+kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;delete
//+kubebuilder:rbac:groups=batch,resources=jobs,verbs=get;list;watch;delete
//+kubebuilder:rbac:groups="",resources=namespaces,verbs=get;list;watch

// Reconcile is part of the main kubernetes reconciliation loop
func (r *KubeKillerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	var kk v1alpha1.KubeKiller
	if err := r.Get(ctx, req.NamespacedName, &kk); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// Check if scheduleAt is set (specific time point execution)
	if kk.Spec.ScheduleAt != nil {
		now := time.Now()
		scheduleTime := kk.Spec.ScheduleAt.Time

		// Check if task has already been executed
		if kk.Status.LastRunTime != nil {
			lastRunTime := kk.Status.LastRunTime.Time
			// If last run time is at or after the scheduled time, task has been executed
			if !lastRunTime.Before(scheduleTime) {
				log.Info().Msgf("Scheduled task at %s has already been executed at %s, skipping",
					scheduleTime.Format(time.RFC3339), lastRunTime.Format(time.RFC3339))
				return ctrl.Result{}, nil
			}
		}

		// If the scheduled time hasn't arrived yet, requeue
		if now.Before(scheduleTime) {
			requeueAfter := scheduleTime.Sub(now)
			log.Info().Msgf("Scheduled task will run at %s, requeuing in %v", scheduleTime.Format(time.RFC3339), requeueAfter)
			return ctrl.Result{RequeueAfter: requeueAfter}, nil
		}

		// Scheduled time has arrived, execute the task
		log.Info().Msgf("Scheduled time %s has arrived, executing task", scheduleTime.Format(time.RFC3339))
	} else {
		// Use interval-based execution (original behavior)
		interval, err := time.ParseDuration(kk.Spec.Interval)
		if err != nil {
			// Default to 5 minutes if parsing fails
			interval = 5 * time.Minute
			log.Warn().Err(err).Msgf("Failed to parse interval, using default 5m")
		}

		// Check if we should run now
		shouldRun := true
		if kk.Status.LastRunTime != nil {
			timeSinceLastRun := time.Since(kk.Status.LastRunTime.Time)
			if timeSinceLastRun < interval {
				shouldRun = false
				// Requeue after the remaining time
				requeueAfter := interval - timeSinceLastRun
				return ctrl.Result{RequeueAfter: requeueAfter}, nil
			}
		}

		if !shouldRun {
			return ctrl.Result{RequeueAfter: interval}, nil
		}
	}

	// Update status to Running
	kk.Status.Phase = "Running"
	now := metav1.Now()
	kk.Status.LastRunTime = &now

	// Execute based on mode
	var result string
	var killed int
	var err2 error

	switch kk.Spec.Mode {
	case "demon":
		result, killed, err2 = r.runDemonMode(ctx, &kk)
	case "illidan":
		result, killed, err2 = r.runIllidanMode(ctx, &kk)
	default:
		result = fmt.Sprintf("Unknown mode: %s", kk.Spec.Mode)
		err2 = fmt.Errorf("unknown mode: %s", kk.Spec.Mode)
	}

	kk.Status.LastRunResult = result
	kk.Status.ResourcesKilled = killed
	if err2 != nil {
		kk.Status.Phase = "Error"
		kk.Status.LastRunResult = fmt.Sprintf("%s: %v", result, err2)
	} else {
		kk.Status.Phase = "Ready"
	}

	// Update status
	if err := r.Status().Update(ctx, &kk); err != nil {
		return ctrl.Result{}, err
	}

	// If scheduleAt was set and task has been executed, don't requeue
	if kk.Spec.ScheduleAt != nil {
		log.Info().Msg("Scheduled task completed, not requeuing")
		return ctrl.Result{}, err2
	}

	// Requeue after interval for interval-based execution
	interval, _ := time.ParseDuration(kk.Spec.Interval)
	if interval == 0 {
		interval = 5 * time.Minute
	}
	return ctrl.Result{RequeueAfter: interval}, err2
}

// runDemonMode runs in demon mode - kills ALL pods at every period
func (r *KubeKillerReconciler) runDemonMode(ctx context.Context, kk *v1alpha1.KubeKiller) (string, int, error) {
	log.Info().Msg("Running in DEMON mode - killing ALL pods")

	namespaces := r.getNamespaces(ctx, kk)
	totalKilled := 0

	for _, ns := range namespaces {
		if ns == "kube-system" {
			continue // Skip kube-system
		}

		k, err := killer.NewPodKiller(ns)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to create PodKiller for namespace %s", ns)
			continue
		}

		if kk.Spec.DryRun {
			k.DryRun()
		}

		// In demon mode, kill all pods
		k.BlackHand()
		if err := k.Kill(); err != nil {
			log.Error().Err(err).Msgf("Failed to kill pods in namespace %s", ns)
			continue
		}

		// Note: This is a simplified count, actual implementation would need to track deletions
		// For demon mode, we assume at least one pod was killed per namespace
		totalKilled++
	}

	result := fmt.Sprintf("Demon mode: killed pods in %d namespaces", len(namespaces))
	return result, totalKilled, nil
}

// runIllidanMode runs in Illidan mode - hunts unhealthy resources
func (r *KubeKillerReconciler) runIllidanMode(ctx context.Context, kk *v1alpha1.KubeKiller) (string, int, error) {
	log.Info().Msg("Running in ILLIDAN mode - hunting unhealthy resources")

	namespaces := r.getNamespaces(ctx, kk)
	resources := kk.Spec.Resources
	if len(resources) == 0 {
		// Default resources to kill in Illidan mode
		resources = []string{"pod", "job", "pvc", "pv", "service", "configmap", "secret"}
	}

	totalKilled := 0
	var errors []error

	for _, ns := range namespaces {
		if ns == "kube-system" {
			continue // Skip kube-system
		}

		for _, resourceType := range resources {
			killed, err := r.killResource(ctx, resourceType, ns, kk.Spec.DryRun)
			if err != nil {
				log.Error().Err(err).Msgf("Failed to kill %s in namespace %s", resourceType, ns)
				errors = append(errors, err)
				continue
			}
			totalKilled += killed
		}
	}

	result := fmt.Sprintf("Illidan mode: killed %d resources across %d namespaces", totalKilled, len(namespaces))
	if len(errors) > 0 {
		return result, totalKilled, fmt.Errorf("encountered %d errors", len(errors))
	}
	return result, totalKilled, nil
}

// killResource kills a specific resource type in a namespace
func (r *KubeKillerReconciler) killResource(ctx context.Context, resourceType, namespace string, dryRun bool) (int, error) {
	var err error
	killed := 0

	switch resourceType {
	case "pod", "p", "po":
		k, err := killer.NewPodKiller(namespace)
		if err != nil {
			return 0, err
		}
		if dryRun {
			k.DryRun()
		}
		err = k.Kill()
		if err == nil {
			killed = 1
		}
	case "job", "jobs":
		k, err := killer.NewJobKiller(namespace)
		if err != nil {
			return 0, err
		}
		if dryRun {
			k.DryRun()
		}
		err = k.Kill()
		if err == nil {
			killed = 1
		}
	case "pvc":
		k, err := killer.NewPVCKiller(namespace)
		if err != nil {
			return 0, err
		}
		if dryRun {
			k.DryRun()
		}
		err = k.Kill()
		if err == nil {
			killed = 1
		}
	case "pv":
		k, err := killer.NewPVKiller()
		if err != nil {
			return 0, err
		}
		if dryRun {
			k.DryRun()
		}
		err = k.Kill()
		if err == nil {
			killed = 1
		}
	case "service", "svc", "s":
		k, err := killer.NewServiceKiller(namespace)
		if err != nil {
			return 0, err
		}
		if dryRun {
			k.DryRun()
		}
		err = k.Kill()
		if err == nil {
			killed = 1
		}
	case "configmap", "cm":
		k, err := killer.NewConfigmapKiller(namespace)
		if err != nil {
			return 0, err
		}
		if dryRun {
			k.DryRun()
		}
		err = k.Kill()
		if err == nil {
			killed = 1
		}
	case "secret", "secrets":
		k, err := killer.NewSecretKiller(namespace)
		if err != nil {
			return 0, err
		}
		if dryRun {
			k.DryRun()
		}
		err = k.Kill()
		if err == nil {
			killed = 1
		}
	default:
		return 0, fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	return killed, err
}

// getNamespaces returns the list of namespaces to operate on
func (r *KubeKillerReconciler) getNamespaces(ctx context.Context, kk *v1alpha1.KubeKiller) []string {
	if len(kk.Spec.Namespaces) > 0 {
		return kk.Spec.Namespaces
	}

	// Get all namespaces
	nsList := &corev1.NamespaceList{}
	if err := r.Client.List(ctx, nsList); err != nil {
		log.Error().Err(err).Msg("Failed to list namespaces")
		return []string{"default"}
	}

	namespaces := make([]string, 0)
	excludeMap := make(map[string]bool)
	for _, exclude := range kk.Spec.ExcludeNamespaces {
		excludeMap[exclude] = true
	}

	for _, ns := range nsList.Items {
		if ns.Name == "kube-system" || excludeMap[ns.Name] {
			continue
		}
		namespaces = append(namespaces, ns.Name)
	}

	return namespaces
}

// SetupWithManager sets up the controller with the Manager.
func (r *KubeKillerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&v1alpha1.KubeKiller{}).
		Complete(r)
}
