package config

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// RenderHeader draws the main app banner and menu selections
func RenderHeader(activeTab int) string {
	figure := `  |    _ _|   \ |  |  / 
  |      |   .  |  . <  
 ____| ___| _|\_| _|\_\ `

	title := Styles.Title.Render(figure)

	var tabs []string
	if activeTab == 0 {
		tabs = append(tabs, Styles.ActiveTab.Render("wifi"))
	} else {
		tabs = append(tabs, Styles.InactiveTab.Render("wifi"))
	}

	if activeTab == 1 {
		tabs = append(tabs, Styles.ActiveTab.Render("bluetooth"))
	} else {
		tabs = append(tabs, Styles.InactiveTab.Render("bluetooth"))
	}

	if activeTab == 2 {
		tabs = append(tabs, Styles.ActiveTab.Render("vpn"))
	} else {
		tabs = append(tabs, Styles.InactiveTab.Render("vpn"))
	}

	spacing := " "
	for i := 1; i < HeaderSpacing; i++ {
		spacing = spacing + " "
	}

	// V2 Change: Horizontally align using layout position method
	tabBox := Styles.TabsBox.Render(strings.Join(tabs, " | "))

	// V2 Change: Vertically align layers using layout position method
	return lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.JoinHorizontal(lipgloss.Bottom, title, tabBox),
		DividerBorder(),
	)
}

func DividerBorder() string {
	if ColorDivider != "" {
		return lipgloss.NewStyle().Foreground(lipgloss.Color(ColorDivider)).Render(strings.Repeat("-", WindowWidth-2))
	}
	return ""
}

func LogBlock(content string) string {
	truncated := Truncate(content, WindowWidth-6)

	return Styles.LogBox.Render(truncated)
}

// PlaceOverlay renders `popup` centered on top of `bg` using Lip Gloss v2.
func PlaceOverlay(bg, popup string) string {
	bgW, bgH := lipgloss.Width(bg), lipgloss.Height(bg)
	popupW, popupH := lipgloss.Width(popup), lipgloss.Height(popup)

	// Center math
	startX := (bgW - popupW) / 2
	startY := (bgH - popupH) / 2
	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}

	// 1. Create the popup layer first
	popupLayer := lipgloss.NewLayer(popup).
		X(startX).
		Y(startY).
		Z(1)

	// 2. Create the background layer standalone (pass nil or leave children empty)
	bgLayer := lipgloss.NewLayer(bg).X(0).Y(0)

	// 3. Explicitly attach the child layer to avoid the variadic unpacking panic
	bgLayer.AddLayers(popupLayer)

	// 4. Render using the root background container
	comp := lipgloss.NewCompositor(bgLayer)

	return comp.Render()
}
