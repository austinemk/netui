// Package wifi for wifi management
package wifi

import (
	"time"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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

type Model struct {
	Client    *DBusClient
	Adapter   AdapterInfo
	Saved     []SavedProfile
	ActiveAPs []AccessPoint

	// Dynamic Layout Elements
	Table    table.Model
	Viewport viewport.Model

	// Navigation & Component UI states
	Cursor     int // Kept for backend array mapping compatibility
	MenuCursor int
	UIState    UIState
	Scanning   bool
	Loading    bool
	Err        error

	// Password handling for secured lines
	SelectedAP    AccessPoint
	PasswordInput string
}

func New() Model {
	// Initialize default columns structure
	columns := []table.Column{
		{Title: "Status", Width: 8},
		{Title: "Network Name (SSID)", Width: 26},
		{Title: "Signal", Width: 8},
		{Title: "Security", Width: 12},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
	)

	// Apply beautiful theme defaults
	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)

	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#3B82F6")). // Uses your blue color from Styles.CursorColor
		Bold(true)
	t.SetStyles(s)

	return Model{
		Client:   &DBusClient{Conn: nil},
		Scanning: false,
		Loading:  true,
		UIState:  StateNormal,
		Table:    t,
		Viewport: viewport.New(0, 0),
	}
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
