package ui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m model) View() tea.View {
	if m.width == 0 {
		v := tea.NewView("loading…")
		v.AltScreen = true
		return v
	}

	if m.width < 40 || m.height < 12 {
		msg := lipgloss.NewStyle().Foreground(m.theme.muted).
			Render("abyss · resize the window to see the interface")
		msg = fitANSI(msg, max(1, m.width))
		v := tea.NewView(lipgloss.Place(
			max(1, m.width),
			max(1, m.height),
			lipgloss.Center,
			lipgloss.Center,
			msg,
		))
		v.AltScreen = true
		v.WindowTitle = "abyss"
		return v
	}

	contentW := m.width - 4
	m.logs.LeftGutterFunc = m.logGutter()

	searchBody := m.searchView(contentW)

	var mainBody string
	if m.visibleRows() == 0 {
		mainBody = strings.Join(m.emptyState(contentW, max(1, m.layout.mainH-2)), "\n")
	} else {
		mainBody = m.activeTable().View()
	}

	searchPanel := renderPanel(m.theme, m.width, m.layout.searchH,
		"⌕ SEARCH", m.searchBadge(), m.focus == focusSearch, searchBody)
	mainPanel := renderPanel(m.theme, m.width, m.layout.mainH,
		"▤ "+strings.ToUpper(tabNames[m.tab]), m.mainBadge(), m.focus == focusMain, mainBody)
	footer := m.footerView(m.width)

	sections := []string{searchPanel, m.tabBar(m.width), mainPanel}
	if m.tab == tabDetails {
		sections = append(sections, renderPanel(m.theme, m.width, m.layout.logH,
			"≡ LOG", m.logBadge(), m.focus == focusLog, m.logs.View()))
	}
	sections = append(sections, footer)

	content := lipgloss.JoinVertical(lipgloss.Left, sections...)

	v := tea.NewView(content)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	v.ReportFocus = true
	v.WindowTitle = m.windowTitle()
	v.ProgramStatus = m.programStatus()
	v.ProgressBar = m.progressBar()
	return v
}

func (m model) searchView(contentW int) string {
	t := m.theme

	hintStyle := lipgloss.NewStyle().Foreground(t.faint)
	hint := ""
	if m.focus == focusSearch && m.search.Value() != "" {
		hint = hintStyle.Render("esc clear")
	} else if m.focus != focusSearch && m.search.Value() != "" {
		hint = hintStyle.Render("tab edit")
	}

	available := contentW
	if hint != "" {
		available -= lipgloss.Width(hint) + 2
	}

	input := m.search
	input.SetWidth(max(8, available-lipgloss.Width(input.Prompt)-1))

	line := input.View()
	if hint != "" {
		gap := contentW - lipgloss.Width(line) - lipgloss.Width(hint)
		if gap < 1 {
			return fitANSI(line, contentW)
		}
		line += strings.Repeat(" ", gap) + hint
	}

	return fitANSI(line, contentW)
}

func (m model) emptyState(width, height int) []string {
	if height < 1 {
		return nil
	}

	blank := strings.Repeat(" ", width)
	lines := make([]string, height)
	for i := range lines {
		lines[i] = blank
	}

	mid := height / 2
	if height > 2 {
		mid = height/2 - 1
	}

	faint := lipgloss.NewStyle().Foreground(m.theme.faint)
	muted := lipgloss.NewStyle().Foreground(m.theme.muted)
	title := faint.Render("∅") + "  " + muted.Render("no results")
	hint := faint.Render("try another search · esc to clear")
	lines[mid] = lipgloss.PlaceHorizontal(width, lipgloss.Center, title)
	if mid+1 < height {
		lines[mid+1] = lipgloss.PlaceHorizontal(width, lipgloss.Center, hint)
	}

	return lines
}

func (m model) searchBadge() string {
	if m.loading {
		return "filtering…"
	}
	if m.query != "" {
		return fmt.Sprintf("%d results", m.visibleRows())
	}
	return fmt.Sprintf("%d packages", m.total())
}

func (m model) mainBadge() string {
	installed := 0
	for _, it := range m.items {
		if it.Installed {
			installed++
		}
	}

	badge := fmt.Sprintf("%d/%d", m.visibleRows(), m.total())
	if m.tab == tabDetails && len(m.items) > 0 {
		badge += fmt.Sprintf(" · %d%% installed", installed*100/len(m.items))
	}
	if pages := m.maxPages(); pages > 1 {
		badge += fmt.Sprintf(" · page %d/%d", m.currentPage(), pages)
	}
	return badge
}

func (m model) logBadge() string {
	return fmt.Sprintf("%d lines", len(m.entries))
}
