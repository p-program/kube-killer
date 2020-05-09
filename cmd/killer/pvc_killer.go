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
	dryRun       bool
	namespace    string
	deleteOption metav1.DeleteOptions
}

// NewPVCKiller NewPVCKiller
// dryRun true: fake killer; flase true killer
// namespace can be ""， empty stands for current namespace
func NewPVCKiller(dryRun bool, namespace string) *PVCKiller {
	k := PVCKiller{
		dryRun:    dryRun,
		namespace: namespace,
	}
	deleteOption := metav1.DeleteOptions{}
	var gracePeriodSeconds int64 = 1
	deleteOption.GracePeriodSeconds = &gracePeriodSeconds
	if k.dryRun {
		deleteOption.DryRun = []string{"All"}
	}
	k.deleteOption = deleteOption
	return &k
}

// PVCDeserveDead Pending/Lost PVC deserve to die
func (k *PVCKiller) PVCDeserveDead(pvc *v1.PersistentVolumeClaim) bool {
	phase := pvc.Status.Phase
	if phase == v1.ClaimBound {
		return false
	}
	return true
}

// KillUnBoundPVC KillUnBoundPVC
func (k *PVCKiller) KillUnBoundPVC() error {
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		return err
	}
	list, err := clientset.CoreV1().PersistentVolumeClaims(k.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, pvc := range list.Items {
		if !k.PVCDeserveDead(&pvc) {
			continue
		}
		log.Info().Msgf("delete pvc: %s", pvc.Name)
		clientset.CoreV1().PersistentVolumeClaims(k.namespace).Delete(context.TODO(), pvc.Name, k.deleteOption)
	}
	return nil
}
