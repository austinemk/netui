package wifi

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/godbus/dbus/v5"
)

func PollWifiTicker() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func TriggerHardwareScanCmd(client *DBusClient) tea.Cmd {
	return func() tea.Msg {
		obj := client.Conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")
		var devices []dbus.ObjectPath
		_ = obj.Call("org.freedesktop.NetworkManager.GetDevices", 0).Store(&devices)

		for _, path := range devices {
			devObj := client.Conn.Object("org.freedesktop.NetworkManager", path)
			devType, _ := devObj.GetProperty("org.freedesktop.NetworkManager.Device.DeviceType")
			if u, ok := devType.Value().(uint32); ok && u == 2 {
				// Fire network hardware card interface probe
				_ = devObj.Call("org.freedesktop.NetworkManager.Device.Wireless.RequestScan", 0, map[string]dbus.Variant{})
				break
			}
		}
		aps, _ := GetActiveAccessPoints(client)
		return ScanFinishedMsg(aps)
	}
}

func IsProfileSaved(profiles []SavedProfile, ssid string) bool {
	for _, p := range profiles {
		if p.Name == ssid {
			return true
		}
	}
	return false
}
