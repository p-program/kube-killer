package killer

import (
	"context"
	"fmt"
	"strings"

	"github.com/p-program/kube-killer/core"
	"github.com/rs/zerolog/log"
	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextensionsclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/util/retry"
)

// CRKiller kills Custom Resources based on group pattern
type CRKiller struct {
	dynamicClient        dynamic.Interface
	crdClient            apiextensionsclient.Interface
	deleteOption         metav1.DeleteOptions
	dryRun               bool
	mafia                bool
	namespace            string
	groupPattern         string
}

// NewCRKiller creates a new CRKiller instance
// groupPattern: the group pattern to match (e.g., "*.example.com" or "example.com")
// namespace can be "" for all namespaces
func NewCRKiller(groupPattern, namespace string) (*CRKiller, error) {
	dynamicClient, err := dynamic.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	crdClient, err := apiextensionsclient.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		return nil, fmt.Errorf("failed to create CRD client: %w", err)
	}

	k := CRKiller{
		dynamicClient: dynamicClient,
		crdClient:     crdClient,
		namespace:     namespace,
		groupPattern:  groupPattern,
	}

	var gracePeriodSeconds int64 = 1
	k.deleteOption = metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}

	return &k, nil
}

func (k *CRKiller) BlackHand() *CRKiller {
	k.mafia = true
	return k
}

func (k *CRKiller) DryRun() *CRKiller {
	k.dryRun = true
	k.deleteOption.DryRun = []string{"All"}
	return k
}

// matchesGroup checks if a group matches the pattern
// Supports wildcard patterns like "*.example.com"
func (k *CRKiller) matchesGroup(group string) bool {
	pattern := k.groupPattern
	if pattern == "" {
		return false
	}

	// Exact match
	if pattern == group {
		return true
	}

	// Wildcard match: *.example.com
	if strings.HasPrefix(pattern, "*.") {
		suffix := strings.TrimPrefix(pattern, "*.")
		return strings.HasSuffix(group, "."+suffix) || group == suffix
	}

	// Prefix match: example.*
	if strings.HasSuffix(pattern, ".*") {
		prefix := strings.TrimSuffix(pattern, ".*")
		return strings.HasPrefix(group, prefix+".")
	}

	return false
}

// getMatchingCRDs returns all CRDs that match the group pattern
func (k *CRKiller) getMatchingCRDs() ([]apiextensionsv1.CustomResourceDefinition, error) {
	crdList, err := k.crdClient.ApiextensionsV1().CustomResourceDefinitions().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list CRDs: %w", err)
	}

	var matchingCRDs []apiextensionsv1.CustomResourceDefinition
	for _, crd := range crdList.Items {
		if k.matchesGroup(crd.Spec.Group) {
			matchingCRDs = append(matchingCRDs, crd)
			log.Info().Msgf("Found matching CRD: %s (group: %s)", crd.Name, crd.Spec.Group)
		}
	}

	return matchingCRDs, nil
}

// getGVRFromCRD converts a CRD to a GroupVersionResource
func (k *CRKiller) getGVRFromCRD(crd apiextensionsv1.CustomResourceDefinition) (schema.GroupVersionResource, error) {
	// Get the storage version (the version marked as storage: true)
	var storageVersion string
	for _, version := range crd.Spec.Versions {
		if version.Storage {
			storageVersion = version.Name
			break
		}
	}

	if storageVersion == "" && len(crd.Spec.Versions) > 0 {
		// Fallback to first version if no storage version is marked
		storageVersion = crd.Spec.Versions[0].Name
	}

	if storageVersion == "" {
		return schema.GroupVersionResource{}, fmt.Errorf("no version found for CRD %s", crd.Name)
	}

	gvr := schema.GroupVersionResource{
		Group:    crd.Spec.Group,
		Version:  storageVersion,
		Resource: crd.Spec.Names.Plural,
	}

	return gvr, nil
}

