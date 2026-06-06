package bluetooth

import (
	"fmt"

	"github.com/austinemk/linktui/pkg/bus"

	tea "charm.land/bubbletea/v2"
	"github.com/godbus/dbus/v5"
)

const (
	dBusPropertyDest = "org.freedesktop.DBus.Properties.Get"
	adapterInterface = "org.bluez.Adapter1"
)

func FetchAdapterInfoCmd() tea.Cmd {
	return func() tea.Msg {
		info, err := FetchAdapterInfo()
		if err != nil {
			return ErrMsg(fmt.Errorf("bluetooth adapter not available: %w", err))
		}
		return AdapterInfoLoadedMsg(info)
	}
}

func FetchAdapterInfo() (AdapterInfo, error) {
	conn := bus.Get()
	obj := conn.Object(bluezInterface, adapterPath)

	var powered, discoverable, pairable bool

	var pv dbus.Variant
	if err := obj.Call(dBusPropertyDest, 0, adapterInterface, "Powered").Store(&pv); err == nil {
		powered, _ = pv.Value().(bool)
	}

	var dv dbus.Variant
	if err := obj.Call(dBusPropertyDest, 0, adapterInterface, "Discoverable").Store(&dv); err == nil {
		discoverable, _ = dv.Value().(bool)
	}

	var prv dbus.Variant
	if err := obj.Call(dBusPropertyDest, 0, adapterInterface, "Pairable").Store(&prv); err == nil {
		pairable, _ = prv.Value().(bool)
	}

	return AdapterInfo{
		Powered:      powered,
		Discoverable: discoverable,
		Pairable:     pairable,
	}, nil
}

func ToggleAdapterPropertyCmd(prop string, currentVal bool) tea.Cmd {
	return func() tea.Msg {
		conn := bus.Get()
		obj := conn.Object(bluezInterface, adapterPath)
		_ = obj.Call(
			"org.freedesktop.DBus.Properties.Set", 0,
			"org.bluez.Adapter1", prop, dbus.MakeVariant(!currentVal),
		)
		return AdapterToggledMsg{}
	}
}
