// Package vpn for vpn stuff management
package vpn

import (
	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/table"
	"github.com/godbus/dbus/v5"
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

type TunnelProfile struct {
	Name           string
	UUID           string
	Type           string
	InterfaceName  string
	Active         bool
	ConnectionPath dbus.ObjectPath
}

// IPInfo holds the public IP and optional location details.
type IPInfo struct {
	PublicIP string
	Country  string
	Region   string
	City     string
	ISP      string
}

type (
	NMStatusMsg      bool
	TunnelsLoadedMsg TunnelsLoadedData
	ActionSuccessMsg string
	ErrMsg           error
	ClearLogMsg      struct{ ID uint64 }
	IPInfoMsg        *IPInfo
)

type TunnelsLoadedData struct {
	Tunnels []TunnelProfile
}

type Model struct {
	NMStatus   bool
	Tunnels    []TunnelProfile
	Table      table.Model
	FilePicker filepicker.Model
	MenuCursor int
	UIState    UIState
	Err        error
	LogID      uint64
	Cursor     int

	// IP display
	IPInfo *IPInfo

	// Form input states
	ActiveField    FormField
	SelectedTunnel TunnelProfile
	FormInputs     map[FormField]string
}
