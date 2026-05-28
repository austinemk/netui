// Package wifi for wifi management
package wifi

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/godbus/dbus/v5"
)

type UIState int

const (
	StateNormal UIState = iota
	StateSavedActionsMenu
	StatePasswordInput
)

type DBusClient struct {
	Conn *dbus.Conn
}

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
	Path     dbus.ObjectPath
}

type SavedProfile struct {
	Name        string
	UUID        string
	AutoConnect bool
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
	Adapter AdapterInfo
	Saved   []SavedProfile
	APs     []AccessPoint
}

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		conn, err := dbus.SystemBus()
		if err != nil {
			return ErrMsg(err)
		}
		m.Client.Conn = conn

		adapter, err := GetAdapterSettings(m.Client)
		if err != nil {
			return ErrMsg(err)
		}
		saved, err := GetSavedProfiles(m.Client)
		if err != nil {
			return ErrMsg(err)
		}
		aps, _ := GetActiveAccessPoints(m.Client)

		return InfoLoadedMsg(InfoLoadedData{Adapter: adapter, Saved: saved, APs: aps})
	}
}
