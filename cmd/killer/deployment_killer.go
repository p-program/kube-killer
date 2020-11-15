package killer

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/p-program/kube-killer/core"
	"k8s.io/client-go/kubernetes"
)

func init() {

}

// var newDeployKillerCommand = &cobra.Command{
// 	Use:   "deploy",
// 	Short: "Kill kubernetes's deploy",
// 	Long:  ``,
// 	Run: func(cmd *cobra.Command, args []string) {
// 		fmt.Print("")
// 	}}

type DeploymentKiller struct {
	client       *kubernetes.Clientset
	deleteOption metav1.DeleteOptions
	dryRun       bool
	mafia        bool
	namespace    string
}

func NewDeploymentKiller(namespace string) (*DeploymentKiller, error) {
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		return nil, err
	}
	k := DeploymentKiller{
		namespace: namespace,
		client:    clientset,
	}
	var gracePeriodSeconds int64 = 1
	k.deleteOption = metav1.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	}
	return &k, nil
}

func (k *DeploymentKiller) BlackHand() *DeploymentKiller {
	k.mafia = true
	return k
}

func (k *DeploymentKiller) DryRun() *DeploymentKiller {
	k.dryRun = true
	k.deleteOption.DryRun = []string{"All"}
	return k
}

func (k *DeploymentKiller) Kill() {

}

func (k *DeploymentKiller) kill2(resources []string) {
	fmt.Printf("bilibili:")
	fmt.Println(resources)
	// resourceType := resources[0]
	resourceName := resources[1]
	// create the clientset
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		panic(err.Error())
	}
	namespace := ""
	err = clientset.CoreV1().Pods(namespace).Delete(context.TODO(), resourceName, metav1.DeleteOptions{})
	// clientset.AppsV1().Deployments()
	if err != nil {
		panic(err.Error())
	}
	// fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
}
