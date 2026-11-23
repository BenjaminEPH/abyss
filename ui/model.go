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
	tab           int
	dark          bool
	termFocused   bool

	search  textinput.Model
	list    table.Model // packages tab
	tbl     table.Model // details tab
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
			{name: "bash", version: "5.2.37", revision: "1", repository: "core", size: "1.8 MiB", description: "The GNU Bourne Again shell"},
			{name: "coreutils", version: "9.5", revision: "2", repository: "core", size: "7.2 MiB", description: "The basic file, shell and text manipulation utilities"},
			{name: "git", version: "2.47.1", revision: "1", repository: "extra", size: "14.5 MiB", description: "The fast distributed version control system", installed: true},
			{name: "go", version: "1.23.4", revision: "1", repository: "extra", size: "242.0 MiB", description: "Core compiler tools for the Go programming language"},
			{name: "neovim", version: "0.10.3", revision: "1", repository: "extra", size: "28.4 MiB", description: "Fork of Vim aiming to improve user experience", installed: true},
			{name: "ripgrep", version: "14.1.1", revision: "1", repository: "extra", size: "5.1 MiB", description: "A search tool that combines the usability of ag with the speed of grep"},
		},
	}
	m.theme = newTheme(m.dark)

	m.search = textinput.New()
	m.search.Prompt = "⌕ "
	m.search.Placeholder = "Search packages…"
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
	m.addLog(levelSuccess, "%d packages loaded", len(m.items))
	m.addLog(levelInfo, "tab changes focus · / search · ? help")
	m.logs.GotoBottom()

	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.RequestBackgroundColor, m.search.Focus())
}
