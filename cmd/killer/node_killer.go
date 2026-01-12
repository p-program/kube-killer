package killer

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/p-program/kube-killer/core"
	v1 "k8s.io/api/core/v1"
	policyv1 "k8s.io/api/policy/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/strategicpatch"
	"k8s.io/client-go/kubernetes"
)

type NodeKiller struct {
	client       *kubernetes.Clientset
	deleteOption metav1.DeleteOptions
	dryRun       bool
	mafia        bool
	nodeName     string
}

func NewNodeKiller(nodeName string) (*NodeKiller, error) {
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		return nil, err
	}
	k := NodeKiller{
		client:   clientset,
		nodeName: nodeName,
	}
	var gracePeriodSeconds int64 = 0
	k.deleteOption = metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}
	return &k, nil
}

func (k *NodeKiller) BlackHand() *NodeKiller {
	k.mafia = true
	return k
}

func (k *NodeKiller) DryRun() *NodeKiller {
	k.dryRun = true
	k.deleteOption.DryRun = []string{"All"}
	return k
}

func (k *NodeKiller) Kill() error {
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		panic(err.Error())
	}
	ctx := context.TODO()

	if !k.mafia {
		// Step 1: kubectl cordon $node - Mark node as unschedulable
		log.Info().Msgf("kubectl cordon %s", k.nodeName)
		err = k.cordonNode(clientset, ctx)
		if err != nil {
			return fmt.Errorf("failed to cordon node %s: %w", k.nodeName, err)
		}

		// Step 2: kubectl drain $node --ignore-daemonsets - Evict all pods except DaemonSet pods
		log.Info().Msgf("kubectl drain %s --ignore-daemonsets", k.nodeName)
		err = k.drainNodePods(clientset, ctx)
		if err != nil {
			return fmt.Errorf("failed to drain node %s: %w", k.nodeName, err)
		}
	}

	// Step 3: kubectl delete $node - Delete the node
	log.Info().Msgf("kubectl delete node %s", k.nodeName)
	err = k.deleteNode(clientset, ctx)
	if err != nil {
		return fmt.Errorf("failed to delete node %s: %w", k.nodeName, err)
	}

	log.Info().Msgf("Successfully completed node deletion process for %s", k.nodeName)
	return nil
}

// cordonNode marks the node as unschedulable
func (k *NodeKiller) cordonNode(clientset *kubernetes.Clientset, ctx context.Context) error {
	getOption := metav1.GetOptions{}
	node, err := clientset.CoreV1().Nodes().Get(ctx, k.nodeName, getOption)
	if err != nil {
		return err
	}

	// Check if already cordoned
	if node.Spec.Unschedulable {
		log.Info().Msgf("Node %s is already cordoned", k.nodeName)
		return nil
	}

	oldData, err := json.Marshal(node)
	if err != nil {
		return err
	}
	node.Spec.Unschedulable = true
	newData, err := json.Marshal(node)
	if err != nil {
		return err
	}

	patchBytes, patchErr := strategicpatch.CreateTwoWayMergePatch(oldData, newData, node)
	if patchErr != nil {
		return patchErr
	}

	patchOptions := metav1.PatchOptions{}
	if k.dryRun {
		patchOptions.DryRun = []string{metav1.DryRunAll}
		log.Info().Msgf("[DRY RUN] Would cordon node %s", k.nodeName)
		return nil
	}

	_, err = clientset.CoreV1().Nodes().Patch(ctx, k.nodeName, types.StrategicMergePatchType, patchBytes, patchOptions)
	return err
}

// drainNodePods evicts all pods from the node except DaemonSet pods
// Similar to: kubectl drain $node --ignore-daemonsets
func (k *NodeKiller) drainNodePods(clientset *kubernetes.Clientset, ctx context.Context) error {
	log.Info().Msgf("Draining pods from node %s (ignoring DaemonSet pods)", k.nodeName)

	// Get all pods on this node
	fieldSelector := "spec.nodeName=" + k.nodeName
	pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return err
	}

	if len(pods.Items) == 0 {
		log.Info().Msgf("No pods found on node %s", k.nodeName)
		return nil
	}

	// First pass: Evict all non-DaemonSet pods
	evictedPods := make([]v1.Pod, 0)
	for _, pod := range pods.Items {
		// Check if pod belongs to a DaemonSet
		if k.isDaemonSetPod(pod) {
			log.Info().Msgf("Skipping DaemonSet pod %s/%s", pod.Namespace, pod.Name)
			continue
		}

		// Skip pods that are already terminating
		if pod.DeletionTimestamp != nil {
			log.Info().Msgf("Pod %s/%s is already terminating, skipping", pod.Namespace, pod.Name)
			evictedPods = append(evictedPods, pod)
			continue
		}

		log.Info().Msgf("Evicting pod %s/%s from node %s", pod.Namespace, pod.Name, k.nodeName)
		err = k.evictPod(clientset, ctx, pod)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to evict pod %s/%s", pod.Namespace, pod.Name)
			// Continue with other pods even if one fails
			continue
		}
		evictedPods = append(evictedPods, pod)
	}

	// Second pass: Wait for all evicted pods to terminate
	if len(evictedPods) > 0 {
		log.Info().Msgf("Waiting for %d pods to terminate on node %s", len(evictedPods), k.nodeName)
		err = k.waitForPodsToTerminate(clientset, ctx, evictedPods)
		if err != nil {
			log.Warn().Err(err).Msgf("Some pods may not have terminated gracefully")
		}
	}

	log.Info().Msgf("Successfully evicted %d pods from node %s", len(evictedPods), k.nodeName)
	return nil
}

