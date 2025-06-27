package k8s

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// check if Service implements NameSpaceLister interface
var _ NameSpaceListerConfigUpdater = (*Service)(nil)

// Service is a struct that provides methods to list Kubernetes namespaces
type Service struct {
	client             kubernetes.Interface
	kubeConfigLocation string
}

// GetKubeConfigLocation returns the location of the kubeconfig file used by the Service
func (s *Service) GetKubeConfigLocation() string {
	return s.kubeConfigLocation
}

// NewService creates a new instance of Service that connects to the Kubernetes API
func NewService() (*Service, error) {

	// Load kubeconfig using standard loading rules - KUBECONFIG or ~/.kube/config
	// we are not looking for an in-cluster configuration
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	kubeconfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		loadingRules,
		&clientcmd.ConfigOverrides{},
	)

	// Create a client config from the kubeconfig
	// This will use the default kubeconfig file if no specific file is provided
	// or if the KUBECONFIG environment variable is not set
	// If the kubeconfig file is not found, it will return an error
	config, err := kubeconfig.ClientConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create client config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernets clientset: %w", err)
	}
	return &Service{
		client:             clientset,
		kubeConfigLocation: loadingRules.GetDefaultFilename(),
	}, nil
}

// ListNamespaces - return the lists all namespaces in the Kubernetes cluster
func (s *Service) ListNamespaces(ctx context.Context) ([]string, error) {
	nss, err := s.client.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %w", err)
	}
	names := make([]string, 0, len(nss.Items))
	for _, ns := range nss.Items {
		names = append(names, ns.Name)
	}
	return names, nil
}
