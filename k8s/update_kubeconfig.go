package k8s

import (
	"context"
	"fmt"

	"k8s.io/client-go/tools/clientcmd"
)

// UpdateKubeConfigWithNamespace updates the kubeconfig to use the specified namespace
func (s *Service) UpdateKubeConfigWithNamespace(ctx context.Context, namespace string) error {
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
