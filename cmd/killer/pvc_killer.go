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

type PVCKiller struct {
	client       *kubernetes.Clientset
	deleteOption metav1.DeleteOptions
	dryRun       bool
	mafia        bool
	half         bool
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

func (k *PVCKiller) SetHalf() *PVCKiller {
	k.half = true
	return k
}

func (k *PVCKiller) DryRun() *PVCKiller {
	k.dryRun = true
	k.deleteOption.DryRun = []string{"All"}
	return k
}

// DeserveDead Pending/Lost PVC deserve to die, or PVC not used by any Pod
func (k *PVCKiller) DeserveDead(resource interface{}) bool {
	if k.mafia {
		return true
	}
	pvc := resource.(v1.PersistentVolumeClaim)
	phase := pvc.Status.Phase
	if phase == v1.ClaimBound {
		// Check if it's used by any Pod
		return false
	}
	return true
}

// Kill Kill unused PVCs
func (k *PVCKiller) Kill() error {
	if k.mafia {
		if k.half {
			return k.KillHalfPVCs()
		}
		return k.KillAllPVCs()
	}
	return k.KillUnusedPVCs()
}

func (k *PVCKiller) KillAllPVCs() error {
	list, err := k.client.CoreV1().PersistentVolumeClaims(k.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, pvc := range list.Items {
		log.Info().Msgf("Deleting pvc %s in namespace %s", pvc.Name, k.namespace)
		err = k.client.CoreV1().PersistentVolumeClaims(k.namespace).Delete(context.TODO(), pvc.Name, k.deleteOption)
		if err != nil {
			log.Err(err)
		}
	}
	return nil
}

func (k *PVCKiller) KillHalfPVCs() error {
	list, err := k.client.CoreV1().PersistentVolumeClaims(k.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	if len(list.Items) == 0 {
		log.Info().Msg("No PVCs to kill")
		return nil
	}

	// Randomly shuffle the PVCs list
	pvcList := list.Items
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(pvcList), func(i, j int) {
		pvcList[i], pvcList[j] = pvcList[j], pvcList[i]
	})

	// Calculate how many PVCs to kill (half, rounded down)
	pvcsToKill := len(pvcList) / 2
	if pvcsToKill == 0 {
		pvcsToKill = 1 // At least kill one PVC if there's only one
	}

	log.Info().Msgf("Killing %d out of %d PVCs", pvcsToKill, len(pvcList))
	for i := 0; i < pvcsToKill; i++ {
		pvc := pvcList[i]
		log.Info().Msgf("Deleting pvc %s in namespace %s", pvc.Name, k.namespace)
		err = k.client.CoreV1().PersistentVolumeClaims(k.namespace).Delete(context.TODO(), pvc.Name, k.deleteOption)
		if err != nil {
			log.Err(err)
		}
	}
	return nil
}

func (k *PVCKiller) KillUnusedPVCs() error {
	// Get all PVCs
	pvcList, err := k.client.CoreV1().PersistentVolumeClaims(k.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}

	// Get all Pods to check which PVCs are in use
	podKiller, err := NewPodKiller(k.namespace)
	if err != nil {
		return err
	}
	pods, err := podKiller.getAllPodsInCurrentNamespace()
	if err != nil {
		return err
	}

	// Build map of used PVCs
	usedPVCs := make(map[string]bool)
	for _, pod := range pods {
		for _, volume := range pod.Spec.Volumes {
			if volume.PersistentVolumeClaim != nil {
				usedPVCs[volume.PersistentVolumeClaim.ClaimName] = true
			}
		}
	}

	// Delete unused PVCs
	for _, pvc := range pvcList.Items {
		// Skip bound PVCs that are in use
		if pvc.Status.Phase == v1.ClaimBound && usedPVCs[pvc.Name] {
			continue
		}
		// Delete unbound or unused PVCs
		if pvc.Status.Phase != v1.ClaimBound || !usedPVCs[pvc.Name] {
			log.Info().Msgf("Deleting unused pvc %s in namespace %s", pvc.Name, k.namespace)
			err = k.client.CoreV1().PersistentVolumeClaims(k.namespace).Delete(context.TODO(), pvc.Name, k.deleteOption)
			if err != nil {
				log.Err(err)
			}
		}
	}
	return nil
}
