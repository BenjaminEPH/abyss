package ui

import (
	"strconv"

	"charm.land/lipgloss/v2"
	"github.com/evertras/bubble-table/table"
)

// activeTable returns the table shown in the current tab.
func (m model) activeTable() table.Model {
	if m.tab == tabDetails {
		return m.tbl
	}
	return m.list
}

func (m *model) setActiveTable(t table.Model) {
	if m.tab == tabDetails {
		m.tbl = t
		return
	}
	m.list = t
}

func (m *model) styledTable(cols []table.Column, rows []table.Row, focused bool) table.Model {
	t := m.theme

	return table.New(cols).
		WithRows(rows).
		WithBaseStyle(lipgloss.NewStyle().Foreground(t.text).Align(lipgloss.Left)).
		WithRowStyleFunc(func(input table.RowStyleFuncInput) lipgloss.Style {
			installed, _ := input.Row.Data[columnKeyInstalled].(bool)

			if input.IsHighlighted {
				return lipgloss.NewStyle().
					Background(t.selectedBg).
					Foreground(t.selectedFg).
					Bold(true)
			}
			if installed {
				return lipgloss.NewStyle().Foreground(t.success)
			}
			return lipgloss.NewStyle()
		}).
		HeaderStyle(lipgloss.NewStyle().
			Foreground(t.accent).
			Bold(true).
			Background(t.surfaceAlt)).
		WithBorderForeground(t.border).
		BorderRounded().
		WithOuterBorder(false).
		WithRowBorder(false).
		WithFooterVisibility(false).
		WithPaginationWrapping(false).
		Filtered(true).
		Focused(focused)
}

func (m *model) rebuildTable() {
	highlightedList := m.list.GetHighlightedRowIndex()
	highlightedDetails := m.tbl.GetHighlightedRowIndex()

	m.list = m.styledTable(m.listColumns(), m.buildListRows(), m.focus == focusMain && m.tab == tabPackages).
		WithFilterInputValue(m.query).
		WithHighlightedRow(highlightedList)

	m.tbl = m.styledTable(m.tableColumns(), m.buildTableRows(), m.focus == focusMain && m.tab == tabDetails).
		WithFilterInputValue(m.query).
		WithHighlightedRow(highlightedDetails)

	m.sizeTable()
}

func (m *model) listColumns() []table.Column {
	return []table.Column{
		table.NewColumn(columnKeyIndex, "#", 5).
			WithStyle(lipgloss.NewStyle().Foreground(m.theme.muted).Align(lipgloss.Right)),
		table.NewColumn(columnKeyStatus, "Status", 11).
			WithFiltered(true),
		table.NewColumn(columnKeyName, "Name", 20).
			WithFiltered(true),
		table.NewFlexColumn(columnKeyDescription, "Package Description", 1).
			WithFiltered(true),
	}
}

func (m *model) tableColumns() []table.Column {
	return []table.Column{
		table.NewColumn(columnKeyName, "Name", 18).
			WithFiltered(true),
		table.NewColumn(columnKeyVersion, "Version", 10).
			WithFiltered(true),
		table.NewColumn(columnKeyRevision, "Revision", 9).
			WithFiltered(true),
		table.NewColumn(columnKeyRepository, "Repository", 12).
			WithFiltered(true).
			WithStyle(lipgloss.NewStyle().Foreground(m.theme.muted)),
		table.NewColumn(columnKeySize, "Size", 9).
			WithStyle(lipgloss.NewStyle().Align(lipgloss.Right)),
		table.NewFlexColumn(columnKeyDescription, "Description", 1).
			WithFiltered(true),
	}
}

func (m *model) buildListRows() []table.Row {
	rows := make([]table.Row, 0, len(m.items))

	for i, it := range m.items {
		status := "available"
		if it.installed {
			status = "installed"
		}

		rows = append(rows, table.NewRow(table.RowData{
			columnKeyID:          i,
			columnKeyIndex:       strconv.Itoa(i + 1),
			columnKeyStatus:      status,
			columnKeyName:        it.name,
			columnKeyDescription: it.description,
			columnKeyInstalled:   it.installed,
		}))
	}

	return rows
}

func (m *model) buildTableRows() []table.Row {
	rows := make([]table.Row, 0, len(m.items))

	for i, it := range m.items {
		rows = append(rows, table.NewRow(table.RowData{
			columnKeyID:          i,
			columnKeyName:        it.name,
			columnKeyVersion:     it.version,
			columnKeyRevision:    it.revision,
			columnKeyRepository:  it.repository,
			columnKeySize:        it.size,
			columnKeyDescription: it.description,
			columnKeyInstalled:   it.installed,
		}))
	}

	return rows
}

func (m *model) refreshTableRows() {
	highlightedList := m.list.GetHighlightedRowIndex()
	highlightedDetails := m.tbl.GetHighlightedRowIndex()

	m.list = m.list.WithRows(m.buildListRows()).WithHighlightedRow(highlightedList)
	m.tbl = m.tbl.WithRows(m.buildTableRows()).WithHighlightedRow(highlightedDetails)
}

func (m *model) sizeTable() {
	contentW := max(10, m.width-4)
	bodyH := max(1, m.layout.mainH-2)
	rows := max(1, bodyH-2)

	// the table still accounts for the hidden outer border
	m.list = m.list.WithTargetWidth(contentW + 2).WithPageSize(rows)
	m.tbl = m.tbl.WithTargetWidth(contentW + 2).WithPageSize(rows)
}

func (m model) visibleRows() int {
	t := m.activeTable()
	return t.TotalRows()
}

func (m model) maxPages() int {
	t := m.activeTable()
	return t.MaxPages()
}

func (m model) currentPage() int {
	t := m.activeTable()
	return t.CurrentPage()
}
