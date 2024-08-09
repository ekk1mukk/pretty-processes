package main

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Kill key.Binding
	Sort key.Binding
}

var customKeys = keyMap{
	Kill: key.NewBinding(
		key.WithKeys("k", "kill process"),
		key.WithHelp("k", "kill process"),
	),
	Sort: key.NewBinding(
		key.WithKeys("s", "Sort by"),
		key.WithHelp("s", "Sort by"),
	),
}

// ShortHelp returns keybindings for the minimal help view.
func (k keyMap) GetCustomShortHelpCommands() []key.Binding {
	return []key.Binding{
		k.Kill, k.Sort,
	}
}

func (k keyMap) GetCustomFullHelpCommands() []key.Binding {
	return []key.Binding{
		k.Kill, k.Sort,
	}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return nil
}

func (k keyMap) ShortHelp() []key.Binding {
	return nil
}
