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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// check if we have more than 1 argument passed to the program
	if len(os.Args) > 2 {
		fmt.Fprintf(os.Stderr, helpMessage, os.Args[0])
		os.Exit(1)
	}

	if len(os.Args) == 2 { // we have exactly one argument, we check what it is
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
			// TODO: do not verify if the namespace exits using the checkIfExists flag,
			// might have to introduce a flag later to control this behaviour
			if err := svc.UpdateKubeConfigWithNamespace(ctx, namespace, false); err != nil {
				// if we cannot switch to the namespace, we print an error message and exit
				fmt.Fprintf(os.Stderr, "\n\n❌ ❌ Error switching to namespace %q: %v\n", namespace, err)
				os.Exit(1)
			}
			fmt.Printf("\n✅ Updated kube config at %q and switched to namespace %q\n",
				svc.GetKubeConfigLocation(), namespace)
			return
		}
	}

	// if we reach here, it means we have no arguments, so we can start the TUI
	namespaces, err := svc.ListNamespaces(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Could not list namespaces:", err)
		os.Exit(1)
	}

	// we have no arguments, so we launch the TUI to switch namespaces
	m := tui.NewModel(svc, namespaces, svc.GetKubeConfigLocation())

	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
