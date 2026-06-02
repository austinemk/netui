package vpn

import (
	"fmt"
	"strings"

	"corntui/config"

	"charm.land/lipgloss/v2"
)

func (m Model) View() string {
	if m.Loading {
		return "\n  🔄 Querying Active secure interfaces via System Bus..."
	}
	if m.Err != nil {
		return fmt.Sprintf("\n  ❌ System Failure Hook: %v", m.Err)
	}

	var segments []string

	// 1. Render table
	segments = append(segments, m.TableBlock())

	// 2. Render Manual Config Input UI View Block
	if m.UIState == StateAddForm {
		segments = append(segments, m.AddFormBlock())
	}

	// 3. Render Native Bubble File Picker Frame View
	if m.UIState == StateImportFile {
		segments = append(segments, m.ImportFileBlock())
	}

	// 4. Modal Window Context Menu Overlays
	if m.UIState == StateActionsMenu {
		if len(m.Tunnels) > 0 {
			segments = append(segments, m.OptionsPopupBlock())
		}
	}

	// 5. Hints
	segments = append(segments, m.HintsBlock())

	return lipgloss.JoinVertical(lipgloss.Left, segments...)
}

func (m Model) AddFormBlock() string {
	var fLines []string
	fLines = append(fLines, config.Styles.Heading.Render("Create WireGuard Interface Profile"))

	fields := []struct {
		id    FormField
		label string
	}{
		{FieldProfileName, "Connection Name : "},
		{FieldInterfaceName, "Interface (wg0) : "},
		{FieldPrivateKey, "Private Key     : "},
		{FieldPeerEndpoint, "Peer Endpoint   : "},
		{FieldPeerPublicKey, "Peer Public Key : "},
	}

	for _, f := range fields {
		cursorPrefix := "  "
		if m.ActiveField == f.id {
			cursorPrefix = "> "
		}
		fLines = append(fLines, fmt.Sprintf("%s%s%s", cursorPrefix, f.label, m.FormInputs[f.id]))
	}

	doneRow := "  [ SUBMIT AND REGISTER ]"
	if m.ActiveField == FieldDone {
		doneRow = "> [ SUBMIT AND REGISTER ]"
	}
	fLines = append(fLines, "\n"+doneRow)
	return lipgloss.JoinVertical(lipgloss.Left, fLines...)
}

func (m Model) ImportFileBlock() string {
	var iLines []string
	iLines = append(iLines, config.Styles.Heading.Render("Select a File"))
	iLines = append(iLines, m.FilePicker.View())
	return lipgloss.JoinVertical(lipgloss.Left, iLines...)
}

func (m Model) TableBlock() string {
	var sections []string
	sections = append(sections, config.Styles.Heading.Render("WireGuard List"))
	sections = append(sections, m.Table.View())
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) OptionsPopupBlock() string {
	target := m.Tunnels[m.Table.Cursor()]
	actLabel := "Activate Link"
	if target.Active {
		actLabel = "Deactivate Link"
	}
	options := []string{actLabel, "Delete Profile"}

	var menuLines []string
	menuLines = append(menuLines, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#EF4444")).Render(fmt.Sprintf("── Actions: %s ──", target.Name)))
	for i, opt := range options {
		if m.MenuCursor == i {
			menuLines = append(menuLines, fmt.Sprintf(" > %s", opt))
		} else {
			menuLines = append(menuLines, fmt.Sprintf("   %s", opt))
		}
	}

	return config.Styles.BoxStyle.Render(strings.Join(menuLines, "\n"))
}

func (m Model) HintsBlock() string {
	actionsHints := ""

	switch m.UIState {
	case StateAddForm:
		actionsHints = "esc: back |  | enter: submit"
	case StateActionsMenu:
		actionsHints = "󰹹: nav | backspace: back "
	case StateImportFile:
		actionsHints = "left/right: nav | backspace: back"
	case StateNormal:
		actionsHints = "n: new | i: import | enter: actions | r: refresh"
	}

	hints := actionsHints + " | q: quit"

	return lipgloss.JoinVertical(
		lipgloss.Center,
		config.DividerBorder(),
		config.Styles.Hints.Render(hints),
	)
}
