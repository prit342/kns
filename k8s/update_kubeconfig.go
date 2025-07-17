package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"
)

// UpdateKubeConfigWithNamespace updates the kubeconfig to use the specified namespace
func (s *Service) UpdateKubeConfigWithNamespace(
	ctx context.Context, // context for the operation
	namespace string, // the namespace to switch to in the kubeconfig
	checkIfExists bool, // whether to check if the namespace exists before updating
) error {
	// Check if the namespace is valid
	if namespace == "" {
		return fmt.Errorf("namespace cannot be empty")
	}
	if checkIfExists {
		// check if namespace actually exists in the kubernets cluster
		_, err := s.client.CoreV1().Namespaces().Get(ctx, namespace, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("error reading namespace '%s': %w", namespace, err)
		}
	}
	// Get the current context
	config, err := clientcmd.LoadFromFile(s.kubeConfigLocation)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %w", err)
	}

	// try to find the current context
	currentContext := config.CurrentContext
	if currentContext == "" {
		return fmt.Errorf("no current context set in kubeconfig")
	}

	// update the namespace for the current context
	config.Contexts[currentContext].Namespace = namespace

	// Save back to file
	return clientcmd.WriteToFile(*config, s.kubeConfigLocation)
}
