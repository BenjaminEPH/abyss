package ui

import (
	"charm.land/bubbles/v2/key"
)

type keyMap struct {
	Focus  key.Binding
	Search key.Binding
	Clear  key.Binding
	Move   key.Binding
	Toggle key.Binding
	Top    key.Binding
	Help   key.Binding
	Quit   key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Focus: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "focus"),
		),
		Search: key.NewBinding(
			key.WithKeys("/", "ctrl+k"),
			key.WithHelp("/", "search"),
		),
		Clear: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "clear"),
		),
		Move: key.NewBinding(
			key.WithKeys("up", "down", "k", "j"),
			key.WithHelp("↑/↓", "move"),
		),
		Toggle: key.NewBinding(
			key.WithKeys("enter", "space"),
			key.WithHelp("⏎", "complete"),
		),
		Top: key.NewBinding(
			key.WithKeys("g", "G", "home", "end"),
			key.WithHelp("g/G", "top/bottom"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "quit"),
		),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Focus, k.Search, k.Toggle, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Focus, k.Search, k.Clear},
		{k.Move, k.Toggle, k.Top},
		{k.Help, k.Quit},
	}
}
