// Package bluetooth for managing bluetooth services
package bluetooth

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
	adapterPath    = "/org/bluez/hci0" // Default bluetooth adapter path
)

// Helper to grab a quick handle on the System Bus
func getSystemBus() (*dbus.Conn, error) {
	return dbus.ConnectSystemBus()
}

type (
	DevicesLoadedMsg     []Device
	ScanStartedMsg       struct{}
	ScanStoppedMsg       struct{}
	AdapterToggledMsg    struct{}
	TickMsg              time.Time
	ErrMsg               error
	ActionSuccessMsg     string // String type matching declaration in device.go
	AdapterInfoLoadedMsg AdapterInfo
)

func (m Model) Init() tea.Cmd {
	return tea.Batch(FetchDevicesCmd(), FetchAdapterInfoCmd())
}

func CleanBluetooth(m Model) bool {
	if m.Scanning {
		_ = ControlScan(false)
	}
	return true
}
