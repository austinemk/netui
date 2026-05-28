package bluetooth

import (
	"netui/config"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

// --- Dedicated Handler Functions ---

func (m Model) handleActionsMenu(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	var cmd tea.Cmd
	switch keyMsg.String() {
	case "up", "k":
		if m.MenuCursor > 0 {
			m.MenuCursor--
		} else {
			m.MenuCursor = len(m.MenuOptions) - 1
		}
	case "down", "j":
		if m.MenuCursor < len(m.MenuOptions)-1 {
			m.MenuCursor++
		} else {
			m.MenuCursor = 0
		}
	case "esc":
		m.UIState = StateNormal
	case "enter":
		if m.MenuCursor >= 0 && m.MenuCursor < len(m.MenuOptions) {
			action := m.MenuOptions[m.MenuCursor]
			cmd = ExecuteActionCmd(action, m.SelectedMac)
		}
		m.UIState = StateNormal
		return m, cmd
	}
	return m, nil
}

func (m Model) handleWindowSize(msg tea.WindowSizeMsg) (Model, tea.Cmd) {
	// Directly constraint size maps inside the table component
	m.Table.SetWidth(config.WindowWidth - 4)
	tableHeight := config.WindowHeight - 12
	m.Table.SetHeight(max(tableHeight, 5))

	// Refresh rows layout structure matching new dimensional width bounds
	m.syncTableRows()
	return m, nil
}

func (m Model) handleDevicesLoaded(msg DevicesLoadedMsg) (Model, tea.Cmd) {
	m.Devices = msg
	m.syncTableRows()
	if m.Scanning {
		return m, PollDevicesTicker()
	}
	return m, nil
}

func (m Model) handleTick() (Model, tea.Cmd) {
	if m.Scanning {
		return m, FetchDevicesCmd()
	}
	return m, nil
}

func (m Model) handleScanStopped() (Model, tea.Cmd) {
	m.Scanning = false
	m.Table.GotoTop()
	m.Cursor = m.Table.Cursor()
	return m, FetchDevicesCmd()
}

func (m *Model) handleAdapterInfoLoaded(msg AdapterInfoLoadedMsg) {
	m.Powered = msg.Powered
	m.Discoverable = msg.Discoverable
	m.Pairable = msg.Pairable
}

func (m Model) handleKeyInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "s":
		if m.Scanning {
			return m, StopScanCmd()
		} else {
			m.Scanning = true
			m.Table.GotoTop()
			m.Cursor = m.Table.Cursor()
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
		return m.handleEnterKey()

	default:
		var cmd tea.Cmd
		m.Table, cmd = m.Table.Update(msg)
		m.Cursor = m.Table.Cursor()
		return m, cmd
	}
	return m, nil
}

func (m Model) handleEnterKey() (Model, tea.Cmd) {
	visibleDevices := m.getFilteredDevices()
	selectedIdx := m.Table.Cursor()
	if len(visibleDevices) == 0 || selectedIdx >= len(visibleDevices) {
		return m, nil
	}

	targetDev := visibleDevices[selectedIdx]
	m.SelectedMac = targetDev.MAC
	m.SelectedDev = targetDev

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

	m.MenuOptions = opts
	m.MenuCursor = 0
	m.UIState = StateActionsMenu
	return m, nil
}

func (m Model) getFilteredDevices() []Device {
	return m.Devices
}

func (m *Model) syncTableRows() {
	visibleDevices := m.getFilteredDevices()
	var rows []table.Row

	wdth := m.Table.Width()
	if wdth <= 0 {
		wdth = config.WindowWidth - 4
		if wdth <= 0 {
			wdth = 50 // Safe fallback minimum initialization width bounds
		}
	}

	m.Table.SetColumns([]table.Column{
		{Title: "Status", Width: wdth / 8},
		{Title: "Device Name", Width: (wdth * 3) / 5},
		{Title: "MAC Address", Width: wdth / 4},
	})

	for _, dev := range visibleDevices {
		statusIcon := "󰂯"
		if dev.Icon != "" {
			statusIcon = dev.Icon
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

	if m.Table.Cursor() >= len(rows) {
		m.Table.GotoTop()
		m.Cursor = m.Table.Cursor()
	}
}
