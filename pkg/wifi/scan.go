package wifi

import (
	"time"

	"github.com/austinemk/linktui/pkg/bus"

	tea "charm.land/bubbletea/v2"
	"github.com/godbus/dbus/v5"
)

const (
	nmDest           = "org.freedesktop.NetworkManager"
	nmPath           = "/org/freedesktop/NetworkManager"
	nmIface          = "org.freedesktop.NetworkManager"
	nmDeviceIface    = "org.freedesktop.NetworkManager.Device"
	nmWifiIface      = "org.freedesktop.NetworkManager.Device.Wireless"
	nmAPiface        = "org.freedesktop.NetworkManager.AccessPoint"
	nmSettingsSvc    = "org.freedesktop.NetworkManager.Settings"
	nmSettingsPath   = "/org/freedesktop/NetworkManager/Settings"
	nmDeviceTypeWifi = uint32(2)
)

func PollWifiTicker() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func TriggerHardwareScanCmd() tea.Cmd {
	return func() tea.Msg {
		conn := bus.Get()
		wifiPath, err := findWifiDevicePath(conn)
		if err == nil {
			wDev := conn.Object(nmDest, wifiPath)
			// RequestScan with empty options map
			wDev.Call(nmWifiIface+".RequestScan", 0, map[string]dbus.Variant{})
		}

		aps, _ := GetActiveAccessPoints()
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

func findWifiDevicePath(conn *dbus.Conn) (dbus.ObjectPath, error) {
	nm := conn.Object(nmDest, nmPath)
	var devicePaths []dbus.ObjectPath
	call := nm.Call(nmIface+".GetDevices", 0)
	if call.Err != nil {
		return "", call.Err
	}
	if err := call.Store(&devicePaths); err != nil {
		return "", err
	}

	for _, dPath := range devicePaths {
		dev := conn.Object(nmDest, dPath)
		v, err := dev.GetProperty(nmDeviceIface + ".DeviceType")
		if err != nil {
			continue
		}
		if devType, ok := v.Value().(uint32); ok && devType == nmDeviceTypeWifi {
			return dPath, nil
		}
	}
	return "", nil
}

func GetActiveAccessPoints() ([]AccessPoint, error) {
	conn := bus.Get()

	wifiPath, err := findWifiDevicePath(conn)
	if err != nil || wifiPath == "" {
		return nil, err
	}

	wDev := conn.Object(nmDest, wifiPath)

	// Get active AP path
	activeAPv, _ := wDev.GetProperty(nmWifiIface + ".ActiveAccessPoint")
	var activeAPPath string
	if activeAPv.Value() != nil {
		if p, ok := activeAPv.Value().(dbus.ObjectPath); ok {
			activeAPPath = string(p)
		}
	}

	// Get all APs
	var apPaths []dbus.ObjectPath
	call := wDev.Call(nmWifiIface+".GetAllAccessPoints", 0)
	if call.Err != nil {
		return nil, call.Err
	}
	if err := call.Store(&apPaths); err != nil {
		return nil, err
	}

	var list []AccessPoint
	for _, apPath := range apPaths {
		ap := conn.Object(nmDest, apPath)

		ssidV, err := ap.GetProperty(nmAPiface + ".Ssid")
		if err != nil {
			continue
		}
		ssidBytes, ok := ssidV.Value().([]byte)
		if !ok || len(ssidBytes) == 0 {
			continue
		}
		ssid := string(ssidBytes)

		strengthV, _ := ap.GetProperty(nmAPiface + ".Strength")
		strength, _ := strengthV.Value().(uint8)

		wpaV, _ := ap.GetProperty(nmAPiface + ".WpaFlags")
		wpaFlags, _ := wpaV.Value().(uint32)

		rsnV, _ := ap.GetProperty(nmAPiface + ".RsnFlags")
		rsnFlags, _ := rsnV.Value().(uint32)

		sec := "open"
		if wpaFlags > 0 {
			sec = "wpa"
		}
		if rsnFlags > 0 {
			sec = "wpa2"
		}

		list = append(list, AccessPoint{
			SSID:     ssid,
			Strength: strength,
			Security: sec,
			IsActive: string(apPath) == activeAPPath,
			APPath:   apPath,
		})
	}
	return list, nil
}
