package cmd

import (
	"context"
	"strings"

	"github.com/p-program/kube-killer/core"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	v1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type Icebox struct {
	client    *kubernetes.Clientset
	dryRun    bool
	namespace string
}

func NewIcebox(namespace string) (*Icebox, error) {
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		return nil, err
	}
	i := Icebox{
		namespace: namespace,
		client:    clientset,
	}
	return &i, nil
}

func (k *Icebox) DryRun() *Icebox {
	k.dryRun = true
	return k
}

// Freeze set deployment/statefulset/ zero size
func (i *Icebox) Freeze(resourceType, resourceName string) error {
	scalePolicy := &autoscalingv1.Scale{Spec: v1.ScaleSpec{Replicas: 0}}
	updateOption := metav1.UpdateOptions{DryRun: []string{"All"}}
	if i.dryRun {
		updateOption.DryRun = []string{"All"}
	}
	switch strings.ToLower(resourceType) {
	case "d", "deploy", "deployment":
		_, err := i.client.AppsV1().Deployments(i.namespace).UpdateScale(context.TODO(), resourceName, scalePolicy, updateOption)
		if err != nil {
			return err
		}
		break
	case "ss", "statefulset":
		_, err := i.client.AppsV1().StatefulSets(i.namespace).UpdateScale(context.TODO(), resourceName, scalePolicy, updateOption)
		if err != nil {
			return err
		}
		break
	}
	return nil
}
