package bluetooth

import (
	"fmt"

	"github.com/austinemk/linktui/pkg/config"

	"charm.land/lipgloss/v2"
)

func (m Model) View() string {
	if !m.BluezStatus {
		return config.Styles.LogBox.Render("bluez service is not responding. Ensure 'bluez' is installed and bluetooth is running")
	}

	var segments []string

	if m.Scanning {
		segments = append(segments, m.ScanningBlock())
	} else {
		segments = append(segments, m.SavedBlock())
	}

	segments = append(segments, m.AdapterBlock())
	segments = append(segments, m.HintsBlock())

	if m.Err != nil {
		segments = append(segments, config.LogBlock(m.Err.Error()))
	}

	background := lipgloss.JoinVertical(lipgloss.Left, segments...)

	if m.UIState == StateActionsMenu {
		return config.PlaceOverlay(background, m.ActionsMenuBlock())
	}
	if m.UIState == StatePasskeyPrompt {
		return config.PlaceOverlay(background, m.PasskeyPromptBlock())
	}

	return background
}

// AdapterBlock for displaying adapter info
func (m Model) AdapterBlock() string {
	lines := []string{fmt.Sprintf("power: %s", map[bool]string{true: "", false: ""}[m.Adapter.Powered])}
	lines = append(lines, fmt.Sprintf("   discoverable: %s", map[bool]string{true: "", false: ""}[m.Adapter.Discoverable]))

	lines = append(lines, fmt.Sprintf("   pairable: %s", map[bool]string{true: "", false: ""}[m.Adapter.Pairable]))

	lines = append(lines, fmt.Sprintf("   state: %s", map[bool]string{true: "discovering", false: "saved"}[m.Scanning]))

	return config.Styles.AdapterInfo.Render(lipgloss.JoinHorizontal(lipgloss.Center, lines...))
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

func (m Model) ActionsMenuBlock() string {
	var menuLines []string
	//menuLines = append(menuLines, lipgloss.NewStyle().Bold(true).Render(fmt.Sprintf("%s Options Menu", m.SelectedDev.Name)), "")

	for i, opt := range m.MenuOptions {
		if m.MenuCursor == i {
			menuLines = append(menuLines, config.Styles.HighlightText.Render(opt))
		} else {
			menuLines = append(menuLines, " "+opt+" ")
		}
	}

	return config.Styles.PopupStyle.Render(
		lipgloss.JoinVertical(
			lipgloss.Center,
			lipgloss.JoinVertical(lipgloss.Center, menuLines...),
			config.Styles.Hints.Render("\n\n esc/backspace: back up/down: nav"),
		),
	)
}

func (m Model) HintsBlock() string {
	actionsHints := ""

	switch m.UIState {
	case StateActionsMenu:
		actionsHints = "j/k: nav | backspace: back "
	case StateNormal:
		actionsHints = "j/k:nav | p:power | d:discoverable | b:pairable | s:discover"
	}

	return lipgloss.JoinVertical(
		lipgloss.Center,
		config.DividerBorder(),
		config.Styles.Hints.Render(actionsHints),
	)
}

// PasskeyPromptBlock builds a clean terminal layout block for the confirmation window
func (m Model) PasskeyPromptBlock() string {
	promptText := fmt.Sprintf("Pairing request from %s\nConfirm Passkey: %06d?", m.SelectedDev.Name, m.CurrentPasskey)

	var yesOpt, noOpt string

	// Apply active highlights depending on current selection position
	if m.MenuCursor == 0 {
		yesOpt = lipgloss.NewStyle().Background(lipgloss.Color("2")).Foreground(lipgloss.Color("15")).Bold(true).Render(" [ YES ] ")
		noOpt = lipgloss.NewStyle().Render("  No  ")
	} else {
		yesOpt = lipgloss.NewStyle().Render("  Yes  ")
		noOpt = lipgloss.NewStyle().Background(lipgloss.Color("9")).Foreground(lipgloss.Color("15")).Bold(true).Render(" [ NO ] ")
	}

	buttons := lipgloss.JoinHorizontal(lipgloss.Center, yesOpt, "    ", noOpt)

	popupContent := lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.NewStyle().Bold(true).Render(promptText),
		"",
		buttons,
	)

	// Wrap inside your structural config BoxStyle
	return config.Styles.PopupStyle.Render(popupContent)
}
