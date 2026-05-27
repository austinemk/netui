package components

import "github.com/charmbracelet/lipgloss"

// RenderFooter draws clean, smart context action guides
func RenderFooter(activeTab int, isPopup bool) string {
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	hints := ""

	if isPopup {
		hints = hints + " esc: close"
	}

	/*switch activeTab {
	case 0: // Wi-Fi Tab Instructions
		return divider + hintStyle.Render("j/k/: navigate list │ Enter: Connect/options | s: scan/off ")
	case 1:
		return divider + hintStyle.Render("r: refresh | q: quit app")
	default:
		return divider + hintStyle.Render(" 1-3: Swap Tabs │ q: Quit App")
	}*/

	hints = hints + "j,k: up/down | enter: connect/options | q: quit"
	hintsView := hintStyle.Render(hints)
	footer := lipgloss.JoinVertical(lipgloss.Left, dividerBorder(), hintsView)

	return footer
}
