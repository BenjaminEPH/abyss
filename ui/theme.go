package ui

import (
	"image/color"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
)

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
