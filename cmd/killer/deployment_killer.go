package killer

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/p-program/kube-killer/core"
	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/util/retry"
)

type DeploymentKiller struct {
	client       *kubernetes.Clientset
	deleteOption metav1.DeleteOptions
	dryRun       bool
	mafia        bool
	half         bool
	namespace    string
}

func NewDeploymentKiller(namespace string) (*DeploymentKiller, error) {
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		return nil, err
	}
	k := DeploymentKiller{
		namespace: namespace,
		client:    clientset,
	}
	var gracePeriodSeconds int64 = 1
	k.deleteOption = metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}
	return &k, nil
}

func (k *DeploymentKiller) BlackHand() *DeploymentKiller {
	k.mafia = true
	return k
}

func (k *DeploymentKiller) DryRun() *DeploymentKiller {
	k.dryRun = true
	k.deleteOption.DryRun = []string{"All"}
	return k
}

func (k *DeploymentKiller) SetHalf() *DeploymentKiller {
	k.half = true
	return k
}

// Kill deletes deployments based on the mode
// - If mafia is true and half is true: deletes half of the deployments randomly
// - If mafia is true: deletes all deployments
// - Otherwise: deletes all deployments (default behavior, similar to kubectl delete deploy)
func (k *DeploymentKiller) Kill() error {
	if k.mafia {
		if k.half {
			return k.KillHalfDeployments()
		}
		return k.KillAllDeployments()
	}
	return k.KillAllDeployments()
}

// KillAllDeployments deletes all deployments in the namespace
func (k *DeploymentKiller) KillAllDeployments() error {
	deployments, err := k.client.AppsV1().Deployments(k.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}

	if len(deployments.Items) == 0 {
		log.Info().Msgf("No deployments found in namespace %s", k.namespace)
		return nil
	}

	log.Info().Msgf("Found %d deployment(s) in namespace %s", len(deployments.Items), k.namespace)

	for _, deploy := range deployments.Items {
		if err := k.deleteDeployment(deploy); err != nil {
			log.Error().Err(err).Msgf("Failed to delete deployment %s", deploy.Name)
			// Continue with other deployments
		}
	}

	return nil
}

// KillHalfDeployments deletes half of the deployments randomly
func (k *DeploymentKiller) KillHalfDeployments() error {
	deployments, err := k.client.AppsV1().Deployments(k.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list deployments: %w", err)
	}

	if len(deployments.Items) == 0 {
		log.Info().Msgf("No deployments found in namespace %s", k.namespace)
		return nil
	}

	// Randomly shuffle the deployments list
	deployList := deployments.Items
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(deployList), func(i, j int) {
		deployList[i], deployList[j] = deployList[j], deployList[i]
	})

	// Calculate how many deployments to kill (half, rounded down)
	deploysToKill := len(deployList) / 2
	if deploysToKill == 0 {
		deploysToKill = 1 // At least kill one deployment if there's only one
	}

	log.Info().Msgf("Killing %d out of %d deployment(s) in namespace %s", deploysToKill, len(deployList), k.namespace)

	for i := 0; i < deploysToKill; i++ {
		deploy := deployList[i]
		if err := k.deleteDeployment(deploy); err != nil {
			log.Error().Err(err).Msgf("Failed to delete deployment %s", deploy.Name)
			// Continue with other deployments
		}
	}

	return nil
}

// deleteDeployment deletes a single deployment
func (k *DeploymentKiller) deleteDeployment(deploy appsv1.Deployment) error {
	deployName := deploy.Name

	if k.dryRun {
		log.Info().Msgf("[DRY RUN] Would delete deployment %s in namespace %s", deployName, k.namespace)
		return nil
	}

	log.Warn().Msgf("Deleting deployment %s in namespace %s", deployName, k.namespace)

	// Use retry for deletion in case of conflicts
	err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
		return k.client.AppsV1().Deployments(k.namespace).Delete(context.TODO(), deployName, k.deleteOption)
	})

	if err != nil {
		return fmt.Errorf("failed to delete deployment %s: %w", deployName, err)
	}

	log.Info().Msgf("Successfully deleted deployment %s in namespace %s", deployName, k.namespace)
	return nil
}
