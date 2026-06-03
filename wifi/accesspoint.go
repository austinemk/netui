package wifi

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Wifx/gonetworkmanager/v3"
)

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
