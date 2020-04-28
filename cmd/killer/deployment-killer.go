package killer

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/p-program/kube-killer/core"
	"github.com/spf13/cobra"
	"k8s.io/client-go/kubernetes"
)

func init() {

}

var newDeployKillerCommand = &cobra.Command{
	Use:   "deploy",
	Short: "Kill kubernetes's deploy",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Print("")
	}}

var newDeploymentKillerCommand = &cobra.Command{
	Use:   "deployment",
	Short: "Kill kubernetes's deployment",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		kill()
	}}

func kill() {

}

func killResource(resources []string) {
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
	if err != nil {
		panic(err.Error())
	}
	// fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
}

func killByCondition() {

}
