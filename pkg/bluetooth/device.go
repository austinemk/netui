package bluetooth

import (
	"fmt"
	"strings"

	"github.com/austinemk/linktui/pkg/bus"

	tea "charm.land/bubbletea/v2"
	"github.com/godbus/dbus/v5"
)

func ConvertMacToPath(mac string) dbus.ObjectPath {
	safeMac := strings.ReplaceAll(mac, ":", "_")
	return dbus.ObjectPath(fmt.Sprintf("%s/dev_%s", adapterPath, safeMac))
}

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
	conn := bus.Get()
	adapter := conn.Object(bluezInterface, adapterPath)
	path := ConvertMacToPath(mac)
	return adapter.Call("org.bluez.Adapter1.RemoveDevice", 0, path).Err
}

func callDeviceMethod(mac string, method string) error {
	conn := bus.Get()
	path := ConvertMacToPath(mac)
	obj := conn.Object(bluezInterface, path)
	return obj.Call(fmt.Sprintf("org.bluez.Device1.%s", method), 0).Err
}

func setDeviceProperty(mac string, propName string, value bool) error {
	conn := bus.Get()
	path := ConvertMacToPath(mac)
	obj := conn.Object(bluezInterface, path)
	return obj.Call(
		"org.freedesktop.DBus.Properties.Set", 0,
		"org.bluez.Device1", propName, dbus.MakeVariant(value),
	).Err
}

func ExecuteActionCmd(action string, mac string) tea.Cmd {
	return func() tea.Msg {
		var err error
		switch action {
		case "Connect":
			_ = Trust(mac)
			err = Connect(mac)
		case "Disconnect":
			err = Disconnect(mac)
		case "Pair":
			_ = Trust(mac)
			err = Pair(mac)
		case "Trust":
			err = Trust(mac)
		case "Distrust":
			err = Distrust(mac)
		case "Remove":
			err = Remove(mac)
		}

		if err != nil {
			return ErrMsg(err)
		}
		return ActionSuccessMsg(fmt.Sprintf("%s executed successfully", action))
	}
}
