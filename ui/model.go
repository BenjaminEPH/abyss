package ui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/evertras/bubble-table/table"
)

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
			{title: "Buy Carrots", tag: "shopping"},
			{title: "Buy celery", tag: "shopping"},
			{title: "Review pull request", tag: "work"},
			{title: "Write release notes", tag: "work", done: true},
			{title: "Water the plants", tag: "home"},
			{title: "Read 20 pages", tag: "personal", done: true},
		},
	}
	m.theme = newTheme(m.dark)

	m.search = textinput.New()
	m.search.Prompt = "⌕ "
	m.search.Placeholder = "Search tasks or tags…"
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

	m.addLog(levelInfo, "abyss started")
	m.addLog(levelSuccess, "%d tasks loaded", len(m.items))
	m.addLog(levelInfo, "tab changes focus · / search · ? help")
	m.logs.GotoBottom()

	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.RequestBackgroundColor, m.search.Focus())
}
