package tui

import (
	"context"
	"fmt"

	"github.com/prit342/kns/k8s"
	"github.com/rivo/tview"
)

// tui is the model for the application, holding the list of namespaces and the selected namespace.
type tui struct {
	namespaces         []string                         // List of namespaces to display
	svc                k8s.NameSpaceListerConfigUpdater // svc is the Kubernetes service used to interact with the cluster
	kubeconfigLocation string                           // file path of the kubeconfig file
	app                *tview.Application
	err                error // error to hold any errors that occur during the application lifecycle
}

func NewTUI(namespaces []string, svc k8s.NameSpaceListerConfigUpdater, kubeconfigLocation string) *tui {
	return &tui{
		namespaces:         namespaces,
		svc:                svc,
		kubeconfigLocation: kubeconfigLocation,
		app:                tview.NewApplication(),
	}
}

// NewAPP initializes a new TUI application with a list of namespaces.
func (t *tui) Run(ctx context.Context) error {
	selectedNamespace := "" // this will hold the user’s selection

	list := tview.NewList()
	for _, ns := range t.namespaces {
		ns := ns // avoid loop capture issue

		list.AddItem(ns, "", 0, func() {
			selectedNamespace = ns
			// call the function to handle the namespace selection
			t.handleNamespaceSelection(ctx, selectedNamespace)
			t.app.Stop()
		})
	}

	list.SetTitle("Select Kubernetes Namespace").
		SetBorder(true).
		SetBackgroundColor(tview.Styles.GraphicsColor).
		SetTitleColor(tview.Styles.TitleColor).
		SetTitleAlign(tview.AlignCenter).
		SetBorderPadding(1, 1, 2, 2).
		SetTitleColor(tview.Styles.TitleColor).
		SetTitleAlign(tview.AlignLeft)

	t.app.SetRoot(list, true).SetFocus(list)

	if err := t.app.Run(); err != nil {
		return fmt.Errorf("error running TUI application: %w", err)
	}
	return t.err
}

// handleNamespaceSelection updates the kubeconfig with the selected namespace and prints a message.
func (t *tui) handleNamespaceSelection(ctx context.Context, ns string) {
	if err := t.svc.UpdateKubeConfigWithNamespace(ctx, ns); err != nil {
		t.err = fmt.Errorf("error updating kubeconfig with namespace %s: %w", ns, err)
	}
}
