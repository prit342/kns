package k8s

import (
	"context"
)

// NameSpaceLister is an interface that defines the method to list namespaces
type NameSpaceListerConfigUpdater interface {
	ListNamespaces(ctx context.Context) ([]string, error)
	GetKubeConfigLocation() string
	UpdateKubeConfigWithNamespace(context.Context, string, bool) error
}
