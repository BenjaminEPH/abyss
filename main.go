package main

import (
	"fmt"
	"image/color"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/evertras/bubble-table/table"
)

const (
	focusSearch = iota
	focusMain
	focusLog
	focusCount
)

const (
	columnKeyID     = "id"
	columnKeyDone   = "done"
	columnKeyTitle  = "title"
	columnKeyTag    = "tag"
	columnKeyIsDone = "is_done"
)

type item struct {
	title string
	tag   string
	done  bool
}

type logLevel int

const (
	levelInfo logLevel = iota
	levelSuccess
	levelWarning
	levelError
)

type logEntry struct {
	at    time.Time
	level logLevel
	msg   string
}

type layout struct {
	searchY, searchH int
	mainY, mainH     int
	logY, logH       int
	footerY, footerH int
}

type filterMsg struct {
	seq   int
	query string
}

type theme struct {
	dark bool

	accent     color.Color
	text       color.Color
	muted      color.Color
	faint      color.Color
	border     color.Color
	surfaceAlt color.Color
	selectedBg color.Color
	selectedFg color.Color
	success    color.Color
	warning    color.Color
	danger     color.Color
	info       color.Color
}

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
			key.WithHelp("tab", "foco"),
		),
		Search: key.NewBinding(
			key.WithKeys("/", "ctrl+k"),
			key.WithHelp("/", "buscar"),
		),
		Clear: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "limpiar"),
		),
		Move: key.NewBinding(
			key.WithKeys("up", "down", "k", "j"),
			key.WithHelp("↑/↓", "mover"),
		),
		Toggle: key.NewBinding(
			key.WithKeys("enter", "space"),
			key.WithHelp("⏎", "completar"),
		),
		Top: key.NewBinding(
			key.WithKeys("g", "G", "home", "end"),
			key.WithHelp("g/G", "inicio/fin"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "ayuda"),
		),
		Quit: key.NewBinding(
			key.WithKeys("q", "ctrl+c"),
			key.WithHelp("q", "salir"),
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

type model struct {
	width, height int
	focus         int
	dark          bool
	termFocused   bool

	search  textinput.Model
	tbl     table.Model
	logs    viewport.Model
	help    help.Model
	spinner spinner.Model
	keys    keyMap

	items   []item
	entries []logEntry

	query     string
	searchSeq int
	loading   bool

	theme  theme
	layout layout
}

func newTheme(dark bool) theme {
	lightDark := lipgloss.LightDark(dark)

	return theme{
		dark:       dark,
		accent:     lightDark(lipgloss.Color("#6D28D9"), lipgloss.Color("#A78BFA")),
		text:       lightDark(lipgloss.Color("#27272A"), lipgloss.Color("#E4E4E7")),
		muted:      lightDark(lipgloss.Color("#71717A"), lipgloss.Color("#A1A1AA")),
		faint:      lightDark(lipgloss.Color("#A1A1AA"), lipgloss.Color("#52525B")),
		border:     lightDark(lipgloss.Color("#D4D4D8"), lipgloss.Color("#3F3F46")),
		surfaceAlt: lightDark(lipgloss.Color("#E4E4E7"), lipgloss.Color("#27272A")),
		selectedBg: lightDark(lipgloss.Color("#DDD6FE"), lipgloss.Color("#312E81")),
		selectedFg: lightDark(lipgloss.Color("#3B0764"), lipgloss.Color("#EDE9FE")),
		success:    lightDark(lipgloss.Color("#15803D"), lipgloss.Color("#4ADE80")),
		warning:    lightDark(lipgloss.Color("#B45309"), lipgloss.Color("#FBBF24")),
		danger:     lightDark(lipgloss.Color("#B91C1C"), lipgloss.Color("#F87171")),
		info:       lightDark(lipgloss.Color("#0E7490"), lipgloss.Color("#67E8F9")),
	}
}

func initialModel() model {
	m := model{
		dark:        true,
		termFocused: true,
		width:       80,
		height:      24,
		keys:        defaultKeyMap(),
		help:        help.New(),
		spinner:     spinner.New(spinner.WithSpinner(spinner.MiniDot)),
		items: []item{
			{title: "Buy Carrots", tag: "compras"},
			{title: "Buy celery", tag: "compras"},
			{title: "Review pull request", tag: "trabajo"},
			{title: "Write release notes", tag: "trabajo", done: true},
			{title: "Water the plants", tag: "casa"},
			{title: "Read 20 pages", tag: "personal", done: true},
		},
	}
	m.theme = newTheme(m.dark)

	m.search = textinput.New()
	m.search.Prompt = "⌕ "
	m.search.Placeholder = "Buscar tareas o etiquetas…"
	m.search.CharLimit = 80
	m.search.Focus()

	m.logs = viewport.New()
	m.logs.SoftWrap = true
	m.logs.FillHeight = true
	m.logs.MouseWheelEnabled = true
	m.logs.KeyMap = logViewportKeys()

	m.applyTheme()
	m.rebuildTable()
	m = m.resize()

	m.addLog(levelInfo, "abyss iniciado")
	m.addLog(levelSuccess, "%d tareas cargadas", len(m.items))
	m.addLog(levelInfo, "tab cambia el foco · / busca · ? ayuda")
	m.logs.GotoBottom()

	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.RequestBackgroundColor, m.search.Focus())
}

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
	case key.Matches(msg, m.keys.Help):
		if m.height < 20 {
			m.addLog(levelWarning, "agrandá la ventana para ver la ayuda completa")
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
		var cmd tea.Cmd
		m.tbl, cmd = m.tbl.Update(msg)
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

	case y >= m.layout.mainY && y < m.layout.mainY+m.layout.mainH:
		cmd := m.setFocus(focusMain)
		// Table body starts after the panel border and the table header.
		row := y - (m.layout.mainY + 1) - 2
		if row >= 0 {
			start, end := m.tbl.VisibleIndices()
			idx := start + row
			if idx >= start && idx <= end && idx < m.tbl.TotalRows() {
				m.tbl = m.tbl.WithHighlightedRow(idx)
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
		switch msg.Button {
		case tea.MouseWheelDown:
			m.tbl = m.tbl.PageDown()
		case tea.MouseWheelUp:
			m.tbl = m.tbl.PageUp()
		}
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
	f := (m.focus + delta + focusCount) % focusCount
	return m.setFocus(f)
}

func (m *model) setFocus(f int) tea.Cmd {
	if f < 0 || f >= focusCount {
		return nil
	}

	m.focus = f
	m.tbl = m.tbl.Focused(f == focusMain)

	if f == focusSearch {
		m.search.CursorEnd()
		return m.search.Focus()
	}

	m.search.Blur()
	return nil
}

func (m *model) applyFilter(q string) {
	q = strings.TrimSpace(q)
	if q == m.query {
		return
	}

	prev := m.query
	m.query = q
	m.tbl = m.tbl.WithFilterInputValue(q)

	if q == "" {
		if prev != "" {
			m.addLog(levelInfo, "filtro limpiado · %d tareas", len(m.items))
		}
		return
	}

	n := m.tbl.TotalRows()
	if n == 0 {
		m.addLog(levelWarning, "sin coincidencias para %q", q)
		return
	}
	m.addLog(levelInfo, "filtro %q · %d coincidencias", q, n)
}

func (m *model) toggleSelected() {
	row := m.tbl.HighlightedRow()
	if row.Data == nil {
		return
	}

	id, ok := row.Data[columnKeyID].(int)
	if !ok || id < 0 || id >= len(m.items) {
		return
	}

	m.items[id].done = !m.items[id].done
	if m.items[id].done {
		m.addLog(levelSuccess, "completada: %s", m.items[id].title)
	} else {
		m.addLog(levelInfo, "reabierta: %s", m.items[id].title)
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

func (m *model) applyTheme() {
	t := m.theme

	searchStyles := textinput.DefaultStyles(t.dark)
	searchStyles.Focused.Prompt = lipgloss.NewStyle().Foreground(t.accent).Bold(true)
	searchStyles.Focused.Text = lipgloss.NewStyle().Foreground(t.text)
	searchStyles.Focused.Placeholder = lipgloss.NewStyle().Foreground(t.faint)
	searchStyles.Blurred.Prompt = lipgloss.NewStyle().Foreground(t.muted)
	searchStyles.Blurred.Text = lipgloss.NewStyle().Foreground(t.muted)
	searchStyles.Blurred.Placeholder = lipgloss.NewStyle().Foreground(t.faint)
	searchStyles.Cursor.Color = t.accent
	m.search.SetStyles(searchStyles)

	helpStyles := help.DefaultStyles(t.dark)
	helpStyles.ShortKey = lipgloss.NewStyle().Foreground(t.accent)
	helpStyles.ShortDesc = lipgloss.NewStyle().Foreground(t.muted)
	helpStyles.ShortSeparator = lipgloss.NewStyle().Foreground(t.faint)
	helpStyles.FullKey = lipgloss.NewStyle().Foreground(t.accent)
	helpStyles.FullDesc = lipgloss.NewStyle().Foreground(t.muted)
	helpStyles.FullSeparator = lipgloss.NewStyle().Foreground(t.faint)
	helpStyles.Ellipsis = lipgloss.NewStyle().Foreground(t.faint)
	m.help.Styles = helpStyles

	m.spinner.Style = lipgloss.NewStyle().Foreground(t.accent)
}

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
		table.NewColumn(columnKeyDone, "Hecho", 5).
			WithStyle(lipgloss.NewStyle().Align(lipgloss.Center)),
		table.NewFlexColumn(columnKeyTitle, "Tarea", 1).
			WithFiltered(true),
		table.NewColumn(columnKeyTag, "Etiqueta", 12).
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

func (m model) resize() model {
	if m.width <= 0 || m.height <= 0 {
		return m
	}

	searchH := 3
	footerH := m.footerHeight()
	bodyH := m.height - searchH - footerH
	if bodyH < 6 {
		bodyH = 6
	}

	logH := bodyH * 40 / 100
	if logH < 3 {
		logH = 3
	}
	if logH > bodyH-3 {
		logH = bodyH - 3
	}

	mainH := bodyH - logH
	if mainH < 3 {
		mainH = 3
		logH = bodyH - mainH
	}

	contentW := max(10, m.width-4)

	m.search.SetWidth(max(8, contentW-4))
	m.logs.SetWidth(max(5, contentW-6))
	m.logs.SetHeight(max(1, logH-2))
	m.help.SetWidth(max(10, m.width-2))

	m.layout = layout{
		searchY: 0,
		searchH: searchH,
		mainY:   searchH,
		mainH:   mainH,
		logY:    searchH + mainH,
		logH:    logH,
		footerY: searchH + mainH + logH,
		footerH: footerH,
	}

	m.sizeTable()
	m.refreshLogs()

	return m
}

func (m model) View() tea.View {
	if m.width == 0 {
		v := tea.NewView("cargando…")
		v.AltScreen = true
		return v
	}

	if m.width < 40 || m.height < 12 {
		msg := lipgloss.NewStyle().Foreground(m.theme.muted).
			Render("abyss · agrandá la ventana para ver la interfaz")
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
	if m.tbl.TotalRows() == 0 {
		mainBody = strings.Join(m.emptyState(contentW, max(1, m.layout.mainH-2)), "\n")
	} else {
		mainBody = m.tbl.View()
	}

	searchPanel := renderPanel(m.theme, m.width, m.layout.searchH,
		"⌕ BUSCAR", m.searchBadge(), m.focus == focusSearch, searchBody)
	mainPanel := renderPanel(m.theme, m.width, m.layout.mainH,
		"▤ TAREAS", m.mainBadge(), m.focus == focusMain, mainBody)
	logPanel := renderPanel(m.theme, m.width, m.layout.logH,
		"≡ REGISTRO", m.logBadge(), m.focus == focusLog, m.logs.View())
	footer := m.footerView(m.width)

	content := lipgloss.JoinVertical(lipgloss.Left,
		searchPanel,
		mainPanel,
		logPanel,
		footer,
	)

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
		hint = hintStyle.Render("esc limpiar")
	} else if m.focus != focusSearch && m.search.Value() != "" {
		hint = hintStyle.Render("tab editar")
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
	title := faint.Render("∅") + "  " + muted.Render("sin resultados")
	hint := faint.Render("probá otra búsqueda · esc para limpiar")
	lines[mid] = lipgloss.PlaceHorizontal(width, lipgloss.Center, title)
	if mid+1 < height {
		lines[mid+1] = lipgloss.PlaceHorizontal(width, lipgloss.Center, hint)
	}

	return lines
}

func (m model) searchBadge() string {
	if m.loading {
		return "filtrando…"
	}
	if m.query != "" {
		return fmt.Sprintf("%d resultados", m.tbl.TotalRows())
	}
	return fmt.Sprintf("%d tareas", len(m.items))
}

func (m model) mainBadge() string {
	done := 0
	for _, it := range m.items {
		if it.done {
			done++
		}
	}

	pct := 0
	if len(m.items) > 0 {
		pct = done * 100 / len(m.items)
	}

	badge := fmt.Sprintf("%d/%d · %d%%", m.tbl.TotalRows(), len(m.items), pct)
	if pages := m.tbl.MaxPages(); pages > 1 {
		badge += fmt.Sprintf(" · pág %d/%d", m.tbl.CurrentPage(), pages)
	}
	return badge
}

func (m model) logBadge() string {
	return fmt.Sprintf("%d líneas", len(m.entries))
}

func (m model) statusView() string {
	t := m.theme

	stateDot := lipgloss.NewStyle().Foreground(t.success).Render("●")
	state := "listo"
	if m.loading {
		stateDot = m.spinner.View()
		state = "filtrando"
	}
	if !m.termFocused {
		stateDot = lipgloss.NewStyle().Foreground(t.warning).Render("○")
		state = "en pausa"
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
			lipgloss.NewStyle().Foreground(t.muted).Render("filtro ")+
				lipgloss.NewStyle().Foreground(t.accent).Render(q))
	}

	parts = append(parts,
		lipgloss.NewStyle().Foreground(t.muted).
			Render(fmt.Sprintf("%d/%d tareas", m.tbl.TotalRows(), len(m.items))),
		lipgloss.NewStyle().Foreground(t.muted).
			Render("foco: "+focusName(m.focus)),
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
		return fmt.Sprintf("abyss · %s (%d)", m.query, m.tbl.TotalRows())
	}
	return fmt.Sprintf("abyss · %d tareas", len(m.items))
}

func (m model) programStatus() *tea.ProgramStatus {
	status := &tea.ProgramStatus{
		State:   tea.ProgramStateIdle,
		App:     "abyss",
		Title:   "abyss",
		Message: "listo",
	}
	if m.loading {
		status.State = tea.ProgramStateWorking
		status.Message = "filtrando tareas"
	}
	return status
}

func (m model) progressBar() *tea.ProgressBar {
	if len(m.items) == 0 {
		return nil
	}

	done := 0
	for _, it := range m.items {
		if it.done {
			done++
		}
	}

	return tea.NewProgressBar(tea.ProgressBarDefault, done*100/len(m.items))
}

func focusName(focus int) string {
	switch focus {
	case focusSearch:
		return "buscar"
	case focusMain:
		return "tareas"
	case focusLog:
		return "registro"
	}
	return "?"
}

func renderPanel(t theme, width, height int, title, badge string, focused bool, body string) string {
	if width < 4 || height < 3 {
		return ""
	}

	inner := width - 2
	borderColor := t.border
	titleColor := t.muted
	if focused {
		borderColor = t.accent
		titleColor = t.accent
	}

	borderStyle := lipgloss.NewStyle().Foreground(borderColor)
	titleStyle := lipgloss.NewStyle().Foreground(titleColor).Bold(focused)
	badgeStyle := lipgloss.NewStyle().Foreground(t.faint)

	titleText := " " + title + " "
	badgeText := ""
	if badge != "" {
		badgeText = " " + badge + " "
	}

	available := inner - 2 // one dash on each side of the title block
	if lipgloss.Width(titleText)+lipgloss.Width(badgeText) > available {
		if badgeText != "" {
			maxBadge := available - lipgloss.Width(titleText)
			if maxBadge >= 4 {
				badgeText = " " + ansi.Truncate(badge, maxBadge-2, "…") + " "
			} else {
				badgeText = ""
			}
		}
		if lipgloss.Width(titleText) > available {
			maxTitle := max(1, available)
			titleText = " " + ansi.Truncate(title, max(1, maxTitle-2), "…") + " "
		}
	}

	fill := available - lipgloss.Width(titleText) - lipgloss.Width(badgeText)
	if fill < 0 {
		fill = 0
	}

	top := borderStyle.Render("╭─") +
		titleStyle.Render(titleText) +
		borderStyle.Render(strings.Repeat("─", fill)) +
		badgeStyle.Render(badgeText) +
		borderStyle.Render("─╮")

	contentW := max(0, width-4)
	bodyLines := strings.Split(body, "\n")
	rows := make([]string, 0, height)
	for i := 0; i < height-2; i++ {
		line := ""
		if i < len(bodyLines) {
			line = bodyLines[i]
		}
		rows = append(rows, borderStyle.Render("│")+
			" "+fitANSI(line, contentW)+" "+
			borderStyle.Render("│"))
	}

	bottom := borderStyle.Render("╰" + strings.Repeat("─", inner) + "╯")

	return top + "\n" + strings.Join(rows, "\n") + "\n" + bottom
}

func fitANSI(s string, width int) string {
	if width <= 0 {
		return ""
	}

	current := lipgloss.Width(s)
	if current > width {
		return ansi.Truncate(s, width, "…")
	}
	if current < width {
		return s + strings.Repeat(" ", width-current)
	}
	return s
}

func logViewportKeys() viewport.KeyMap {
	keys := viewport.DefaultKeyMap()
	keys.PageDown = key.NewBinding(key.WithKeys("pgdown", "f"))
	keys.PageUp = key.NewBinding(key.WithKeys("pgup", "b"))
	keys.HalfPageDown = key.NewBinding(key.WithKeys("d", "ctrl+d"))
	keys.HalfPageUp = key.NewBinding(key.WithKeys("u", "ctrl+u"))
	return keys
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}
}
