package wifi

import (
	"fmt"
	"strings"

	"netui/components"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.Loading {
		return "\n Connecting to System Bus Interfaces..."
	}
	if m.Err != nil {
		return fmt.Sprintf("\n  ❌ Error: %v", m.Err)
	}

	var segments []string

	// 2. Conditional Interface Block Rendering
	if m.Scanning {
		apBlock := "\n Nearby Access Points\n\n" + m.Table.View() + "\n"
		segments = append(segments, lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(apBlock))
		segments = append(segments, lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true).Render(" scanning active"))

	} else {
		adapterBlock := fmt.Sprintf(
			"\n  Settings\n  Interface:    %s\n  Link Status:  %s\n  Power:  %s [p: switch]\n",
			m.Adapter.Interface, m.Adapter.State, map[bool]string{true: "󰤨  on", false: "󰤭  off"}[m.Adapter.Enabled],
		)
		segments = append(segments, lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render(adapterBlock))

		savedBlock := "󰆓 Saved networks\n" + m.Table.View()
		segments = append(segments, lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Render(savedBlock))
	}

	screen := lipgloss.JoinVertical(lipgloss.Left, segments...)

	// 3. Popup Overlay Processing
	if m.UIState == StatePasswordInput {
		hiddenPassword := strings.Repeat("*", len(m.PasswordInput))
		box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#EF4444")).Padding(1, 3).Margin(1, 2)
		popup := box.Render(fmt.Sprintf("Enter Password for: %s\n\n %s_", m.SelectedAP.SSID, hiddenPassword))
		return lipgloss.JoinVertical(lipgloss.Center, screen, popup)
	}

	if m.UIState == StateSavedActionsMenu {
		options := []string{"autoconnect/off", "forget"}
		popup := components.RenderOptionsPopup(m.SelectedSaved.Name, options, m.MenuCursor)
		return popup
	}

	return screen
}
