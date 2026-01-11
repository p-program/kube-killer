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
	"k8s.io/client-go/util/retry"
)

// CRDKiller kills CustomResourceDefinitions based on group pattern
type CRDKiller struct {
	crdClient    apiextensionsclient.Interface
	deleteOption metav1.DeleteOptions
	dryRun       bool
	mafia        bool
	groupPattern string
}

// NewCRDKiller creates a new CRDKiller instance
// groupPattern: the group pattern to match (e.g., "*.example.com" or "example.com")
// Note: CRD is cluster-scoped, so namespace is not needed
func NewCRDKiller(groupPattern string) (*CRDKiller, error) {
	crdClient, err := apiextensionsclient.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		return nil, fmt.Errorf("failed to create CRD client: %w", err)
	}

	k := CRDKiller{
		crdClient:    crdClient,
		groupPattern: groupPattern,
	}

	var gracePeriodSeconds int64 = 1
	k.deleteOption = metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}

	return &k, nil
}

func (k *CRDKiller) BlackHand() *CRDKiller {
	k.mafia = true
	return k
}

func (k *CRDKiller) DryRun() *CRDKiller {
	k.dryRun = true
	k.deleteOption.DryRun = []string{"All"}
	return k
}

// matchesGroup checks if a group matches the pattern
// Supports wildcard patterns like "*.example.com"
func (k *CRDKiller) matchesGroup(group string) bool {
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
func (k *CRDKiller) getMatchingCRDs() ([]apiextensionsv1.CustomResourceDefinition, error) {
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

// deleteCRD deletes a single CRD
func (k *CRDKiller) deleteCRD(crd apiextensionsv1.CustomResourceDefinition) error {
	crdName := crd.Name
	group := crd.Spec.Group

	if k.dryRun {
		log.Info().Msgf("[DRY RUN] Would delete CRD %s (group: %s)", crdName, group)
		return nil
	}

	log.Warn().Msgf("Deleting CRD %s (group: %s)", crdName, group)

	// Use retry for deletion in case of conflicts
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return k.crdClient.ApiextensionsV1().CustomResourceDefinitions().Delete(context.TODO(), crdName, k.deleteOption)
	})

	if err != nil {
		return fmt.Errorf("failed to delete CRD %s: %w", crdName, err)
	}

	log.Info().Msgf("Successfully deleted CRD %s (group: %s)", crdName, group)
	return nil
}

// Kill deletes all CRDs matching the group pattern
func (k *CRDKiller) Kill() error {
	log.Info().Msgf("Starting CRD kill operation for group pattern: %s", k.groupPattern)

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

	// Delete each matching CRD
	totalDeleted := 0
	for _, crd := range matchingCRDs {
		if err := k.deleteCRD(crd); err != nil {
			log.Error().Err(err).Msgf("Failed to delete CRD %s", crd.Name)
			// Continue with other CRDs
		} else {
			totalDeleted++
		}
	}

	log.Info().Msgf("Completed CRD kill operation. Processed %d CRD(s)", totalDeleted)
	return nil
}


