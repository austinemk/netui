package vpn

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/Wifx/gonetworkmanager/v3"
)

func FetchTunnelsCmd(client gonetworkmanager.NetworkManager) tea.Cmd {
	return func() tea.Msg {
		t, err := GetVPNConnections(client)
		if err != nil {
			return ErrMsg(err)
		}
		return TunnelsLoadedMsg(TunnelsLoadedData{Tunnels: t, Client: client})
	}
}

func ToggleTunnelCmd(client gonetworkmanager.NetworkManager, tunnel TunnelProfile, activate bool) tea.Cmd {
	return func() tea.Msg {
		if activate {
			_, err := client.ActivateConnection(tunnel.Connection, nil, nil)
			if err != nil {
				return ErrMsg(err)
			}
		} else {
			activeConns, err := client.GetPropertyActiveConnections()
			if err == nil {
				for _, aConn := range activeConns {
					uuid, _ := aConn.GetPropertyUUID()
					if uuid == tunnel.UUID {
						err = client.DeactivateConnection(aConn)
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

func DeleteTunnelCmd(tunnel TunnelProfile) tea.Cmd {
	return func() tea.Msg {
		// Ensure connection object exists before calling methods on it
		if tunnel.Connection == nil {
			return ErrMsg(fmt.Errorf("cannot delete: connection reference is missing"))
		}

		err := tunnel.Connection.Delete()
		if err != nil {
			return ErrMsg(fmt.Errorf("failed to delete profile: %v", err))
		}

		return ActionSuccessMsg("WireGuard Profile deleted successfully!")
	}
}
