package bluetooth

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/godbus/dbus/v5"
)

func FetchAdapterInfoCmd(client *BlueZClient) tea.Cmd {
	return func() tea.Msg {
		if client == nil || client.Conn == nil {
			return ErrMsg(fmt.Errorf("bluetooth adapter not available"))
		}
		info, err := FetchAdapterInfo(client)
		if err != nil {
			return ErrMsg(err)
		}
		return AdapterInfoLoadedMsg(info)
	}
}

func FetchAdapterInfo(client *BlueZClient) (AdapterInfo, error) {
	obj := client.Conn.Object(bluezInterface, adapterPath)

	var powered, discoverable, pairable bool

	var pVariant dbus.Variant
	if err := obj.Call("org.freedesktop.DBus.Properties.Get", 0, "org.bluez.Adapter1", "Powered").Store(&pVariant); err == nil {
		powered, _ = pVariant.Value().(bool)
	}

	var dVariant dbus.Variant
	if err := obj.Call("org.freedesktop.DBus.Properties.Get", 0, "org.bluez.Adapter1", "Discoverable").Store(&dVariant); err == nil {
		discoverable, _ = dVariant.Value().(bool)
	}

	var prVariant dbus.Variant
	if err := obj.Call("org.freedesktop.DBus.Properties.Get", 0, "org.bluez.Adapter1", "Pairable").Store(&prVariant); err == nil {
		pairable, _ = prVariant.Value().(bool)
	}

	return AdapterInfo{
		Powered:      powered,
		Discoverable: discoverable,
		Pairable:     pairable,
	}, nil
}

func ToggleAdapterPropertyCmd(client *BlueZClient, prop string, currentVal bool) tea.Cmd {
	return func() tea.Msg {
		obj := client.Conn.Object(bluezInterface, adapterPath)

		typedValue := !currentVal
		_ = obj.Call("org.freedesktop.DBus.Properties.Set", 0, "org.bluez.Adapter1", prop, dbus.MakeVariant(typedValue))

		return AdapterToggledMsg{}
	}
}
