package vpn

import (
	"fmt"
	"strings"

	"linktui/config"

	"charm.land/lipgloss/v2"
)

func (m Model) View() string {
	if m.Client == nil {
		return config.Styles.LogBox.Render("nm Client not loaded yet")
	}

	var segments []string

	switch m.UIState {
	case StateAddForm:
		segments = append(segments, m.AddFormBlock())
	case StateImportFile:
		segments = append(segments, m.ImportFileBlock())
	default:
		segments = append(segments, m.TableBlock())
	}

	if m.IPInfo != nil {
		segments = append(segments, m.IPInfoBlock())
	}

	segments = append(segments, m.HintsBlock())

	if m.Err != nil {
		segments = append(segments, config.LogBlock(m.Err.Error()))
	}

	background := lipgloss.JoinVertical(lipgloss.Left, segments...)

	if m.UIState == StateActionsMenu && len(m.Tunnels) > 0 {
		return config.PlaceOverlay(background, m.OptionsPopupBlock())
	}

	return background
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
	menuLines = append(menuLines, fmt.Sprintf("── Actions: %s ──", target.Name))
	for i, opt := range options {
		if m.MenuCursor == i {
			menuLines = append(menuLines, fmt.Sprintf(" > %s", opt))
		} else {
			menuLines = append(menuLines, fmt.Sprintf("   %s", opt))
		}
	}

	return config.Styles.PopupStyle.Render(strings.Join(menuLines, "\n"))
}

func (m Model) IPInfoBlock() string {
	lines := []string{"IP:"}
	lines = append(lines, m.IPInfo.PublicIP+",")

	if m.IPInfo.ISP != "" {
		lines = append(lines, config.Truncate(m.IPInfo.ISP, 10)+",")
	}

	if m.IPInfo.City != "" {
		lines = append(lines, config.Truncate(m.IPInfo.City, 10)+",")
	}

	if m.IPInfo.Country != "" {
		lines = append(lines, config.Truncate(m.IPInfo.Country, 10))
	}

	return config.Styles.AdapterInfo.Render(strings.Join(lines, " "))
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
		actionsHints = "n: new | i: import | enter: actions | p: IPinfo"
	}

	hints := actionsHints + " | q: quit"

	return lipgloss.JoinVertical(
		lipgloss.Center,
		config.DividerBorder(),
		config.Styles.Hints.Render(hints),
	)
}
