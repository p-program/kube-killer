package killer

import (
	"github.com/p-program/kube-killer/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type NodeKiller struct {
	client       *kubernetes.Clientset
	deleteOption metav1.DeleteOptions
	dryRun       bool
	mafia        bool
}

func NewNodeKiller() (*NodeKiller, error) {
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
