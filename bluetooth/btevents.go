package bluetooth

import (
	"math"

	"netui/config"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
)

// --- Dedicated Handler Functions ---

// handleActionsMenu handles inputs when the context popup menu is active
func (m Model) handleActionsMenu(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil // No longer using *m
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
		m.Table.SetHeight(int(math.Floor(config.TabBodyHeight * 0.8)))

	case "enter":
		if m.MenuCursor >= 0 && m.MenuCursor < len(m.MenuOptions) {
			action := m.MenuOptions[m.MenuCursor]
			cmd = ExecuteActionCmd(m.Client, action, m.SelectedMac)
		}
		m.UIState = StateNormal
		return m, cmd
	}
	return m, nil
}

// handleDevicesLoaded processes paired updates from DBus
func (m Model) handleDevicesLoaded(msg PairedDevicesLoadedMsg) (Model, tea.Cmd) {
	m.Devices = []Device(msg)
	m = m.syncTableRows() // Updates and captures the modified layout copy
	return m, nil
}

// handleDiscoveredLoaded processes active radio scanning discoveries from DBus
func (m Model) handleDiscoveredLoaded(msg DiscoveredDevicesLoadedMsg) (Model, tea.Cmd) {
	m.Devices = []Device(msg)
	m = m.syncTableRows()

	return m, nil
}

// handleAdapterInfoLoaded updates active system radios configurations
func (m Model) handleAdapterInfoLoaded(info AdapterInfo) (Model, tea.Cmd) {
	m.Powered = info.Powered
	m.Discoverable = info.Discoverable
	m.Pairable = info.Pairable
	return m, nil
}

// handleScanStopped changes states back to standard rendering tracks
func (m Model) handleScanStopped() (Model, tea.Cmd) {
	m.Scanning = false
	return m, LoadPairedDevicesCmd(m.Client)
}

// handleTick is called by our background poll routine to constantly request objects
func (m Model) handleTick() (Model, tea.Cmd) {
	if m.Scanning {
		return m, tea.Batch(DiscoverDevicesCmd(m.Client), PollBluetoothTicker())
	}
	return m, tea.Batch(LoadPairedDevicesCmd(m.Client), PollBluetoothTicker())
}

// handleNormalStateNavigation acts on user interaction strings inside primary frames
func (m Model) handleNormalStateNavigation(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "s", " ":
		if m.Scanning {
			return m, StopScanCmd(m.Client)
		}
		return m, StartScanCmd(m.Client)

	case "p":
		return m, ToggleAdapterPropertyCmd(m.Client, "Powered", !m.Powered)

	case "d":
		return m, ToggleAdapterPropertyCmd(m.Client, "Discoverable", !m.Discoverable)

	case "enter":
		visibleDevices := m.getFilteredDevices()
		selectedIdx := m.Table.Cursor()

		if selectedIdx < 0 || selectedIdx >= len(visibleDevices) {
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
	return m, nil
}

// getFilteredDevices returns the direct items array because separation logic happens inside commands
func (m Model) getFilteredDevices() []Device {
	return m.Devices
}

// syncTableRows maps data structures onto bubbletea table UI dimensions
// Note: Changed from pointer receiver to value receiver returning Model
func (m Model) syncTableRows() Model {
	visibleDevices := m.getFilteredDevices()
	var rows []table.Row

	// DO NOT clear or reset SetColumns inside here to prevent visual glitching

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
	return m // Return the mutated copy back
}
