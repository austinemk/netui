package bluetooth

import (
	"time"

	"github.com/austinemk/linktui/pkg/bus"

	tea "charm.land/bubbletea/v2"
	"github.com/godbus/dbus/v5"
)

func StartDiscoveryCmd() tea.Cmd {
	return func() tea.Msg {
		if err := StartDiscovery(); err != nil {
			return ErrMsg(err)
		}
		dvs, _ := DiscoverDevices()
		return ScanFinishedMsg(dvs)
	}
}

func ContinueDiscoveryCmd() tea.Cmd {
	return func() tea.Msg {
		dvs, _ := DiscoverDevices()
		return ScanFinishedMsg(dvs)
	}
}

func LoadPairedDevicesCmd() tea.Cmd {
	return func() tea.Msg {
		dvs, _ := LoadPairedDevices()
		ap, _ := FetchAdapterInfo()
		return InfoLoadedMsg(InfoLoadedData{
			Adapter: ap,
			Devices: dvs,
		})
	}
}

func StopDiscoveryCmd() tea.Cmd {
	return func() tea.Msg {
		StopDiscovery()
		return DiscoveryStoppedMsg{}
	}
}

func StartDiscovery() error {
	conn := bus.Get()
	obj := conn.Object(bluezInterface, adapterPath)
	return obj.Call("org.bluez.Adapter1.StartDiscovery", 0).Err
}

func StopDiscovery() {
	conn := bus.Get()
	obj := conn.Object(bluezInterface, adapterPath)
	_ = obj.Call("org.bluez.Adapter1.StopDiscovery", 0)
}

func FetchAllBlueZObjects() ([]Device, error) {
	conn := bus.Get()
	obj := conn.Object(bluezInterface, dbus.ObjectPath("/"))

	var nodes map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	if err := obj.Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&nodes); err != nil {
		return nil, err
	}

	var devices []Device
	for path, interfaces := range nodes {
		props, exists := interfaces["org.bluez.Device1"]
		if !exists {
			continue
		}

		dev := Device{}

		if addr, ok := props["Address"].Value().(string); ok {
			dev.MAC = addr
		}
		if dev.MAC == "" {
			continue
		}

		if name, ok := props["Name"].Value().(string); ok {
			dev.Name = name
		} else if alias, ok := props["Alias"].Value().(string); ok {
			dev.Name = alias
		} else {
			dev.Name = "Unknown Device"
		}

		dev.Icon = ""
		if iconName, ok := props["Icon"].Value().(string); ok {
			dev.Icon = FromString(iconName).String()
		} else if cod, ok := props["Class"].Value().(uint32); ok {
			dev.Icon = FromClassOfDevice(cod).String()
		} else if codInt, ok := props["Class"].Value().(int32); ok {
			dev.Icon = FromClassOfDevice(uint32(codInt)).String()
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

		_ = path
		devices = append(devices, dev)
	}

	return devices, nil
}

func LoadPairedDevices() ([]Device, error) {
	devices, err := FetchAllBlueZObjects()
	if err != nil {
		return nil, err
	}
	var pairedOnly []Device
	for _, d := range devices {
		if d.Paired {
			pairedOnly = append(pairedOnly, d)
		}
	}
	return pairedOnly, nil
}

func DiscoverDevices() ([]Device, error) {
	devices, err := FetchAllBlueZObjects()
	if err != nil {
		return nil, err
	}
	var discoveredOnly []Device
	for _, d := range devices {
		if !d.Paired {
			discoveredOnly = append(discoveredOnly, d)
		}
	}
	return discoveredOnly, nil
}

func PollBluetoothTicker() tea.Cmd {
	return tea.Tick(4*time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}
