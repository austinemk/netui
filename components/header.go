package components

import (
	"netui/config"

	"github.com/charmbracelet/lipgloss"
	//"github.com/common-nighthawk/go-figure"
)

// RenderHeader draws the main app banner and menu selections
func RenderHeader(activeTab int) string {
	//banner := figure.NewFigure("netui", "larry3d", true).String()
	title := config.Styles.Title.Render("NETUI") + "\n"

	var t1, t2, t3 string
	if activeTab == 0 {
		t1 = config.Styles.ActiveTab.Render("[1] Wi-Fi")
	} else {
		t1 = config.Styles.InactiveTab.Render("[1] Wi-Fi")
	}

	if activeTab == 1 {
		t2 = config.Styles.ActiveTab.Render("[2] Bluetooth")
	} else {
		t2 = config.Styles.InactiveTab.Render("[2] Bluetooth")
	}

	if activeTab == 2 {
		t3 = config.Styles.ActiveTab.Render("[3] VPN")
	} else {
		t3 = config.Styles.InactiveTab.Render("[3] VPN")
	}

	tabs := lipgloss.JoinHorizontal(lipgloss.Top, t1, t2, t3)
	//return title + tabs
	return lipgloss.JoinVertical(
		lipgloss.Left,
		title,
		tabs,
		dividerBorder(),
	)
}
