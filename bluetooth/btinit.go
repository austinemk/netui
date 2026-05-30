// Package bluetooth for managing bluetooth services
package bluetooth

import (
	"time"

	tea "charm.land/bubbletea/v2"
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

func NewBlueZClient() (*BlueZClient, error) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return nil, err
	}
	return &BlueZClient{Conn: conn}, nil
}

type (
	PairedDevicesLoadedMsg     []Device // Clean Track A
	DiscoveredDevicesLoadedMsg []Device // Clean Track B
	ScanStartedMsg             struct{}
	ScanStoppedMsg             struct{}
	AdapterToggledMsg          struct{}
	TickMsg                    time.Time
	ErrMsg                     error
	ActionSuccessMsg           string
	AdapterInfoLoadedMsg       AdapterInfo
)

func (m Model) Init() tea.Cmd {
	if m.Client == nil {
		return nil
	}
	return tea.Batch(LoadPairedDevicesCmd(m.Client), FetchAdapterInfoCmd(m.Client))
}

// Clean gracefully stops any hardware discovery and closes the system bus connection to prevent memory leaks.
func (m Model) Clean() {
	if m.Client == nil || m.Client.Conn == nil {
		return
	}

	// 1. If the hardware is actively discovering devices, tell BlueZ to stop immediately
	if m.Scanning {
		obj := m.Client.Conn.Object(bluezInterface, adapterPath)
		// Send a direct synchronous DBus call to ensure it hits the OS before the binary exits
		_ = obj.Call("org.bluez.Adapter1.StopDiscovery", 0)
	}

	// 2. Close the D-Bus connection completely to clear system RAM and file descriptors
	_ = m.Client.Conn.Close()
}
