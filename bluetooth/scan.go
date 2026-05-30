package bluetooth

import (
	"fmt"
	"os"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/godbus/dbus/v5"
)

// StartScanCmd tells BlueZ to turn on the physical Bluetooth radio scanning
func StartScanCmd(client *BlueZClient) tea.Cmd {
	return func() tea.Msg {
		obj := client.Conn.Object(bluezInterface, adapterPath)
		call := obj.Call("org.bluez.Adapter1.StartDiscovery", 0)
		if call.Err != nil {
			return ErrMsg(call.Err)
		}
		return ScanStartedMsg{}
	}
}

// StopScanCmd tells BlueZ to shut down the physical radio scanning immediately
func StopScanCmd(client *BlueZClient) tea.Cmd {
	return func() tea.Msg {
		obj := client.Conn.Object(bluezInterface, adapterPath)
		_ = obj.Call("org.bluez.Adapter1.StopDiscovery", 0)
		return ScanStoppedMsg{}
	}
}

// FetchAllBlueZObjects remains as the single source of truth from the OS cache
func FetchAllBlueZObjects(client *BlueZClient) ([]Device, error) {
	obj := client.Conn.Object(bluezInterface, dbus.ObjectPath("/"))
	var nodes map[dbus.ObjectPath]map[string]map[string]dbus.Variant

	err := obj.Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&nodes)
	if err != nil {
		logToFile("❌ DBUS ERROR calling GetManagedObjects: %v", err)
		return nil, err
	}

	logToFile("📬 Received %d total DBus object paths from ObjectManager", len(nodes))

	var devices []Device
	for path, interfaces := range nodes {
		props, exists := interfaces["org.bluez.Device1"]
		if !exists {
			// This path is not a Bluetooth device (might be an adapter or agent)
			continue
		}

		logToFile("🔍 Inspecting device path: %s", path)

		dev := Device{}

		if addr, ok := props["Address"].Value().(string); ok {
			dev.MAC = addr
		}

		if dev.MAC == "" {
			logToFile("⚠️  Skipping path %s because MAC Address is empty", path)
			continue
		}

		if name, ok := props["Name"].Value().(string); ok {
			dev.Name = name
		} else if alias, ok := props["Alias"].Value().(string); ok {
			dev.Name = alias
		} else {
			dev.Name = "Unknown Device"
		}

		if paired, ok := props["Paired"].Value().(bool); ok {
			dev.Paired = paired
		}
		if connected, ok := props["Connected"].Value().(bool); ok {
			dev.Connected = connected
		}
		if trusted, ok := props["Trusted"].Value().(bool); ok {
			dev.Trusted = trusted
		}

		logToFile("✅ Successfully parsed: Name='%s' MAC='%s' Paired=%t Connected=%t", dev.Name, dev.MAC, dev.Paired, dev.Connected)

		devices = append(devices, dev)
	}

	logToFile("📦 Total devices loaded into memory from scan block: %d", len(devices))
	return devices, nil
}

func LoadPairedDevicesCmd(client *BlueZClient) tea.Cmd {
	return func() tea.Msg {
		logToFile("📥 Command Triggered: LoadPairedDevicesCmd")
		devices, err := FetchAllBlueZObjects(client)
		if err != nil {
			return ErrMsg(err)
		}

		var pairedOnly []Device
		for _, d := range devices {
			if d.Paired {
				pairedOnly = append(pairedOnly, d)
			}
		}
		logToFile("💾 Filtering SAVED table: showing %d paired out of %d total devices", len(pairedOnly), len(devices))
		return PairedDevicesLoadedMsg(pairedOnly)
	}
}

func DiscoverDevicesCmd(client *BlueZClient) tea.Cmd {
	return func() tea.Msg {
		logToFile("📡 Command Triggered: DiscoverDevicesCmd")
		devices, err := FetchAllBlueZObjects(client)
		if err != nil {
			return ErrMsg(err)
		}

		var discoveredOnly []Device
		for _, d := range devices {
			if !d.Paired {
				discoveredOnly = append(discoveredOnly, d)
			}
		}
		logToFile("🌐 Filtering DISCOVERED table: showing %d unpaired out of %d total devices", len(discoveredOnly), len(devices))
		return DiscoveredDevicesLoadedMsg(discoveredOnly)
	}
}

func PollBluetoothTicker() tea.Cmd {
	return tea.Tick(4*time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func logToFile(format string, v ...interface{}) {
	f, err := os.OpenFile("debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	msg := fmt.Sprintf(format, v...)
	fmt.Fprintf(f, "[%s] %s\n", time.Now().Format("15:04:05"), msg)
}
