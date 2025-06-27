package tui

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/prit342/kns/k8s"
)

const (
	listHeight   = 20
	defaultWidth = 20
)

var (
	titleStyle        = lipgloss.NewStyle().MarginLeft(2)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
	quitTextStyle     = lipgloss.NewStyle().Margin(1, 0, 2, 4)
)

type item string

func (i item) FilterValue() string { return string(i) }

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(item)
	if !ok {
		return
	}

	str := fmt.Sprintf("%d. %s", index+1, i)

	fn := itemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

// model holds the state of the application
type model struct {
	list               list.Model                       // list is the list of namespaces to choose from
	choice             string                           // choice is the selected namespace
	quitting           bool                             // quitting indicates whether the user has chosen to exit the program
	svc                k8s.NameSpaceListerConfigUpdater // svc is the Kubernetes service used to interact with the cluster
	err                error                            // err is used to store any error that occurs during the program execution
	kubeconfigLocation string                           // file path of the kubeconfig file
}

// Based on the example of lists at https://github.com/charmbracelet/bubbletea/tree/main/tutorials/commands
//
// Init initializes the model and returns a command to run
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles user input and updates the model accordingly
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter", " ":
			i, ok := m.list.SelectedItem().(item)
			if ok {
				m.choice = string(i)
			}
			if err := m.svc.UpdateKubeConfigWithNamespace(context.Background(), m.choice); err != nil {
				m.err = fmt.Errorf("failed to update kubeconfig with namespace %s: %w", m.choice, err)
				m.quitting = true
			}

			return m, tea.Quit
		}
	}

	return m, cmd
}

// View renders the current state of the model as a string
// It displays the list of namespaces, any error messages, or a confirmation message
func (m model) View() string {

	// we check if we had any errors during update
	if m.err != nil {
		return quitTextStyle.Render(fmt.Sprintf("\n\nError: %s", m.err))
	}
	// if a namespace was selected, we display a confirmation message
	if m.choice != "" {
		return quitTextStyle.Render(
			fmt.Sprintf("\n✅ switched namespace to '%s' in kubeconfig file '%s'", m.choice, m.kubeconfigLocation),
		)
	}
	// if the user is quitting without making any changes, we display a message
	if m.quitting {
		return quitTextStyle.Render(
			fmt.Sprintf("\n\n✅ no updated were made to kubeconfig file at %s", m.kubeconfigLocation),
		)
	}
	return "\n" + m.list.View()
}

// NewModel initializes a new model with the provided service and list of namespaces
func NewModel(svc k8s.NameSpaceListerConfigUpdater, namespaces []string, kubeconfigFileLocation string) model {

	var items []list.Item
	for _, v := range namespaces {
		items = append(items, item(v))
	}

	l := list.New(items, itemDelegate{}, defaultWidth, listHeight)
	l.Title = "Select kubernetes namespace"
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	l.KeyMap.Quit = key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	)
	m := model{
		list:               l,
		svc:                svc,
		kubeconfigLocation: kubeconfigFileLocation,
	}

	return m
}
