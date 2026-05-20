package wifi

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/godbus/dbus/v5"
)

func ConnectToAccessPoint(client *DBusClient, ap AccessPoint, password string) tea.Cmd {
	return func() tea.Msg {
		nm := client.Conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")

		// Find active Wi-Fi Device reference
		var devices []dbus.ObjectPath
		_ = nm.Call("org.freedesktop.NetworkManager.GetDevices", 0).Store(&devices)
		var targetDevice dbus.ObjectPath
		for _, path := range devices {
			devObj := client.Conn.Object("org.freedesktop.NetworkManager", path)
			devType, _ := devObj.GetProperty("org.freedesktop.NetworkManager.Device.DeviceType")
			if u, ok := devType.Value().(uint32); ok && u == 2 {
				targetDevice = path
				break
			}
		}

		if password == "" {
			var activeConn dbus.ObjectPath
			_ = nm.Call("org.freedesktop.NetworkManager.ActivateConnection", 0, dbus.ObjectPath("/"), targetDevice, ap.Path).Store(&activeConn)
			return ActionSuccessMsg("Connecting to saved network...")
		} else {
			// Create a transient connection map
			connectionSettings := map[string]map[string]dbus.Variant{
				"connection": {
					"id":   dbus.MakeVariant(ap.SSID),
					"type": dbus.MakeVariant("802-11-wireless"),
					// 0x1 represents NM_CONNECTION_FLAG_NOT_SAVED (Transient / In-memory only)
					"flags": dbus.MakeVariant(uint32(1)),
				},
				"802-11-wireless": {
					"ssid": dbus.MakeVariant([]byte(ap.SSID)),
				},
				"802-11-wireless-security": {
					"key-mgmt": dbus.MakeVariant("wpa-psk"),
					"psk":      dbus.MakeVariant(password),
				},
			}

			var path dbus.ObjectPath
			var activeConn dbus.ObjectPath

			// Call AddAndActivateConnection. Because flags=1, this profile is NOT saved to disk yet.
			err := nm.Call("org.freedesktop.NetworkManager.AddAndActivateConnection", 0, connectionSettings, targetDevice, ap.Path).Store(&path, &activeConn)
			if err != nil {
				return ErrMsg(err)
			}

			// Launch a background monitoring routine to see if authentication succeeds
			go monitorConnectionState(client, activeConn, path)

			return ActionSuccessMsg("Authenticating...")
		}
	}
}

// monitorConnectionState watches the active connection to see if it successfully hits state 2 (ACTIVATED)
func monitorConnectionState(client *DBusClient, activeConnPath dbus.ObjectPath, connPath dbus.ObjectPath) {
	activeObj := client.Conn.Object("org.freedesktop.NetworkManager", activeConnPath)

	// Poll the state for up to 15 seconds to check if authentication passes
	for i := 0; i < 15; i++ {
		time.Sleep(1 * time.Second)

		stateVal, err := activeObj.GetProperty("org.freedesktop.NetworkManager.Connection.Active.State")
		if err != nil {
			break
		}

		state, ok := stateVal.Value().(uint32)
		if !ok {
			continue
		}

		// State 2 == NM_ACTIVE_CONNECTION_STATE_ACTIVATED (Success!)
		if state == 2 {
			// The password was correct! Now commit the transient connection permanently to disk.
			settingsObj := client.Conn.Object("org.freedesktop.NetworkManager", connPath)
			// Call Save() on the Settings.Connection object to persist it
			_ = settingsObj.Call("org.freedesktop.NetworkManager.Settings.Connection.Save", 0)
			return
		}

		// State 4 == NM_ACTIVE_CONNECTION_STATE_DEACTIVATING / Failed
		if state == 4 {
			break
		}
	}

	// If it timed out or hit a failed state, we do absolutely nothing.
	// Because it was transient, NetworkManager automatically discards the profile.
}

func ToggleAutoConnectCmd(client *DBusClient, uuid string, auto bool) tea.Cmd {
	return func() tea.Msg {
		path, err := findConnectionPathByUUID(client, uuid)
		if err == nil {
			connObj := client.Conn.Object("org.freedesktop.NetworkManager", path)
			var settingsMap map[string]map[string]dbus.Variant
			_ = connObj.Call("org.freedesktop.NetworkManager.Settings.Connection.GetSettings", 0).Store(&settingsMap)

			settingsMap["connection"]["autoconnect"] = dbus.MakeVariant(auto)
			_ = connObj.Call("org.freedesktop.NetworkManager.Settings.Connection.Update", 0, settingsMap)
		}
		return ActionSuccessMsg("AutoConnect Updated")
	}
}

