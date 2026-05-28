package bluetooth

import (
	"sync"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/godbus/dbus/v5"
)

// scanConn holds the D-Bus connection that issued StartDiscovery.
// BlueZ ties the discovery session to the originating connection — closing
// it immediately cancels discovery. We keep it open until StopScan is called.
var (
	scanConn   *dbus.Conn
	scanConnMu sync.Mutex
)

func StartScanCmd() tea.Cmd {
	return func() tea.Msg {
		scanConnMu.Lock()
		defer scanConnMu.Unlock()

		// Close any leftover session from a previous scan
		if scanConn != nil {
			_ = scanConn.Close()
			scanConn = nil
		}

		conn, err := getSystemBus()
		if err != nil {
			return ErrMsg(err)
		}

		obj := conn.Object(bluezInterface, adapterPath)
		call := obj.Call("org.bluez.Adapter1.StartDiscovery", 0)
		if call.Err != nil {
			_ = conn.Close()
			return ErrMsg(call.Err)
		}

		// Keep the connection alive — closing it would end the BlueZ
		// discovery session immediately.
		scanConn = conn
		return ScanStartedMsg{}
	}
}

func StopScanCmd() tea.Cmd {
	return func() tea.Msg {
		scanConnMu.Lock()
		defer scanConnMu.Unlock()

		if scanConn != nil {
			obj := scanConn.Object(bluezInterface, adapterPath)
			_ = obj.Call("org.bluez.Adapter1.StopDiscovery", 0)
			_ = scanConn.Close()
			scanConn = nil
		}

		return ScanStoppedMsg{}
	}
}

func FetchDevicesCmd() tea.Cmd {
	return func() tea.Msg {
		devs, err := FetchCachedDevices()
		if err != nil {
			return ErrMsg(err)
		}
		return DevicesLoadedMsg(devs)
	}
}

func PollDevicesTicker() tea.Cmd {
	return tea.Tick(10*time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func getFilteredDevices(devices []Device, scanning bool) []Device {
	var filtered []Device
	for _, dev := range devices {
		if scanning {
			filtered = append(filtered, dev)
		} else {
			if dev.Paired {
				filtered = append(filtered, dev)
			}
		}
	}
	return filtered
}

// getFilteredDevices is kept as a method for compatibility with view.go and init.go
/*func (m Model) getFilteredDevices() []Device {
	return getFilteredDevices(m.Devices, m.Scanning)
}*/

func ControlScan(turnOn bool) error {
	conn, err := getSystemBus()
	if err != nil {
		return err
	}
	defer conn.Close()

	obj := conn.Object(bluezInterface, adapterPath)

	var method string
	if turnOn {
		method = "org.bluez.Adapter1.StartDiscovery"
	} else {
		method = "org.bluez.Adapter1.StopDiscovery"
	}

	call := obj.Call(method, 0)
	return call.Err
}

func FetchCachedDevices() ([]Device, error) {
	conn, err := getSystemBus()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	obj := conn.Object(bluezInterface, "/")
	var nodes map[dbus.ObjectPath]map[string]map[string]dbus.Variant

	err = obj.Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&nodes)
	if err != nil {
		return nil, err
	}

	var devices []Device

	for _, interfaces := range nodes {
		if props, exists := interfaces["org.bluez.Device1"]; exists {
			dev := Device{}

			if addr, ok := props["Address"].Value().(string); ok {
				dev.MAC = addr
			}

			if name, ok := props["Name"].Value().(string); ok {
				dev.Name = name
			} else if alias, ok := props["Alias"].Value().(string); ok {
				dev.Name = alias
			} else {
				dev.Name = dev.MAC
			}

			if connected, ok := props["Connected"].Value().(bool); ok {
				dev.Connected = connected
			}
			if paired, ok := props["Paired"].Value().(bool); ok {
				dev.Paired = paired
			}
			if trusted, ok := props["Trusted"].Value().(bool); ok {
				dev.Trusted = trusted
			}
			if icon, ok := props["Icon"].Value().(string); ok {
				dev.Icon = string(FromString(icon))
			}

			devices = append(devices, dev)
		}
	}

	return devices, nil
}
