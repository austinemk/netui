package vpn

import (
	tea "charm.land/bubbletea/v2"
	"github.com/Wifx/gonetworkmanager/v3"
)

type UIState int

const (
	StateNormal UIState = iota
	StateActionsMenu
	StateAddForm
	StateImportFile
)

type FormField int

const (
	FieldProfileName FormField = iota
	FieldInterfaceName
	FieldPrivateKey
	FieldPeerEndpoint
	FieldPeerPublicKey
	FieldDone
)

type DBusClient struct {
	NM gonetworkmanager.NetworkManager
}

type TunnelProfile struct {
	Name       string
	UUID       string
	Type       string
	Active     bool
	Connection gonetworkmanager.Connection
}

type (
	TunnelsLoadedMsg []TunnelProfile
	ActionSuccessMsg string
	ErrMsg           error
)

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.FilePicker.Init(), // Provisions directory lookup internal commands cleanly
		func() tea.Msg {
			nm, err := gonetworkmanager.NewNetworkManager()
			if err != nil {
				return ErrMsg(err)
			}

			tempClient := &DBusClient{NM: nm}
			t, err := GetVPNConnections(tempClient)
			if err != nil {
				return ErrMsg(err)
			}

			return TunnelsLoadedMsg(t)
		},
	)
}
