package wifi

import (
	"context"
	"fmt"
	"time"

	"github.com/austinemk/linktui/pkg/bus"

	tea "charm.land/bubbletea/v2"
	"github.com/godbus/dbus/v5"
	"github.com/google/uuid"
)

const (
	nmActiveConnIface   = "org.freedesktop.NetworkManager.Connection.Active"
	nmActiveStateActiv  = uint32(2) // NM_ACTIVE_CONNECTION_STATE_ACTIVATED
	nmActiveStateDeactv = uint32(4) // NM_ACTIVE_CONNECTION_STATE_DEACTIVATED
	nmActiveStateDeactg = uint32(3) // NM_ACTIVE_CONNECTION_STATE_DEACTIVATING
)

func ConnectToAccessPoint(ctx context.Context, ap AccessPoint, password string) tea.Cmd {
	return func() tea.Msg {
		conn := bus.Get()

		wifiPath, err := findWifiDevicePath(conn)
		if err != nil || wifiPath == "" {
			return ErrMsg(fmt.Errorf("no Wi-Fi device found"))
		}

		if password == "" {
			return connectSavedProfile(conn, ap, wifiPath)
		}
		return connectWithPassword(ctx, conn, ap, password, wifiPath)
	}
}

func connectSavedProfile(conn *dbus.Conn, ap AccessPoint, devicePath dbus.ObjectPath) tea.Msg {
	settings := conn.Object(nmDest, nmSettingsPath)

	var connPaths []dbus.ObjectPath
	call := settings.Call(nmSettingsSvc+".ListConnections", 0)
	if call.Err != nil {
		return ErrMsg(call.Err)
	}
	if err := call.Store(&connPaths); err != nil {
		return ErrMsg(err)
	}

	var matchedPath dbus.ObjectPath
	for _, cPath := range connPaths {
		cObj := conn.Object(nmDest, cPath)
		var sMap map[string]map[string]dbus.Variant
		call := cObj.Call(nmConnSettingsIface+".GetSettings", 0)
		if call.Err != nil {
			continue
		}
		if err := call.Store(&sMap); err != nil {
			continue
		}
		if cSettings, ok := sMap["connection"]; ok {
			if id, _ := cSettings["id"].Value().(string); id == ap.SSID {
				matchedPath = cPath
				break
			}
		}
	}

	if matchedPath == "" {
		return ErrMsg(fmt.Errorf("no saved profile found for %q", ap.SSID))
	}

	nm := conn.Object(nmDest, nmPath)
	call = nm.Call(
		nmIface+".ActivateConnection", 0,
		matchedPath, devicePath, dbus.ObjectPath("/"),
	)
	if call.Err != nil {
		return ErrMsg(call.Err)
	}

	return ActionSuccessMsg("Connecting to " + ap.SSID + "...")
}

func connectWithPassword(ctx context.Context, conn *dbus.Conn, ap AccessPoint, password string, devicePath dbus.ObjectPath) tea.Msg {
	newUUID, _ := uuid.NewUUID()

	connectionSettings := map[string]map[string]dbus.Variant{
		"connection": {
			"id":          dbus.MakeVariant(ap.SSID),
			"type":        dbus.MakeVariant("802-11-wireless"),
			"uuid":        dbus.MakeVariant(newUUID.String()),
			"autoconnect": dbus.MakeVariant(true),
		},
		"802-11-wireless": {
			"ssid": dbus.MakeVariant([]byte(ap.SSID)),
			"mode": dbus.MakeVariant("infrastructure"),
		},
		"802-11-wireless-security": {
			"key-mgmt": dbus.MakeVariant("wpa-psk"),
			"psk":      dbus.MakeVariant(password),
		},
		"ipv4": {"method": dbus.MakeVariant("auto")},
		"ipv6": {"method": dbus.MakeVariant("ignore")},
	}

	nm := conn.Object(nmDest, nmPath)
	var activeConnPath dbus.ObjectPath
	call := nm.Call(
		nmIface+".AddAndActivateConnection", 0,
		connectionSettings, devicePath, dbus.ObjectPath("/"),
	)
	if call.Err != nil {
		return ErrMsg(call.Err)
	}
	// Returns (connection path, active connection path)
	var savedConnPath dbus.ObjectPath
	if err := call.Store(&savedConnPath, &activeConnPath); err != nil {
		return ErrMsg(err)
	}

	return monitorConnectionState(ctx, conn, activeConnPath, savedConnPath, ap.SSID)
}

func monitorConnectionState(ctx context.Context, conn *dbus.Conn, activeConnPath, savedConnPath dbus.ObjectPath, ssid string) tea.Msg {
	deleteProfile := func() {
		if savedConnPath == "" {
			return
		}
		obj := conn.Object(nmDest, savedConnPath)
		obj.Call(nmConnSettingsIface+".Delete", 0)
	}

	const maxAttempts = 15
	for i := 0; i < maxAttempts; i++ {
		select {
		case <-ctx.Done():
			deleteProfile()
			return nil
		case <-time.After(1 * time.Second):
		}

		aObj := conn.Object(nmDest, activeConnPath)
		sv, err := aObj.GetProperty(nmActiveConnIface + ".State")
		if err != nil {
			deleteProfile()
			return ErrMsg(fmt.Errorf("lost connection to NetworkManager: %w", err))
		}

		state, _ := sv.Value().(uint32)
		switch state {
		case nmActiveStateActiv:
			return ActionSuccessMsg("Connected!")
		case nmActiveStateDeactv, nmActiveStateDeactg:
			deleteProfile()
			return ErrMsg(fmt.Errorf("wrong password for %q — profile not saved", ssid))
		}
	}

	deleteProfile()
	return ErrMsg(fmt.Errorf("timed out connecting — check password and try again"))
}
