package killer

import (
	"context"
	"math/rand"
	"time"

	"github.com/p-program/kube-killer/core"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/pager"
)

type ConfigmapKiller struct {
	client       *kubernetes.Clientset
	deleteOption metav1.DeleteOptions
	dryRun       bool
	mafia        bool
	half         bool
	namespace    string
}

func NewConfigmapKiller(namespace string) (*ConfigmapKiller, error) {
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		return nil, err
	}
	k := ConfigmapKiller{
		namespace: namespace,
		client:    clientset,
	}
	var gracePeriodSeconds int64 = 1
	k.deleteOption = metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}
	return &k, nil
}

func (k *ConfigmapKiller) DryRun() *ConfigmapKiller {
	k.dryRun = true
	k.deleteOption.DryRun = []string{"All"}
	return k
}

func (k *ConfigmapKiller) BlackHand() *ConfigmapKiller {
	k.mafia = true
	return k
}

func (k *ConfigmapKiller) SetHalf() *ConfigmapKiller {
	k.half = true
	return k
}

func (k *ConfigmapKiller) Kill() error {
	if k.mafia {
		if k.half {
			return k.KillHalfConfigMaps()
		}
		return k.KillAllConfigMaps()
	}
	return k.KillUnusedConfigMaps()
}

func (k *ConfigmapKiller) KillAllConfigMaps() error {
	configMaps, err := k.getAllConfigMapsInCurrentNamespace()
	if err != nil {
		return err
	}
	for _, cm := range configMaps {
		log.Info().Msgf("Deleting configmap %s in namespace %s", cm.Name, k.namespace)
		err = k.client.CoreV1().ConfigMaps(k.namespace).Delete(context.TODO(), cm.Name, k.deleteOption)
		if err != nil {
			log.Err(err)
		}
	}
	return nil
}

func (k *ConfigmapKiller) KillHalfConfigMaps() error {
	configMaps, err := k.getAllConfigMapsInCurrentNamespace()
	if err != nil {
		return err
	}
	if len(configMaps) == 0 {
		log.Info().Msg("No configmaps to kill")
		return nil
	}
	
	// Randomly shuffle the configmaps list
	cmList := make([]*v1.ConfigMap, len(configMaps))
	copy(cmList, configMaps)
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(cmList), func(i, j int) {
		cmList[i], cmList[j] = cmList[j], cmList[i]
	})
	
	// Calculate how many configmaps to kill (half, rounded down)
	cmsToKill := len(cmList) / 2
	if cmsToKill == 0 {
		cmsToKill = 1 // At least kill one configmap if there's only one
	}
	
	log.Info().Msgf("Killing %d out of %d configmaps", cmsToKill, len(cmList))
	for i := 0; i < cmsToKill; i++ {
		cm := cmList[i]
		log.Info().Msgf("Deleting configmap %s in namespace %s", cm.Name, k.namespace)
		err = k.client.CoreV1().ConfigMaps(k.namespace).Delete(context.TODO(), cm.Name, k.deleteOption)
		if err != nil {
			log.Err(err)
		}
	}
	return nil
}

func (k *ConfigmapKiller) KillUnusedConfigMaps() error {
	podKiller, err := NewPodKiller(k.namespace)
	if err != nil {
		return err
	}
	pods, err := podKiller.getAllPodsInCurrentNamespace()
	if err != nil {
		return err
	}
	log.Info().Msgf("Retrieving ConfigMaps in %s...", k.namespace)
	configMaps, err := k.getAllConfigMapsInCurrentNamespace()
	if err != nil {
		return err
	}
	unusedConfigMaps, err := k.detectUnusedConfigMaps(pods, configMaps)
	if err != nil {
		return err
	}
	for _, cm := range unusedConfigMaps {
		log.Info().Msgf("Deleting unused configmap %s in namespace %s", cm.Name, k.namespace)
		err = k.client.CoreV1().ConfigMaps(k.namespace).Delete(context.TODO(), cm.Name, k.deleteOption)
		if err != nil {
			log.Err(err)
		}
	}
	return nil
}

func (k *ConfigmapKiller) getAllConfigMapsInCurrentNamespace() ([]*v1.ConfigMap, error) {
	p := pager.New(pager.SimplePageFunc(func(opts metav1.ListOptions) (runtime.Object, error) {
		list, err := k.client.CoreV1().ConfigMaps(k.namespace).List(context.TODO(), opts)
		if err != nil {
			return nil, errors.Wrap(err, "cannot retrieve configmaps")
		}
		return list, nil
	}))
	p.PageSize = 500
	ctx := context.Background()
	configMaps := []*v1.ConfigMap{}
	err := p.EachListItem(ctx, metav1.ListOptions{}, func(obj runtime.Object) error {
		cm, ok := obj.(*v1.ConfigMap)
		if !ok {
			return errors.Errorf("this is not a configmap: %#v", obj)
		}
		configMaps = append(configMaps, cm)
		return nil
	})
	if err != nil {
		return []*v1.ConfigMap{}, errors.Wrap(err, "cannot iterate configmaps")
	}
	return configMaps, nil
}

func (k *ConfigmapKiller) detectUnusedConfigMaps(pods []*v1.Pod, configMaps []*v1.ConfigMap) ([]*v1.ConfigMap, error) {
	usedConfigMapNames := map[string]bool{}
	
	// Check configmaps used by Pods
	for _, pod := range pods {
		for _, container := range pod.Spec.Containers {
			for _, envFrom := range container.EnvFrom {
				if envFrom.ConfigMapRef != nil {
					usedConfigMapNames[envFrom.ConfigMapRef.Name] = true
				}
			}
			for _, env := range container.Env {
				if env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil {
					usedConfigMapNames[env.ValueFrom.ConfigMapKeyRef.Name] = true
				}
			}
		}
		for _, volume := range pod.Spec.Volumes {
			if volume.ConfigMap != nil {
				usedConfigMapNames[volume.ConfigMap.Name] = true
			}
			if volume.Projected != nil {
				for _, source := range volume.Projected.Sources {
					if source.ConfigMap != nil {
						usedConfigMapNames[source.ConfigMap.Name] = true
					}
				}
			}
		}
	}
	
	unused := []*v1.ConfigMap{}
	for _, cm := range configMaps {
		if !usedConfigMapNames[cm.Name] {
			unused = append(unused, cm)
		}
	}

	return unused, nil
}
