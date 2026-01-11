package killer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/p-program/kube-killer/core"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
)

type NamespaceKiller struct {
	client        *kubernetes.Clientset
	dynamicClient dynamic.Interface
	deleteOption  metav1.DeleteOptions
	dryRun        bool
	mafia         bool
	namespace     string
	force         bool
}

func NewNamespaceKiller(namespace string) (*NamespaceKiller, error) {
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	k := NamespaceKiller{
		namespace:     namespace,
		client:        clientset,
		dynamicClient: dynamicClient,
	}
	var gracePeriodSeconds int64 = 1
	k.deleteOption = metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}
	return &k, nil
}

func (k *NamespaceKiller) DryRun() *NamespaceKiller {
	k.dryRun = true
	k.deleteOption.DryRun = []string{"All"}
	return k
}

func (k *NamespaceKiller) BlackHand() *NamespaceKiller {
	k.mafia = true
	return k
}

// Force sets force deletion mode, which will remove finalizers if namespace is stuck
func (k *NamespaceKiller) Force() *NamespaceKiller {
	k.force = true
	return k
}

// Kill deletes the namespace, handling stuck Terminating namespaces
func (k *NamespaceKiller) Kill() error {
	ctx := context.TODO()

	// Get namespace to check its status
	ns, err := k.client.CoreV1().Namespaces().Get(ctx, k.namespace, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get namespace %s: %w", k.namespace, err)
	}

	// Check if namespace is already in Terminating state
	if ns.Status.Phase == v1.NamespaceTerminating {
		log.Warn().Msgf("Namespace %s is already in Terminating state", k.namespace)
		return k.handleTerminatingNamespace(ctx, ns)
	}

	// If namespace is not in Terminating, try normal deletion
	if !k.dryRun {
		log.Info().Msgf("Deleting namespace %s", k.namespace)
		err = k.client.CoreV1().Namespaces().Delete(ctx, k.namespace, k.deleteOption)
		if err != nil {
			return fmt.Errorf("failed to delete namespace %s: %w", k.namespace, err)
		}

		// Wait a bit and check if it's stuck
		time.Sleep(2 * time.Second)
		ns, err = k.client.CoreV1().Namespaces().Get(ctx, k.namespace, metav1.GetOptions{})
		if err != nil {
			// Namespace might be deleted already
			log.Info().Msgf("Namespace %s deleted successfully", k.namespace)
			return nil
		}

		if ns.Status.Phase == v1.NamespaceTerminating {
			log.Warn().Msgf("Namespace %s is stuck in Terminating state", k.namespace)
			return k.handleTerminatingNamespace(ctx, ns)
		}
	} else {
		log.Info().Msgf("[DRY RUN] Would delete namespace %s", k.namespace)
	}

	return nil
}

// handleTerminatingNamespace handles namespace stuck in Terminating state
// This is inspired by knsk's approach
func (k *NamespaceKiller) handleTerminatingNamespace(ctx context.Context, ns *v1.Namespace) error {
	log.Info().Msgf("Handling terminating namespace %s", k.namespace)

	// Step 1: Delete all resources in the namespace
	if err := k.deleteAllResourcesInNamespace(ctx); err != nil {
		log.Warn().Err(err).Msgf("Error deleting resources in namespace %s", k.namespace)
		// Continue anyway if force mode is enabled
		if !k.force {
			return fmt.Errorf("failed to delete all resources: %w", err)
		}
	}

	// Step 2: Wait a bit and check if namespace is still terminating
	time.Sleep(3 * time.Second)
	ns, err := k.client.CoreV1().Namespaces().Get(ctx, k.namespace, metav1.GetOptions{})
	if err != nil {
		// Namespace might be deleted
		log.Info().Msgf("Namespace %s deleted successfully", k.namespace)
		return nil
	}

	if ns.Status.Phase == v1.NamespaceTerminating {
		log.Warn().Msgf("Namespace %s is still in Terminating state after resource cleanup", k.namespace)
		if k.force {
			// Step 3: Remove finalizers if force mode is enabled
			return k.removeFinalizers(ctx, ns)
		}
		log.Info().Msgf("Use --mafia flag to force remove finalizers")
		return fmt.Errorf("namespace %s is still terminating", k.namespace)
	}

	log.Info().Msgf("Namespace %s deleted successfully", k.namespace)
	return nil
}

