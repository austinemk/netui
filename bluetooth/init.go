// Package bluetooth for managing bluetooth services
package bluetooth

import (
	"time"

	"netui/components"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

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

type Model struct {
	Devices  []Device
	Cursor   int
	Scanning bool
	Err      error

	// Adapter current states
	Powered      bool
	Discoverable bool
	Pairable     bool

	// Embedded context options menu
	PopupMenu   components.OptionsPopupModel
	SelectedMac string

	// Scrolling Components
	Table    table.Model
	Viewport viewport.Model
}

func New() Model {
	// Initialize columns for data grid structure
	columns := []table.Column{
		{Title: " ", Width: 3},
		{Title: "Device Name", Width: 30},
		{Title: "MAC Address", Width: 20},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
	)

	// Custom row selection styling
	s := table.DefaultStyles()
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(true)
	t.SetStyles(s)

	return Model{
		Scanning:  false,
		PopupMenu: components.NewOptionsPopup("", []string{}),
		Table:     t,
		Viewport:  viewport.New(0, 0),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(FetchDevicesCmd(), FetchAdapterInfoCmd())
}

func CleanBluetooth(m Model) bool {
	if m.Scanning {
		_ = ControlScan(false)
	}
	return true
}
