// Package config for app configurations
package config

// Enforced window constraint thresholds
var (
	WindowWidth  = 70
	WindowHeight = 25
)

// All dependent layout items MUST be vars so they can be recalculated
var (
	OtherContentHeight = 10
	ListHeight         = WindowHeight - OtherContentHeight
	ListWidth          = WindowWidth - 2
	ListHeightHalf     = ListHeight / 2
	ListHeightQuarter  = ListHeight / 4
	ListWidthHalf      = ListWidth / 2
	ListWidthQuarter   = ListWidth / 4
	ListWidthEigth     = ListWidth / 8
	ListWidthSixteenth = ListWidth / 16
)

// Header
var (
	HeaderSpacing = (WindowWidth - 20) / 8
)

// Popup box layout variables
var (
	PopupWidth  = (ListWidth * 3) / 5
	PopupHeight = ListWidthHalf
)

func Truncate(s string, max int) string {
	runes := []rune(s)

	if len(runes) <= max {
		return s
	}

	return string(runes[:max-3]) + "..."
}
