package wifi

import (
	"fmt"
	"strings"

	"netui/config"

	"charm.land/lipgloss/v2"
)

func (m Model) View() string {
	if m.Loading {
		return "Connecting to System Bus Interfaces..."
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

	segments = append(segments, m.adapterBlock())

	// V2: Join lines now use alignment positioning directly
	return lipgloss.JoinVertical(lipgloss.Left, segments...)
}

func (m Model) adapterBlock() string {
	linkStat := false
	if m.Adapter.State == "Connected" {
		linkStat = true
	}
	intface := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).
		Render(fmt.Sprintf("device: %s", m.Adapter.Interface))
	connected := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).
		Render(fmt.Sprintf("  connected: %s", map[bool]string{true: "", false: ""}[linkStat]))
	power := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).
		Render(fmt.Sprintf("  power: %s", map[bool]string{true: "󰤨 ", false: "󰤭 "}[m.Adapter.Enabled]))
	scan := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Italic(true).
		Render(fmt.Sprintf("  status: %s", map[bool]string{true: "scanning", false: "saved"}[m.Scanning]))

	// V2: Horizontally combine block lines using Alignment
	return lipgloss.NewStyle().Render(
		lipgloss.JoinHorizontal(
			lipgloss.Center,
			intface,
			connected,
			power,
			scan,
		),
	)
}

func (m Model) ScanningBlock() string {
	title := config.Styles.Heading.Render("Nearby Access Points")
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
	title := config.Styles.Heading.Render("󰆓 Saved networks")
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
