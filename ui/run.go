package ui

import tea "charm.land/bubbletea/v2"

// Run starts the TUI and blocks until it exits.
func Run() error {
	_, err := tea.NewProgram(initialModel()).Run()
	return err
}
