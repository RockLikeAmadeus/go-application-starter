package tui

import "charm.land/lipgloss/v2"

var (
	accent = lipgloss.Color("#7C3AED")
	muted  = lipgloss.Color("#6B7280")
	danger = lipgloss.Color("#DC2626")

	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FFFFFF")).Background(accent).Padding(0, 1)
	selectedStyle = lipgloss.NewStyle().Foreground(accent).Bold(true)
	itemStyle     = lipgloss.NewStyle().PaddingLeft(2)
	helpStyle     = lipgloss.NewStyle().Foreground(muted)
	errorStyle    = lipgloss.NewStyle().Foreground(danger)
	frameStyle    = lipgloss.NewStyle().Padding(1, 2)
)
