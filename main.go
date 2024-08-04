package main

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/v4/process"
)

var docStyle = lipgloss.NewStyle().Margin(1, 2)

// Define an item struct that implements the list.Item interface
type item struct {
	pid  int32
	name string
}

// Implement the list.Item interface
func (i item) Title() string       { return fmt.Sprintf("(%d) %s", i.pid, i.name) }
func (i item) Description() string { return i.name }
func (i item) FilterValue() string { return i.name }

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
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
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
	var processItems []list.Item

	processes, err := process.Processes()
	if err != nil {
		log.Fatalf("Error fetching processes: %s", err)
	}

	// Create a slice of processInfo to store PID and Name
	type processInfo struct {
		PID  int32
		Name string
	}

	var processList []processInfo

	for _, proc := range processes {
		name, err := proc.Name()
		if err != nil {
			log.Printf("Error fetching process name: %s", err)
			continue
		}
		processList = append(processList, processInfo{PID: proc.Pid, Name: name})
	}

	// Sort the processList by PID in descending order
	sort.Slice(processList, func(i, j int) bool {
		return processList[i].PID > processList[j].PID
	})

	// Convert sorted processList to list.Items
	for _, proc := range processList {
		processItems = append(processItems, item{pid: proc.PID, name: proc.Name})
	}

	return processItems
}

func main() {
	items := getProcesses()

	// Initialize the Bubble Tea list with a delegate
	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, 0, 0) // Set width and height to 0, allowing dynamic resizing
	l.Title = "pretty-processes v0.0.1"
	l.SetShowStatusBar(true)    // Enable status bar for better UX
	l.SetFilteringEnabled(true) // Enable filtering

	m := model{list: l}

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
