package config

import (
	"os"

	"github.com/BurntSushi/toml"
)

// AppConfig mirrors the config.toml file structure
type AppConfig struct {
	Window struct {
		Width  int `toml:"width"`
		Height int `toml:"height"`
	} `toml:"window"`

	Colors struct {
		Foreground          string `toml:"foreground"`
		Background          string `toml:"background"`
		Accent              string `timl:"accent"`
		Highlight           string `toml:"highlight"`
		HighlightBackground string `toml:"Highlight_background"`
		Muted               string `toml:"muted"`
		Border              string `toml:"border"`
		LogBackground       string `toml:"log_background"`
		Cursor              string `toml:"cursor"`
	} `toml:"colors"`
}

// LoadConfig opens the TOML file, maps configurations, and initializes styles
func LoadConfig(filePath string) error {
	cfg := AppConfig{}

	// 1. Assign current file values as fallbacks in case file elements are missing
	cfg.Window.Width = WindowWidth
	cfg.Window.Height = WindowHeight
	cfg.Colors.Foreground = ColorForeground
	cfg.Colors.Background = ColorBackground
	cfg.Colors.Border = ColorBorder
	cfg.Colors.Accent = ColorAccent
	cfg.Colors.Highlight = ColorHighlight
	cfg.Colors.HighlightBackground = ColorHighlightBackground
	cfg.Colors.Muted = ColorMuted
	cfg.Colors.LogBackground = ColorLogBackground
	cfg.Colors.Cursor = ColorCursor

	// 2. Decode the TOML file if it is found on disk
	if _, err := os.Stat(filePath); err == nil {
		if _, err := toml.DecodeFile(filePath, &cfg); err != nil {
			return err
		}
	}

	// 3. Update global window constraint variables
	WindowWidth = cfg.Window.Width
	WindowHeight = cfg.Window.Height

	// CRITICAL: Recalculate dependent grid layout numbers based on the new dimensions
	ListHeight = WindowHeight - OtherContentHeight
	ListWidth = WindowWidth - 2
	ListHeightHalf = ListHeight / 2
	ListHeightQuarter = ListHeight / 4
	ListWidthHalf = ListWidth / 2
	ListWidthQuarter = ListWidth / 4
	ListWidthEigth = ListWidth / 8
	ListWidthSixteenth = ListWidth / 16
	HeaderSpacing = (WindowWidth - 20) / 8

	PopupWidth = (ListWidth * 3) / 5
	PopupHeight = ListWidthHalf

	// 4. Update style colors
	ColorForeground = cfg.Colors.Foreground
	ColorBackground = cfg.Colors.Background
	ColorBorder = cfg.Colors.Border
	ColorLogBackground = cfg.Colors.LogBackground
	ColorAccent = cfg.Colors.Accent
	ColorHighlight = cfg.Colors.Highlight
	ColorHighlightBackground = cfg.Colors.HighlightBackground
	ColorMuted = cfg.Colors.Muted
	ColorCursor = cfg.Colors.Cursor

	// 5. Build Lipgloss styles with the newly mapped sizes and themes
	InitStyles()

	return nil
}
