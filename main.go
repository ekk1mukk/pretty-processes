package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

// main function to run the Bubble Tea program
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

// getTitle returns the current time formatted as a string.
func getTitle() string {
	return fmt.Sprintf("pretty-processes v0.0.4 | %s", time.Now().Format("15:04:05"))
}
