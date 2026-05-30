// Package wifi for wifi management
package wifi

import (
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/Wifx/gonetworkmanager/v3"
)

type UIState int

const (
	StateNormal UIState = iota
	StateSavedActionsMenu
	StatePasswordInput
)

type AdapterInfo struct {
	Interface string
	State     string
	Enabled   bool
}

type AccessPoint struct {
	SSID     string
	Strength uint8
	Security string
	IsActive bool
	AP       gonetworkmanager.AccessPoint
}

type SavedProfile struct {
	Name        string
	UUID        string
	AutoConnect bool
	Settings    gonetworkmanager.Connection
}

// Bubble Tea Message Definitions
type (
	InfoLoadedMsg     InfoLoadedData
	ScanFinishedMsg   []AccessPoint
	TickMsg           time.Time
	ErrMsg            error
	AdapterToggledMsg struct{}
	ActionSuccessMsg  string
)

type InfoLoadedData struct {
	Client  gonetworkmanager.NetworkManager // <-- Add this field
	Adapter AdapterInfo
	Saved   []SavedProfile
	APs     []AccessPoint
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		nm, err := gonetworkmanager.NewNetworkManager()
		if err != nil {
			return ErrMsg(err)
		}

		adapter, err := GetAdapterSettings(nm)
		if err != nil {
			return ErrMsg(err)
		}

		saved, err := GetSavedProfiles(nm)
		if err != nil {
			return ErrMsg(err)
		}

		aps, err := GetActiveAccessPoints(nm)
		if err != nil {
			return ErrMsg(err)
		}

		return InfoLoadedMsg(InfoLoadedData{
			Client:  nm, // <-- Pass it along here
			Adapter: adapter,
			Saved:   saved,
			APs:     aps,
		})
	}
}

// Clean gracefully halts background procedures, context loops, and stops leaks
func (m Model) Clean() {
	// 1. Cancel the background context to instantly terminate the monitor goroutine
	if m.Cancel != nil {
		m.Cancel()
	}

	// 2. Explicitly stop the hardware scanning loop state flag
	m.Scanning = false

	// 3. Clean up inner component state instances
	m.PassInput.Reset()
	m.Table.SetRows(nil)
}
