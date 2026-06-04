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

type TunnelProfile struct {
	Name          string
	UUID          string
	Type          string
	InterfaceName string // kernel interface name, e.g. "wg0", "vpn0", "proton"
	Active        bool
	Connection    gonetworkmanager.Connection
}

// IPInfo holds the public IP and optional location details.
// Nil means not yet fetched. Location fields are empty until the user presses p.
type IPInfo struct {
	PublicIP string
	Country  string
	Region   string
	City     string
	ISP      string
}

type (
	TunnelsLoadedMsg TunnelsLoadedData
	ActionSuccessMsg string
	ErrMsg           error
	ClearLogMsg      struct{ ID uint64 }
	IPInfoMsg        *IPInfo
)

type TunnelsLoadedData struct {
	Tunnels []TunnelProfile
	Client  gonetworkmanager.NetworkManager
}

type Model struct {
	Client     gonetworkmanager.NetworkManager
	Tunnels    []TunnelProfile
	Table      table.Model
	FilePicker filepicker.Model // Integrated Native File Picker Component
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
