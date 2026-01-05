package killer

import (
	"context"

	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// getNamespaces returns a list of namespaces to process
// If namespace is empty and allNamespaces is true, returns all namespaces except kube-system
// If namespace is specified, returns only that namespace
func getNamespaces(client *kubernetes.Clientset, namespace string, allNamespaces bool) ([]string, error) {
	if !allNamespaces && namespace != "" {
		return []string{namespace}, nil
	}
	
	// Get all namespaces
	nsList, err := client.CoreV1().Namespaces().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	namespaces := []string{}
	for _, ns := range nsList.Items {
		// Exclude kube-system as per kubectl-reap behavior
		if ns.Name != "kube-system" {
			namespaces = append(namespaces, ns.Name)
		}
	}
	
	return namespaces, nil
}

// processNamespaces processes a function for each namespace
func processNamespaces(client *kubernetes.Clientset, namespace string, allNamespaces bool, fn func(string) error) error {
	namespaces, err := getNamespaces(client, namespace, allNamespaces)
	if err != nil {
		return err
	}
	
	for _, ns := range namespaces {
		log.Info().Msgf("Processing namespace: %s", ns)
		if err := fn(ns); err != nil {
			log.Warn().Err(err).Msgf("Error processing namespace %s", ns)
			// Continue with other namespaces
		}
	}
	
	return nil
}

