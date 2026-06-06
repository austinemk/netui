package app

import (
	"github.com/austinemk/linktui/pkg/config"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func (m *AppModel) View() tea.View {
	if m.SizeError != "" {
		boxStyle := lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("#EF4444")).Padding(2, 4)
		v := tea.NewView(boxStyle.Render(m.SizeError))
		v.AltScreen = true
		return v
	}

	header := config.RenderHeader(int(m.ActiveTab))

	var body string
	if !m.BusReady {
		body = config.Styles.LogBox.Render("D-Bus client not loaded yet")
	} else {
		if !m.LoadedTabs[m.ActiveTab] {
			body = "\n  Loading interface details..."
		} else {
			switch m.ActiveTab {
			case WifiTab:
				body = m.WifiView.View()
			case BluetoothTab:
				body = m.BtView.View()
			case VpnTab:
				body = m.VpnView.View()
			}
		}
	}

	var logView string
	if m.LogMessage != "" {
		logView = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Render("\n[LOG] " + m.LogMessage)
	}

	mainLayout := lipgloss.JoinVertical(lipgloss.Left, header, body, logView)
	mainLayout = config.Styles.Container.Render(mainLayout)
	v := tea.NewView(mainLayout)
	v.AltScreen = true
	return v
}
