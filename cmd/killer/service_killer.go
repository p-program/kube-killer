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
	dryRun       bool
	namespace    string
	deleteOption metav1.DeleteOptions
}

func NewServiceKiller(dryRun bool, namespace string) *ServiceKiller {
	k := ServiceKiller{
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

func (k *ServiceKiller) serviceDeserveDead(service *v1.Service) bool {
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

func (k *ServiceKiller) Kill(name string) error {
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		return err
	}
	clientset.CoreV1().Services(k.namespace).Delete(context.TODO(), name, k.deleteOption)
	return nil
}
