package ui

import (
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
)

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if m.loading {
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m = m.resize()
		return m, tea.Batch(cmds...)

	case tea.BackgroundColorMsg:
		m.dark = msg.IsDark()
		m.theme = newTheme(m.dark)
		m.applyTheme()
		m.rebuildTable()
		m.refreshLogs()
		m = m.resize()
		return m, tea.Batch(cmds...)

	case tea.FocusMsg:
		m.termFocused = true
		return m, tea.Batch(cmds...)

	case tea.BlurMsg:
		m.termFocused = false
		return m, tea.Batch(cmds...)

	case filterMsg:
		if msg.seq != m.searchSeq {
			return m, tea.Batch(cmds...)
		}
		m.loading = false
		m.applyFilter(msg.query)
		return m, tea.Batch(cmds...)

	case tea.KeyPressMsg:
		cmds = append(cmds, m.handleKey(msg))
		return m, tea.Batch(cmds...)

	case tea.MouseClickMsg:
		cmds = append(cmds, m.handleClick(msg))
		return m, tea.Batch(cmds...)

	case tea.MouseWheelMsg:
		cmds = append(cmds, m.handleWheel(msg))
		return m, tea.Batch(cmds...)
	}

	// Forward non-key messages (paste, cursor blink, ...) to the focused widget.
	switch m.focus {
	case focusSearch:
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		cmds = append(cmds, cmd)
	case focusLog:
		var cmd tea.Cmd
		m.logs, cmd = m.logs.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) handleKey(msg tea.KeyPressMsg) tea.Cmd {
	s := msg.String()

	if s == "ctrl+c" {
		return tea.Quit
	}
	if s == "tab" {
		return m.cycleFocus(1)
	}
	if s == "shift+tab" || s == "backtab" {
		return m.cycleFocus(-1)
	}

	if m.focus == focusSearch {
		switch s {
		case "alt+left":
			return m.setTab(m.tab - 1)
		case "alt+right":
			return m.setTab(m.tab + 1)
		case "esc":
			if m.search.Value() != "" {
				m.search.SetValue("")
				m.searchSeq++
				m.loading = false
				m.applyFilter("")
				return nil
			}
			return m.setFocus(focusMain)
		case "enter":
			q := m.search.Value()
			m.searchSeq++
			m.loading = false
			m.applyFilter(q)
			return m.setFocus(focusMain)
		}

		prev := m.search.Value()
		var cmd tea.Cmd
		m.search, cmd = m.search.Update(msg)
		if m.search.Value() != prev {
			m.searchSeq++
			m.loading = true
			seq := m.searchSeq
			q := m.search.Value()
			return tea.Batch(
				cmd,
				m.spinner.Tick,
				tea.Tick(130*time.Millisecond, func(time.Time) tea.Msg {
					return filterMsg{seq: seq, query: q}
				}),
			)
		}
		return cmd
	}

	switch {
	case key.Matches(msg, m.keys.Tabs):
		switch s {
		case "[", "alt+left":
			return m.setTab(m.tab - 1)
		case "]", "alt+right":
			return m.setTab(m.tab + 1)
		default:
			return m.setTab(int(s[0] - '1'))
		}
	case key.Matches(msg, m.keys.Help):
		if m.height < 20 {
			m.addLog(levelWarning, "resize the window to see full help")
			return nil
		}
		m.help.ShowAll = !m.help.ShowAll
		*m = m.resize()
		return nil
	case key.Matches(msg, m.keys.Quit):
		return tea.Quit
	case key.Matches(msg, m.keys.Search):
		return m.setFocus(focusSearch)
	case key.Matches(msg, m.keys.Clear):
		if m.query != "" {
			m.search.SetValue("")
			m.applyFilter("")
		} else {
			return m.setFocus(focusSearch)
		}
		return nil
	}

	switch m.focus {
	case focusMain:
		if s == "enter" || s == "space" {
			m.toggleSelected()
			return nil
		}
		t := m.activeTable()
		var cmd tea.Cmd
		t, cmd = t.Update(msg)
		m.setActiveTable(t)
		return cmd

	case focusLog:
		switch s {
		case "home", "g":
			m.logs.GotoTop()
			return nil
		case "end", "G":
			m.logs.GotoBottom()
			return nil
		}
		var cmd tea.Cmd
		m.logs, cmd = m.logs.Update(msg)
		return cmd
	}

	return nil
}

func (m *model) handleClick(msg tea.MouseClickMsg) tea.Cmd {
	if msg.Button != tea.MouseLeft {
		return nil
	}

	y := msg.Y
	switch {
	case y >= m.layout.searchY && y < m.layout.searchY+m.layout.searchH:
		return m.setFocus(focusSearch)

	case y >= m.layout.tabsY && y < m.layout.tabsY+m.layout.tabsH:
		if i := m.tabAt(msg.X); i >= 0 {
			return m.setTab(i)
		}
		return nil

	case y >= m.layout.mainY && y < m.layout.mainY+m.layout.mainH:
		cmd := m.setFocus(focusMain)
		// Table body starts after the panel border and the table header.
		row := y - (m.layout.mainY + 1) - 2
		if row >= 0 {
			t := m.activeTable()
			start, end := t.VisibleIndices()
			idx := start + row
			if idx >= start && idx <= end && idx < t.TotalRows() {
				m.setActiveTable(t.WithHighlightedRow(idx))
			}
		}
		return cmd

	case y >= m.layout.logY && y < m.layout.logY+m.layout.logH:
		return m.setFocus(focusLog)
	}

	return nil
}

func (m *model) handleWheel(msg tea.MouseWheelMsg) tea.Cmd {
	y := msg.Y
	switch {
	case y >= m.layout.mainY && y < m.layout.mainY+m.layout.mainH:
		m.setFocus(focusMain)
		t := m.activeTable()
		switch msg.Button {
		case tea.MouseWheelDown:
			t = t.PageDown()
		case tea.MouseWheelUp:
			t = t.PageUp()
		}
		m.setActiveTable(t)
		return nil

	case y >= m.layout.logY && y < m.layout.logY+m.layout.logH:
		cmd := m.setFocus(focusLog)
		var viewportCmd tea.Cmd
		m.logs, viewportCmd = m.logs.Update(msg)
		return tea.Batch(cmd, viewportCmd)
	}

	return nil
}

func (m *model) cycleFocus(delta int) tea.Cmd {
	f := m.focus
	for range focusCount {
		f = (f + delta + focusCount) % focusCount
		if f != focusLog || m.tab == tabDetails {
			break
		}
	}
	return m.setFocus(f)
}

// setTab switches the content shown below the search bar.
func (m *model) setTab(i int) tea.Cmd {
	if i < 0 || i >= tabCount || i == m.tab {
		return nil
	}

	m.tab = i
	*m = m.resize()
	m.addLog(levelInfo, "tab: %s", tabNames[i])

	if m.focus == focusLog && m.tab == tabPackages {
		return m.setFocus(focusMain)
	}
	return m.setFocus(m.focus)
}

func (m *model) setFocus(f int) tea.Cmd {
	if f < 0 || f >= focusCount {
		return nil
	}

	if f == focusLog && m.tab == tabPackages {
		return nil
	}

	m.focus = f
	m.list = m.list.Focused(f == focusMain && m.tab == tabPackages)
	m.tbl = m.tbl.Focused(f == focusMain && m.tab == tabDetails)

	if f == focusSearch {
		m.search.CursorEnd()
		return m.search.Focus()
	}

	m.search.Blur()
	return nil
}
