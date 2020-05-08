package killer

import (
	"context"
	"fmt"

	"github.com/p-program/kube-killer/core"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type PVKiller struct {
	client *kubernetes.Clientset
}

func NewPVKiller() (*PVKiller, error) {
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		return nil, err
	}
	k := PVKiller{client: clientset}
	return &k, err
}

// KillUnBoundPV kill Released PV
func (k *PVKiller) KillUnBoundPV() error {
	volumeList, err := k.getAllPV()
	if err != nil {
		return err
	}
	for i := 0; i < len(volumeList.Items); i++ {
		volume := volumeList.Items[i]
		volumeName := volume.Name
		// volume.Status.Phase

		// var gracePeriodSeconds int64 = 0
		// deleteOption := metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds}
		// k.client.CoreV1().PersistentVolumes().Delete(context.TODO(), volumeName, deleteOption)
		fmt.Println(volumeName)
	}
	// var gracePeriodSeconds int64 = 0
	// deleteOption := metav1.DeleteOptions{GracePeriodSeconds: &gracePeriodSeconds}
	// listOption := metav1.ListOptions{FieldSelector: "status.phase=Released"}
	// err := k.client.CoreV1().PersistentVolumes().DeleteCollection(context.TODO(), deleteOption, listOption)
	return err
}

func (k *PVKiller) deletePV(v1.PersistentVolumePhase) {

}

func (k *PVKiller) getAllPV() (*v1.PersistentVolumeList, error) {
	return k.client.CoreV1().PersistentVolumes().List(context.TODO(), metav1.ListOptions{})
}
