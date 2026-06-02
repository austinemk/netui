// Package config for app configurations
package config

// Enforced window constraint thresholds
const (
	WindowWidth   = 70
	WindowHeight  = 25
	TabBodyHeight = WindowHeight - 7
	TabBodyWidth  = WindowWidth - 2
)

// TableHeight for setting table height
const (
	MinTableHeight = 7
	TableHeight    = 7
)

// popup box
const (
	PopupWidth  = WindowWidth / 2
	PopupHeight = WindowHeight / 2
	PopupHpos   = WindowWidth / 4
	PopupVpos   = (WindowHeight * 2) / 5
)

/*func PopupWindow(content string) string {
	return Styles.SuccessLog.Render(
		lipgloss.Place(
			PopupWidth,
			PopupHeight,
			PopupHpos,
			PopupVpos,
			content,
		),
	)
}*/

func Truncate(s string, max int) string {
	runes := []rune(s)

	if len(runes) <= max {
		return s
	}

	return string(runes[:max-3]) + "..."
}
