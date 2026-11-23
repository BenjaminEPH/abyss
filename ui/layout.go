package ui

func (m model) resize() model {
	if m.width <= 0 || m.height <= 0 {
		return m
	}

	searchH := 3
	tabsH := 1
	footerH := m.footerHeight()
	bodyH := m.height - searchH - tabsH - footerH
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

	// The log viewport keeps its size even while hidden so it stays valid
	// when the user switches back to the details tab.
	m.search.SetWidth(max(8, contentW-4))
	m.logs.SetWidth(max(5, contentW-6))
	m.logs.SetHeight(max(1, logH-2))
	m.help.SetWidth(max(10, m.width-2))

	// The packages tab has no log panel: the table takes the whole body.
	if m.tab == tabPackages {
		mainH = bodyH
		logH = 0
	}

	mainY := searchH + tabsH
	m.layout = layout{
		searchY: 0,
		searchH: searchH,
		tabsY:   searchH,
		tabsH:   tabsH,
		mainY:   mainY,
		mainH:   mainH,
		logY:    mainY + mainH,
		logH:    logH,
		footerY: mainY + mainH + logH,
		footerH: footerH,
	}

	m.sizeTable()
	m.refreshLogs()

	return m
}
