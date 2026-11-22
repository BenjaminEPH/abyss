package ui

import (
	"charm.land/lipgloss/v2"
	"github.com/evertras/bubble-table/table"
)

func (m *model) rebuildTable() {
	t := m.theme
	highlighted := m.tbl.GetHighlightedRowIndex()

	m.tbl = table.New(m.tableColumns()).
		WithRows(m.buildTableRows()).
		WithBaseStyle(lipgloss.NewStyle().Foreground(t.text).Align(lipgloss.Left)).
		WithRowStyleFunc(func(input table.RowStyleFuncInput) lipgloss.Style {
			done, _ := input.Row.Data[columnKeyIsDone].(bool)

			if input.IsHighlighted {
				return lipgloss.NewStyle().
					Background(t.selectedBg).
					Foreground(t.selectedFg).
					Bold(true)
			}
			if done {
				return lipgloss.NewStyle().
					Foreground(t.faint).
					Strikethrough(true)
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
		Focused(m.focus == focusMain)

	m.tbl = m.tbl.WithFilterInputValue(m.query)
	m.tbl = m.tbl.WithHighlightedRow(highlighted)
	m.sizeTable()
}

func (m *model) tableColumns() []table.Column {
	return []table.Column{
		table.NewColumn(columnKeyDone, "Done", 5).
			WithStyle(lipgloss.NewStyle().Align(lipgloss.Center)),
		table.NewFlexColumn(columnKeyTitle, "Task", 1).
			WithFiltered(true),
		table.NewColumn(columnKeyTag, "Tag", 12).
			WithFiltered(true).
			WithStyle(lipgloss.NewStyle().Foreground(m.theme.muted)),
	}
}

func (m *model) buildTableRows() []table.Row {
	rows := make([]table.Row, 0, len(m.items))

	for i, it := range m.items {
		glyph := "○"
		glyphStyle := lipgloss.NewStyle().Foreground(m.theme.muted)
		if it.done {
			glyph = "✓"
			glyphStyle = lipgloss.NewStyle().Foreground(m.theme.success).Bold(true)
		}

		rows = append(rows, table.NewRow(table.RowData{
			columnKeyID:     i,
			columnKeyDone:   table.NewStyledCell(glyph, glyphStyle),
			columnKeyTitle:  it.title,
			columnKeyTag:    "#" + it.tag,
			columnKeyIsDone: it.done,
		}))
	}

	return rows
}

func (m *model) refreshTableRows() {
	highlighted := m.tbl.GetHighlightedRowIndex()
	m.tbl = m.tbl.WithRows(m.buildTableRows())
	m.tbl = m.tbl.WithHighlightedRow(highlighted)
}

func (m *model) sizeTable() {
	contentW := max(10, m.width-4)
	bodyH := max(1, m.layout.mainH-2)
	rows := max(1, bodyH-2)

	m.tbl = m.tbl.
		WithTargetWidth(contentW + 2). // the table still accounts for the hidden outer border
		WithPageSize(rows)
}
