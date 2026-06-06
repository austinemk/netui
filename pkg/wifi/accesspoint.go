package wifi

import (
	"github.com/austinemk/linktui/pkg/bus"

	tea "charm.land/bubbletea/v2"
	"github.com/godbus/dbus/v5"
)

const (
	nmConnSettingsIface = "org.freedesktop.NetworkManager.Settings.Connection"
)

func ToggleAutoConnectCmd(uuid string, auto bool) tea.Cmd {
	return func() tea.Msg {
		conn := bus.Get()
		cPath, sMap, err := findConnectionByUUID(conn, uuid)
		if err != nil {
			return ErrMsg(err)
		}

		cSettings := sMap["connection"]
		patch := map[string]map[string]dbus.Variant{
			"connection": {
				"id":          cSettings["id"],
				"uuid":        cSettings["uuid"],
				"type":        cSettings["type"],
				"autoconnect": dbus.MakeVariant(auto),
			},
		}

		cObj := conn.Object(nmDest, cPath)
		call := cObj.Call(nmConnSettingsIface+".Update", 0, patch)
		if call.Err != nil {
			return ErrMsg(call.Err)
		}

		return ActionSuccessMsg("AutoConnect Updated")
	}
}

func ForgetProfileCmd(uuid string) tea.Cmd {
	return func() tea.Msg {
		conn := bus.Get()
		cPath, _, err := findConnectionByUUID(conn, uuid)
		if err != nil {
			return ErrMsg(err)
		}

		cObj := conn.Object(nmDest, cPath)
		call := cObj.Call(nmConnSettingsIface+".Delete", 0)
		if call.Err != nil {
			return ErrMsg(call.Err)
		}

		return ActionSuccessMsg("Profile Removed")
	}
}

func GetSavedProfiles() ([]SavedProfile, error) {
	conn := bus.Get()
	settings := conn.Object(nmDest, nmSettingsPath)

	var connPaths []dbus.ObjectPath
	call := settings.Call(nmSettingsSvc+".ListConnections", 0)
	if call.Err != nil {
		return nil, call.Err
	}
	if err := call.Store(&connPaths); err != nil {
		return nil, err
	}

	var saved []SavedProfile
	for _, cPath := range connPaths {
		cObj := conn.Object(nmDest, cPath)

		var sMap map[string]map[string]dbus.Variant
		call := cObj.Call(nmConnSettingsIface+".GetSettings", 0)
		if call.Err != nil {
			continue
		}
		if err := call.Store(&sMap); err != nil {
			continue
		}

		connSettings, ok := sMap["connection"]
		if !ok {
			continue
		}

		cType, _ := connSettings["type"].Value().(string)
		if cType != "802-11-wireless" {
			continue
		}

		auto := true
		if v, ok := connSettings["autoconnect"]; ok {
			if bVal, ok := v.Value().(bool); ok {
				auto = bVal
			}
		}

		id, _ := connSettings["id"].Value().(string)
		uid, _ := connSettings["uuid"].Value().(string)

		saved = append(saved, SavedProfile{
			Name:           id,
			UUID:           uid,
			AutoConnect:    auto,
			ConnectionPath: cPath,
		})
	}
	return saved, nil
}

// findConnectionByUUID returns the object path and settings map for a connection by UUID.
func findConnectionByUUID(conn *dbus.Conn, uuid string) (dbus.ObjectPath, map[string]map[string]dbus.Variant, error) {
	settings := conn.Object(nmDest, nmSettingsPath)

	var connPaths []dbus.ObjectPath
	call := settings.Call(nmSettingsSvc+".ListConnections", 0)
	if call.Err != nil {
		return "", nil, call.Err
	}
	if err := call.Store(&connPaths); err != nil {
		return "", nil, err
	}

	for _, cPath := range connPaths {
		cObj := conn.Object(nmDest, cPath)
		var sMap map[string]map[string]dbus.Variant
		call := cObj.Call(nmConnSettingsIface+".GetSettings", 0)
		if call.Err != nil {
			continue
		}
		if err := call.Store(&sMap); err != nil {
			continue
		}

		connSettings, ok := sMap["connection"]
		if !ok {
			continue
		}

		if u, _ := connSettings["uuid"].Value().(string); u == uuid {
			return cPath, sMap, nil
		}
	}
	return "", nil, nil
}
