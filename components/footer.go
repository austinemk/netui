package components

import "github.com/charmbracelet/lipgloss"

// RenderFooter draws clean, smart context action guides
func RenderFooter(activeTab int, isPopup bool) string {
	divider := lipgloss.NewStyle().Foreground(lipgloss.Color("#374151")).Render("\n────────────────────────────────────────────────────────────────────────\n ")
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))

	if isPopup {
		return divider + hintStyle.Render(" Esc: Close Overlay │ Enter: Submit Input │ Backspace: Delete")
	}

	switch activeTab {
	case 0: // Wi-Fi Tab Instructions
		return divider + hintStyle.Render(" ↔ Arrows/Tab: Subtabs │ j/k: Select Network │ Enter: Connect │ q: Quit")
	case 1:
		return divider + hintStyle.Render("r: refresh | q: quit app")
	default:
		return divider + hintStyle.Render(" 1-3: Swap Tabs │ q: Quit App")
	}
}
