package killer

import (
	"context"

	"github.com/p-program/kube-killer/core"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// StatefulSetKiller StatefulSet must die ！
type StatefulSetKiller struct {
	client       *kubernetes.Clientset
	deleteOption metav1.DeleteOptions
	dryRun       bool
	mafia        bool
	namespace    string
}

func NewStatefulSetKiller(namespace string) (*StatefulSetKiller, error) {
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		return nil, err
	}
	k := StatefulSetKiller{
		namespace: namespace,
		client:    clientset,
	}
	var gracePeriodSeconds int64 = 0
	k.deleteOption = metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}
	return &k, nil
}

func (k *StatefulSetKiller) DryRun() *StatefulSetKiller {
	k.dryRun = true
	k.deleteOption.DryRun = []string{"All"}
	return k
}

func (k *StatefulSetKiller) BlackHand() *StatefulSetKiller {
	k.mafia = true
	return k
}

func (k *StatefulSetKiller) Kill() error {
	if k.mafia {
		return k.KillAllStatefulSets()
	}
	return k.KillAllStatefulSets()
}

func (k *StatefulSetKiller) KillAllStatefulSets() error {
	log.Warn().Msg("KillAllStatefulSets")
	statefulSets, err := k.client.AppsV1().StatefulSets(k.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for _, sts := range statefulSets.Items {
		log.Warn().Msgf("deleting statefulset %s in namespace %s", sts.Name, k.namespace)
		err = k.client.AppsV1().StatefulSets(k.namespace).Delete(context.TODO(), sts.Name, k.deleteOption)
		if err != nil {
			log.Error().Err(err)
		}
	}
	return nil
}
