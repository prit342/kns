package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/prit342/kns/k8s"
	"github.com/prit342/kns/tui"
)

var (
	// Version is the version of the application.
	version = "dev"
	// BuildTime is the time when the application was built.
	buildDate = "unknown"
	// GitCommit is the git commit hash when the application was built.
	gitCommit = "unknown"
)

const helpMessage = `
Usage: %s [k8s-namespace|version|help]
  k8s-namespace: Switch to the specified Kubernetes namespace
  version: Show the version of kns
  help: Show this help message
  no arguments will launch the TUI to switch namespaces
`
const versionMessage = `%s - Kubernetes Namespace Switcher
Version: %s
Build Date: %s
Git Commit: %s
`

func main() {
	svc, err := k8s.NewService()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating Kubernetes service:", err)
		os.Exit(1)
	}

	// Create a context with a timeout for the operations
	// 15 seconds should be enough for most operations
	ctx, cancel := context.WithTimeout(context.Background(), 14*time.Second)
	defer cancel()

	// check if we have more than 1 argument passed to the program
	if len(os.Args) > 2 {
		fmt.Fprintf(os.Stderr, helpMessage, os.Args[0])
		os.Exit(1)
	}

	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "version":
			fmt.Printf(versionMessage, os.Args[0], version, buildDate, gitCommit)
			os.Exit(0)
		case "help":
			fmt.Printf(helpMessage, os.Args[0])
			os.Exit(0)
		default:
			// try to switch to the specified namespace
			namespace := os.Args[1]
			if err := svc.UpdateKubeConfigWithNamespace(ctx, namespace); err != nil {
				fmt.Fprintf(os.Stderr, "\n\nError switching to namespace '%s': %v\n", namespace, err)
				os.Exit(1)
			}
			fmt.Printf("\nupdated kubeconfig and switched to namespace '%s'\n", namespace)
			os.Exit(0)
		}
	}

	// if we reach here, it means we have no arguments, so we can start the TUI
	namespaces, err := svc.ListNamespaces(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not list namespaces:", err)
		os.Exit(1)
	}

	kubeConfigLocation := svc.GetKubeConfigLocation(ctx)
	// we have no arguments, so we launch the TUI to switch namespaces
	app := tui.NewTUI(namespaces, svc, kubeConfigLocation)
	if err := app.Run(ctx); err != nil {
		fmt.Printf("\nError running program: %s", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully updated kubeconfig at %s", kubeConfigLocation)

}
