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

func GetActiveAccessPoints(nm gonetworkmanager.NetworkManager) ([]AccessPoint, error) {
	devices, err := nm.GetDevices()
	if err != nil {
		return nil, err
	}

	var wDev gonetworkmanager.DeviceWireless
	for _, dev := range devices {
		devType, _ := dev.GetPropertyDeviceType()
		if devType == gonetworkmanager.NmDeviceTypeWifi {
			wDev, err = gonetworkmanager.NewDeviceWireless(dev.GetPath())
			break
		}
	}

	if wDev == nil {
		return nil, nil
	}

	apPaths, err := wDev.GetAllAccessPoints()
	if err != nil {
		return nil, err
	}

	activeAp, _ := wDev.GetPropertyActiveAccessPoint()
	var activeApPath string
	if activeAp != nil {
		activeApPath = string(activeAp.GetPath())
	}

	var list []AccessPoint
	for _, ap := range apPaths {
		ssid, _ := ap.GetPropertySSID()
		if ssid == "" {
			continue
		}
		strength, _ := ap.GetPropertyStrength()
		wpaFlags, _ := ap.GetPropertyWPAFlags()
		rsnFlags, _ := ap.GetPropertyRSNFlags()

		sec := "open"
		if wpaFlags > 0 {
			sec = "wpa"
		}
		if rsnFlags > 0 {
			sec = "wpa/2"
		}

		list = append(list, AccessPoint{
			SSID:     ssid,
			Strength: strength,
			Security: sec,
			IsActive: string(ap.GetPath()) == activeApPath,
			AP:       ap,
		})
	}
	return list, nil
}
