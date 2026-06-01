// Package bluetooth for managing bluetooth services
package bluetooth

import (
	"time"

	"charm.land/bubbles/v2/table"
	"github.com/godbus/dbus/v5"
)

type UIState int

const (
	StateNormal UIState = iota
	StateActionsMenu
)

type Device struct {
	MAC       string
	Name      string
	Connected bool
	Paired    bool
	Trusted   bool
	Icon      string
}

type AdapterInfo struct {
	Powered      bool
	Discoverable bool
	Pairable     bool
}

const (
	bluezInterface = "org.bluez"
	adapterPath    = "/org/bluez/hci0"
)

type BlueZClient struct {
	Conn *dbus.Conn
}

type (
	InfoLoadedMsg        InfoLoadedData
	DiscoveryStoppedMsg  struct{}
	AdapterToggledMsg    struct{}
	ScanFinishedMsg      []Device
	TickMsg              time.Time
	ErrMsg               error
	ActionSuccessMsg     string
	AdapterInfoLoadedMsg AdapterInfo
)

type InfoLoadedData struct {
	Client  *BlueZClient
	Adapter AdapterInfo
	Devices []Device
}

type Model struct {
	Client      *BlueZClient
	Adapter     AdapterInfo
	Devices     []Device
	Table       table.Model
	Cursor      int
	MenuCursor  int
	UIState     UIState
	Scanning    bool
	Err         error
	SelectedMac string
	SelectedDev Device
	MenuOptions []string
}
