package vpn

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

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

// FetchIPWithGeoCmd fetches the public IP and location in one shot.
// Only called when the user explicitly presses p.
func FetchIPWithGeoCmd(settleDelay time.Duration) tea.Cmd {
	return func() tea.Msg {
		if settleDelay > 0 {
			time.Sleep(settleDelay)
		}
		info := &IPInfo{}
		httpClient := &http.Client{Timeout: 5 * time.Second}

		// Public IP
		resp, err := httpClient.Get("https://api.ipify.org")
		if err == nil {
			defer resp.Body.Close()
			body, _ := io.ReadAll(resp.Body)
			info.PublicIP = strings.TrimSpace(string(body))
		}

		// Geo — ip-api.com is free, no key, 45 req/min
		if info.PublicIP != "" {
			resp2, err := httpClient.Get(fmt.Sprintf(
				"http://ip-api.com/json/%s?fields=country,regionName,city,isp,status",
				info.PublicIP,
			))
			if err == nil {
				defer resp2.Body.Close()
				var result struct {
					Status  string `json:"status"`
					Country string `json:"country"`
					Region  string `json:"regionName"`
					City    string `json:"city"`
					ISP     string `json:"isp"`
				}
				if json.NewDecoder(resp2.Body).Decode(&result) == nil && result.Status == "success" {
					info.Country = result.Country
					info.Region = result.Region
					info.City = result.City
					info.ISP = result.ISP
				}
			}
		}

		return IPInfoMsg(info)
	}
}
