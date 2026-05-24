package config

import "github.com/charmbracelet/lipgloss"

// CustomStyles aggregates the look and feel elements of netui
type CustomStyles struct {
	Container     lipgloss.Style
	Title         lipgloss.Style
	ActiveTab     lipgloss.Style
	InactiveTab   lipgloss.Style
	ActiveSub     lipgloss.Style
	InactiveSub   lipgloss.Style
	HighlightText lipgloss.Style
	LogFrame      lipgloss.Style
	CursorColor   lipgloss.Style
	Notice        lipgloss.Style
}

// Styles is the globally accessible style blueprint
var Styles CustomStyles

func init() {
	// Base application boundaries
	Styles.Container = lipgloss.NewStyle().
		Margin(1, 2).
		Width(72).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8"))

	// Banner titles
	Styles.Title = lipgloss.NewStyle().
		Background(lipgloss.Color("#2563EB")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Padding(0, 1)

	// Tab structures
	Styles.ActiveTab = lipgloss.NewStyle().
		Background(lipgloss.Color("8")).
		Foreground(lipgloss.Color("2")).
		Padding(0, 1)

	Styles.InactiveTab = lipgloss.NewStyle().
		Background(lipgloss.Color("#1F2937")).
		Foreground(lipgloss.Color("#9CA3AF")).
		Padding(0, 1)

	// Subtabs structures
	Styles.ActiveSub = lipgloss.NewStyle().
		Foreground(lipgloss.Color("4")).
		Bold(true).
		Padding(0, 1)

	Styles.InactiveSub = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF")).
		Padding(0, 1)

	Styles.HighlightText = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4"))

	// Status context indicators
	Styles.LogFrame = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F59E0B")).
		Italic(true)

	Styles.CursorColor = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3B82F6"))

	// Noficication
	Styles.Notice = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).Italic(true)
}