// evictPod evicts a pod using the Eviction API
func (k *NodeKiller) evictPod(clientset *kubernetes.Clientset, ctx context.Context, pod v1.Pod) error {
	eviction := &policyv1.Eviction{
		ObjectMeta: metav1.ObjectMeta{
			Name:      pod.Name,
			Namespace: pod.Namespace,
		},
	}

	if k.dryRun {
		log.Info().Msgf("[DRY RUN] Would evict pod %s/%s", pod.Namespace, pod.Name)
		return nil
	}

	// Use PolicyV1 Evictions API which respects PodDisruptionBudgets and allows graceful termination
	err := clientset.PolicyV1().Evictions(pod.Namespace).Evict(ctx, eviction)
	if err != nil {
		return fmt.Errorf("failed to evict pod %s/%s: %w", pod.Namespace, pod.Name, err)
	}

	return nil
}

// waitForPodsToTerminate waits for all pods to be terminated
func (k *NodeKiller) waitForPodsToTerminate(clientset *kubernetes.Clientset, ctx context.Context, pods []v1.Pod) error {
	timeout := 5 * time.Minute
	interval := 5 * time.Second

	startTime := time.Now()
	for time.Since(startTime) < timeout {
		allTerminated := true
		remainingPods := 0

		for _, pod := range pods {
			// Check if pod still exists
			currentPod, err := clientset.CoreV1().Pods(pod.Namespace).Get(ctx, pod.Name, metav1.GetOptions{})
			if err != nil {
				// Pod doesn't exist anymore, consider it terminated
				continue
			}

			// Check if pod is terminated
			if currentPod.DeletionTimestamp == nil && currentPod.Status.Phase != v1.PodSucceeded && currentPod.Status.Phase != v1.PodFailed {
				allTerminated = false
				remainingPods++
			}
		}

		if allTerminated {
			log.Info().Msgf("All pods have been terminated on node %s", k.nodeName)
			return nil
		}

		if remainingPods > 0 {
			log.Info().Msgf("Waiting for %d pods to terminate on node %s...", remainingPods, k.nodeName)
		}

		time.Sleep(interval)
	}

	return fmt.Errorf("timeout waiting for pods to terminate on node %s", k.nodeName)
}

// deleteNode deletes the node from the cluster
func (k *NodeKiller) deleteNode(clientset *kubernetes.Clientset, ctx context.Context) error {
	deleteOptions := metav1.DeleteOptions{}
	if k.dryRun {
		deleteOptions.DryRun = []string{metav1.DryRunAll}
		log.Info().Msgf("[DRY RUN] Would delete node %s", k.nodeName)
		return nil
	}

	err := clientset.CoreV1().Nodes().Delete(ctx, k.nodeName, deleteOptions)
	if err != nil {
		return fmt.Errorf("failed to delete node %s: %w", k.nodeName, err)
	}

	log.Info().Msgf("Node %s deletion initiated", k.nodeName)
	return nil
}

// isDaemonSetPod checks if a pod belongs to a DaemonSet by examining OwnerReferences
func (k *NodeKiller) isDaemonSetPod(pod v1.Pod) bool {
	for _, ownerRef := range pod.OwnerReferences {
		if ownerRef.Kind == "DaemonSet" {
			return true
		}
		// Check if it's owned by a ReplicaSet that belongs to a DaemonSet
		// This is a more thorough check, but for simplicity, we'll check direct ownership
		if ownerRef.Kind == "ReplicaSet" {
			// Get the ReplicaSet to check its owner
			rs, err := k.client.AppsV1().ReplicaSets(pod.Namespace).Get(context.TODO(), ownerRef.Name, metav1.GetOptions{})
			if err == nil {
				for _, rsOwnerRef := range rs.OwnerReferences {
					if rsOwnerRef.Kind == "DaemonSet" {
						return true
					}
				}
			}
		}
	}
	return false
}
