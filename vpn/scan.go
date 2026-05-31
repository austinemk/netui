package vpn

import tea "charm.land/bubbletea/v2"

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