// deleteCRsForCRD deletes all CRs for a given CRD
func (k *CRKiller) deleteCRsForCRD(crd apiextensionsv1.CustomResourceDefinition) error {
	gvr, err := k.getGVRFromCRD(crd)
	if err != nil {
		return fmt.Errorf("failed to get GVR for CRD %s: %w", crd.Name, err)
	}

	if crd.Spec.Scope == apiextensionsv1.ClusterScoped {
		// Cluster-scoped resource - no namespace needed
		resourceInterface := k.dynamicClient.Resource(gvr)
		log.Info().Msgf("Processing cluster-scoped CRD: %s (GVR: %s)", crd.Name, gvr.String())
		return k.deleteCRsInNamespace(resourceInterface, gvr, crd)
	} else {
		// Namespace-scoped resource
		if k.namespace == "" {
			// Need to process all namespaces
			return k.deleteCRsInAllNamespaces(gvr, crd)
		}
		resourceInterface := k.dynamicClient.Resource(gvr).Namespace(k.namespace)
		log.Info().Msgf("Processing namespaced CRD: %s in namespace %s (GVR: %s)", crd.Name, k.namespace, gvr.String())
		return k.deleteCRsInNamespace(resourceInterface, gvr, crd)
	}
}

// deleteCRsInAllNamespaces deletes CRs across all namespaces
func (k *CRKiller) deleteCRsInAllNamespaces(gvr schema.GroupVersionResource, crd apiextensionsv1.CustomResourceDefinition) error {
	// Get all namespaces
	clientset, err := getKubernetesClient()
	if err != nil {
		return err
	}

	nsList, err := clientset.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list namespaces: %w", err)
	}

	for _, ns := range nsList.Items {
		// Skip kube-system
		if ns.Name == "kube-system" {
			continue
		}

		log.Info().Msgf("Processing namespace: %s for CRD: %s", ns.Name, crd.Name)
		resourceInterface := k.dynamicClient.Resource(gvr).Namespace(ns.Name)
		if err := k.deleteCRsInNamespace(resourceInterface, gvr, crd); err != nil {
			log.Warn().Err(err).Msgf("Error deleting CRs in namespace %s", ns.Name)
			// Continue with other namespaces
		}
	}

	return nil
}

// deleteCRsInNamespace deletes all CRs in a specific namespace
func (k *CRKiller) deleteCRsInNamespace(resourceInterface dynamic.ResourceInterface, gvr schema.GroupVersionResource, crd apiextensionsv1.CustomResourceDefinition) error {
	// List all CRs
	crList, err := resourceInterface.List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list CRs for %s: %w", gvr.String(), err)
	}

	if len(crList.Items) == 0 {
		log.Info().Msgf("No CRs found for %s", gvr.String())
		return nil
	}

	log.Info().Msgf("Found %d CRs for %s", len(crList.Items), gvr.String())

	// Delete each CR
	for _, cr := range crList.Items {
		crName := cr.GetName()
		crNamespace := cr.GetNamespace()
		
		// Format resource identifier for logging
		var resourceID string
		if crNamespace != "" {
			resourceID = fmt.Sprintf("%s/%s", crNamespace, crName)
		} else {
			resourceID = crName // Cluster-scoped resource
		}
		
		if k.dryRun {
			log.Info().Msgf("[DRY RUN] Would delete CR %s (GVR: %s)", resourceID, gvr.String())
		} else {
			log.Warn().Msgf("Deleting CR %s (GVR: %s)", resourceID, gvr.String())
			
			// Use retry for deletion in case of conflicts
			err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
				return resourceInterface.Delete(context.TODO(), crName, k.deleteOption)
			})
			
			if err != nil {
				log.Error().Err(err).Msgf("Failed to delete CR %s", resourceID)
				// Continue with other CRs
			} else {
				log.Info().Msgf("Successfully deleted CR %s", resourceID)
			}
		}
	}

	return nil
}

// Kill deletes all CRs matching the group pattern
func (k *CRKiller) Kill() error {
	log.Info().Msgf("Starting CR kill operation for group pattern: %s", k.groupPattern)

	// Get all matching CRDs
	matchingCRDs, err := k.getMatchingCRDs()
	if err != nil {
		return fmt.Errorf("failed to get matching CRDs: %w", err)
	}

	if len(matchingCRDs) == 0 {
		log.Warn().Msgf("No CRDs found matching group pattern: %s", k.groupPattern)
		return nil
	}

	log.Info().Msgf("Found %d CRD(s) matching group pattern: %s", len(matchingCRDs), k.groupPattern)

	// Delete CRs for each matching CRD
	totalDeleted := 0
	for _, crd := range matchingCRDs {
		if err := k.deleteCRsForCRD(crd); err != nil {
			log.Error().Err(err).Msgf("Failed to delete CRs for CRD %s", crd.Name)
			// Continue with other CRDs
		} else {
			totalDeleted++
		}
	}

	log.Info().Msgf("Completed CR kill operation. Processed %d CRD(s)", totalDeleted)
	return nil
}

