package vpn

import (
	"crypto/rand"
	"fmt"
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

	// Fetch active connection UUIDs safely using the wrapper
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

		// Filter for wireguard and vpn tunnels
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
		// gonetworkmanager uses map[string]map[string]interface{}
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