// deleteAllResourcesInNamespace deletes all resources in the namespace
func (k *NamespaceKiller) deleteAllResourcesInNamespace(ctx context.Context) error {
	log.Info().Msgf("Deleting all resources in namespace %s", k.namespace)

	// Get discovery client to find all API resources
	discoveryClient := k.client.Discovery()

	// Get all API resources
	apiResourceList, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		log.Warn().Err(err).Msg("Failed to get server preferred resources, continuing with partial list")
	}

	// Resources to skip (these are managed by Kubernetes or shouldn't be deleted)
	skipResources := map[string]bool{
		"events":               true,
		"events.events.k8s.io": true,
		"bindings":             true,
		"endpoints":            true,
		"limitranges":          true,
		"resourcequotas":       true,
		"podtemplates":         true,
		"serviceaccounts":      true,
		"secrets":              true, // Will be handled separately if needed
		"configmaps":           true, // Will be handled separately if needed
	}

	// Process each API group
	for _, apiResourceList := range apiResourceList {
		for _, apiResource := range apiResourceList.APIResources {
			// Skip subresources
			if strings.Contains(apiResource.Name, "/") {
				continue
			}

			// Skip resources that shouldn't be deleted
			if skipResources[apiResource.Name] || skipResources[fmt.Sprintf("%s.%s", apiResource.Name, apiResource.Group)] {
				continue
			}

			// Skip if not namespaced
			if !apiResource.Namespaced {
				continue
			}

			// Build GVR
			gv, err := schema.ParseGroupVersion(apiResourceList.GroupVersion)
			if err != nil {
				log.Warn().Err(err).Msgf("Failed to parse group version %s", apiResourceList.GroupVersion)
				continue
			}

			gvr := schema.GroupVersionResource{
				Group:    gv.Group,
				Version:  gv.Version,
				Resource: apiResource.Name,
			}

			// Delete resources
			if err := k.deleteResourcesByGVR(ctx, gvr); err != nil {
				log.Warn().Err(err).Msgf("Failed to delete resources %s", gvr.String())
				// Continue with other resources
			}
		}
	}

	return nil
}

// deleteResourcesByGVR deletes all resources of a specific GVR in the namespace
func (k *NamespaceKiller) deleteResourcesByGVR(ctx context.Context, gvr schema.GroupVersionResource) error {
	resourceInterface := k.dynamicClient.Resource(gvr).Namespace(k.namespace)

	// List all resources
	list, err := resourceInterface.List(ctx, metav1.ListOptions{})
	if err != nil {
		// Some resources might not be accessible, skip them
		return nil
	}

	if len(list.Items) == 0 {
		return nil
	}

	log.Info().Msgf("Found %d resources of type %s in namespace %s", len(list.Items), gvr.String(), k.namespace)

	// Delete each resource
	for _, item := range list.Items {
		name := item.GetName()
		if k.dryRun {
			log.Info().Msgf("[DRY RUN] Would delete %s/%s", gvr.String(), name)
		} else {
			log.Info().Msgf("Deleting %s/%s", gvr.String(), name)
			err := resourceInterface.Delete(ctx, name, k.deleteOption)
			if err != nil {
				log.Warn().Err(err).Msgf("Failed to delete %s/%s", gvr.String(), name)
				// Continue with other resources
			}
		}
	}

	return nil
}

// removeFinalizers removes finalizers from the namespace to force deletion
func (k *NamespaceKiller) removeFinalizers(ctx context.Context, ns *v1.Namespace) error {
	if len(ns.Spec.Finalizers) == 0 {
		log.Info().Msgf("No finalizers to remove from namespace %s", k.namespace)
		return nil
	}

	log.Warn().Msgf("Removing finalizers from namespace %s: %v", k.namespace, ns.Spec.Finalizers)

	if k.dryRun {
		log.Info().Msgf("[DRY RUN] Would remove finalizers from namespace %s", k.namespace)
		return nil
	}

	// Create a copy and remove finalizers
	nsCopy := ns.DeepCopy()
	nsCopy.Spec.Finalizers = []v1.FinalizerName{}

	// Update the namespace
	_, err := k.client.CoreV1().Namespaces().Finalize(ctx, nsCopy, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to remove finalizers from namespace %s: %w", k.namespace, err)
	}

	log.Info().Msgf("Successfully removed finalizers from namespace %s", k.namespace)
	return nil
}
