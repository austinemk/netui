package config

import (
	"charm.land/lipgloss/v2"
)

func DividerBorder() string {
	divider := ""
	for i := 1; i < WindowWidth; i++ {
		divider = divider + "-"
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(divider)
}

func LogBlock(content string) string {
	truncated := Truncate(content, TabBodyWidth-4)

	return Styles.LogBox.Render(truncated)
}
