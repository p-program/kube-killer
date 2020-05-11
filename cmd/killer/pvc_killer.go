package killer

import (
	"context"

	"github.com/p-program/kube-killer/core"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PVCKiller struct {
	client       *kubernetes.Clientset
	deleteOption metav1.DeleteOptions
	dryRun       bool
	mafia        bool
	namespace    string
}

// NewPVCKiller NewPVCKiller
// namespace can be ""， empty stands for current namespace
func NewPVCKiller(namespace string) (*PVCKiller, error) {
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		return nil, err
	}
	k := PVCKiller{
		namespace: namespace,
		client:    clientset,
	}
	var gracePeriodSeconds int64 = 1
	k.deleteOption = metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}
	return &k, nil
}

func (k *PVCKiller) BlackHand() *PVCKiller {
	k.mafia = true
	return k
}

func (k *PVCKiller) DryRun() *PVCKiller {
	k.dryRun = true
	k.deleteOption.DryRun = []string{"All"}
	return k
}

// DeserveDead Pending/Lost PVC deserve to die
func (k *PVCKiller) DeserveDead(resource interface{}) bool {
	if k.mafia {
		return true
	}
	pvc := resource.(v1.PersistentVolumeClaim)
	phase := pvc.Status.Phase
	if phase == v1.ClaimBound {
		return false
	}
	return true
}

// Kill Kill
func (k *PVCKiller) Kill() error {
	list, err := k.client.CoreV1().PersistentVolumeClaims(k.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, pvc := range list.Items {
		if !k.DeserveDead(pvc) {
			continue
		}
		log.Info().Msgf("delete pvc %s in namespace %s ", pvc.Name, pvc.Namespace)
		k.client.CoreV1().PersistentVolumeClaims(k.namespace).Delete(context.TODO(), pvc.Name, k.deleteOption)
	}
	return nil
}
