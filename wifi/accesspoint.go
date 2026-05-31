package wifi

import (
	"context" // <-- 1. ADD THIS IMPORT
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Wifx/gonetworkmanager/v3"
)

// <-- 2. UPDATE THIS SIGNATURE TO ACCEPT A CONTEXT
func ConnectToAccessPoint(ctx context.Context, nm gonetworkmanager.NetworkManager, ap AccessPoint, password string) tea.Cmd {
	return func() tea.Msg {
		devices, err := nm.GetDevices()
		if err != nil {
			return ErrMsg(err)
		}

		var targetDevice gonetworkmanager.Device
		for _, dev := range devices {
			devType, _ := dev.GetPropertyDeviceType()
			if devType == gonetworkmanager.NmDeviceTypeWifi {
				targetDevice = dev
				break
			}
		}

		if targetDevice == nil {
			return ErrMsg(err)
		}

		if password == "" {
			settings, err := gonetworkmanager.NewSettings()
			if err != nil {
				return ErrMsg(err)
			}

			connections, err := settings.ListConnections()
			if err != nil {
				return ErrMsg(err)
			}

			var matchedConnection gonetworkmanager.Connection
			for _, conn := range connections {
				sMap, err := conn.GetSettings()
				if err != nil {
					continue
				}
				if connSettings, ok := sMap["connection"]; ok {
					if connSettings["id"] == ap.SSID {
						matchedConnection = conn
						break
					}
				}
			}

			if matchedConnection == nil {
				return ErrMsg(err)
			}

			_, err = nm.ActivateConnection(matchedConnection, targetDevice, nil)
			if err != nil {
				return ErrMsg(err)
			}
			return ActionSuccessMsg("Connecting to saved network...")
		} else {
			connectionSettings := map[string]map[string]interface{}{
				"connection": {
					"id":    ap.SSID,
					"type":  "802-11-wireless",
					"flags": uint32(1),
				},
				"802-11-wireless": {
					"ssid": []byte(ap.SSID),
				},
				"802-11-wireless-security": {
					"key-mgmt": "wpa-psk",
					"psk":      password,
				},
			}

			activeConn, err := nm.AddAndActivateConnection(connectionSettings, targetDevice)
			if err != nil {
				return ErrMsg(err)
			}

			// <-- 3. PASS THE CTX DOWN TO THE MONITOR GOROUTINE
			go monitorConnectionState(ctx, activeConn)

			return ActionSuccessMsg("Authenticating...")
		}
	}
}

// <-- 4. UPDATE THIS TO WATCH THE CONTEXT FOR EARLY CLOSURE
func monitorConnectionState(ctx context.Context, activeConn gonetworkmanager.ActiveConnection) {
	for i := 0; i < 15; i++ {
		// Use a select block to watch for the application closing early
		select {
		case <-ctx.Done():
			// The user closed the app, and Clean() was called!
			// Exit immediately and stop leaking resources.
			return
		default:
			time.Sleep(1 * time.Second)
		}

		state, err := activeConn.GetPropertyState()
		if err != nil {
			break
		}

		if state == gonetworkmanager.NmActiveConnectionStateActivated {
			return
		}

		if state == gonetworkmanager.NmActiveConnectionStateDeactivating {
			break
		}
	}
}

// ... Keep ToggleAutoConnectCmd, ForgetProfileCmd, GetActiveAccessPoints, and GetSavedProfiles as they were
func ToggleAutoConnectCmd(nm gonetworkmanager.NetworkManager, uuid string, auto bool) tea.Cmd {
	return func() tea.Msg {
		settings, err := gonetworkmanager.NewSettings()
		if err == nil {
			connections, err := settings.ListConnections()
			if err == nil {
				for _, conn := range connections {
					sMap, err := conn.GetSettings()
					if err == nil {
						if cSettings, ok := sMap["connection"]; ok {
							if cSettings["uuid"] == uuid {
								sMap["connection"]["autoconnect"] = auto
								_ = conn.Update(sMap)
								break
							}
						}
					}
				}
			}
		}
		return ActionSuccessMsg("AutoConnect Updated")
	}
}

func ForgetProfileCmd(nm gonetworkmanager.NetworkManager, uuid string) tea.Cmd {
	return func() tea.Msg {
		settings, err := gonetworkmanager.NewSettings()
		if err == nil {
			connections, err := settings.ListConnections()
			if err == nil {
				for _, conn := range connections {
					sMap, err := conn.GetSettings()
					if err == nil {
						if cSettings, ok := sMap["connection"]; ok {
							if cSettings["uuid"] == uuid {
								_ = conn.Delete()
								break
							}
						}
					}
				}
			}
		}
		return ActionSuccessMsg("Profile Removed")
	}
}

func GetSavedProfiles(nm gonetworkmanager.NetworkManager) ([]SavedProfile, error) {
	settings, err := gonetworkmanager.NewSettings()
	if err != nil {
		return nil, err
	}

	connections, err := settings.ListConnections()
	if err != nil {
		return nil, err
	}

	var saved []SavedProfile
	for _, conn := range connections {
		sMap, err := conn.GetSettings()
		if err != nil {
			continue
		}

		connSettings, ok := sMap["connection"]
		if !ok {
			continue
		}

		if connSettings["type"] == "802-11-wireless" {
			auto := true
			if val, ok := connSettings["autoconnect"]; ok {
				if bVal, ok := val.(bool); ok {
					auto = bVal
				}
			}
			saved = append(saved, SavedProfile{
				Name:        connSettings["id"].(string),
				UUID:        connSettings["uuid"].(string),
				AutoConnect: auto,
				Settings:    conn,
			})
		}
	}
	return saved, nil
}
