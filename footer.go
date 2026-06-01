package main

import (
	"corntui/config"

	"charm.land/lipgloss/v2"
)

// RenderFooter draws clean, smart context action guides
func RenderFooter(activeTab int, isPopup bool) string {
	tabhints := ""

	switch activeTab {
	case 0: // Wi-Fi Tab Instructions
		tabhints = "p: toggle power | s: start discovering"
	case 1:
		tabhints = "p: power | d: discoverable | b: pairable | s: scan new devices"
	default:
		tabhints = " 1-3: Swap Tabs │ q: Quit App"
	}

	hints := "j,k: up/down | enter: connect/options | esc: close popup | q: quit"

	// V2 Change: Vertically align using layout position method
	footer := lipgloss.JoinVertical(
		lipgloss.Center,
		config.DividerBorder(),
		config.Styles.TabHints.Render(tabhints),
		config.Styles.Hints.Render(hints),
	)

	return footer
}
