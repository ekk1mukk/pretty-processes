package main

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Define styling for different parts of the UI
var (
	appStyle               = lipgloss.NewStyle().Margin(1, 2)
	itemTitleStyle         = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#FF79C6")).Bold(true)
	itemDescStyle          = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#8BE9FD"))
	selectedItemTitleStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#50FA7B")).Bold(true)
	selectedItemDescStyle  = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#FFB86C")).Bold(true)
)

// itemDelegate defines how to render each list item.
type itemDelegate struct{}

// Height returns the height of the list item.
func (d itemDelegate) Height() int { return 3 }

// Spacing returns the spacing between list items.
func (d itemDelegate) Spacing() int { return 1 }

// Update handles updates for the list item (unused here).
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

// Render renders the list item.
func (d itemDelegate) Render(w io.Writer, m list.Model, index int, i list.Item) {
	it, ok := i.(item)
	if !ok {
		return
	}

	title := itemTitleStyle.Render(it.Title())
	desc := itemDescStyle.Render(it.Description())

	if index == m.Index() {
		title = selectedItemTitleStyle.Render("> " + it.Title())
		desc = selectedItemDescStyle.Render(it.Description())
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

type refreshMsg struct{}

// model represents the Bubble Tea model for the application.
type model struct {
	list     list.Model
	quitting bool
}

// Init initializes the model and starts the refresh process.
func (m model) Init() tea.Cmd {
	return tea.Batch(m.refreshProcesses())
}

// Update updates the model based on incoming messages.
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetSize(msg.Width, msg.Height-appStyle.GetVerticalFrameSize())
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}

	case refreshMsg:
		filter := m.list.FilterInput.Value()

		newItems := getProcesses()
		m.list.SetItems(newItems)

		m.list.FilterInput.SetValue(filter)

		m.list.Title = getTitle()

		return m, m.refreshProcesses()
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the current view of the model.
func (m model) View() string {
	if m.quitting {
		return "Bye!\n"
	}

	return appStyle.Render(m.list.View())
}

// refreshProcesses returns a command to refresh the process list.
func (m *model) refreshProcesses() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(3 * time.Second)
		return refreshMsg{}
	}
}
