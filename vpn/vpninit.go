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
	// Wrapper instance instead of raw dbus connection
	NM gonetworkmanager.NetworkManager
}

type TunnelProfile struct {
	Name       string
	UUID       string
	Type       string
	Active     bool
	Connection gonetworkmanager.Connection // Wrapper reference
}

type (
	TunnelsLoadedMsg []TunnelProfile
	ActionSuccessMsg string
	ErrMsg           error
)

type Model struct {
	Client     *DBusClient
	Tunnels    []TunnelProfile
	Cursor     int
	MenuCursor int
	UIState    UIState
	Loading    bool
	Err        error

	// Form input states
	ActiveField FormField
	FormInputs  map[FormField]string
}

func New() Model {
	return Model{
		Client:     &DBusClient{NM: nil},
		Loading:    true,
		UIState:    StateNormal,
		FormInputs: make(map[FormField]string),
	}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		nm, err := gonetworkmanager.NewNetworkManager()
		if err != nil {
			return ErrMsg(err)
		}

		// DO NOT do: m.Client.NM = nm here if m is a value receiver.
		// Instead, pass it or handle assignment inside your VPN sub-view's Update()
		// when handling TunnelsLoadedMsg, or handle it inside app.go before Init() runs.

		// Creating a temp client wrapper to fetch data safely
		tempClient := &DBusClient{NM: nm}
		t, err := GetVPNConnections(tempClient)
		if err != nil {
			return ErrMsg(err)
		}

		return TunnelsLoadedMsg(t)
	}
}
