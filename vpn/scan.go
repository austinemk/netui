package vpn

import (
	tea "charm.land/bubbletea/v2"
)

// Helper command to fetch fresh tunnel state profiles asynchronously
func FetchTunnelsCmd(client *DBusClient) tea.Cmd {
	return func() tea.Msg {
		t, err := GetVPNConnections(client)
		if err != nil {
			return ErrMsg(err)
		}
		return TunnelsLoadedMsg(t)
	}
}

func ToggleTunnelCmd(client *DBusClient, tunnel TunnelProfile, activate bool) tea.Cmd {
	return func() tea.Msg {
		if activate {
			// Find an available specific device if needed, or send nil for generic setups
			_, err := client.NM.ActivateConnection(tunnel.Connection, nil, nil)
			if err != nil {
				return ErrMsg(err)
			}
		} else {
			activeConns, err := client.NM.GetPropertyActiveConnections()
			if err == nil {
				for _, aConn := range activeConns {
					uuid, _ := aConn.GetPropertyUUID()
					if uuid == tunnel.UUID {
						err = client.NM.DeactivateConnection(aConn)
						if err != nil {
							return ErrMsg(err)
						}
						break
					}
				}
			}
		}
		return ActionSuccessMsg("VPN Activation/Deactivation State updated!")
	}
}

// m.Update(msg tea.Msg) can remain completely unchanged!

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// State Intercept 1: Form Inputs Logic Management Loop
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

	// State Intercept 2: Normal Actions Dialog Panel
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
				targetTunnel := m.Tunnels[m.Cursor]
				var cmd tea.Cmd
				if m.MenuCursor == 0 {
					cmd = ToggleTunnelCmd(m.Client, targetTunnel, !targetTunnel.Active)
				}
				m.UIState = StateNormal
				m.Loading = true
				return m, cmd
			}
		}
		return m, nil
	}

	// Standard Dashboard Loop Processing
	switch msg := msg.(type) {
	case TunnelsLoadedMsg:
		m.Tunnels = msg
		m.Loading = false // Breaks out of the querying freeze cleanly
		return m, nil

	case ActionSuccessMsg:
		return m, FetchTunnelsCmd(m.Client)

	case ErrMsg:
		m.Err = msg
		m.Loading = false
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.Tunnels)-1 {
				m.Cursor++
			}
		case "enter":
			if len(m.Tunnels) > 0 {
				m.MenuCursor = 0
				m.UIState = StateActionsMenu
			}
		case "n":
			m.UIState = StateAddForm
			m.ActiveField = FieldProfileName
			m.FormInputs = make(map[FormField]string)
		case "r":
			m.Loading = true
			return m, FetchTunnelsCmd(m.Client)
		}
	}

	return m, nil
}
