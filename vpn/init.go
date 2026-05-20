package vpn

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/godbus/dbus/v5"
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
	Conn *dbus.Conn
}

type TunnelProfile struct {
	Name   string
	UUID   string
	Type   string
	Active bool
	Path   dbus.ObjectPath
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
		Client:     &DBusClient{Conn: nil},
		Loading:    true,
		UIState:    StateNormal,
		FormInputs: make(map[FormField]string),
	}
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		conn, err := dbus.SystemBus()
		if err != nil {
			return ErrMsg(err)
		}
		m.Client.Conn = conn

		tunnels, err := GetVPNConnections(m.Client)
		if err != nil {
			return ErrMsg(err)
		}
		return TunnelsLoadedMsg(tunnels)
	}
}
