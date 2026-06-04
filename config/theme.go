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
	TabsBox       lipgloss.Style
	BodyText      lipgloss.Style
	PopupStyle    lipgloss.Style
	HighlightText lipgloss.Style
	CursorColor   lipgloss.Style
	Hints         lipgloss.Style
	AdapterInfo   lipgloss.Style
}

// Styles is the globally accessible style blueprint
var Styles CustomStyles

// Color variables that will be populated by defaults or TOML
var (
	ColorForeground          = ""
	ColorBackground          = ""
	ColorBorder              = "8"
	ColorAccent              = "5"
	ColorMuted               = "8"
	ColorHighlight           = "1"
	ColorHighlightBackground = ""
	ColorPopupBackground     = ""
	ColorLogBackground       = "#eba0ac"
	ColorCursor              = "#3B82F6"
)

// InitStyles constructs the styles dynamically (Renamed from init)
func InitStyles() {
	// Base application boundaries
	Styles.Container = lipgloss.NewStyle().
		Italic(true).
		Padding(0, 1).
		Width(int(float64(WindowWidth))).
		Height(int(float64(WindowHeight)))

		// Apply foreground if provided
	if ColorForeground != "" {
		Styles.Container = Styles.Container.Border(lipgloss.NormalBorder()).Foreground(lipgloss.Color(ColorForeground))
	}
	if ColorBorder != "" {
		Styles.Container = Styles.Container.Border(lipgloss.NormalBorder()).BorderForeground(lipgloss.Color(ColorBorder))
	}

	// Apply background if provided (otherwise remains transparent)
	if ColorBackground != "" {
		Styles.Container = Styles.Container.Background(lipgloss.Color(ColorBackground))
	}

	// Banner titles
	Styles.Title = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorAccent)).
		Bold(true).
		Padding(0, 1)

		// Contextual Headers
	Styles.Heading = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorMuted)).
		Underline(true).
		Bold(true).
		BorderBottom(true).
		Italic(true)

	Styles.LogBox = lipgloss.NewStyle().
		Background(lipgloss.Color(ColorLogBackground)).
		Padding(0, 1).
		Foreground(lipgloss.Color("#1e1e2e")).
		Bold(true).
		Width(WindowWidth - 2).
		AlignHorizontal(lipgloss.Center)

	// Tab structures
	Styles.ActiveTab = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorAccent)).
		Bold(true)

	Styles.InactiveTab = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorMuted)).
		Bold(true)

	Styles.TabsBox = lipgloss.NewStyle().Padding(0, 1).
		AlignHorizontal(lipgloss.Center).
		Height(1).Border(lipgloss.NormalBorder()).
		BorderBottom(true).BorderTop(true).BorderLeft(false).BorderRight(false).
		BorderForeground(lipgloss.Color(ColorBorder))

	// popup  structures
	Styles.PopupStyle = lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Padding(1).
		Width(int(math.Floor(float64(ListWidth)*0.6))).
		Margin(0, int(math.Floor(float64(ListWidth)*0.15))).
		Height(ListHeightHalf)

	if ColorBorder != "" {
		Styles.PopupStyle = Styles.PopupStyle.BorderForeground(lipgloss.Color(ColorBorder))
	} else {
		Styles.PopupStyle = Styles.PopupStyle.BorderForeground(lipgloss.Color("8"))
	}

	if ColorPopupBackground != "" {
		Styles.PopupStyle = Styles.PopupStyle.Background(lipgloss.Color(ColorPopupBackground))
	}

	// highlight structures
	Styles.HighlightText = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorHighlight)).
		Padding(0, 1)

	if ColorHighlightBackground != "" {
		Styles.HighlightText = Styles.HighlightText.Background(lipgloss.Color(ColorHighlightBackground))
	}

	Styles.CursorColor = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorCursor))

	Styles.Hints = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorMuted)).
		Bold(true)

	Styles.AdapterInfo = lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorAccent))
}
