package wifi

import (
	"context"
	"fmt"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Wifx/gonetworkmanager/v3"
	"github.com/google/uuid"
)

// ConnectToAccessPoint dispatches to the saved-profile or new-password path.
func ConnectToAccessPoint(ctx context.Context, nm gonetworkmanager.NetworkManager, ap AccessPoint, password string) tea.Cmd {
	return func() tea.Msg {
		device, err := findWifiDevice(nm)
		if err != nil {
			return ErrMsg(err)
		}

		if password == "" {
			return connectSavedProfile(nm, ap, device)
		}
		return connectWithPassword(ctx, nm, ap, password, device)
	}
}

// findWifiDevice returns the first Wi-Fi device found, or an error.
func findWifiDevice(nm gonetworkmanager.NetworkManager) (gonetworkmanager.Device, error) {
	devices, err := nm.GetDevices()
	if err != nil {
		return nil, err
	}

	for _, dev := range devices {
		devType, _ := dev.GetPropertyDeviceType()
		if devType == gonetworkmanager.NmDeviceTypeWifi {
			return dev, nil
		}
	}

	return nil, fmt.Errorf("no Wi-Fi device found")
}

// connectSavedProfile activates an existing saved profile by SSID.
func connectSavedProfile(nm gonetworkmanager.NetworkManager, ap AccessPoint, device gonetworkmanager.Device) tea.Msg {
	settings, err := gonetworkmanager.NewSettings()
	if err != nil {
		return ErrMsg(err)
	}

	connections, err := settings.ListConnections()
	if err != nil {
		return ErrMsg(err)
	}

	var matched gonetworkmanager.Connection
	for _, conn := range connections {
		sMap, err := conn.GetSettings()
		if err != nil {
			continue
		}
		if cSettings, ok := sMap["connection"]; ok {
			if cSettings["id"] == ap.SSID {
				matched = conn
				break
			}
		}
	}

	if matched == nil {
		return ErrMsg(fmt.Errorf("no saved profile found for %q", ap.SSID))
	}

	if _, err = nm.ActivateConnection(matched, device, nil); err != nil {
		return ErrMsg(err)
	}

	return ActionSuccessMsg("Connecting to " + ap.SSID + "...")
}

// connectWithPassword creates a new profile, activates it, then monitors the
// result — deleting the profile if authentication fails.
func connectWithPassword(ctx context.Context, nm gonetworkmanager.NetworkManager, ap AccessPoint, password string, device gonetworkmanager.Device) tea.Msg {
	newUUID, _ := uuid.NewUUID()

	connectionSettings := map[string]map[string]interface{}{
		"connection": {
			"id":          ap.SSID,
			"type":        "802-11-wireless",
			"uuid":        newUUID.String(),
			"autoconnect": true,
		},
		"802-11-wireless": {
			"ssid": []byte(ap.SSID),
			"mode": "infrastructure",
		},
		"802-11-wireless-security": {
			"key-mgmt": "wpa-psk",
			"psk":      password,
		},
		"ipv4": {"method": "auto"},
		"ipv6": {"method": "ignore"},
	}

	// AddAndActivateConnection persists immediately — get the connection handle
	activeConn, err := nm.AddAndActivateConnection(connectionSettings, device)
	if err != nil {
		return ErrMsg(err)
	}

	// Get the saved connection object before monitoring, so we can delete it on failure
	conn, err := activeConn.GetPropertyConnection()
	if err != nil {
		return ErrMsg(err)
	}

	return monitorConnectionState(ctx, activeConn, conn, ap.SSID)
}

// monitorConnectionState polls NM until the connection activates, fails, or
// times out — returning a tea.Msg so every outcome reaches Update().
// On any failure it deletes the just-created profile so nothing is saved.
func monitorConnectionState(ctx context.Context, activeConn gonetworkmanager.ActiveConnection, conn gonetworkmanager.Connection, ssid string) tea.Msg {
	const maxAttempts = 15
	for i := 0; i < maxAttempts; i++ {
		select {
		case <-ctx.Done():
			_ = conn.Delete()
			return nil
		case <-time.After(1 * time.Second):
		}

		state, err := activeConn.GetPropertyState()
		if err != nil {
			_ = conn.Delete()
			return ErrMsg(fmt.Errorf("lost connection to NetworkManager: %w", err))
		}

		switch state {
		case gonetworkmanager.NmActiveConnectionStateActivated:
			return ActionSuccessMsg("Connected!")

		case gonetworkmanager.NmActiveConnectionStateDeactivating,
			gonetworkmanager.NmActiveConnectionStateDeactivated:
			_ = conn.Delete()
			return ErrMsg(fmt.Errorf("wrong password for %q — profile not saved", ssid))
		}
	}

	_ = conn.Delete()
	return ErrMsg(fmt.Errorf("timed out connecting — check password and try again"))
}
