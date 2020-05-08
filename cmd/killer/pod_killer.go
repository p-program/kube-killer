package killer

import (
	"context"

	"github.com/p-program/kube-killer/core"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PodKiller See https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/
type PodKiller struct {
}

func NewPodKiller() *PodKiller {
	k := PodKiller{}
	return &k
}

// KillNonRunningPods kill Evicted,Completed pods
// TODO:need to test pod.Status.Phase=Terminating | Pending
func (k *PodKiller) KillNonRunningPods() {
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		panic(err.Error())
	}
	deleteOption := metav1.DeleteOptions{}
	*deleteOption.GracePeriodSeconds = int64(1)
	namespace := ""
	listOption := metav1.ListOptions{FieldSelector: "status.phase!=Running"}
	clientset.CoreV1().Pods(namespace).DeleteCollection(context.TODO(), deleteOption, listOption)
}

// KillAllPods delete all pods
func (k *PodKiller) KillAllPods() {
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		panic(err.Error())
	}
	deleteOption := metav1.DeleteOptions{}
	*deleteOption.GracePeriodSeconds = int64(1)
	namespace := ""
	// metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: "test-rc-static=true"}
	clientset.CoreV1().Pods(namespace).DeleteCollection(context.TODO(), deleteOption, metav1.ListOptions{})
}
