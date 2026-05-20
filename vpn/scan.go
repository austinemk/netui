package vpn

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/godbus/dbus/v5"
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
		nm := client.Conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")

		if activate {
			var activeConn dbus.ObjectPath
			err := nm.Call("org.freedesktop.NetworkManager.ActivateConnection", 0, tunnel.Path, dbus.ObjectPath("/"), dbus.ObjectPath("/")).Store(&activeConn)
			if err != nil {
				return ErrMsg(err)
			}
		} else {
			var activePaths []dbus.ObjectPath
			activeConnsVal, err := nm.GetProperty("org.freedesktop.NetworkManager.ActiveConnections")
			if err == nil {
				activePaths, _ = activeConnsVal.Value().([]dbus.ObjectPath)
				for _, aPath := range activePaths {
					aObj := client.Conn.Object("org.freedesktop.NetworkManager", aPath)
					uuidProp, _ := aObj.GetProperty("org.freedesktop.NetworkManager.Connection.Active.Uuid")
					if uStr, ok := uuidProp.Value().(string); ok && uStr == tunnel.UUID {
						_ = nm.Call("org.freedesktop.NetworkManager.DeactivateConnection", 0, aPath)
						break
					}
				}
			}
		}
		return ActionSuccessMsg("State change completed")
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// State Intercept 1: Form Wizard Processing
	if m.UIState == StateAddForm {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "esc":
				m.UIState = StateNormal
			case "tab", "down", "j":
				if m.ActiveField < FieldDone {
					m.ActiveField++
				}
			case "shift+tab", "up", "k":
				if m.ActiveField > FieldProfileName {
					m.ActiveField--
				}
			case "backspace":
				currInput := m.FormInputs[m.ActiveField]
				if len(currInput) > 0 {
					m.FormInputs[m.ActiveField] = currInput[:len(currInput)-1]
				}
			case "enter":
				if m.ActiveField == FieldDone {
					// FIXED: Execute creation, set loading, and explicitly chain the D-Bus re-fetch command
					m.UIState = StateNormal
					m.Loading = true
					return m, tea.Batch(AddWireguardProfileCmd(m.Client, m.FormInputs), FetchTunnelsCmd(m.Client))
				}
				m.ActiveField++
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
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
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
		// FIXED: Call the structured D-Bus collector explicitly instead of returning a naked nested callback
		return m, FetchTunnelsCmd(m.Client)

	case ErrMsg:
		m.Err = msg
		m.Loading = false
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "a":
			m.UIState = StateAddForm
			m.ActiveField = FieldProfileName
			m.FormInputs = make(map[FormField]string)
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
				m.UIState = StateActionsMenu
				m.MenuCursor = 0
			}
		}
	}
	return m, nil
}
