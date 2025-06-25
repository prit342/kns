package k8s

import (
	"context"
	"fmt"
	"path/filepath"

	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// check if Service implements NameSpaceLister interface
var _ NameSpaceListerConfigUpdater = (*Service)(nil)

// Service is a struct that provides methods to list Kubernetes namespaces
type Service struct {
	client             *kubernetes.Clientset
	kubeConfigLocation string
}

// GetKubeConfigLocation returns the location of the kubeconfig file used by the Service
func (s *Service) GetKubeConfigLocation() string {
	return s.kubeConfigLocation
}

// NewService creates a new instance of Service that connects to the Kubernetes API
func NewService() (*Service, error) {
	var config *rest.Config
	var err error

	kubeconfig := os.Getenv("KUBECONFIG")
	// first try to use the KUBECONFIG environment variable
	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		// if KUBECONFIG is not set, try to use the default kubeconfig file
		// get the home directory to construct the default kubeconfig path
		// this is usually located at $HOME/.kube/config
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("could not get home directory: %w", err)
		}
		kubeconfig = filepath.Join(homeDir, ".kube", "config")
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to build config from kubeconfig file %s: %w", kubeconfig, err)
		}
	}

	if err != nil {
		// if we cannot read the kubeconfig file, try to use in-cluster config
		// extra step to check if we are running inside a Kubernetes pod which
		// means we do not have a kubeconfig file
		if _, err = rest.InClusterConfig(); err == nil {
			// InClusterConfig is used when running inside a Kubernetes pod
			return nil, fmt.Errorf("in-cluster configs are not supported, we need kubeconfig file to update: %w", err)
		}
		return nil, fmt.Errorf("failed to read kubernetes config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernets clientset: %w", err)
	}
	return &Service{
		client:             clientset,
		kubeConfigLocation: kubeconfig,
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