func ForgetProfileCmd(client *DBusClient, uuid string) tea.Cmd {
	return func() tea.Msg {
		path, err := findConnectionPathByUUID(client, uuid)
		if err == nil {
			connObj := client.Conn.Object("org.freedesktop.NetworkManager", path)
			_ = connObj.Call("org.freedesktop.NetworkManager.Settings.Connection.Delete", 0)
		}
		return ActionSuccessMsg("Profile Removed")
	}
}

func findConnectionPathByUUID(client *DBusClient, uuid string) (dbus.ObjectPath, error) {
	settings := client.Conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager/Settings")
	var connections []dbus.ObjectPath
	_ = settings.Call("org.freedesktop.NetworkManager.Settings.ListConnections", 0).Store(&connections)
	for _, path := range connections {
		connObj := client.Conn.Object("org.freedesktop.NetworkManager", path)
		var sMap map[string]map[string]dbus.Variant
		_ = connObj.Call("org.freedesktop.NetworkManager.Settings.Connection.GetSettings", 0).Store(&sMap)
		if sMap["connection"]["uuid"].Value().(string) == uuid {
			return path, nil
		}
	}
	return dbus.ObjectPath(""), error(nil)
}

func GetActiveAccessPoints(client *DBusClient) ([]AccessPoint, error) {
	obj := client.Conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")
	var devices []dbus.ObjectPath
	_ = obj.Call("org.freedesktop.NetworkManager.GetDevices", 0).Store(&devices)
	var wifiPath dbus.ObjectPath
	for _, path := range devices {
		devObj := client.Conn.Object("org.freedesktop.NetworkManager", path)
		devType, _ := devObj.GetProperty("org.freedesktop.NetworkManager.Device.DeviceType")
		if u, ok := devType.Value().(uint32); ok && u == 2 {
			wifiPath = path
			break
		}
	}
	if wifiPath == "" {
		return nil, nil
	}

	wifiDev := client.Conn.Object("org.freedesktop.NetworkManager", wifiPath)
	var apPaths []dbus.ObjectPath
	_ = wifiDev.Call("org.freedesktop.NetworkManager.Device.Wireless.GetAllAccessPoints", 0).Store(&apPaths)
	activeApProp, _ := wifiDev.GetProperty("org.freedesktop.NetworkManager.Device.Wireless.ActiveAccessPoint")
	activeApPath, _ := activeApProp.Value().(dbus.ObjectPath)

	var list []AccessPoint
	for _, path := range apPaths {
		apObj := client.Conn.Object("org.freedesktop.NetworkManager", path)
		ssidProp, _ := apObj.GetProperty("org.freedesktop.NetworkManager.AccessPoint.Ssid")
		strengthProp, _ := apObj.GetProperty("org.freedesktop.NetworkManager.AccessPoint.Strength")
		wpaProp, _ := apObj.GetProperty("org.freedesktop.NetworkManager.AccessPoint.WpaFlags")
		rsnProp, _ := apObj.GetProperty("org.freedesktop.NetworkManager.AccessPoint.RsnFlags")

		ssidBytes, ok := ssidProp.Value().([]uint8)
		if !ok || len(ssidBytes) == 0 {
			continue
		}
		strength, _ := strengthProp.Value().(uint8)

		sec := "open"
		if wpa, ok := wpaProp.Value().(uint32); ok && wpa > 0 {
			sec = "wpa"
		}
		if rsn, ok := rsnProp.Value().(uint32); ok && rsn > 0 {
			sec = "wpa2/3"
		}

		list = append(list, AccessPoint{
			SSID:     string(ssidBytes),
			Strength: strength,
			Security: sec,
			IsActive: path == activeApPath,
			Path:     path,
		})
	}
	return list, nil
}

func GetSavedProfiles(client *DBusClient) ([]SavedProfile, error) {
	settings := client.Conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager/Settings")
	var connections []dbus.ObjectPath
	if err := settings.Call("org.freedesktop.NetworkManager.Settings.ListConnections", 0).Store(&connections); err != nil {
		return nil, err
	}
	var saved []SavedProfile
	for _, path := range connections {
		connObj := client.Conn.Object("org.freedesktop.NetworkManager", path)
		var sMap map[string]map[string]dbus.Variant
		if err := connObj.Call("org.freedesktop.NetworkManager.Settings.Connection.GetSettings", 0).Store(&sMap); err != nil {
			continue
		}
		connSettings := sMap["connection"]
		if connSettings["type"].Value().(string) == "802-11-wireless" {
			auto := true
			if val, ok := connSettings["autoconnect"]; ok {
				auto = val.Value().(bool)
			}
			saved = append(saved, SavedProfile{
				Name:        connSettings["id"].Value().(string),
				UUID:        connSettings["uuid"].Value().(string),
				AutoConnect: auto,
			})
		}
	}
	return saved, nil
}
