package bluetooth

import (
	"fmt"
	"strings"

	"netui/config"

	"github.com/charmbracelet/lipgloss"
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
	if m.Scanning {
		segments = append(segments, lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true).Render("󱉶 scanning active ..."))
	}
	segments = append(segments, m.AdapterBlock())

	return lipgloss.JoinVertical(lipgloss.Left, segments...)
}

// AdapterBlock for displaying adapter info
func (m Model) AdapterBlock() string {
	powStatus := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).
		Render(fmt.Sprintf("power: %s", map[bool]string{true: "", false: ""}[m.Powered]))
	discStatus := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).
		Render(fmt.Sprintf("   discoverable: %s", map[bool]string{true: "", false: ""}[m.Discoverable]))

	pairStatus := lipgloss.NewStyle().Foreground(lipgloss.Color("4")).
		Render(fmt.Sprintf("   pairable: %s", map[bool]string{true: "", false: ""}[m.Pairable]))

	adapterBlock := lipgloss.JoinHorizontal(
		lipgloss.Center,
		powStatus,
		discStatus,
		pairStatus,
	)

	return adapterBlock
}

func (m Model) ScanningBlock() string {
	title := "Discovered devices\n"
	table := m.Table.View()

	return lipgloss.NewStyle().Render(
		lipgloss.JoinVertical(lipgloss.Left, title, table),
	)
}

func (m Model) SavedBlock() string {
	title := "󰆓 Known Paired Storage Devices\n"
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
