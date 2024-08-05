// view.go

package main

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Define styling for different parts of the UI
var (
	appStyle               = lipgloss.NewStyle().Margin(0, 2)
	titleStyle             = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF79C6")).Background(lipgloss.Color("#282A36")).Padding(1, 2).Align(lipgloss.Center).MarginBottom(1)
	itemTitleStyle         = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#FF79C6")).Bold(true)
	itemDescStyle          = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#8BE9FD"))
	selectedItemTitleStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#50FA7B")).Bold(true)
	selectedItemDescStyle  = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#FFB86C")).Bold(true)
	helpStyle              = lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")).Padding(1).Margin(0, 3)
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
	it, ok := i.(processItem)
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
	keys     keyMap
	help     help.Model
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
		// Calculate available height for the list
		availableHeight := msg.Height - appStyle.GetVerticalFrameSize() - titleHeight() - helpHeight(m)
		m.list.SetSize(msg.Width, availableHeight)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
			return m, nil
		case key.Matches(msg, m.keys.Quit):
			m.quitting = true
			return m, tea.Quit
		}

	case refreshMsg:
		filter := m.list.FilterInput.Value()
		newItems := getProcesses()

		// Preserve filtering when refreshing items
		if filter != "" {
			m.list.SetItems(filterItems(newItems, filter))
		} else {
			m.list.SetItems(newItems)
		}

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

	// Render title at the top
	title := titleStyle.Render(m.list.Title)

	// Render list and help below title
	listView := appStyle.Render(m.list.View())
	helpView := helpStyle.Render(m.help.View(m.keys))

	return fmt.Sprintf("%s\n%s\n%s", title, listView, helpView)
}

// titleHeight returns the height of the title.
func titleHeight() int {
	return 3 // Adjust this based on the actual height of your title styling
}

// helpHeight calculates the height of the help view for layout purposes.
func helpHeight(m model) int {
	if m.help.ShowAll {
		return len(m.keys.FullHelp()[0]) // Return the number of key bindings in the help view
	}
	return 1
}

// refreshProcesses returns a command to refresh the process list.
func (m *model) refreshProcesses() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(3 * time.Second)
		return refreshMsg{}
	}
}

// filterItems filters items based on the provided filter value.
func filterItems(items []list.Item, filter string) []list.Item {
	var filteredItems []list.Item
	for _, item := range items {
		if it, ok := item.(processItem); ok {
			if it.FilterValue() == filter {
				filteredItems = append(filteredItems, it)
			}
		}
	}
	return filteredItems
}
