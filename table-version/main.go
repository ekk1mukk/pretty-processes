package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/shirou/gopsutil/v4/process"
)

// ProcessItemDetails holds the details of a process item
type ProcessItemDetails struct {
	PID          string
	Name         string
	CPU          string
	RAM          string
	RAMAmount    string
	PPID         string
	CreationDate string
}

// processItem represents a single process entry.
type processItem struct {
	pid          int32
	name         string
	cmdline      string
	ram          float32
	ramAmount    uint64
	cpu          float64
	ppid         int32
	creationDate int64
}

var (
	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

	headerStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#282A36")).
			Foreground(lipgloss.Color("#FF79C6")).
			Bold(true).
			Padding(0, 1)

	cellStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#44475a")).
			Foreground(lipgloss.Color("#f8f8f2")).
			Padding(0, 1)

	selectedStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#6272A4")).
			Foreground(lipgloss.Color("#50FA7B")).
			Bold(true).
			Padding(0, 1)

	footerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#f8f8f2")).
			Bold(true).
			Margin(0, 0, 0, 3).
			Padding(0, 0, 0, 0)

	tableData       []ProcessItemDetails
	refreshInterval = 3 * time.Second
)

type model struct {
	table    table.Model
	quitting bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Adjust the table height to match the terminal height
		m.table.SetHeight(msg.Height - 6) // Leave some room for borders and footer

	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "r":
			m.table.SetRows(createTableRows(getProcesses()))
		case "pgup":
			m.table.MoveUp(m.table.Height()) // Move to the top
		case "pgdn":
			m.table.MoveDown(m.table.Height()) // Move to the bottom
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}
	return baseStyle.Render(m.table.View()) + "\n" + footerStyle.Render("Press 'q' to quit | 'r' to refresh | 'PgUp' to top | 'PgDn' to bottom")
}

func main() {
	columns := []table.Column{
		{Title: "PID", Width: 7},
		{Title: "Name", Width: 40},
		{Title: "CPU (%)", Width: 8},
		{Title: "RAM (%)", Width: 8},
		{Title: "RAM (Used)", Width: 10},
		{Title: "PPID", Width: 7},
		{Title: "Creation Date", Width: 20},
	}

	processes := getProcesses()
	rows := createTableRows(processes)

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)

	// Customizing the table styles
	s := table.DefaultStyles()

	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("190")).
		Background(lipgloss.Color("57")).
		Bold(true)
	t.SetStyles(s)

	m := model{t, false}

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

// getProcesses retrieves and returns a list of current processes.
func getProcesses() []processItem {
	ctx := context.Background()
	processes, err := process.Processes()
	if err != nil {
		log.Fatalf("Error fetching processes: %s", err)
	}

	var processList []processItem

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

		processList = append(processList, processItem{
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

	// Sort by PID
	sort.Slice(processList, func(i, j int) bool {
		return processList[i].pid > processList[j].pid
	})

	return processList
}

// createTableRows converts process data to table rows.
func createTableRows(processes []processItem) []table.Row {
	var rows []table.Row

	for _, proc := range processes {
		rows = append(rows, table.Row{
			strconv.Itoa(int(proc.pid)),
			proc.name,
			fmt.Sprintf("%.2f", proc.cpu),
			fmt.Sprintf("%.2f", proc.ram),
			formatBytes(proc.ramAmount),
			strconv.Itoa(int(proc.ppid)),
			time.Unix(proc.creationDate/1000, 0).Format("2006-01-02 15:04:05"),
		})
	}

	return rows
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
