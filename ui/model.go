package ui

import (
	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"github.com/BenjaminEPH/abyss/packages"
	"github.com/BenjaminEPH/abyss/types"
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

	packages []types.Package // packages tab
	items    []types.Item    // details tab
	entries  []logEntry

	query     string
	searchSeq int
	loading   bool

	theme  theme
	layout layout
}

func initialModel() model {
	installed, loadErr := packages.QueryInstalled()

	m := model{
		packages:    installed,
		dark:        true,
		termFocused: true,
		width:       80,
		height:      24,
		keys:        defaultKeyMap(),
		help:        help.New(),
		spinner:     spinner.New(spinner.WithSpinner(spinner.MiniDot)),
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
	if loadErr != nil {
		m.addLog(levelError, "xbps-query failed: %v", loadErr)
	} else {
		m.addLog(levelSuccess, "%d packages loaded", len(m.packages))
	}
	m.addLog(levelInfo, "tab changes focus · / search · ? help")
	m.logs.GotoBottom()

	return m
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tea.RequestBackgroundColor, m.search.Focus())
}
