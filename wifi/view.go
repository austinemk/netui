package wifi

import (
	"fmt"
	"strings"

	"netui/config"

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
		segments = append(segments, m.ScanningBlock())
	} else {
		segments = append(segments, m.SavedBlock())
	}

	// 3. Popup Overlay Processing
	if m.UIState == StatePasswordInput {
		segments = append(segments, m.PasswordBlock())
	}

	if m.UIState == StateSavedActionsMenu {
		segments = append(segments, m.OptionsBlock())
	}
	if m.Scanning {
		segments = append(segments, lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true).Render(" scanning active"))
	}
	segments = append(segments, m.adapterBlock())

	return lipgloss.JoinVertical(lipgloss.Left, segments...)
}

func (m Model) adapterBlock() string {
	linkStat := false
	if m.Adapter.State == "Connected" {
		linkStat = true
	}
	intface := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).
		Render(fmt.Sprintf("interface: %s", m.Adapter.Interface))
	connected := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).
		Render(fmt.Sprintf("  connected: %s", map[bool]string{true: "", false: ""}[linkStat]))
	power := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).
		Render(fmt.Sprintf("  power: %s", map[bool]string{true: "󰤨  on", false: "󰤭  off"}[m.Adapter.Enabled]))

	return lipgloss.NewStyle().Render(
		lipgloss.JoinHorizontal(
			lipgloss.Left,
			intface,
			connected,
			power,
		),
	)
}

func (m Model) ScanningBlock() string {
	title := "Nearby Access Points"
	table := m.Table.View()
	return lipgloss.NewStyle().Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			table,
		),
	)
}

func (m Model) SavedBlock() string {
	title := "󰆓 Saved networks\n"
	table := m.Table.View()

	return lipgloss.NewStyle().Render(
		lipgloss.JoinVertical(
			lipgloss.Left,
			title,
			table,
		),
	)
}

func (m Model) PasswordBlock() string {
	passContent := fmt.Sprintf("Enter Password for: %s\n\n %s", m.SelectedAP.SSID, m.PassInput.View())
	return config.Styles.BoxStyle.Render(passContent)
}

func (m Model) OptionsBlock() string {
	options := []string{"autoconnect/off", "forget"}
	var menuLines []string
	menuLines = append(menuLines, lipgloss.NewStyle().Render(fmt.Sprintf("%s options", m.SelectedSaved.Name)))
	for i, opt := range options {
		if m.MenuCursor == i {
			menuLines = append(menuLines, config.Styles.HighlightText.Render(opt))
		} else {
			menuLines = append(menuLines, opt)
		}
	}
	popup := config.Styles.BoxStyle.Render(strings.Join(menuLines, "\n"))
	return popup
}
