package killer

import (
	"context"
	"math/rand"
	"time"

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
	half         bool
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

func (k *PVKiller) SetHalf() *PVKiller {
	k.half = true
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
		// Available PVs are not bound, can be deleted
		return true
	}
	if phase == v1.VolumeBound {
		// Bound PVs are in use, should not be deleted
		return false
	}
	// Released, Failed volumes can be deleted
	return true
}

// Kill UnBoundPV or PVs not satisfying any PVCs
func (k *PVKiller) Kill() error {
	if k.mafia {
		if k.half {
			return k.KillHalfPVs()
		}
		return k.KillAllPVs()
	}
	return k.KillUnusedPVs()
}

func (k *PVKiller) KillAllPVs() error {
	volumeList, err := k.getAllPV()
	if err != nil {
		return err
	}
	for _, volume := range volumeList.Items {
		log.Info().Msgf("Deleting PV %s (phase: %s)", volume.Name, volume.Status.Phase)
		err = k.deletePV(volume.Name, k.dryRun)
		if err != nil {
			log.Warn().Err(err)
			continue
		}
	}
	return nil
}

func (k *PVKiller) KillHalfPVs() error {
	volumeList, err := k.getAllPV()
	if err != nil {
		return err
	}
	if len(volumeList.Items) == 0 {
		log.Info().Msg("No PVs to kill")
		return nil
	}

	// Randomly shuffle the PVs list
	pvList := volumeList.Items
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(pvList), func(i, j int) {
		pvList[i], pvList[j] = pvList[j], pvList[i]
	})

	// Calculate how many PVs to kill (half, rounded down)
	pvsToKill := len(pvList) / 2
	if pvsToKill == 0 {
		pvsToKill = 1 // At least kill one PV if there's only one
	}

	log.Info().Msgf("Killing %d out of %d PVs", pvsToKill, len(pvList))
	for i := 0; i < pvsToKill; i++ {
		volume := pvList[i]
		log.Info().Msgf("Deleting PV %s (phase: %s)", volume.Name, volume.Status.Phase)
		err = k.deletePV(volume.Name, k.dryRun)
		if err != nil {
			log.Warn().Err(err)
			continue
		}
	}
	return nil
}

func (k *PVKiller) KillUnusedPVs() error {
	volumeList, err := k.getAllPV()
	if err != nil {
		return err
	}

	// Get all PVCs across all namespaces to check which PVs are in use
	// Note: PVs are cluster-scoped, so we need to check all namespaces
	pvcList, err := k.client.CoreV1().PersistentVolumeClaims("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		log.Warn().Err(err).Msg("Failed to list PVCs, will only check PV phase")
	}

	// Build map of PVs that are bound to PVCs
	usedPVs := make(map[string]bool)
	for _, pvc := range pvcList.Items {
		if pvc.Spec.VolumeName != "" {
			usedPVs[pvc.Spec.VolumeName] = true
		}
	}

	for _, volume := range volumeList.Items {
		volumeName := volume.Name
		phase := volume.Status.Phase

		// Skip bound volumes that are in use
		if phase == v1.VolumeBound && usedPVs[volumeName] {
			continue
		}

		// Delete available, released, or failed volumes
		if phase == v1.VolumeAvailable || phase == v1.VolumeReleased || phase == v1.VolumeFailed {
			log.Info().Msgf("Deleting unused PV %s (phase: %s)", volumeName, phase)
			err = k.deletePV(volumeName, k.dryRun)
			if err != nil {
				log.Warn().Err(err)
				continue
			}
		}
	}
	return nil
}

func (k *PVKiller) deletePV(name string, dryRun bool) error {
	return k.client.CoreV1().PersistentVolumes().Delete(context.TODO(), name, k.deleteOption)
}

func (k *PVKiller) getAllPV() (*v1.PersistentVolumeList, error) {
	return k.client.CoreV1().PersistentVolumes().List(context.TODO(), metav1.ListOptions{})
}
