package vpn

import (
	tea "charm.land/bubbletea/v2"
)

// --- CORE LIFECYCLE ---

// handleKeyPress to handle global keys
func (m Model) handleKeyPress(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	if !m.NMStatus {
		return m, nil
	}

	switch msg.String() {
	case "enter":
		if len(m.Tunnels) > 0 {
			m.MenuCursor = 0
			m.UIState = StateActionsMenu
		}
	case "n":
		if m.UIState == StateNormal {
			m.UIState = StateAddForm
			m.ActiveField = FieldProfileName
			m.FormInputs = make(map[FormField]string)
		} else {
			m.UIState = StateNormal
		}
	case "i":
		if m.UIState == StateNormal {
			m.UIState = StateImportFile //[cite: 1]
		} else {
			m.UIState = StateNormal
		}
	case "r": //[cite: 1]
		return m, FetchTunnelsCmd()
	case "p":
		return m, FetchIPWithGeoCmd(0)
	default:
		var cmd tea.Cmd
		m.Table, cmd = m.Table.Update(msg)
		m.Cursor = m.Table.Cursor()
		return m, cmd
	}

	return m, nil
}

// --- DATA TRANSFORMERS / HELPERS ---

// --- SUB-HANDLERS ---

func (m Model) handleFormState(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "esc":
		m.UIState = StateNormal
		return m, nil
	case "up":
		if m.ActiveField > FieldProfileName {
			m.ActiveField--
		}
	case "down":
		if m.ActiveField < FieldDone {
			m.ActiveField++
		} else {
			m.ActiveField = FieldProfileName
		}
	case "enter":
		if m.ActiveField == FieldDone {
			m.UIState = StateNormal
			return m, CreateWireGuardProfileCmd(m.FormInputs)
		}
		if m.ActiveField < FieldDone {
			m.ActiveField++
		}
	case "backspace":
		curr := m.FormInputs[m.ActiveField]
		if len(curr) > 0 {
			m.FormInputs[m.ActiveField] = curr[:len(curr)-1]
		}
	default:
		if len(keyMsg.String()) == 1 {
			m.FormInputs[m.ActiveField] += keyMsg.String()
		}
	}
	return m, nil
}

func (m Model) handleFilePickerState(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "esc" {
		m.UIState = StateNormal
		return m, nil
	}

	m.FilePicker, cmd = m.FilePicker.Update(msg)

	if didSelect, selectedPath := m.FilePicker.DidSelectFile(msg); didSelect {
		m.UIState = StateNormal
		return m, ImportWireGuardFileCmd(selectedPath)
	}

	return m, cmd
}

func (m Model) handleActionsMenuState(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	var cmd tea.Cmd
	switch keyMsg.String() {
	case "up", "k":
		if m.MenuCursor > 0 {
			m.MenuCursor--
		}
	case "down", "j":
		if m.MenuCursor < 2 { // Changed from 1 to 2 to allow scrolling down to Delete
			m.MenuCursor++
		}
	case "esc", "backspace":
		m.UIState = StateNormal
	case "enter":
		if len(m.Tunnels) > 0 {
			targetTunnel := m.Tunnels[m.Table.Cursor()]

			switch m.MenuCursor {
			case 0: // Toggle Activation State
				cmd = ToggleTunnelCmd(targetTunnel, !targetTunnel.Active)
			case 1: // Delete Profile State
				cmd = DeleteTunnelCmd(targetTunnel)
			}
			m.UIState = StateNormal
			return m, cmd
		}
	}
	return m, nil
}
