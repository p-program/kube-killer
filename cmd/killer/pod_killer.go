package killer

import (
	"context"

	"github.com/p-program/kube-killer/core"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/pager"
)

// PodKiller See https://kubernetes.io/docs/concepts/workloads/pods/pod-lifecycle/
type PodKiller struct {
	client       *kubernetes.Clientset
	deleteOption metav1.DeleteOptions
	dryRun       bool
	mafia        bool
	namespace    string
}

// NewPodKiller NewPodKiller
// namespace can be ""， empty stands for current namespace
func NewPodKiller(namespace string) (*PodKiller, error) {
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		return nil, err
	}
	k := PodKiller{
		namespace: namespace,
		client:    clientset,
	}
	log.Info().Msgf("namespace:%s", namespace)
	var gracePeriodSeconds int64 = 1
	k.deleteOption = metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}
	return &k, nil
}

func (k *PodKiller) BlackHand() *PodKiller {
	k.mafia = true
	return k
}

func (k *PodKiller) DryRun() *PodKiller {
	k.dryRun = true
	k.deleteOption.DryRun = []string{"All"}
	return k
}

func (k *PodKiller) DeserveDead(resource interface{}) bool {
	if k.mafia {
		return true
	}
	//because FieldSelector support "status.phase!=Running" , set true here
	return true
}

func (k *PodKiller) Kill() error {
	if k.mafia {
		return k.KillAllPods()
	}
	return k.KillNonRunningPods()
}

// KillNonRunningPods kill Evicted,Completed pods
// TODO:need to test pod.Status.Phase=Terminating | Pending
func (k *PodKiller) KillNonRunningPods() error {
	log.Warn().Msg("KillNonRunningPods")
	listOption := metav1.ListOptions{FieldSelector: "status.phase!=Running"}
	pList, err := k.client.CoreV1().Pods(k.namespace).List(context.TODO(), listOption)
	if err != nil {
		return err
	}
	for _, pod := range pList.Items {
		podName := pod.ObjectMeta.Name
		log.Warn().Msgf("delete pod: %s in namespace %s ", podName, pod.Namespace)
		err = k.client.CoreV1().Pods(pod.Namespace).Delete(context.TODO(), podName, k.deleteOption)
		if err != nil {
			log.Err(err)
		}
	}
	return nil
}

func (k *PodKiller) getPods(labelMap map[string]string) (*v1.PodList, error) {
	listOptions := metav1.ListOptions{}
	if len(labelMap) > 0 {
		labelSelector := labels.Set(labelMap).AsSelector()
		listOptions.LabelSelector = labelSelector.String()
	}
	return k.client.CoreV1().Pods(k.namespace).List(context.TODO(), listOptions)
}

func (k *PodKiller) getAllPodsInCurrentNamespace() ([]*v1.Pod, error) {
	p := pager.New(pager.SimplePageFunc(func(opts metav1.ListOptions) (runtime.Object, error) {
		list, err := k.client.CoreV1().Pods(namespace).List(context.TODO(), opts)
		if err != nil {
			return nil, errors.Wrap(err, "cannot retrieve pods")
		}
		return list, nil
	}))
	p.PageSize = 500
	ctx := context.Background()
	pods := []*v1.Pod{}
	err := p.EachListItem(ctx, metav1.ListOptions{}, func(obj runtime.Object) error {
		pod, ok := obj.(*v1.Pod)
		if !ok {
			return errors.Errorf("this is not a pod: %#v", obj)
		}
		pods = append(pods, pod)
		return nil
	})
	if err != nil {
		return []*v1.Pod{}, errors.Wrap(err, "cannot iterate secrets")
	}
	return pods, nil
}

func (k *PodKiller) KillHalfPods() error {
	//TODO
	return nil
}

// KillAllPods delete all pods
func (k *PodKiller) KillAllPods() error {
	log.Warn().Msg("KillAllPods")
	pods, err := k.getPods(nil)
	if err != nil {
		return err
	}
	for i := 0; i < len(pods.Items); i++ {
		pod := pods.Items[i]
		podName := pod.Name
		log.Warn().Msgf("delete pod %s in namespace %s ", podName, pod.Namespace)
		err = k.client.CoreV1().Pods(k.namespace).Delete(context.TODO(), podName, k.deleteOption)
		if err != nil {
			log.Error().Err(err)
		}
	}
	// will return {SelfLink:"", ResourceVersion:"", Continue:"", RemainingItemCount:(*int64)(nil)}, Status:"Failure", Message:"the server does not allow this method on the requested resource", Reason:"MethodNotAllowed", Details:(*v1.StatusDetails)(0xc000454180), Code:405}
	//return k.client.CoreV1().Pods(k.namespace).DeleteCollection(context.TODO(), k.deleteOption, metav1.ListOptions{})
	return nil
}
