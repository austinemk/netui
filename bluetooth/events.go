package bluetooth

import (
	"netui/components"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// 1. Intercept for Active Popup Controls
	if m.PopupMenu.Active {
		m.PopupMenu, cmd = m.PopupMenu.Update(msg)
		if selectMsg, ok := msg.(components.OptionSelectedMsg); ok {
			m.PopupMenu.Active = false
			return m, ExecuteActionCmd(selectMsg.Option, m.SelectedMac)
		}
		return m, cmd
	}

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		// Slicing application height limits gracefully
		m.Viewport.Width = msg.Width
		m.Viewport.Height = msg.Height

		// Table constraints padding limits
		m.Table.SetWidth(msg.Width - 4)
		// Reserve roughly 10 lines for headers, options, and adapters toggles
		tableHeight := msg.Height - 10
		if tableHeight < 3 {
			tableHeight = 3
		}
		m.Table.SetHeight(tableHeight)

	case ScanStartedMsg:
		return m, tea.Batch(FetchDevicesCmd(), PollDevicesTicker())

	case DevicesLoadedMsg:
		m.Devices = msg
		m.syncTableRows() // Hydrate rows with fresh devices
		if m.Scanning {
			return m, PollDevicesTicker()
		}

	case TickMsg:
		if m.Scanning {
			return m, FetchDevicesCmd()
		}

	case ScanStoppedMsg:
		m.Scanning = false
		m.Table.GotoTop() // Safe reset cursor
		return m, FetchDevicesCmd()

	case AdapterInfoLoadedMsg:
		m.Powered = msg.Powered
		m.Discoverable = msg.Discoverable
		m.Pairable = msg.Pairable

	case AdapterToggledMsg:
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
				m.Table.GotoTop()
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

		case "enter":
			visibleDevices := m.getFilteredDevices()
			selectedIdx := m.Table.Cursor()
			if len(visibleDevices) == 0 || selectedIdx >= len(visibleDevices) {
				return m, nil
			}

			targetDev := visibleDevices[selectedIdx]
			m.SelectedMac = targetDev.MAC

			var opts []string
			if targetDev.Connected {
				opts = append(opts, "Disconnect")
			} else {
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

			if !m.Scanning {
				opts = append(opts, "Remove")
			}

			m.PopupMenu = components.NewOptionsPopup(targetDev.Name, opts)
			m.PopupMenu.Active = true
			return m, nil
		}
	}

	// 2. Route messages to the standard table controller (handles arrow highlights)
	m.Table, cmd = m.Table.Update(msg)
	cmds = append(cmds, cmd)

	// 3. Route messages to viewport (handles scrollbars page transitions)
	m.Viewport, cmd = m.Viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// Helper to convert internal device array indices cleanly into standard table rows
func (m *Model) syncTableRows() {
	visibleDevices := m.getFilteredDevices()
	var rows []table.Row

	for _, dev := range visibleDevices {
		statusIcon := IconGenericBluetooth.String()
		if dev.Icon != "" {
			statusIcon = FromString(dev.Icon).String()
		}
		if dev.Connected {
			statusIcon = ""
		}

		rows = append(rows, table.Row{
			statusIcon,
			dev.Name,
			dev.MAC,
		})
	}
	m.Table.SetRows(rows)
}
