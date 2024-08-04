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

var appStyle = lipgloss.NewStyle().Margin(1, 2)
var itemTitleStyle = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#DDDDDD")).Bold(true)
var itemDescStyle = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#3F88F5"))
var selectedItemTitleStyle = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#BAFFDF")).Bold(true)
var selectedItemDescStyle = lipgloss.NewStyle().PaddingLeft(4).Foreground(lipgloss.Color("#42BFDD")).Bold(true)

type item struct {
	pid       int32
	name      string
	cmdline   string
	ram       float32
	ramAmount uint64
	cpu       float64
	ppid      int32
}

func (i item) Title() string { return fmt.Sprintf("(%d) %s", i.pid, i.name) }

func (i item) Description() string {
	return fmt.Sprintf("CMD: %s\nRAM: (%.2f%%) %d CPU: %.2f%% PPID: %d", i.cmdline, i.ram, i.ramAmount, i.cpu, i.ppid)
}

func (i item) FilterValue() string { return i.name + i.cmdline }

type itemDelegate struct{}

func (d itemDelegate) Height() int { return 3 }

func (d itemDelegate) Spacing() int { return 1 }

func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

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

func (m model) Init() tea.Cmd {
	return tea.Batch(m.refreshProcesses())
}

func refreshCmd() tea.Cmd {
	return func() tea.Msg {
		time.Sleep(3 * time.Second)
		return refreshMsg{}
	}
}

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
		m.list.Title = "pretty-processes v0.0.3 | " + getCurrentTime()
		return m, refreshCmd()
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return "Bye!\n"
	}

	return appStyle.Render(m.list.View())
}

func getCurrentTime() string {
	return time.Now().Format("15:04:05")
}

func getProcesses() []list.Item {
	ctx := context.Background()
	processes, err := process.Processes()
	if err != nil {
		log.Fatalf("Error fetching processes: %s", err)
	}

	var processList []item

	for _, proc := range processes {
		name, err := proc.Name()
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
			log.Printf("Error fetching CPU usage: %s", err)
			continue
		}

		cpuPercentage, err := proc.CPUPercentWithContext(ctx)
		if err != nil {
			log.Printf("Error fetching CPU usage: %s", err)
			continue
		}

		ppid, err := proc.Ppid()
		if err != nil {
			log.Printf("Error fetching PPID: %s", err)
			continue
		}

		processList = append(processList, item{
			pid:       proc.Pid,
			name:      name,
			cmdline:   cmdline,
			ram:       ramPercentage,
			ramAmount: ramAmount.RSS,
			cpu:       cpuPercentage,
			ppid:      ppid,
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

func (m *model) refreshProcesses() tea.Cmd {
	return func() tea.Msg {
		return refreshMsg{}
	}
}

func main() {
	items := getProcesses()

	delegate := itemDelegate{}
	l := list.New(items, delegate, 0, 0)
	l.Title = "pretty-processes v0.0.3 | " + getCurrentTime()
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
