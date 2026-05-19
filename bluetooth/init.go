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

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if m.PopupMenu.Active {
		var cmd tea.Cmd
		m.PopupMenu, cmd = m.PopupMenu.Update(msg)

		if selectMsg, ok := msg.(components.OptionSelectedMsg); ok {
			m.PopupMenu.Active = false
			return m, ExecuteActionCmd(selectMsg.Option, m.SelectedMac)
		}
		return m, cmd
	}

	switch msg := msg.(type) {

	case ScanStartedMsg:
		// BlueZ has confirmed discovery is active — safe to start polling now
		return m, tea.Batch(FetchDevicesCmd(), PollDevicesTicker())

	case DevicesLoadedMsg:
		m.Devices = msg
		// FIXED: If we are actively scanning, keep the ticker sequence looping
		if m.Scanning {
			return m, PollDevicesTicker()
		}

	case TickMsg:
		if m.Scanning {
			// FIXED: The tick message now requests fresh D-Bus state data.
			// This tells BlueZ to feed newly discovered signals directly back into DevicesLoadedMsg.
			return m, FetchDevicesCmd()
		}

	case ScanStoppedMsg:
		m.Scanning = false
		m.Cursor = 0
		// Fetch one final time to solidify state listings locally
		return m, FetchDevicesCmd()

	case AdapterInfoLoadedMsg:
		m.Powered = msg.Powered
		m.Discoverable = msg.Discoverable
		m.Pairable = msg.Pairable

	case AdapterToggledMsg:
		// Re-fetch adapter state so Powered/Discoverable/Pairable update in the UI
		return m, FetchAdapterInfoCmd()

	case ActionSuccessMsg:
		m.Err = nil
		return m, tea.Batch(FetchDevicesCmd(), FetchAdapterInfoCmd())

	case ErrMsg:
		m.Err = msg

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			if m.Scanning {
				_ = ControlScan(false)
			}
			return m, tea.Quit

		case "s":
			if m.Scanning {
				return m, StopScanCmd()
			} else {
				m.Scanning = true
				m.Cursor = 0
				m.Err = nil
				return m, StartScanCmd()
			}

		case "p":
			if !m.Scanning {
				return m, ToggleAdapterPropertyCmd("Powered", m.Powered)
			}
		case "d":
			if !m.Scanning {
				return m, ToggleAdapterPropertyCmd("Discoverable", m.Discoverable)
			}
		case "b":
			if !m.Scanning {
				return m, ToggleAdapterPropertyCmd("Pairable", m.Pairable)
			}

		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			filteredCount := len(m.getFilteredDevices())
			if m.Cursor < filteredCount-1 {
				m.Cursor++
			}

		case "enter":
			visibleDevices := m.getFilteredDevices()
			if len(visibleDevices) == 0 || m.Cursor >= len(visibleDevices) {
				return m, nil
			}

			targetDev := visibleDevices[m.Cursor]
			m.SelectedMac = targetDev.MAC

			var opts []string
			if targetDev.Connected {
				opts = append(opts, "Disconnect")
			} else {
				// Show contextual pairing options conditionally
				if !targetDev.Paired {
					opts = append(opts, "Pair")
				} else {
					opts = append(opts, "Connect")
				}
			}

			if targetDev.Trusted {
				opts = append(opts, "Distrust")
			} else {
				opts = append(opts, "Trust")
			}

			// Add destructive option only in Offline Saved Mode to avoid accidents
			if !m.Scanning {
				opts = append(opts, "Remove")
			}

			m.PopupMenu = components.NewOptionsPopup(targetDev.Name, opts)
			m.PopupMenu.Active = true
		}
	}

	return m, nil
}
