// Package bluetooth for managing bluetooth services
package bluetooth

import (
	"time"

	"netui/components"

	tea "github.com/charmbracelet/bubbletea"
)

type (
	DevicesLoadedMsg     []Device
	ScanStartedMsg       struct{}
	ScanStoppedMsg       struct{}
	AdapterToggledMsg    struct{}
	TickMsg              time.Time
	ErrMsg               error
	ActionSuccessMsg     string
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
}

func New() Model {
	return Model{Scanning: false, PopupMenu: components.NewOptionsPopup("", []string{})}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(FetchDevicesCmd(), FetchAdapterInfoCmd())
}

func CleanBluetooth(m Model) bool {
	if m.Scanning {
		StopScanCmd()
	}
	return true
}
