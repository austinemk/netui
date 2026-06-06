package bluetooth

import (
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
	case "esc", "backspace":
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

func (m Model) handlePasskeyPrompt(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		// If another asynchronous background update occurs, don't drop the context
		return m, nil
	}

	switch keyMsg.String() {
	case "left", "h":
		m.MenuCursor = 0 // Yes
	case "right", "l":
		m.MenuCursor = 1 // No
	case "enter":
		if m.ActiveRespChan != nil {
			if m.MenuCursor == 0 {
				m.ActiveRespChan <- true // Signals "Accept" back to Agent.RequestConfirmation
			} else {
				m.ActiveRespChan <- false // Signals "Reject"
			}
		}

		// Reset state frame back to normal layout view
		m.UIState = StateNormal
		m.ActiveRespChan = nil

		// Keep listening for future incoming agent challenges
		return m, ListenForAgentRequests()
	}
	return m, nil
}

func (m Model) handleInfoLoaded(msg InfoLoadedMsg) (Model, tea.Cmd) {
	m.Adapter = msg.Adapter
	m.Devices = msg.Devices
	m.syncTableRows()
	if m.Scanning {
		return m, PollBluetoothTicker()
	}
	return m, nil
}

func (m Model) handleScanFinished(msg ScanFinishedMsg) (Model, tea.Cmd) {
	m.Devices = msg
	m.syncTableRows()
	if m.Scanning {
		return m, PollBluetoothTicker()
	}
	return m, nil
}

// handleTick is called by our background poll routine to constantly request objects
func (m Model) handleTick() (Model, tea.Cmd) {
	if m.Scanning {
		return m, ContinueDiscoveryCmd()
	}
	return m, nil
}

func (m Model) handleAdapterOrActionSuccess() (Model, tea.Cmd) {
	return m, func() tea.Msg {
		a, _ := FetchAdapterInfo()
		d, _ := LoadPairedDevices()
		return InfoLoadedMsg(InfoLoadedData{Adapter: a, Devices: d})
	}
}

// handleNormalStateNavigation acts on user interaction strings inside primary frames
func (m Model) handleKeyPress(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	if !m.BluezStatus {
		return m, nil
	}
	switch msg.String() {
	case "s":
		m.Scanning = !m.Scanning
		m.Devices = nil
		if m.Scanning {
			return m, StartDiscoveryCmd()
		}
		return m, tea.Batch(StopDiscoveryCmd(), LoadPairedDevicesCmd())

	case "p":
		if !m.Scanning {
			return m, ToggleAdapterPropertyCmd("Powered", m.Adapter.Powered)
		}

	case "d":
		return m, ToggleAdapterPropertyCmd("Discoverable", m.Adapter.Discoverable)

	case "b":
		return m, ToggleAdapterPropertyCmd("Pairable", m.Adapter.Pairable)

	case "enter":
		selectedIdx := m.Table.Cursor()

		if selectedIdx < 0 || selectedIdx >= len(m.Devices) {
			return m, nil
		}

		targetDev := m.Devices[selectedIdx]
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

	default:
		var cmd tea.Cmd
		m.Table, cmd = m.Table.Update(msg)
		m.Cursor = m.Table.Cursor()
		return m, cmd

	}
	return m, nil
}
