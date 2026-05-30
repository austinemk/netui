package wifi

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Wifx/gonetworkmanager/v3"
)

func PollWifiTicker() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func TriggerHardwareScanCmd(nm gonetworkmanager.NetworkManager) tea.Cmd {
	return func() tea.Msg {
		devices, err := nm.GetDevices()
		if err == nil {
			for _, dev := range devices {
				devType, _ := dev.GetPropertyDeviceType()
				if devType == gonetworkmanager.NmDeviceTypeWifi {
					wDev, err := gonetworkmanager.NewDeviceWireless(dev.GetPath())
					if err == nil {
						// Clean programmatic trigger instead of raw .Call()
						_ = wDev.RequestScan()
					}
					break
				}
			}
		}
		aps, _ := GetActiveAccessPoints(nm)
		return ScanFinishedMsg(aps)
	}
}

func IsProfileSaved(profiles []SavedProfile, ssid string) bool {
	for _, p := range profiles {
		if p.Name == ssid {
			return true
		}
	}
	return false
}
