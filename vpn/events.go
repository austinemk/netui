package vpn

import (
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	// State Intercept 1: Form Inputs Manual Mode Handling
	if m.UIState == StateAddForm {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
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
		}
		return m, nil
	}

	// State Intercept 2: File Picker Active Routing Navigation
	if m.UIState == StateImportFile {
		// Allow manual exit cleanly with escape key
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok && keyMsg.String() == "esc" {
			m.UIState = StateNormal
			return m, nil
		}

		// Update picker state FIRST so internal selection state is committed
		m.FilePicker, cmd = m.FilePicker.Update(msg)

		// THEN check if a file was selected — order matters, Update must run first
		if didSelect, selectedPath := m.FilePicker.DidSelectFile(msg); didSelect {
			m.UIState = StateNormal
			m.Loading = true
			return m, ImportWireGuardFileCmd(m.Client, selectedPath)
		}

		return m, cmd
	}

	// State Intercept 3: System Context Modal Menu Operations
	if m.UIState == StateActionsMenu {
		if keyMsg, ok := msg.(tea.KeyPressMsg); ok {
			switch keyMsg.String() {
			case "up", "k":
				if m.MenuCursor > 0 {
					m.MenuCursor--
				}
			case "down", "j":
				if m.MenuCursor < 1 {
					m.MenuCursor++
				}
			case "esc":
				m.UIState = StateNormal
			case "enter":
				if len(m.Tunnels) > 0 {
					targetTunnel := m.Tunnels[m.Table.Cursor()]
					if m.MenuCursor == 0 {
						cmd = ToggleTunnelCmd(m.Client, targetTunnel, !targetTunnel.Active)
					}
					m.UIState = StateNormal
					m.Loading = true
					return m, cmd
				}
			}
		}
		return m, nil
	}

	// Core Application Loop Lifecycle Messages
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Height = msg.Height
		m.FilePicker.AutoHeight = true
		return m, nil

	case TunnelsLoadedMsg:
		m.Tunnels = msg
		m.Loading = false

		var rows []table.Row
		for _, t := range m.Tunnels {
			status := "Inactive"
			if t.Active {
				status = "Active 󰌆"
			}
			rows = append(rows, table.Row{t.Name, t.Type, status})
		}
		m.Table.SetRows(rows)
		return m, nil

	case ActionSuccessMsg:
		return m, FetchTunnelsCmd(m.Client)

	case ErrMsg:
		m.Err = msg
		m.Loading = false
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "enter":
			if len(m.Tunnels) > 0 {
				m.MenuCursor = 0
				m.UIState = StateActionsMenu
			}
		case "n":
			m.UIState = StateAddForm
			m.ActiveField = FieldProfileName
			m.FormInputs = make(map[FormField]string)
		case "i":
			m.UIState = StateImportFile
		case "r":
			m.Loading = true
			return m, FetchTunnelsCmd(m.Client)
		}
	}

	m.Table, cmd = m.Table.Update(msg)
	return m, cmd
}
