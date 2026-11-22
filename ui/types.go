package ui

import (
	"time"
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
