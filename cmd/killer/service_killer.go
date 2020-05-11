package killer

import (
	"context"
	"strings"

	"github.com/p-program/kube-killer/core"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
)

type ServiceKiller struct {
	client       *kubernetes.Clientset
	deleteOption metav1.DeleteOptions
	dryRun       bool
	mafia        bool
	namespace    string
}

func NewServiceKiller(namespace string) (*ServiceKiller, error) {
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		return nil, err
	}
	k := ServiceKiller{
		namespace: namespace,
		client:    clientset,
	}
	var gracePeriodSeconds int64 = 1
	k.deleteOption = metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}
	return &k, nil
}

func (k *ServiceKiller) DryRun() *ServiceKiller {
	k.dryRun = true
	k.deleteOption.DryRun = []string{"All"}
	return k
}

func (k *ServiceKiller) BlackHand() *ServiceKiller {
	k.mafia = true
	return k
}

func (k *ServiceKiller) Kill() error {
	services, err := k.client.CoreV1().Services(k.namespace).List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	for i := 0; i < len(services.Items); i++ {
		service := services.Items[i]
		if !k.DeserveDead(service) {
			continue
		}
		log.Warn().Msgf("deleting service %s in namespace %s", service.Name, service.Namespace)
		err = k.client.CoreV1().Services(service.Namespace).Delete(context.TODO(), service.Name, k.deleteOption)
		if err != nil {
			log.Error().Err(err)
		}
	}
	return nil
}

func (k *ServiceKiller) DeserveDead(resource interface{}) bool {
	if k.mafia {
		return true
	}
	service := resource.(v1.Service)
	clusterIP := service.Spec.ClusterIP
	isHeadlessService := strings.EqualFold("None", clusterIP)
	if isHeadlessService {
		// Headless Services must keep
		return false
	}
	selector := service.Spec.Selector
	pods, err := k.getPodsWithLabels(selector)
	if err != nil {
		log.Error().Err(err)
		// 饶你一马
		return false
	}
	if len(pods.Items) > 0 {
		return false
	}
	return true
}

func (k *ServiceKiller) getPodsWithLabels(labelMap map[string]string) (*v1.PodList, error) {
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		return nil, err
	}
	listOptions := metav1.ListOptions{}
	// labelSelector := labels.Set(map[string]string{"name": rcName}).AsSelector()
	labelSelector := labels.Set(labelMap).AsSelector()
	listOptions.LabelSelector = labelSelector.String()
	return clientset.CoreV1().Pods(k.namespace).List(context.TODO(), listOptions)
}
