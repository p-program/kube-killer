package killer

import (
	"context"
	"encoding/json"

	"github.com/rs/zerolog/log"

	"github.com/p-program/kube-killer/core"
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
	k := NodeKiller{client: clientset}
	var gracePeriodSeconds int64 = 0
	k.deleteOption = metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}
	return &k, err
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
		//# TODO:驱逐除了ds以外所有的pod
		//kubectl drain $node   --ignore-daemonsets
	}
	//kubectl delete $node
	return nil
}
