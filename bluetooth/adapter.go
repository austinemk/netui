package bluetooth

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/godbus/dbus/v5"
)

func FetchAdapterInfoCmd() tea.Cmd {
	return func() tea.Msg {
		info, err := FetchAdapterInfo()
		if err != nil {
			return ErrMsg(err)
		}
		return AdapterInfoLoadedMsg(info)
	}
}

func FetchAdapterInfo() (AdapterInfo, error) {
	conn, err := getSystemBus()
	if err != nil {
		return AdapterInfo{}, err
	}
	defer conn.Close()

	obj := conn.Object(bluezInterface, adapterPath)

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

func ToggleAdapterPropertyCmd(prop string, currentVal bool) tea.Cmd {
	return func() tea.Msg {
		conn, err := getSystemBus()
		if err != nil {
			return ErrMsg(err)
		}
		defer conn.Close()

		obj := conn.Object(bluezInterface, adapterPath)

		// Pass the bool as a typed variable so godbus serialises it as "b"
		// on the wire. Passing !currentVal directly through ...interface{}
		// inside MakeVariant causes godbus to double-wrap the variant into
		// v(v(b)) which BlueZ silently rejects.
		var newVal bool = !currentVal
		err = obj.Call("org.freedesktop.DBus.Properties.Set", 0, "org.bluez.Adapter1", prop, dbus.MakeVariant(newVal)).Err
		if err != nil {
			return ErrMsg(err)
		}

		return AdapterToggledMsg{}
	}
}
