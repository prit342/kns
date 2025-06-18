package main

import (
	"context"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/prit342/kns/k8s"
	"github.com/prit342/kns/tui"
)

func main() {
	svc, err := k8s.NewService()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating Kubernetes service:", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	namespaces, err := svc.ListNamespaces(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not list namespaces:", err)
		os.Exit(1)
	}

	m := tui.NewModel(svc, namespaces)

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
