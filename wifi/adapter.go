package wifi

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/godbus/dbus/v5"
)

func GetAdapterSettings(client *DBusClient) (AdapterInfo, error) {
	obj := client.Conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")
	wirelessEnabled, err := obj.GetProperty("org.freedesktop.NetworkManager.WirelessEnabled")
	if err != nil {
		return AdapterInfo{}, err
	}

	var devices []dbus.ObjectPath
	_ = obj.Call("org.freedesktop.NetworkManager.GetDevices", 0).Store(&devices)

	for _, path := range devices {
		devObj := client.Conn.Object("org.freedesktop.NetworkManager", path)
		devType, _ := devObj.GetProperty("org.freedesktop.NetworkManager.Device.DeviceType")
		if u, ok := devType.Value().(uint32); ok && u == 2 {
			iface, _ := devObj.GetProperty("org.freedesktop.NetworkManager.Device.Interface")
			state, _ := devObj.GetProperty("org.freedesktop.NetworkManager.Device.State")

			stateStr := "Disconnected"
			if s, ok := state.Value().(uint32); ok && s == 100 {
				stateStr = "Connected"
			}
			return AdapterInfo{
				Interface: iface.Value().(string),
				State:     stateStr,
				Enabled:   wirelessEnabled.Value().(bool),
			}, nil
		}
	}
	return AdapterInfo{Interface: "Unknown", State: "Missing", Enabled: false}, nil
}

func TogglePowerCmd(client *DBusClient, enable bool) tea.Cmd {
	return func() tea.Msg {
		obj := client.Conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")
		err := obj.SetProperty("org.freedesktop.NetworkManager.WirelessEnabled", dbus.MakeVariant(enable))
		if err != nil {
			return ErrMsg(err)
		}
		return AdapterToggledMsg{}
	}
}
