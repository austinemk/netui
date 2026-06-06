// Package wifi for wifi management
package wifi

import (
	"context"
	"time"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	"github.com/godbus/dbus/v5"
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
	APPath   dbus.ObjectPath
}

type SavedProfile struct {
	Name           string
	UUID           string
	AutoConnect    bool
	ConnectionPath dbus.ObjectPath
}

// Bubble Tea Message Definitions
type (
	NMStatusMsg       bool
	InfoLoadedMsg     InfoLoadedData
	ScanFinishedMsg   []AccessPoint
	TickMsg           time.Time
	ErrMsg            error
	ClearLogMsg       struct{ ID uint64 }
	AdapterToggledMsg struct{}
	ActionSuccessMsg  string
)

type InfoLoadedData struct {
	Adapter AdapterInfo
	Saved   []SavedProfile
	APs     []AccessPoint
}

// model struct
type Model struct {
	NMStatus  bool
	Adapter   AdapterInfo
	Saved     []SavedProfile
	ActiveAPs []AccessPoint

	// Context for graceful cleanup
	Ctx    context.Context
	Cancel context.CancelFunc

	// Dynamic Layout Elements
	Table     table.Model
	PassInput textinput.Model

	// Navigation & Component UI states
	Cursor     int
	MenuCursor int
	UIState    UIState
	Scanning   bool
	Err        error
	LogID      uint64

	// Password handling for secured lines
	SelectedAP    AccessPoint
	SelectedSaved SavedProfile
	PasswordInput string
	MenuOptions   []string
}
