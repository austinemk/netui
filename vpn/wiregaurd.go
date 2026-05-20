package vpn

import (
	"crypto/rand"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/godbus/dbus/v5"
)

func GetVPNConnections(client *DBusClient) ([]TunnelProfile, error) {
	settings := client.Conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager/Settings")
	var connections []dbus.ObjectPath
	if err := settings.Call("org.freedesktop.NetworkManager.Settings.ListConnections", 0).Store(&connections); err != nil {
		return nil, err
	}

	nm := client.Conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager")
	activeConnsVal, err := nm.GetProperty("org.freedesktop.NetworkManager.ActiveConnections")
	var activePaths []dbus.ObjectPath
	if err == nil {
		activePaths, _ = activeConnsVal.Value().([]dbus.ObjectPath)
	}

	activeUUIDs := make(map[string]bool)
	for _, aPath := range activePaths {
		aObj := client.Conn.Object("org.freedesktop.NetworkManager", aPath)
		uuidProp, err := aObj.GetProperty("org.freedesktop.NetworkManager.Connection.Active.Uuid")
		if err == nil {
			if uStr, ok := uuidProp.Value().(string); ok {
				activeUUIDs[uStr] = true
			}
		}
	}

	var tunnels []TunnelProfile
	for _, path := range connections {
		connObj := client.Conn.Object("org.freedesktop.NetworkManager", path)
		var sMap map[string]map[string]dbus.Variant
		if err := connObj.Call("org.freedesktop.NetworkManager.Settings.Connection.GetSettings", 0).Store(&sMap); err != nil {
			continue
		}

		connSettings := sMap["connection"]
		cType, okType := connSettings["type"].Value().(string)

		if okType && (strings.Contains(cType, "vpn") || strings.Contains(cType, "wireguard")) {
			cID, _ := connSettings["id"].Value().(string)
			cUUID, _ := connSettings["uuid"].Value().(string)

			tunnels = append(tunnels, TunnelProfile{
				Name:   cID,
				UUID:   cUUID,
				Type:   cType,
				Active: activeUUIDs[cUUID],
				Path:   path,
			})
		}
	}
	return tunnels, nil
}

// AddWireguardProfileCmd constructs and stores a native NetworkManager WireGuard connection profile
func AddWireguardProfileCmd(client *DBusClient, inputs map[FormField]string) func() tea.Msg {
	return func() tea.Msg {
		settings := client.Conn.Object("org.freedesktop.NetworkManager", "/org/freedesktop/NetworkManager/Settings")

		uuid, err := generateRandomUUID()
		if err != nil {
			return ErrMsg(err)
		}

		// Prepare standard outer properties dictionary
		connectionSettings := map[string]map[string]dbus.Variant{
			"connection": {
				"id":             dbus.MakeVariant(inputs[FieldProfileName]),
				"uuid":           dbus.MakeVariant(uuid),
				"type":           dbus.MakeVariant("wireguard"),
				"interface-name": dbus.MakeVariant(inputs[FieldInterfaceName]),
				"autoconnect":    dbus.MakeVariant(false),
			},
			"wireguard": {
				"private-key": dbus.MakeVariant(inputs[FieldPrivateKey]),
				"listen-port": dbus.MakeVariant(uint32(51820)),
			},
		}

		// If a peer endpoint is specified, append a standard map breakdown representation
		if inputs[FieldPeerPublicKey] != "" && inputs[FieldPeerEndpoint] != "" {
			peer := map[string]dbus.Variant{
				"public-key": dbus.MakeVariant(inputs[FieldPeerPublicKey]),
				"endpoint":   dbus.MakeVariant(inputs[FieldPeerEndpoint]),
			}
			// NetworkManager expects an array of peer dictionary maps (aa{sv})
			connectionSettings["wireguard"]["peers"] = dbus.MakeVariant([]map[string]dbus.Variant{peer})
		}

		var newConnPath dbus.ObjectPath
		err = settings.Call("org.freedesktop.NetworkManager.Settings.AddConnection", 0, connectionSettings).Store(&newConnPath)
		if err != nil {
			return ErrMsg(fmt.Errorf("D-Bus profile write rejection: %v", err))
		}

		return ActionSuccessMsg("WireGuard Profile Created successfully!")
	}
}

func generateRandomUUID() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:]), nil
}
