package ui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
)

func (m *model) refreshLogs() {
	if m.logs.Width() <= 0 {
		return
	}

	atBottom := m.logs.AtBottom()
	lines := make([]string, 0, len(m.entries))
	for _, entry := range m.entries {
		lines = append(lines, m.renderLogEntry(entry))
	}
	m.logs.SetContentLines(lines)

	if atBottom {
		m.logs.GotoBottom()
	}
}

func (m model) renderLogEntry(entry logEntry) string {
	t := m.theme

	label, levelColor := "INFO", t.info
	switch entry.level {
	case levelSuccess:
		label, levelColor = "OK", t.success
	case levelWarning:
		label, levelColor = "WARN", t.warning
	case levelError:
		label, levelColor = "ERR", t.danger
	}

	timestamp := lipgloss.NewStyle().Foreground(t.faint).Render(entry.at.Format("15:04:05"))
	badge := lipgloss.NewStyle().
		Foreground(levelColor).
		Bold(true).
		Render(fmt.Sprintf("%-4s", label))
	message := lipgloss.NewStyle().
		Foreground(t.text).
		Render(strings.ReplaceAll(entry.msg, "\n", " "))

	return timestamp + "  " + badge + "  " + message
}

func (m model) logGutter() viewport.GutterFunc {
	t := m.theme

	return func(ctx viewport.GutterContext) string {
		if ctx.Soft {
			return lipgloss.NewStyle().Foreground(t.faint).Render("    │ ")
		}
		if ctx.Index >= ctx.TotalLines {
			return "      "
		}
		return lipgloss.NewStyle().
			Foreground(t.faint).
			Render(fmt.Sprintf("%3d │ ", ctx.Index+1))
	}
}

func logViewportKeys() viewport.KeyMap {
	keys := viewport.DefaultKeyMap()
	keys.PageDown = key.NewBinding(key.WithKeys("pgdown", "f"))
	keys.PageUp = key.NewBinding(key.WithKeys("pgup", "b"))
	keys.HalfPageDown = key.NewBinding(key.WithKeys("d", "ctrl+d"))
	keys.HalfPageUp = key.NewBinding(key.WithKeys("u", "ctrl+u"))
	return keys
}
