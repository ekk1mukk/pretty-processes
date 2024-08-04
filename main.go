package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/v4/process"
)

// Define styling for different parts of the UI
var (
	appStyle   = lipgloss.NewStyle().Margin(1, 2)
	titleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF79C6")).Background(lipgloss.Color("#282A36")).
			Padding(1, 2).MarginBottom(1).Border(lipgloss.RoundedBorder(), true, true, false, true).Align(lipgloss.Center)
	itemTitleStyle         = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#FF79C6")).Bold(true)
	itemDescStyle          = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#8BE9FD"))
	selectedItemTitleStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("#50FA7B")).Bold(true)
	selectedItemDescStyle  = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#FFB86C")).Bold(true)
	statusBarStyle         = lipgloss.NewStyle().Foreground(lipgloss.Color("#F1FA8C")).Background(lipgloss.Color("#282A36")).Bold(true).Padding(0, 1)
)

type item struct {
	pid          int32
	name         string
	cmdline      string
	ram          float32
	ramAmount    uint64
	cpu          float64
	ppid         int32
	creationDate int64
}

// Title returns the formatted title for the list item.
func (i item) Title() string { return fmt.Sprintf("(%d) %s", i.pid, i.name) }

// Description returns a formatted description for the list item, including RAM and CPU usage.
func (i item) Description() string {
	ramUsage := fmt.Sprintf("%.2f%%", i.ram)
	ramAmount := formatBytes(i.ramAmount)
	cpuUsage := fmt.Sprintf("%.2f%%", i.cpu)
	timeObj := time.Unix(i.creationDate/1000, 0)
	return fmt.Sprintf("CMD: %s\nRAM: %s (%s) | CPU: %s | PPID: %d | Creation date : %s", i.cmdline, ramUsage, ramAmount, cpuUsage, i.ppid, timeObj.Format("2024-01-01 15:04:05"))
}

// FilterValue returns the filterable string value for the list item.
func (i item) FilterValue() string { return i.name + i.cmdline }

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
		newItems := getProcesses()
		m.list.SetItems(newItems)
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

// getTitle returns the current time formatted as a string.
func getTitle() string {
	return fmt.Sprintf("pretty-processes v0.0.4 | %s", time.Now().Format("15:04:05"))
}

// formatBytes converts bytes to a human-readable string with appropriate units.
func formatBytes(b uint64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := uint64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(b)/float64(div), "KMGTPE"[exp])
}

// getProcesses retrieves and returns a list of current processes.
func getProcesses() []list.Item {
	ctx := context.Background()
	processes, err := process.Processes()
	if err != nil {
		log.Fatalf("Error fetching processes: %s", err)
	}

	var processList []item

	for _, proc := range processes {
		name, err := proc.NameWithContext(ctx)
		if err != nil {
			log.Printf("Error fetching process name: %s", err)
			continue
		}

		cmdline, err := proc.CmdlineWithContext(ctx)
		if err != nil {
			log.Printf("Error fetching cmdline: %s", err)
			continue
		}

		ramPercentage, err := proc.MemoryPercentWithContext(ctx)
		if err != nil {
			log.Printf("Error fetching RAM usage: %s", err)
			continue
		}

		ramAmount, err := proc.MemoryInfoExWithContext(ctx)
		if err != nil {
			log.Printf("Error fetching RAM amount: %s", err)
			continue
		}

		cpuPercentage, err := proc.CPUPercentWithContext(ctx)
		if err != nil {
			log.Printf("Error fetching CPU usage: %s", err)
			continue
		}

		ppid, err := proc.PpidWithContext(ctx)
		if err != nil {
			log.Printf("Error fetching PPID: %s", err)
			continue
		}

		creationDate, err := proc.CreateTimeWithContext(ctx)
		if err != nil {
			log.Printf("Error fetching the creation date of processes: %s", err)
			continue
		}

		processList = append(processList, item{
			pid:          proc.Pid,
			name:         name,
			cmdline:      cmdline,
			ram:          ramPercentage,
			ramAmount:    ramAmount.RSS,
			cpu:          cpuPercentage,
			ppid:         ppid,
			creationDate: creationDate,
		})
	}

	sort.Slice(processList, func(i, j int) bool {
		return processList[i].pid > processList[j].pid
	})

	var processItems []list.Item

	for _, proc := range processList {
		processItems = append(processItems, proc)
	}

	return processItems
}

// refreshProcesses returns a command to refresh the process list.
func (m *model) refreshProcesses() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(3 * time.Second)
		return refreshMsg{}
	}
}

func main() {
	items := getProcesses()

	delegate := itemDelegate{}
	l := list.New(items, delegate, 0, 0)
	l.Title = getTitle()
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowTitle(true)

	m := model{list: l}

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
