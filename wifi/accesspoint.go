package wifi

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Wifx/gonetworkmanager/v3"
	"github.com/godbus/dbus/v5"
)

func ToggleAutoConnectCmd(nm gonetworkmanager.NetworkManager, uuid string, auto bool) tea.Cmd {
	return func() tea.Msg {
		settings, err := gonetworkmanager.NewSettings()
		if err != nil {
			return ErrMsg(err)
		}

		connections, err := settings.ListConnections()
		if err != nil {
			return ErrMsg(err)
		}

		for _, conn := range connections {
			sMap, err := conn.GetSettings()
			if err != nil {
				continue
			}

			cSettings, ok := sMap["connection"]
			if !ok {
				continue
			}

			if cSettings["uuid"] == uuid {
				// Only send the sections we're touching — avoids dbus
				// round-trip failures on complex types like ipv6.addresses
				patch := map[string]map[string]interface{}{
					"connection": {
						"id":          cSettings["id"],
						"uuid":        cSettings["uuid"],
						"type":        cSettings["type"],
						"autoconnect": dbus.MakeVariant(auto),
					},
				}
				if err := conn.Update(patch); err != nil {
					return ErrMsg(err)
				}
				break
			}
		}

		return ActionSuccessMsg("AutoConnect Updated")
	}
}

func ForgetProfileCmd(nm gonetworkmanager.NetworkManager, uuid string) tea.Cmd {
	return func() tea.Msg {
		settings, err := gonetworkmanager.NewSettings()
		if err != nil {
			return ErrMsg(err)
		}

		connections, err := settings.ListConnections()
		if err != nil {
			return ErrMsg(err)
		}

		for _, conn := range connections {
			sMap, err := conn.GetSettings()
			if err != nil {
				continue
			}

			cSettings, ok := sMap["connection"]
			if !ok {
				continue
			}

			if cSettings["uuid"] == uuid {
				if err := conn.Delete(); err != nil {
					return ErrMsg(err)
				}
				break
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
				// GetSettings returns dbus.Variant — unwrap it
				if v, ok := val.(dbus.Variant); ok {
					if bVal, ok := v.Value().(bool); ok {
						auto = bVal
					}
				} else if bVal, ok := val.(bool); ok {
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
