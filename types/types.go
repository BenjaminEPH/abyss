// Package types holds the data types shared between the UI and the rest of abyss.
package types

// Package is a row of the installed packages table (tab 1).
type Package struct {
	Status      string
	Name        string
	Description string
}

// Item is a single package entry shown in the UI.
type Item struct {
	Name        string
	Version     string
	Revision    string
	Repository  string
	Size        string
	Description string
	Installed   bool
}
