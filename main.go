package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/v4/process"
)

// Define styling for different parts of the UI
var docStyle = lipgloss.NewStyle().Margin(1, 2) // Add margins to avoid clipping
var itemTitleStyle = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#DDDDDD")).Bold(true)
var itemDescStyle = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#3F88F5"))
var selectedItemTitleStyle = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#BAFFDF")).Bold(true)
var selectedItemDescStyle = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#42BFDD")).Bold(true)

// Define an item struct that implements the list.Item interface
type item struct {
	pid     int32
	name    string
	cmdline string
	ram     float32
	cpu     float64
}

// Implement the list.Item interface
func (i item) Title() string { return fmt.Sprintf("(%d) %s", i.pid, i.name) }

func (i item) Description() string {
	// Multi-line description with RAM and CPU usage
	return fmt.Sprintf("CMD: %s\nRAM: %.2f%% CPU: %.2f%%", i.cmdline, i.ram, i.cpu)
}

func (i item) FilterValue() string { return i.name + i.cmdline }

// Custom delegate to handle multi-line items
type itemDelegate struct{}

// Set the height to match the number of lines in the item's description
func (d itemDelegate) Height() int { return 3 } // Adjusted to match the number of lines in the item

func (d itemDelegate) Spacing() int { return 1 }

func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, i list.Item) {
	it, ok := i.(item)
	if !ok {
		return
	}

	title := itemTitleStyle.Render(it.Title())
	desc := itemDescStyle.Render(it.Description())

	// Apply selected item styling if the current item is selected
	if index == m.Index() {
		// Use the selectedItemStyle to render the selected item in bold
		title = selectedItemTitleStyle.Render("> " + it.Title())
		desc = selectedItemDescStyle.Render(it.Description())
	}

	// Write the styled output to the io.Writer
	fmt.Fprintf(w, "%s\n%s", title, desc) // Removed extra newline for concise display
}

// Define the model struct for Bubble Tea
type model struct {
	list     list.Model
	quitting bool
}

// Init initializes the program
func (m model) Init() tea.Cmd {
	return nil
}

// Update handles incoming messages and updates the model state
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Adjust the list dimensions based on window size
		m.list.SetSize(msg.Width, msg.Height-docStyle.GetVerticalFrameSize())
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the current view
func (m model) View() string {
	if m.quitting {
		return "Bye!\n"
	}

	return docStyle.Render(m.list.View())
}

// getProcesses retrieves and sorts the processes by PID in descending order
func getProcesses() []list.Item {
	processes, err := process.Processes()
	if err != nil {
		log.Fatalf("Error fetching processes: %s", err)
	}

	var processList []item

	for _, proc := range processes {
		name, err := proc.Name()
		cmdline, err := proc.Cmdline()
		ramPercentage, err := proc.MemoryPercent()
		cpuPercentage, err := proc.CPUPercent()

		if err != nil {
			log.Printf("Error fetching process info: %s", err)
			continue
		}
		processList = append(processList, item{pid: proc.Pid, name: name, cmdline: cmdline, ram: ramPercentage, cpu: cpuPercentage})
	}

	// Sort the processList by PID in descending order
	sort.Slice(processList, func(i, j int) bool {
		return processList[i].pid > processList[j].pid
	})

	var processItems []list.Item

	// Convert sorted processList to list.Items
	for _, proc := range processList {
		processItems = append(processItems, proc)
	}

	return processItems
}

func main() {
	items := getProcesses()

	// Initialize the Bubble Tea list with a delegate
	delegate := itemDelegate{}
	l := list.New(items, delegate, 0, 0) // Initialize with zero to auto-resize later
	l.Title = "pretty-processes v0.0.2"
	l.SetShowStatusBar(true)    // Enable status bar for better UX
	l.SetFilteringEnabled(true) // Enable filtering
	l.SetShowTitle(true)

	m := model{list: l}

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
