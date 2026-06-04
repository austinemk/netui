package main

import (
	"strings"

	"corntui/config"

	"charm.land/lipgloss/v2"
)

// RenderHeader draws the main app banner and menu selections
func RenderHeader(activeTab int) string {
	figure := `   __|   _ \  _ \   \ | 
  (     (   |   /  .  | 
 \___| \___/ _|_\ _|\_| `

	title := config.Styles.Title.Render(figure)

	var tabs []string
	if activeTab == 0 {
		tabs = append(tabs, config.Styles.ActiveTab.Render("wifi"))
	} else {
		tabs = append(tabs, config.Styles.InactiveTab.Render("wifi"))
	}

	if activeTab == 1 {
		tabs = append(tabs, config.Styles.ActiveTab.Render("bluetooth"))
	} else {
		tabs = append(tabs, config.Styles.InactiveTab.Render("bluetooth"))
	}

	if activeTab == 2 {
		tabs = append(tabs, config.Styles.ActiveTab.Render("vpn"))
	} else {
		tabs = append(tabs, config.Styles.InactiveTab.Render("vpn"))
	}

	spacing := " "
	for i := 1; i < config.HeaderSpacing; i++ {
		spacing = spacing + " "
	}

	// V2 Change: Horizontally align using layout position method
	tabBox := config.Styles.TabsBox.Render(strings.Join(tabs, " | "))

	// V2 Change: Vertically align layers using layout position method
	return lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Bottom, title, tabBox),
		config.DividerBorder(),
	)
}
