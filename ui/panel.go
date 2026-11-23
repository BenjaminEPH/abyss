package ui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

func focusName(focus int) string {
	switch focus {
	case focusSearch:
		return "search"
	case focusMain:
		return "packages"
	case focusLog:
		return "log"
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
