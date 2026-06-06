package wifi

import (
	tea "charm.land/bubbletea/v2"
)

// --- Dedicated Handler Functions ---

func (m Model) handlePasswordInput(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		var cmd tea.Cmd
		m.PassInput, cmd = m.PassInput.Update(msg)
		return m, cmd
	}

	switch keyMsg.String() {
	case "esc":
		m.UIState = StateNormal
		m.PassInput.Reset()
		return m, nil

	case "enter":
		passwordValue := m.PassInput.Value()
		cmd := ConnectToAccessPoint(m.Ctx, m.SelectedAP, passwordValue)

		m.UIState = StateNormal
		m.PassInput.Reset()
		return m, cmd

	default:
		var cmd tea.Cmd
		m.PassInput, cmd = m.PassInput.Update(msg)
		return m, cmd
	}
}

func (m Model) handleSavedActionsMenu(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
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
		if m.MenuCursor < 1 {
			m.MenuCursor++
		}
	case "esc":
		m.UIState = StateNormal

	case "enter":
		idx := m.Table.Cursor()
		if idx >= 0 && idx < len(m.Saved) {
			prof := m.Saved[idx]
			if m.MenuCursor == 0 {
				cmd = ToggleAutoConnectCmd(prof.UUID, !prof.AutoConnect)
			} else {
				cmd = ForgetProfileCmd(prof.UUID)
			}
		}
		m.UIState = StateNormal

		return m, cmd
	}
	return m, nil
}

func (m Model) handleInfoLoaded(msg InfoLoadedMsg) (Model, tea.Cmd) {
	m.Adapter = msg.Adapter
	m.Saved = msg.Saved
	m.ActiveAPs = msg.APs
	m.syncTableRows()
	if m.Scanning {
		return m, PollWifiTicker()
	}
	return m, nil
}

func (m Model) handleScanFinished(msg ScanFinishedMsg) (Model, tea.Cmd) {
	m.ActiveAPs = msg
	m.syncTableRows()
	if m.Scanning {
		return m, PollWifiTicker()
	}
	return m, nil
}

func (m Model) handleTick() (Model, tea.Cmd) {
	if m.Scanning {
		return m, TriggerHardwareScanCmd()
	}
	return m, nil
}

func (m Model) handleAdapterOrActionSuccess() (Model, tea.Cmd) {
	return m, func() tea.Msg {
		a, _ := GetAdapterSettings()
		s, _ := GetSavedProfiles()
		aps, _ := GetActiveAccessPoints()
		return InfoLoadedMsg(InfoLoadedData{Adapter: a, Saved: s, APs: aps})
	}
}

func (m Model) handleKeyInput(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	if !m.NMStatus {
		return m, nil
	}

	switch msg.String() {
	case "s":
		m.Scanning = !m.Scanning
		m.Table.GotoTop()
		m.syncTableRows()

		if m.Scanning {
			return m, TriggerHardwareScanCmd()
		}

	case "p":
		if !m.Scanning {
			return m, TogglePowerCmd(!m.Adapter.Enabled)
		}

	case "enter":
		idx := m.Table.Cursor()
		if m.Scanning {
			if idx < 0 || len(m.ActiveAPs) == 0 || idx >= len(m.ActiveAPs) {
				return m, nil
			}
			m.SelectedAP = m.ActiveAPs[idx]

			if m.SelectedAP.Security == "open" || IsProfileSaved(m.Saved, m.SelectedAP.SSID) {
				return m, ConnectToAccessPoint(m.Ctx, m.SelectedAP, "")
			}
			m.UIState = StatePasswordInput
		} else {
			if idx >= 0 && len(m.Saved) > 0 && idx < len(m.Saved) {
				m.UIState = StateSavedActionsMenu
				m.SelectedSaved = m.Saved[idx]
				m.MenuCursor = 0
			}
		}

	default:
		var cmd tea.Cmd
		m.Table, cmd = m.Table.Update(msg)
		m.Cursor = m.Table.Cursor()
		return m, cmd
	}
	return m, nil
}
