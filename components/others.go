package components

import (
	"netui/config"

	"charm.land/lipgloss/v2"
)

func dividerBorder() string {
	divider := ""
	for i := 1; i < config.WindowWidth; i++ {
		divider = divider + "-"
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Render(divider)
}

/*func RenderOptionsPopup(title string, options []string, cursor int) string {
	var menulines []string
	menulines = append(menulines, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15")).Italic(true).Render(fmt.Sprintf("%s options \n\n", title)))
	//boxStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("8")).Padding(1, 2)

	for i, opt := range options {
		if cursor == i {
			menulines = append(menulines, config.Styles.HighlightText.Render(opt))
		} else {
			menulines = append(menulines, opt)
		}
	}

	return config.Styles.PopupBoxStyle.Render(lipgloss.Place(
		config.PopupWidth,
		config.PopupHeight,
		config.PopupHpos,
		config.PopupVpos,
		strings.Join(menulines, "\n"),
	))
}*/
