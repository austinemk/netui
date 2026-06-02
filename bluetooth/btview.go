package bluetooth

import (
	"fmt"
	"strings"

	"corntui/config"

	"charm.land/lipgloss/v2"
)

func (m Model) View() string {
	if m.Client == nil {
		return "Bluez client is nil"
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

	// Add the Passkey Prompt popup view to your layout stack
	if m.UIState == StatePasskeyPrompt {
		segments = append(segments, m.PasskeyPromptBlock())
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

// PasskeyPromptBlock builds a clean terminal layout block for the confirmation window
// PasskeyPromptBlock builds a clean terminal layout block for the confirmation window
func (m Model) PasskeyPromptBlock() string {
	// 👇 FIX: Swap hardcoded '123456' with your live state property 'm.CurrentPasskey'
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
	return config.Styles.BoxStyle.Render(popupContent)
}
