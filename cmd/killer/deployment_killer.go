package killer

import (
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
