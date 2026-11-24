package ui

import (
	"fmt"
	"strings"
	"time"
)

func (m *model) applyFilter(q string) {
	q = strings.TrimSpace(q)
	if q == m.query {
		return
	}

	prev := m.query
	m.query = q
	m.list = m.list.WithFilterInputValue(q)
	m.tbl = m.tbl.WithFilterInputValue(q)

	if q == "" {
		if prev != "" {
			m.addLog(levelInfo, "filter cleared · %d packages", m.total())
		}
		return
	}

	n := m.visibleRows()
	if n == 0 {
		m.addLog(levelWarning, "no matches for %q", q)
		return
	}
	m.addLog(levelInfo, "filter %q · %d matches", q, n)
}

// toggleSelected installs/removes the highlighted package. The packages tab
// is a read-only view of what xbps reports, so only the details tab toggles.
func (m *model) toggleSelected() {
	if m.tab != tabDetails {
		return
	}

	row := m.activeTable().HighlightedRow()
	if row.Data == nil {
		return
	}

	id, ok := row.Data[columnKeyID].(int)
	if !ok || id < 0 || id >= len(m.items) {
		return
	}

	m.items[id].Installed = !m.items[id].Installed
	if m.items[id].Installed {
		m.addLog(levelSuccess, "installed: %s", m.items[id].Name)
	} else {
		m.addLog(levelInfo, "removed: %s", m.items[id].Name)
	}
	m.refreshTableRows()
}

func (m *model) addLog(level logLevel, format string, args ...any) {
	atBottom := m.logs.AtBottom()
	m.entries = append(m.entries, logEntry{
		at:    time.Now(),
		level: level,
		msg:   fmt.Sprintf(format, args...),
	})

	const maxLogEntries = 300
	if len(m.entries) > maxLogEntries {
		m.entries = append([]logEntry(nil), m.entries[len(m.entries)-maxLogEntries:]...)
	}

	m.refreshLogs()
	if atBottom {
		m.logs.GotoBottom()
	}
}
