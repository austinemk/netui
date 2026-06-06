package wifi

import (
	"fmt"

	"github.com/austinemk/linktui/pkg/bus"

	tea "charm.land/bubbletea/v2"
	"github.com/godbus/dbus/v5"
)

func GetAdapterSettings() (AdapterInfo, error) {
	conn := bus.Get()
	nm := conn.Object(nmDest, nmPath)

	// Wireless enabled
	wev, err := nm.GetProperty(nmIface + ".WirelessEnabled")
	if err != nil {
		return AdapterInfo{}, err
	}
	wirelessEnabled, _ := wev.Value().(bool)

	wifiPath, err := findWifiDevicePath(conn)
	if err != nil || wifiPath == "" {
		return AdapterInfo{Interface: "Unknown", State: "Missing", Enabled: false},
			fmt.Errorf("no wireless adapter found")
	}

	wDev := conn.Object(nmDest, wifiPath)

	ifaceV, _ := wDev.GetProperty(nmDeviceIface + ".Interface")
	iface, _ := ifaceV.Value().(string)

	stateV, _ := wDev.GetProperty(nmDeviceIface + ".State")
	state, _ := stateV.Value().(uint32)

	stateStr := "Disconnected"
	if state == 100 { // NM_DEVICE_STATE_ACTIVATED = 100
		stateStr = "Connected"
	}

	return AdapterInfo{
		Interface: iface,
		State:     stateStr,
		Enabled:   wirelessEnabled,
	}, nil
}

func TogglePowerCmd(enable bool) tea.Cmd {
	return func() tea.Msg {
		conn := bus.Get()
		nm := conn.Object(nmDest, nmPath)

		call := nm.Call(
			"org.freedesktop.DBus.Properties.Set", 0,
			nmIface, "WirelessEnabled", dbus.MakeVariant(enable),
		)
		if call.Err != nil {
			return ErrMsg(call.Err)
		}
		return AdapterToggledMsg{}
	}
}
