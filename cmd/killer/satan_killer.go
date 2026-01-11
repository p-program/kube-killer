package killer

import (
	"context"
	"fmt"

	"github.com/p-program/kube-killer/core"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// SatanKiller aggregates all resource killers to delete all excess/unused Kubernetes resources
// This is a comprehensive cleanup operation that removes:
// - Non-running pods (Completed, Failed, Evicted, etc.)
// - Completed/Failed jobs
// - Unused ConfigMaps
// - Unused Secrets
// - Services without matching pods
// - Unused PVCs
// - Unused PVs (Available, Released, Failed)
type SatanKiller struct {
	client       *kubernetes.Clientset
	namespace    string
	dryRun       bool
	allNamespaces bool
}

// NewSatanKiller creates a new SatanKiller instance
// namespace can be "" for all namespaces
func NewSatanKiller(namespace string, allNamespaces bool) (*SatanKiller, error) {
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	k := SatanKiller{
		client:        clientset,
		namespace:     namespace,
		allNamespaces: allNamespaces,
	}

	return &k, nil
}

// DryRun enables dry-run mode
func (k *SatanKiller) DryRun() *SatanKiller {
	k.dryRun = true
	return k
}

// Kill executes all resource killers in sequence to delete excess/unused resources
func (k *SatanKiller) Kill() error {
	log.Warn().Msg("!!!WARNING!!!: SATAN mode - This will delete ALL excess/unused Kubernetes resources!")
	log.Info().Msg("Starting comprehensive cleanup of excess/unused Kubernetes resources...")

	var totalErrors []error
	namespaces := []string{k.namespace}

	// Get list of namespaces to process
	if k.allNamespaces {
		nsList, err := k.client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
		if err != nil {
			return fmt.Errorf("failed to list namespaces: %w", err)
		}
		namespaces = []string{}
		for _, ns := range nsList.Items {
			// Skip kube-system
			if ns.Name != "kube-system" {
				namespaces = append(namespaces, ns.Name)
			}
		}
	}

	// Process each namespace
	for _, ns := range namespaces {
		if ns == "" {
			ns = "default"
		}
		log.Info().Msgf("Processing namespace: %s", ns)

		// 1. Kill non-running pods
		if err := k.killPods(ns); err != nil {
			log.Error().Err(err).Msgf("Failed to kill pods in namespace %s", ns)
			totalErrors = append(totalErrors, fmt.Errorf("pods in %s: %w", ns, err))
		}

		// 2. Kill completed/failed jobs
		if err := k.killJobs(ns); err != nil {
			log.Error().Err(err).Msgf("Failed to kill jobs in namespace %s", ns)
			totalErrors = append(totalErrors, fmt.Errorf("jobs in %s: %w", ns, err))
		}

		// 3. Kill unused ConfigMaps
		if err := k.killConfigMaps(ns); err != nil {
			log.Error().Err(err).Msgf("Failed to kill ConfigMaps in namespace %s", ns)
			totalErrors = append(totalErrors, fmt.Errorf("configmaps in %s: %w", ns, err))
		}

		// 4. Kill unused Secrets
		if err := k.killSecrets(ns); err != nil {
			log.Error().Err(err).Msgf("Failed to kill Secrets in namespace %s", ns)
			totalErrors = append(totalErrors, fmt.Errorf("secrets in %s: %w", ns, err))
		}

		// 5. Kill services without matching pods
		if err := k.killServices(ns); err != nil {
			log.Error().Err(err).Msgf("Failed to kill Services in namespace %s", ns)
			totalErrors = append(totalErrors, fmt.Errorf("services in %s: %w", ns, err))
		}

		// 6. Kill unused PVCs
		if err := k.killPVCs(ns); err != nil {
			log.Error().Err(err).Msgf("Failed to kill PVCs in namespace %s", ns)
			totalErrors = append(totalErrors, fmt.Errorf("pvcs in %s: %w", ns, err))
		}
	}

	// 7. Kill unused PVs (cluster-scoped, only once)
	if err := k.killPVs(); err != nil {
		log.Error().Err(err).Msg("Failed to kill PVs")
		totalErrors = append(totalErrors, fmt.Errorf("pvs: %w", err))
	}

	if len(totalErrors) > 0 {
		log.Warn().Msgf("Completed with %d error(s)", len(totalErrors))
		return fmt.Errorf("encountered %d error(s) during cleanup", len(totalErrors))
	}

	log.Info().Msg("Successfully completed comprehensive cleanup of excess/unused Kubernetes resources")
	return nil
}

func (k *SatanKiller) killPods(namespace string) error {
	podKiller, err := NewPodKiller(namespace)
	if err != nil {
		return err
	}
	if k.dryRun {
		podKiller.DryRun()
	}
	log.Info().Msgf("Killing non-running pods in namespace %s", namespace)
	return podKiller.Kill()
}

func (k *SatanKiller) killJobs(namespace string) error {
	jobKiller, err := NewJobKiller(namespace)
	if err != nil {
		return err
	}
	if k.dryRun {
		jobKiller.DryRun()
	}
	log.Info().Msgf("Killing completed/failed jobs in namespace %s", namespace)
	return jobKiller.Kill()
}

func (k *SatanKiller) killConfigMaps(namespace string) error {
	configMapKiller, err := NewConfigmapKiller(namespace)
	if err != nil {
		return err
	}
	if k.dryRun {
		configMapKiller.DryRun()
	}
	log.Info().Msgf("Killing unused ConfigMaps in namespace %s", namespace)
	return configMapKiller.Kill()
}

func (k *SatanKiller) killSecrets(namespace string) error {
	secretKiller, err := NewSecretKiller(namespace)
	if err != nil {
		return err
	}
	if k.dryRun {
		secretKiller.DryRun()
	}
	log.Info().Msgf("Killing unused Secrets in namespace %s", namespace)
	return secretKiller.Kill()
}

func (k *SatanKiller) killServices(namespace string) error {
	serviceKiller, err := NewServiceKiller(namespace)
	if err != nil {
		return err
	}
	if k.dryRun {
		serviceKiller.DryRun()
	}
	log.Info().Msgf("Killing services without matching pods in namespace %s", namespace)
	return serviceKiller.Kill()
}

func (k *SatanKiller) killPVCs(namespace string) error {
	pvcKiller, err := NewPVCKiller(namespace)
	if err != nil {
		return err
	}
	if k.dryRun {
		pvcKiller.DryRun()
	}
	log.Info().Msgf("Killing unused PVCs in namespace %s", namespace)
	return pvcKiller.Kill()
}

func (k *SatanKiller) killPVs() error {
	pvKiller, err := NewPVKiller()
	if err != nil {
		return err
	}
	if k.dryRun {
		pvKiller.DryRun()
	}
	log.Info().Msg("Killing unused PVs (cluster-scoped)")
	return pvKiller.Kill()
}
