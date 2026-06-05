package bluetooth

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/godbus/dbus/v5"
)

func ConvertMacToPath(mac string) dbus.ObjectPath {
	safeMac := strings.ReplaceAll(mac, ":", "_")
	return dbus.ObjectPath(fmt.Sprintf("%s/dev_%s", adapterPath, safeMac))
}

func Connect(client *BlueZClient, mac string) error {
	return callDeviceMethod(client, mac, "Connect")
}

func Disconnect(client *BlueZClient, mac string) error {
	return callDeviceMethod(client, mac, "Disconnect")
}

func Pair(client *BlueZClient, mac string) error {
	return callDeviceMethod(client, mac, "Pair")
}

func Trust(client *BlueZClient, mac string) error {
	return setDeviceProperty(client, mac, "Trusted", true)
}

func Distrust(client *BlueZClient, mac string) error {
	return setDeviceProperty(client, mac, "Trusted", false)
}

func Remove(client *BlueZClient, mac string) error {
	adapter := client.Conn.Object(bluezInterface, adapterPath)
	path := ConvertMacToPath(mac)

	call := adapter.Call("org.bluez.Adapter1.RemoveDevice", 0, path)
	return call.Err
}

func callDeviceMethod(client *BlueZClient, mac string, method string) error {
	path := ConvertMacToPath(mac)
	obj := client.Conn.Object(bluezInterface, path)

	call := obj.Call(fmt.Sprintf("org.bluez.Device1.%s", method), 0)
	return call.Err
}

func setDeviceProperty(client *BlueZClient, mac string, propName string, value bool) error {
	path := ConvertMacToPath(mac)
	obj := client.Conn.Object(bluezInterface, path)

	call := obj.Call("org.freedesktop.DBus.Properties.Set", 0, "org.bluez.Device1", propName, dbus.MakeVariant(value))
	return call.Err
}

func ExecuteActionCmd(client *BlueZClient, action string, mac string) tea.Cmd {
	return func() tea.Msg {
		var err error
		switch action {
		case "Connect":
			_ = Trust(client, mac)
			err = Connect(client, mac)
		case "Disconnect":
			err = Disconnect(client, mac)
		case "Pair":
			_ = Trust(client, mac)
			err = Pair(client, mac)
		case "Trust":
			err = Trust(client, mac)
		case "Distrust":
			err = Distrust(client, mac)
		case "Remove":
			err = Remove(client, mac)
		}

		if err != nil {
			return ErrMsg(err)
		}
		return ActionSuccessMsg(fmt.Sprintf("%s executed successfully", action))
	}
}
