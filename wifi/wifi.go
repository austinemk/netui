// Package wifi for wifi management
package wifi

import (
	"context"
	"time"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
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

// model struct
type Model struct {
	Client    gonetworkmanager.NetworkManager
	Adapter   AdapterInfo
	Saved     []SavedProfile
	ActiveAPs []AccessPoint

	// Context for graceful cleanup
	Ctx    context.Context
	Cancel context.CancelFunc // <-- Add this to track background tasks

	// Dynamic Layout Elements
	Table     table.Model
	PassInput textinput.Model

	// Navigation & Component UI states
	Cursor     int // Kept for backend array mapping compatibility
	MenuCursor int
	UIState    UIState
	Scanning   bool
	Loading    bool
	Err        error

	// Password handling for secured lines
	SelectedAP    AccessPoint
	SelectedSaved SavedProfile
	PasswordInput string
	MenuOptions   []string
}
