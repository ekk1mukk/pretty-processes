// main.go
package main

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	items := getProcesses()

	delegate := itemDelegate{}
	l := list.New(items, delegate, 0, 0)
	l.SetShowTitle(true)
	l.Title = getTitle()
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true) // Disable the default help view
	l.AdditionalFullHelpKeys = customKeys.GetCustomFullHelpCommands

	m := model{
		title: getTitle(),
		list:  l,
		help:  help.New(),
	}

	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}

// getTitle returns the current time formatted as a string.
func getTitle() string {
	return fmt.Sprintf("pretty-processes v0.0.7 | %s", time.Now().Format("15:04:05"))
}
