package cmd

import (
	"fmt"

	"github.com/p-program/kube-killer/core"
	"github.com/spf13/cobra"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

var resources []string

func newKillCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kill",
		Short: "Kill kubernetes's resource",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			kill(resources)
		}}
	flags := cmd.Flags()
	flags.StringArrayVarP(&resources, "kill", "k", nil, "kill resource")
	return cmd
}

func kill(resources []string) {
	fmt.Printf("bilibili:")
	fmt.Println(resources)
	resourceType := resources[0]
	resourceName := resources[1]
	// create the clientset
	clientset, err := kubernetes.NewForConfig(core.GLOBAL_KUBERNETES_CONFIG)
	if err != nil {
		panic(err.Error())
	}
	pods, err := clientset.CoreV1().Pods("").List(metav1.ListOptions{})
	clientset.CoreV1().Pods("").Delete(resourceName)
	if err != nil {
		panic(err.Error())
	}
	fmt.Printf("There are %d pods in the cluster\n", len(pods.Items))
}
