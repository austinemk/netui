package vpn

import (
	"bufio"
	"crypto/rand"
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/Wifx/gonetworkmanager/v3"
)

func GetVPNConnections(client *DBusClient) ([]TunnelProfile, error) {
	settings, err := gonetworkmanager.NewSettings()
	if err != nil {
		return nil, err
	}

	connections, err := settings.ListConnections()
	if err != nil {
		return nil, err
	}

	activeConns, err := client.NM.GetPropertyActiveConnections()
	activeUUIDs := make(map[string]bool)
	if err == nil {
		for _, aConn := range activeConns {
			uuid, err := aConn.GetPropertyUUID()
			if err == nil && uuid != "" {
				activeUUIDs[uuid] = true
			}
		}
	}

	var tunnels []TunnelProfile
	for _, conn := range connections {
		sMap, err := conn.GetSettings()
		if err != nil {
			continue
		}

		connSettings, hasConn := sMap["connection"]
		if !hasConn {
			continue
		}

		cType, _ := connSettings["type"].(string)
		cUUID, _ := connSettings["uuid"].(string)
		cName, _ := connSettings["id"].(string)

		if cType == "wireguard" || cType == "vpn" {
			tunnels = append(tunnels, TunnelProfile{
				Name:       cName,
				UUID:       cUUID,
				Type:       strings.ToUpper(cType),
				Active:     activeUUIDs[cUUID],
				Connection: conn,
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

func CreateWireGuardProfileCmd(client *DBusClient, inputs map[FormField]string) tea.Cmd {
	return func() tea.Msg {
		settings, err := gonetworkmanager.NewSettings()
		if err != nil {
			return ErrMsg(err)
		}

		uuid := generateUUID()
		connectionSettings := map[string]map[string]interface{}{
			"connection": {
				"id":             inputs[FieldProfileName],
				"uuid":           uuid,
				"type":           "wireguard",
				"interface-name": inputs[FieldInterfaceName],
				"autoconnect":    false,
			},
			"wireguard": {
				"private-key": inputs[FieldPrivateKey],
				"listen-port": uint32(51820),
			},
		}

		if inputs[FieldPeerPublicKey] != "" && inputs[FieldPeerEndpoint] != "" {
			peer := map[string]interface{}{
				"public-key": inputs[FieldPeerPublicKey],
				"endpoint":   inputs[FieldPeerEndpoint],
			}
			connectionSettings["wireguard"]["peers"] = []map[string]interface{}{peer}
		}

		_, err = settings.AddConnection(connectionSettings)
		if err != nil {
			return ErrMsg(fmt.Errorf("profile write rejection: %v", err))
		}

		return ActionSuccessMsg("WireGuard Profile Created successfully!")
	}
}

// ImportWireGuardFileCmd reads a local config file and submits it to NetworkManager
func ImportWireGuardFileCmd(client *DBusClient, path string) tea.Cmd {
	return func() tea.Msg {
		file, err := os.Open(path)
		if err != nil {
			return ErrMsg(fmt.Errorf("failed to open file: %v", err))
		}
		defer file.Close()

		inputs := make(map[FormField]string)
		// Extract profile name from filename without extension
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

		// Re-use standard profile writer logic using extracted configurations
		return CreateWireGuardProfileCmd(client, inputs)()
	}
}
