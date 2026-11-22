package ui

import ()

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
