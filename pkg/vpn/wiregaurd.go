package vpn

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"os"
	"strings"

	"github.com/austinemk/linktui/pkg/bus"

	tea "charm.land/bubbletea/v2"
	"github.com/godbus/dbus/v5"
)

const (
	nmSettingsDest = "org.freedesktop.NetworkManager"
	nmSettingsPath = "/org/freedesktop/NetworkManager/Settings"
	nmSettingsSvc  = "org.freedesktop.NetworkManager.Settings"
)

func GetVPNConnections() ([]TunnelProfile, error) {
	conn := bus.Get()
	settings := conn.Object(nmSettingsDest, nmSettingsPath)

	// List all connection paths
	var connPaths []dbus.ObjectPath
	call := settings.Call(nmSettingsSvc+".ListConnections", 0)
	if call.Err != nil {
		return nil, call.Err
	}
	if err := call.Store(&connPaths); err != nil {
		return nil, err
	}

	// Get active connection UUIDs
	nm := conn.Object(nmSettingsDest, nmPath)
	activeUUIDs := make(map[string]bool)
	v, err := nm.GetProperty(nmIface + ".ActiveConnections")
	if err == nil {
		if activePaths, ok := v.Value().([]dbus.ObjectPath); ok {
			for _, aPath := range activePaths {
				aObj := conn.Object(nmSettingsDest, aPath)
				uv, err := aObj.GetProperty(nmConnIface + ".Uuid")
				if err == nil {
					if uuid, ok := uv.Value().(string); ok && uuid != "" {
						activeUUIDs[uuid] = true
					}
				}
			}
		}
	}

	var tunnels []TunnelProfile
	for _, cPath := range connPaths {
		cObj := conn.Object(nmSettingsDest, cPath)

		var sMap map[string]map[string]dbus.Variant
		call := cObj.Call(nmSettingsIface+".GetSettings", 0)
		if call.Err != nil {
			continue
		}
		if err := call.Store(&sMap); err != nil {
			continue
		}

		connSettings, ok := sMap["connection"]
		if !ok {
			continue
		}

		cType, _ := connSettings["type"].Value().(string)
		cUUID, _ := connSettings["uuid"].Value().(string)
		cName, _ := connSettings["id"].Value().(string)

		if cType == "wireguard" || cType == "vpn" {
			tunnels = append(tunnels, TunnelProfile{
				Name:           cName,
				UUID:           cUUID,
				Type:           strings.ToUpper(cType),
				Active:         activeUUIDs[cUUID],
				ConnectionPath: cPath,
			})
		}
	}

	return tunnels, nil
}

func generateUUID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func CreateWireGuardProfileCmd(inputs map[FormField]string) tea.Cmd {
	return func() tea.Msg {
		conn := bus.Get()
		settings := conn.Object(nmSettingsDest, nmSettingsPath)

		uuid := generateUUID()
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

		if inputs[FieldPeerPublicKey] != "" && inputs[FieldPeerEndpoint] != "" {
			peer := map[string]dbus.Variant{
				"public-key": dbus.MakeVariant(inputs[FieldPeerPublicKey]),
				"endpoint":   dbus.MakeVariant(inputs[FieldPeerEndpoint]),
			}
			connectionSettings["wireguard"]["peers"] = dbus.MakeVariant([]map[string]dbus.Variant{peer})
		}

		call := settings.Call(nmSettingsSvc+".AddConnection", 0, connectionSettings)
		if call.Err != nil {
			return ErrMsg(fmt.Errorf("profile write rejection: %v", call.Err))
		}

		return ActionSuccessMsg("WireGuard Profile Created successfully!")
	}
}

// ImportWireGuardFileCmd reads a local config file and submits it to NetworkManager
func ImportWireGuardFileCmd(path string) tea.Cmd {
	return func() tea.Msg {
		file, err := os.Open(path)
		if err != nil {
			return ErrMsg(fmt.Errorf("failed to open file: %v", err))
		}
		defer file.Close()

		inputs := make(map[FormField]string)
		info, err := file.Stat()
		if err == nil {
			name := info.Name()
			if idx := strings.LastIndex(name, "."); idx != -1 {
				name = name[:idx]
			}
			inputs[FieldProfileName] = name
			inputs[FieldInterfaceName] = name
		}

		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "#") || line == "" {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.ToLower(strings.TrimSpace(parts[0]))
			val := strings.TrimSpace(parts[1])

			switch key {
			case "privatekey":
				inputs[FieldPrivateKey] = val
			case "publickey":
				inputs[FieldPeerPublicKey] = val
			case "endpoint":
				inputs[FieldPeerEndpoint] = val
			}
		}

		if inputs[FieldPrivateKey] == "" {
			return ErrMsg(fmt.Errorf("invalid config file: missing PrivateKey"))
		}

		return CreateWireGuardProfileCmd(inputs)()
	}
}
