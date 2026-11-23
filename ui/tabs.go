package ui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

const tabBarIndent = 1

func tabLabel(i int) string {
	return fmt.Sprintf(" %d %s ", i+1, tabNames[i])
}

// tabBar renders the single-line tab strip shown under the search bar.
func (m model) tabBar(width int) string {
	t := m.theme

	active := lipgloss.NewStyle().
		Background(t.selectedBg).
		Foreground(t.selectedFg).
		Bold(true)
	inactive := lipgloss.NewStyle().Foreground(t.muted)

	parts := make([]string, 0, tabCount)
	for i := range tabCount {
		style := inactive
		if i == m.tab {
			style = active
		}
		parts = append(parts, style.Render(tabLabel(i)))
	}

	line := strings.Repeat(" ", tabBarIndent) + strings.Join(parts, " ")
	return fitANSI(line, width)
}

// tabAt returns the tab index under the given column, or -1.
func (m model) tabAt(x int) int {
	start := tabBarIndent
	for i := range tabCount {
		w := lipgloss.Width(tabLabel(i))
		if x >= start && x < start+w {
			return i
		}
		start += w + 1
	}
	return -1
}
