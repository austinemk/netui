package wifi

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"github.com/Wifx/gonetworkmanager/v3"
)

func GetAdapterSettings(nm gonetworkmanager.NetworkManager) (AdapterInfo, error) {
	wirelessEnabled, err := nm.GetPropertyWirelessEnabled()
	if err != nil {
		return AdapterInfo{}, err
	}

	devices, err := nm.GetDevices()
	if err != nil {
		return AdapterInfo{}, err
	}

	for _, dev := range devices {
		devType, err := dev.GetPropertyDeviceType()
		if err != nil {
			continue
		}

		if devType == gonetworkmanager.NmDeviceTypeWifi {
			wDev, err := gonetworkmanager.NewDeviceWireless(dev.GetPath())
			if err != nil {
				continue
			}

			iface, _ := wDev.GetPropertyInterface()
			state, _ := wDev.GetPropertyState()

			stateStr := "Disconnected"
			if state == gonetworkmanager.NmDeviceStateActivated {
				stateStr = "Connected"
			}
			return AdapterInfo{
				Interface: iface,
				State:     stateStr,
				Enabled:   wirelessEnabled,
			}, nil
		}
	}
	return AdapterInfo{Interface: "Unknown", State: "Missing", Enabled: false}, fmt.Errorf("no wireless adapter found")
}

func TogglePowerCmd(nm gonetworkmanager.NetworkManager, enable bool) tea.Cmd {
	return func() tea.Msg {
		err := nm.SetPropertyWirelessEnabled(enable)
		if err != nil {
			return ErrMsg(err)
		}
		return AdapterToggledMsg{}
	}
}
