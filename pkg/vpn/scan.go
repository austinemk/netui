package vpn

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/austinemk/linktui/pkg/bus"

	tea "charm.land/bubbletea/v2"
	"github.com/godbus/dbus/v5"
)

const (
	nmDest          = "org.freedesktop.NetworkManager"
	nmPath          = "/org/freedesktop/NetworkManager"
	nmIface         = "org.freedesktop.NetworkManager"
	nmConnIface     = "org.freedesktop.NetworkManager.Connection.Active"
	nmSettingsIface = "org.freedesktop.NetworkManager.Settings.Connection"
	nmPropIface     = "org.freedesktop.DBus.Properties"
)

func FetchTunnelsCmd() tea.Cmd {
	return func() tea.Msg {
		t, err := GetVPNConnections()
		if err != nil {
			return ErrMsg(err)
		}
		return TunnelsLoadedMsg(TunnelsLoadedData{Tunnels: t})
	}
}

func ToggleTunnelCmd(tunnel TunnelProfile, activate bool) tea.Cmd {
	return func() tea.Msg {
		conn := bus.Get()
		nm := conn.Object(nmDest, nmPath)

		if activate {
			call := nm.Call(
				nmIface+".ActivateConnection", 0,
				tunnel.ConnectionPath,
				dbus.ObjectPath("/"),
				dbus.ObjectPath("/"),
			)
			if call.Err != nil {
				return ErrMsg(call.Err)
			}
		} else {
			// Find the active connection with matching UUID and deactivate it
			var activeConns []dbus.ObjectPath
			if err := nm.Call(nmPropIface+".Get", 0, nmIface, "ActiveConnections").Store(&activeConns); err != nil {
				// try direct property
				v, err := nm.GetProperty(nmIface + ".ActiveConnections")
				if err != nil {
					return ErrMsg(err)
				}
				activeConns, _ = v.Value().([]dbus.ObjectPath)
			}

			for _, aPath := range activeConns {
				aObj := conn.Object(nmDest, aPath)
				v, err := aObj.GetProperty(nmConnIface + ".Uuid")
				if err != nil {
					continue
				}
				uuid, _ := v.Value().(string)
				if uuid == tunnel.UUID {
					call := nm.Call(nmIface+".DeactivateConnection", 0, aPath)
					if call.Err != nil {
						return ErrMsg(call.Err)
					}
					break
				}
			}
		}

		return ActionSuccessMsg("VPN Activation/Deactivation State updated!")
	}
}

func DeleteTunnelCmd(tunnel TunnelProfile) tea.Cmd {
	return func() tea.Msg {
		if tunnel.ConnectionPath == "" {
			return ErrMsg(fmt.Errorf("cannot delete: connection reference is missing"))
		}

		conn := bus.Get()
		obj := conn.Object(nmDest, tunnel.ConnectionPath)
		call := obj.Call(nmSettingsIface+".Delete", 0)
		if call.Err != nil {
			return ErrMsg(fmt.Errorf("failed to delete profile: %v", call.Err))
		}

		return ActionSuccessMsg("WireGuard Profile deleted successfully!")
	}
}

// FetchIPWithGeoCmd fetches the public IP and location in one shot.
// Uses curl to avoid pulling in net/http + entire TLS stack.
func FetchIPWithGeoCmd(settleDelay time.Duration) tea.Cmd {
	return func() tea.Msg {
		if settleDelay > 0 {
			time.Sleep(settleDelay)
		}
		info := &IPInfo{}

		// Public IP
		out, err := exec.Command("curl", "-s", "--max-time", "5", "https://api.ipify.org").Output()
		if err == nil {
			info.PublicIP = strings.TrimSpace(string(out))
		}

		// Geo lookup
		if info.PublicIP != "" {
			out2, err := exec.Command(
				"curl", "-s", "--max-time", "5",
				fmt.Sprintf("http://ip-api.com/json/%s?fields=country,regionName,city,isp,status", info.PublicIP),
			).Output()
			if err == nil {
				var result struct {
					Status  string `json:"status"`
					Country string `json:"country"`
					Region  string `json:"regionName"`
					City    string `json:"city"`
					ISP     string `json:"isp"`
				}
				if json.Unmarshal(out2, &result) == nil && result.Status == "success" {
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
