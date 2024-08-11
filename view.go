package main

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	appStyle   = lipgloss.NewStyle().PaddingLeft(2)
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).MarginTop(1).Underline(true).Blink(true)

	selectedItemTitleStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#D81159")).Bold(true)
	selectedItemDescStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("#D81159")).Bold(true)

	pidStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555"))
	nameStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD"))
	ramStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B"))
	cpuStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C"))
	creationDateStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#01BAEF"))
	cmdlineStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9"))

	arrowsStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Bold(true)

	pidStyleSelected          = lipgloss.NewStyle().Foreground(lipgloss.Color("#FF5555")).Bold(true).Underline(true)
	nameStyleSelected         = lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD")).Bold(true).Underline(true)
	ramStyleSelected          = lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Bold(true).Underline(true)
	cpuStyleSelected          = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFB86C")).Bold(true).Underline(true)
	creationDateStyleSelected = lipgloss.NewStyle().Foreground(lipgloss.Color("#01BAEF")).Bold(true).Underline(true)
	cmdlineStyleSelected      = lipgloss.NewStyle().Foreground(lipgloss.Color("#BD93F9")).Bold(true).Underline(true)

	selectedBoldStyle = lipgloss.NewStyle().Bold(true)
)

type itemDelegate struct{}

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, i list.Item) {
	it, ok := i.(processItem)
	if !ok {
		return
	}

	// Apply styles to different data types
	pid := fmt.Sprintf("%d", it.pid)
	name := it.name
	ram := fmt.Sprintf("RAM: %.2f%% (%s)", it.ram, formatBytes(it.ramAmount))
	cpu := fmt.Sprintf("CPU: %.2f%%", it.cpu)
	creationDate := it.creationDate
	cmdline := it.cmdline

	// Apply bold style to the entire line if the item is selected
	var styledLine string
	if index == m.Index() {
		styledLine = fmt.Sprintf("%s %s %s %s %s %s %s",
			arrowsStyle.Render(">"),
			pidStyleSelected.Render(pid),
			nameStyleSelected.Render(name),
			ramStyleSelected.Render(ram),
			cpuStyleSelected.Render(cpu),
			creationDateStyleSelected.Render(time.Unix(creationDate/1000, 0).Format("2006-01-02 15:04:05")),
			cmdlineStyleSelected.Render(cmdline))
	} else {
		styledLine = fmt.Sprintf("%s %s %s %s %s %s",
			pidStyle.Render(pid),
			nameStyle.Render(name),
			ramStyle.Render(ram),
			cpuStyle.Render(cpu),
			creationDateStyle.Render(time.Unix(creationDate/1000, 0).Format("2006-01-02 15:04:05")),
			cmdlineStyle.Render(cmdline))
	}

	fmt.Fprint(w, styledLine) // Use Fprint instead of Fprintln to avoid adding newlines
}

func getTitleHeight() int {
	return lipgloss.Height(titleStyle.Render(getTitle()))
}

type refreshMsg struct{}

type model struct {
	title    string
	list     list.Model
	help     help.Model
	quitting bool
}

func (m model) Init() tea.Cmd {
	return m.refreshProcesses()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Calculate available height for the list
		h, v := appStyle.GetFrameSize()
		availableHeight := msg.Height - v - getTitleHeight() - lipgloss.Height(m.help.View(customKeys))
		m.list.SetSize(msg.Width-h, availableHeight)
		return m, nil

	case tea.KeyMsg:
		switch {
		case msg.String() == "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		}

	case refreshMsg:
		filter := m.list.FilterInput.Value()
		newItems := getProcesses()

		// Preserve filtering when refreshing items
		if filter == "" {
			m.list.SetItems(newItems)
		}

		m.list.Title = getTitle()

		return m, m.refreshProcesses()
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return "Bye!\n"
	}

	// Render title at the top
	//titleView := titleStyle.Render(m.list.Title)
	m.list.Styles.Title = titleStyle
	listView := appStyle.Render(m.list.View())

	return fmt.Sprintf("%s\n", listView)
}

func (m *model) refreshProcesses() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(1 * time.Second)
		return refreshMsg{}
	}
}
