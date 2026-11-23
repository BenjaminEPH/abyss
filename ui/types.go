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
	columnKeyID          = "id"
	columnKeyName        = "name"
	columnKeyVersion     = "version"
	columnKeyRevision    = "revision"
	columnKeyRepository  = "repository"
	columnKeySize        = "size"
	columnKeyDescription = "description"
	columnKeyInstalled   = "installed"
)

type item struct {
	name        string
	version     string
	revision    string
	repository  string
	size        string
	description string
	installed   bool
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
