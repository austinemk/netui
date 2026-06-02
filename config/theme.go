package config

import (
	"math"

	"charm.land/lipgloss/v2"
)

// CustomStyles aggregates the look and feel elements of netui
type CustomStyles struct {
	Container     lipgloss.Style
	Title         lipgloss.Style
	Heading       lipgloss.Style
	LogBox        lipgloss.Style
	ActiveTab     lipgloss.Style
	InactiveTab   lipgloss.Style
	BodyText      lipgloss.Style
	BoxStyle      lipgloss.Style
	InfoText      lipgloss.Style
	HighlightText lipgloss.Style
	CursorColor   lipgloss.Style
	Hints         lipgloss.Style
	TabHints      lipgloss.Style
}

// Styles is the globally accessible style blueprint
var Styles CustomStyles

func init() {
	// Base application boundaries
	Styles.Container = lipgloss.NewStyle().
		Margin(1, 2).
		Italic(true).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8"))

	// Banner titles
	Styles.Title = lipgloss.NewStyle().
		Background(lipgloss.Color("#2563EB")).
		Foreground(lipgloss.Color("#FFFFFF")).
		Bold(true).
		Padding(0, 1)

	// Contextual Headers (New additions)
	Styles.Heading = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Underline(true).
		Bold(true).
		BorderBottom(true).
		Italic(true)

	Styles.LogBox = lipgloss.NewStyle().
		Background(lipgloss.Color("#eba0ac")).
		Padding(0, 1).
		Foreground(lipgloss.Color("#1e1e2e")).
		Bold(true).Width(TabBodyWidth).AlignHorizontal(lipgloss.Center)

	// Tab structures
	Styles.ActiveTab = lipgloss.NewStyle().
		Background(lipgloss.Color("8")).
		Foreground(lipgloss.Color("2")).
		Padding(0, 1)

	Styles.InactiveTab = lipgloss.NewStyle().
		Background(lipgloss.Color("#1F2937")).
		Foreground(lipgloss.Color("#9CA3AF")).
		Padding(0, 1)

	// BodyText
	//Styles.BodyText = lipgloss.NewStyle().Foreground(lipgloss.Color())

	// Subtabs structures
	Styles.BoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("3")).
		Padding(1).
		Width(int(math.Floor(TabBodyWidth*0.6))).
		Margin(0, int(math.Floor(TabBodyWidth*0.15))).
		Height(int(math.Floor(TabBodyHeight * 0.4)))

	Styles.InfoText = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)

	Styles.HighlightText = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("7")).Background(lipgloss.Color("8")).Padding(0, 1)

	Styles.CursorColor = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#3B82F6"))

	Styles.Hints = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Bold(true)
	Styles.TabHints = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Italic(true)
}
