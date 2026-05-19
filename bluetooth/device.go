package bluetooth

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
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

func setDeviceProperty(mac string, propName string, value bool) error {
	conn, err := getSystemBus()
	if err != nil {
		return err
	}
	defer conn.Close()

	path := ConvertMacToPath(mac)
	obj := conn.Object(bluezInterface, path)

	var typedValue bool = value
	call := obj.Call("org.freedesktop.DBus.Properties.Set", 0, "org.bluez.Device1", propName, dbus.MakeVariant(typedValue))
	return call.Err
}

// Fixed mapping to handle actions cleanly from dynamic entries
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
		return ActionSuccessMsg(fmt.Sprintf("%s action successful", action))
	}
}
