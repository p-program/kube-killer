package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var GLOBAL_KUBERNETES_CONFIG *rest.Config

func init() {
	initKubernetesConfig()
}

func initKubernetesConfig() {
	var kubeconfig string
	home := homeDir()
	if home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	} else {
		err := errors.New("kubeconfig not found")
		panic(err)
	}
	fmt.Printf("kubeconfig: %s \n", kubeconfig)
	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		fmt.Print(err)
	}
	GLOBAL_KUBERNETES_CONFIG = config
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
