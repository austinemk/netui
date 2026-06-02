package vpn

import (
	"math"

	"corntui/config"

	tea "charm.land/bubbletea/v2"
)

// --- CORE LIFECYCLE ---

func (m Model) handleCoreLifecycle(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case TunnelsLoadedMsg:
		m.Tunnels = msg.Tunnels
		m.Client = msg.Client
		m.Loading = false

		// Map backend tunnels data cleanly to the UI table rows
		m.syncTableRows()
		return m, nil

	case ActionSuccessMsg:
		return m, FetchTunnelsCmd(m.Client)

	case ErrMsg:
		m.Err = msg
		m.Loading = false
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)

	}

	m.Table, cmd = m.Table.Update(msg)
	return m, cmd
}

// handleKeyPress to handle global keys
func (m Model) handleKeyPress(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if len(m.Tunnels) > 0 { //[cite: 1]
			m.MenuCursor = 0
			m.Table.SetHeight(int(math.Floor(config.TabBodyHeight * 0.4)))
			m.UIState = StateActionsMenu //[cite: 1]
		}
	case "n":
		m.Table.SetHeight(int(math.Floor(config.TabBodyHeight * 0.3)))
		m.UIState = StateAddForm                  //[cite: 1]
		m.ActiveField = FieldProfileName          //[cite: 1]
		m.FormInputs = make(map[FormField]string) //[cite: 1]
	case "i":
		m.Table.SetHeight(int(math.Floor(config.TabBodyHeight * 0.3)))
		m.UIState = StateImportFile //[cite: 1]
	case "r":
		m.Loading = true                    //[cite: 1]
		return m, FetchTunnelsCmd(m.Client) //[cite: 1]
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
		m.Table.SetHeight(int(math.Floor(config.TabBodyHeight * 0.8)))
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
			m.Table.SetHeight(int(math.Floor(config.TabBodyHeight * 0.85)))
			m.UIState = StateNormal
			m.Loading = true
			return m, CreateWireGuardProfileCmd(m.Client, m.FormInputs)
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
		m.Table.SetHeight(int(math.Floor(config.TabBodyHeight * 0.8)))
		m.UIState = StateNormal
		return m, nil
	}

	m.FilePicker, cmd = m.FilePicker.Update(msg)

	if didSelect, selectedPath := m.FilePicker.DidSelectFile(msg); didSelect {
		m.UIState = StateNormal
		m.Loading = true
		return m, ImportWireGuardFileCmd(m.Client, selectedPath)
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
	case "esc":
		m.Table.SetHeight(int(math.Floor(config.TabBodyHeight * 0.85)))
		m.UIState = StateNormal
	case "enter":
		if len(m.Tunnels) > 0 {
			targetTunnel := m.Tunnels[m.Table.Cursor()]

			switch m.MenuCursor {
			case 0: // Toggle Activation State
				cmd = ToggleTunnelCmd(m.Client, targetTunnel, !targetTunnel.Active)
			case 1: // Delete Profile State
				cmd = DeleteTunnelCmd(m.Client, targetTunnel)
			}
			m.Table.SetHeight(int(math.Floor(config.TabBodyHeight * 0.85)))
			m.UIState = StateNormal
			m.Loading = true
			return m, cmd
		}
	}
	return m, nil
}
