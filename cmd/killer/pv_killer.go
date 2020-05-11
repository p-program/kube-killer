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
	client       *kubernetes.Clientset
	deleteOption metav1.DeleteOptions
	dryRun       bool
	mafia        bool
}

func NewPVKiller() (*PVKiller, error) {
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		return nil, err
	}
	k := PVKiller{client: clientset}
	var gracePeriodSeconds int64 = 0
	k.deleteOption = metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}
	return &k, err
}

func (k *PVKiller) BlackHand() *PVKiller {
	k.mafia = true
	return k
}

func (k *PVKiller) DryRun() *PVKiller {
	k.dryRun = true
	k.deleteOption.DryRun = []string{"All"}
	return k
}

// DeserveDead true stands for dead
func (k *PVKiller) DeserveDead(resource interface{}) bool {
	if k.mafia {
		return true
	}
	pv := resource.(v1.PersistentVolume)
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

// Kill UnBoundPV
func (k *PVKiller) Kill() error {
	volumeList, err := k.getAllPV()
	if err != nil {
		return err
	}
	for i := 0; i < len(volumeList.Items); i++ {
		volume := volumeList.Items[i]
		volumeName := volume.Name
		if !k.DeserveDead(&volume) {
			continue
		}
		log.Info().Msgf("Volume Info { volumeName: %s ;volume.Status.Phase: %s }", volumeName, volume.Status.Phase)
		err = k.deletePV(volumeName, k.dryRun)
		if err != nil {
			//log but continue
			log.Warn().Err(err)
			continue
		}
	}
	return err
}

func (k *PVKiller) deletePV(name string, dryRun bool) error {
	return k.client.CoreV1().PersistentVolumes().Delete(context.TODO(), name, k.deleteOption)
}

func (k *PVKiller) getAllPV() (*v1.PersistentVolumeList, error) {
	return k.client.CoreV1().PersistentVolumes().List(context.TODO(), metav1.ListOptions{})
}
