package bluetooth

import (
	"fmt"
	"strings"

	"github.com/godbus/dbus/v5"
)

type Device struct {
	MAC       string
	Name      string
	Connected bool
	Paired    bool
	Trusted   bool
	Icon      string
}

const (
	bluezInterface = "org.bluez"
	adapterPath    = "/org/bluez/hci0" // Default bluetooth adapter path
)

// Helper to grab a quick handle on the System Bus
func getSystemBus() (*dbus.Conn, error) {
	return dbus.ConnectSystemBus()
}

// ControlScan non-blockingly toggles background discovery state using native D-Bus calls.
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

// FetchCachedDevices Iterate through the managed objects looking for paths that implement org.bluez.Device1
// FetchCachedDevices reads the BlueZ ObjectManager hierarchy to extract device listings instantly.
func FetchCachedDevices() ([]Device, error) {
	conn, err := getSystemBus()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	// Use ObjectManager to fetch ALL managed objects from BlueZ in one single call
	obj := conn.Object(bluezInterface, "/")
	var nodes map[dbus.ObjectPath]map[string]map[string]dbus.Variant

	err = obj.Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&nodes)
	if err != nil {
		return nil, err
	}

	var devices []Device

	// Fix: Use a blank identifier (_) for the object path since we don't explicitly read it
	for _, interfaces := range nodes {
		if props, exists := interfaces["org.bluez.Device1"]; exists {
			dev := Device{}

			// Extract MAC address from Address property
			if addr, ok := props["Address"].Value().(string); ok {
				dev.MAC = addr
			}

			// Extract Name (Fallback to Alias or Address if Name is empty)
			if name, ok := props["Name"].Value().(string); ok {
				dev.Name = name
			} else if alias, ok := props["Alias"].Value().(string); ok {
				dev.Name = alias
			} else {
				dev.Name = dev.MAC
			}

			// Extract States
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
				dev.Icon = icon
			}

			devices = append(devices, dev)
		}
	}

	// Fallback mock data if BlueZ is completely empty/inaccessible (like local non-linux testing)
	if len(devices) == 0 {
		return []Device{}, nil
	}

	return devices, nil
}

// ConvertMacToPath turns "00:11:22:33:44:55" into BlueZ object path format "/org/bluez/hci0/dev_00_11_22_33_44_55"
func ConvertMacToPath(mac string) dbus.ObjectPath {
	safeMac := strings.ReplaceAll(mac, ":", "_")
	return dbus.ObjectPath(fmt.Sprintf("%s/dev_%s", adapterPath, safeMac))
}

// Connect Global action wrapper functions called by view.go actions popup menu
func Connect(mac string) error {
	return callDeviceMethod(mac, "Connect")
}

func Disconnect(mac string) error {
	return callDeviceMethod(mac, "Disconnect")
}

func Pair(mac string) error {
	return callDeviceMethod(mac, "Pair")
}

func Trust(mac string) error {
	return setDeviceProperty(mac, "Trusted", true)
}

func Distrust(mac string) error {
	return setDeviceProperty(mac, "Trusted", false)
}

func Remove(mac string) error {
	conn, err := getSystemBus()
	if err != nil {
		return err
	}
	defer conn.Close()

	adapter := conn.Object(bluezInterface, adapterPath)
	path := ConvertMacToPath(mac)

	call := adapter.Call("org.bluez.Adapter1.RemoveDevice", 0, path)
	return call.Err
}

// Helper to trigger commands directly on a device object interface
func callDeviceMethod(mac string, method string) error {
	conn, err := getSystemBus()
	if err != nil {
		return err
	}
	defer conn.Close()

	path := ConvertMacToPath(mac)
	obj := conn.Object(bluezInterface, path)

	call := obj.Call(fmt.Sprintf("org.bluez.Device1.%s", method), 0)
	return call.Err
}

// Helper to edit BlueZ boolean flags (like Trusted)
func setDeviceProperty(mac string, propName string, value bool) error {
	conn, err := getSystemBus()
	if err != nil {
		return err
	}
	defer conn.Close()

	path := ConvertMacToPath(mac)
	obj := conn.Object(bluezInterface, path)

	call := obj.Call("org.freedesktop.DBus.Properties.Set", 0, "org.bluez.Device1", propName, dbus.MakeVariant(value))
	return call.Err
}
