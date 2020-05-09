package killer

import (
	"context"

	"github.com/p-program/kube-killer/core"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// PodKiller See https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/
type PodKiller struct {
	dryRun       bool
	namespace    string
	deleteOption metav1.DeleteOptions
}

// NewPodKiller NewPodKiller
// dryRun true: fake killer; flase true killer
// namespace can be ""， empty stands for current namespace
func NewPodKiller(dryRun bool, namespace string) *PodKiller {
	k := PodKiller{
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

// KillNonRunningPods kill Evicted,Completed pods
// TODO:need to test pod.Status.Phase=Terminating | Pending
func (k *PodKiller) KillNonRunningPods() error {
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		return err
	}
	listOption := metav1.ListOptions{FieldSelector: "status.phase!=Running"}
	pList, err := clientset.CoreV1().Pods(k.namespace).List(context.TODO(), listOption)
	if err != nil {
		return err
	}
	for _, pod := range pList.Items {
		podName := pod.ObjectMeta.Name
		log.Info().Msgf("delete pod: %s", podName)
		err = clientset.CoreV1().Pods(k.namespace).Delete(context.TODO(), podName, k.deleteOption)
		if err != nil {
			clientset.CoreV1().Pods(k.namespace).Delete(context.TODO(), podName, k.deleteOption)
		}
	}
	return nil
}

// KillAllPods delete all pods
func (k *PodKiller) KillAllPods() error {
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		return err
	}
	// metav1.DeleteOptions{}, metav1.ListOptions{LabelSelector: "test-rc-static=true"}
	return clientset.CoreV1().Pods(k.namespace).DeleteCollection(context.TODO(), k.deleteOption, metav1.ListOptions{})
}
