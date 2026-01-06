package killer

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog/log"

	"github.com/p-program/kube-killer/core"
	v1 "k8s.io/api/core/v1"
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
	if !k.mafia {
		//kubectl cordon $node
		log.Info().Msgf("kubectl cordon %s", k.nodeName)
		getOption := metav1.GetOptions{}
		node, err := clientset.CoreV1().Nodes().Get(context.TODO(), k.nodeName, getOption)
		if err != nil {
			return err
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
		}
		_, err = clientset.CoreV1().Nodes().Patch(context.TODO(), k.nodeName, types.StrategicMergePatchType, patchBytes, patchOptions)
		if err != nil {
			return err
		}
		// Drain all pods from the node except DaemonSet pods
		// Similar to: kubectl drain $node --ignore-daemonsets
		err = k.drainNodePods(clientset)
		if err != nil {
			return err
		}
	}
	//kubectl delete $node
	return nil
}

// drainNodePods evicts all pods from the node except DaemonSet pods
// Similar to: kubectl drain $node --ignore-daemonsets
func (k *NodeKiller) drainNodePods(clientset *kubernetes.Clientset) error {
	log.Info().Msgf("Draining pods from node %s (ignoring DaemonSet pods)", k.nodeName)

	// Get all pods on this node
	fieldSelector := "spec.nodeName=" + k.nodeName
	pods, err := clientset.CoreV1().Pods("").List(context.TODO(), metav1.ListOptions{
		FieldSelector: fieldSelector,
	})
	if err != nil {
		return err
	}

	evictedCount := 0
	for _, pod := range pods.Items {
		// Check if pod belongs to a DaemonSet
		if k.isDaemonSetPod(pod) {
			log.Info().Msgf("Skipping DaemonSet pod %s/%s", pod.Namespace, pod.Name)
			continue
		}

		log.Info().Msgf("Evicting pod %s/%s from node %s", pod.Namespace, pod.Name, k.nodeName)
		err = clientset.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), pod.Name, k.deleteOption)
		if err != nil {
			log.Error().Err(err).Msgf("Failed to evict pod %s/%s", pod.Namespace, pod.Name)
			continue
		}
		evictedCount++
	}

	log.Info().Msgf("Evicted %d pods from node %s", evictedCount, k.nodeName)
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
