package bluetooth

import (
	"fmt"
	"strings"

	"corntui/config"

	"charm.land/lipgloss/v2"
)

func (m Model) View() string {
	if m.Err != nil {
		return fmt.Sprintf("\n  ❌ Bluetooth Interface Error: %v", m.Err)
	}

	var segments []string

	// 1. Conditional Interface Settings Block Rendering
	if m.Scanning {
		segments = append(segments, m.ScanningBlock())
	} else {
		segments = append(segments, m.SavedBlock())
	}

	// 2. Structural Inline Context Popups
	if m.UIState == StateActionsMenu {
		segments = append(segments, m.ActionsMenuBlock())
	}

	segments = append(segments, m.AdapterBlock())
	segments = append(segments, m.HintsBlock())

	// V2: Layout compositions use enum alignment methods
	return lipgloss.JoinVertical(lipgloss.Left, segments...)
}

// AdapterBlock for displaying adapter info
func (m Model) AdapterBlock() string {
	powStatus := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).
		Render(fmt.Sprintf("power: %s", map[bool]string{true: "", false: ""}[m.Adapter.Powered]))
	discStatus := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).
		Render(fmt.Sprintf("   discoverable: %s", map[bool]string{true: "", false: ""}[m.Adapter.Discoverable]))

	pairStatus := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).
		Render(fmt.Sprintf("   pairable: %s", map[bool]string{true: "", false: ""}[m.Adapter.Pairable]))

	scanStatus := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Italic(true).
		Render(fmt.Sprintf("   state: %s", map[bool]string{true: "discovering", false: "saved"}[m.Scanning]))

	// V2: Join strings using direct horizontal alignment methods
	adapterBlock := lipgloss.JoinHorizontal(
		lipgloss.Center,
		powStatus,
		discStatus,
		pairStatus,
		scanStatus,
	)

	return adapterBlock
}

func (m Model) ScanningBlock() string {
	title := config.Styles.Heading.Render("Discovered devices")
	table := m.Table.View()

	return lipgloss.NewStyle().Render(
		lipgloss.JoinVertical(lipgloss.Left, title, table),
	)
}

func (m Model) SavedBlock() string {
	title := config.Styles.Heading.Render("󰆓 Known Paired Storage Devices")
	table := m.Table.View()

	return lipgloss.NewStyle().Render(
		lipgloss.JoinVertical(lipgloss.Left, title, table),
	)
}

func (m Model) PasswordBlock() string {
	return "me"
}

func (m Model) ActionsMenuBlock() string {
	var menuLines []string
	menuLines = append(menuLines, lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("%s Options Menu", m.SelectedDev.Name)), "")

	for i, opt := range m.MenuOptions {
		if m.MenuCursor == i {
			menuLines = append(menuLines, config.Styles.HighlightText.Render(" > "+opt))
		} else {
			menuLines = append(menuLines, "   "+opt)
		}
	}
	return config.Styles.BoxStyle.Render(strings.Join(menuLines, "\n"))
}

func (m Model) HintsBlock() string {
	actionsHints := ""

	switch m.UIState {
	case StateActionsMenu:
		actionsHints = "j/k: nav | backspace: back "
	case StateNormal:
		actionsHints = "j/k: nav | p: power | d: discoverable | b: pairable"
	}

	hints := actionsHints + " | q: quit"

	return lipgloss.JoinVertical(
		lipgloss.Center,
		config.DividerBorder(),
		config.Styles.Hints.Render(hints),
	)
}
