package main

import (
	"corntui/config"

	"charm.land/lipgloss/v2"
	//"github.com/common-nighthawk/go-figure"
)

// RenderHeader draws the main app banner and menu selections
func RenderHeader(activeTab int) string {
	//banner := figure.NewFigure("netui", "larry3d", true).String()
	//title := config.Styles.Title.Render("NETUI") + "\n"

	var t1, t2, t3 string
	if activeTab == 0 {
		t1 = config.Styles.ActiveTab.Render("WiFi")
	} else {
		t1 = config.Styles.InactiveTab.Render("WiFi")
	}

	if activeTab == 1 {
		t2 = config.Styles.ActiveTab.Render("BlUETOOTH")
	} else {
		t2 = config.Styles.InactiveTab.Render("BlUETOOTH")
	}

	if activeTab == 2 {
		t3 = config.Styles.ActiveTab.Render("VPN")
	} else {
		t3 = config.Styles.InactiveTab.Render("VPN")
	}

	divider := "|"

	spacing := " "
	for i := 1; i < config.HeaderSpacing; i++ {
		spacing = spacing + " "
	}

	// V2 Change: Horizontally align using layout position method
	tabs := config.Styles.TabsBox.Render(lipgloss.JoinHorizontal(lipgloss.Center,
		t1, spacing, divider, spacing, t2, spacing, divider, spacing, t3))

	// V2 Change: Vertically align layers using layout position method
	return lipgloss.JoinVertical(
		lipgloss.Left,
		// title,
		"\n",
		tabs,
		config.DividerBorder(),
	)
}
