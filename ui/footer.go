package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func (m model) statusView() string {
	t := m.theme

	stateDot := lipgloss.NewStyle().Foreground(t.success).Render("●")
	state := "ready"
	if m.loading {
		stateDot = m.spinner.View()
		state = "filtering"
	}
	if !m.termFocused {
		stateDot = lipgloss.NewStyle().Foreground(t.warning).Render("○")
		state = "paused"
	}

	separator := lipgloss.NewStyle().Foreground(t.faint).Render("  ·  ")
	parts := []string{
		stateDot + " " + lipgloss.NewStyle().Foreground(t.text).Render(state),
	}

	if m.query != "" {
		q := m.query
		if lipgloss.Width(q) > 18 {
			q = ansi.Truncate(q, 18, "…")
		}
		parts = append(parts,
			lipgloss.NewStyle().Foreground(t.muted).Render("filter ")+
				lipgloss.NewStyle().Foreground(t.accent).Render(q))
	}

	parts = append(parts,
		lipgloss.NewStyle().Foreground(t.muted).
			Render(fmt.Sprintf("%d/%d packages", m.visibleRows(), len(m.items))),
		lipgloss.NewStyle().Foreground(t.muted).
			Render("focus: "+focusName(m.focus)),
	)

	return strings.Join(parts, separator)
}

func (m model) footerView(width int) string {
	if width <= 0 {
		return ""
	}

	showAll := m.help.ShowAll && m.height >= 20
	helpView := m.help
	helpView.ShowAll = showAll

	statusRaw := m.statusView()
	status := fitANSI(statusRaw, width-2)

	if !showAll {
		available := width - lipgloss.Width(statusRaw) - 3
		helpView.SetWidth(max(1, available))
		helpText := helpView.View(m.keys)

		gap := width - lipgloss.Width(statusRaw) - lipgloss.Width(helpText) - 2
		if gap < 1 {
			maxHelp := max(0, width-lipgloss.Width(statusRaw)-2)
			helpText = ansi.Truncate(helpText, maxHelp, "…")
			gap = max(0, width-lipgloss.Width(statusRaw)-lipgloss.Width(helpText)-2)
		}

		return fitANSI(" "+statusRaw+strings.Repeat(" ", gap)+helpText+" ", width)
	}

	helpView.SetWidth(max(1, width-2))
	helpText := helpView.View(m.keys)

	lines := []string{" " + status}
	for _, line := range strings.Split(helpText, "\n") {
		lines = append(lines, " "+line)
	}

	want := m.footerHeight()
	for len(lines) < want {
		lines = append(lines, "")
	}
	if len(lines) > want {
		lines = lines[:want]
	}
	for i := range lines {
		lines[i] = fitANSI(lines[i], width)
	}

	return strings.Join(lines, "\n")
}

func (m model) footerHeight() int {
	if m.help.ShowAll && m.height >= 20 {
		return 1 + m.fullHelpRows()
	}
	return 1
}

func (m model) fullHelpRows() int {
	rows := 0
	for _, group := range m.keys.FullHelp() {
		rows = max(rows, len(group))
	}
	return rows
}

func (m model) windowTitle() string {
	if m.query != "" {
		return fmt.Sprintf("abyss · %s (%d)", m.query, m.visibleRows())
	}
	return fmt.Sprintf("abyss · %d packages", len(m.items))
}

func (m model) programStatus() *tea.ProgramStatus {
	status := &tea.ProgramStatus{
		State:   tea.ProgramStateIdle,
		App:     "abyss",
		Title:   "abyss",
		Message: "ready",
	}
	if m.loading {
		status.State = tea.ProgramStateWorking
		status.Message = "filtering packages"
	}
	return status
}

func (m model) progressBar() *tea.ProgressBar {
	if len(m.items) == 0 {
		return nil
	}

	installed := 0
	for _, it := range m.items {
		if it.installed {
			installed++
		}
	}

	return tea.NewProgressBar(tea.ProgressBarDefault, installed*100/len(m.items))
}
