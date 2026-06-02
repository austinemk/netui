// Package vpn for vpn stuff management
package vpn

import (
	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/table"
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
	TunnelsLoadedMsg TunnelsLoadedData
	ActionSuccessMsg string
	ErrMsg           error
)

type TunnelsLoadedData struct {
	Tunnels []TunnelProfile
	Client  *DBusClient
}

type Model struct {
	Client     *DBusClient
	Tunnels    []TunnelProfile
	Table      table.Model
	FilePicker filepicker.Model // Integrated Native File Picker Component
	MenuCursor int
	UIState    UIState
	Loading    bool
	Err        error
	Cursor     int

	// Form input states
	ActiveField    FormField
	SelectedTunnel TunnelProfile
	FormInputs     map[FormField]string
}
