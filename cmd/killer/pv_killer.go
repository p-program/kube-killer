package killer

import (
	"context"

	"github.com/p-program/kube-killer/core"
	"github.com/rs/zerolog/log"
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
func (k *PVKiller) KillUnBoundPV(dryRun bool) error {
	volumeList, err := k.getAllPV()
	if err != nil {
		return err
	}
	for i := 0; i < len(volumeList.Items); i++ {
		volume := volumeList.Items[i]
		volumeName := volume.Name
		if !k.PVDeserveDead(&volume) {
			continue
		}
		log.Info().Msgf("Volume Info { volumeName: %s ;volume.Status.Phase: %s }", volumeName, volume.Status.Phase)
		err = k.deletePV(volumeName, dryRun)
		if err != nil {
			//log but continue
			log.Warn().Err(err)
			continue
		}

	}
	return err
}

// PVDeserveDead true stands for dead
func (k *PVKiller) PVDeserveDead(pv *v1.PersistentVolume) bool {
	phase := pv.Status.Phase
	if phase == v1.VolumePending {
		return false
	}
	if phase == v1.VolumeAvailable {
		return false
	}
	if phase == v1.VolumeBound {
		return false
	}
	return true
}

func (k *PVKiller) deletePV(name string, dryRun bool) error {
	var gracePeriodSeconds int64 = 0
	deleteOption := metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}
	if dryRun {
		deleteOption.DryRun = []string{"All"}
	}
	return k.client.CoreV1().PersistentVolumes().Delete(context.TODO(), name, deleteOption)
}

func (k *PVKiller) getAllPV() (*v1.PersistentVolumeList, error) {
	return k.client.CoreV1().PersistentVolumes().List(context.TODO(), metav1.ListOptions{})
}
