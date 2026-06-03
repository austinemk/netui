package wifi

import (
	"corntui/config"

	tea "charm.land/bubbletea/v2"
)

// --- Dedicated Handler Functions ---

// V2 Fix: Changed parameter type and type assertion to tea.KeyPressMsg
func (m Model) handlePasswordInput(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyPressMsg)
	if !ok {
		var cmd tea.Cmd
		m.PassInput, cmd = m.PassInput.Update(msg)
		return m, cmd
	}

	switch keyMsg.String() {
	case "backspace", "esc":
		m.UIState = StateNormal
		m.PassInput.Reset()
		m.Table.SetHeight(config.ListHeight)
		return m, nil

	case "enter":
		passwordValue := m.PassInput.Value()
		cmd := ConnectToAccessPoint(m.Ctx, m.Client, m.SelectedAP, passwordValue)

		m.UIState = StateNormal
		m.Table.SetHeight(config.ListHeight)
		m.PassInput.Reset()
		return m, cmd

	default:
		var cmd tea.Cmd
		m.PassInput, cmd = m.PassInput.Update(msg)
		return m, cmd
	}
}

// V2 Fix: Changed type assertion to tea.KeyPressMsg
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
		m.Table.SetHeight(config.ListHeight)

	case "enter":
		idx := m.Table.Cursor()
		if idx >= 0 && idx < len(m.Saved) {
			prof := m.Saved[idx]
			if m.MenuCursor == 0 {
				cmd = ToggleAutoConnectCmd(m.Client, prof.UUID, !prof.AutoConnect)
			} else {
				cmd = ForgetProfileCmd(m.Client, prof.UUID)
			}
		}
		m.UIState = StateNormal
		m.Table.SetHeight(config.ListHeight)

		return m, cmd
	}
	return m, nil
}

func (m Model) handleInfoLoaded(msg InfoLoadedMsg) (Model, tea.Cmd) {
	m.Client = msg.Client // <-- Persist the initialized client to your state!
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
		return m, TriggerHardwareScanCmd(m.Client)
	}
	return m, nil
}

func (m Model) handleAdapterOrActionSuccess() (Model, tea.Cmd) {
	return m, func() tea.Msg {
		a, _ := GetAdapterSettings(m.Client)
		s, _ := GetSavedProfiles(m.Client)
		aps, _ := GetActiveAccessPoints(m.Client)
		return InfoLoadedMsg(InfoLoadedData{Client: m.Client, Adapter: a, Saved: s, APs: aps})
	}
}

// V2 Fix: Changed argument type to tea.KeyPressMsg
func (m Model) handleKeyInput(msg tea.KeyPressMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "s":
		m.Scanning = !m.Scanning
		m.Table.GotoTop()
		m.syncTableRows()

		if m.Scanning {
			return m, TriggerHardwareScanCmd(m.Client)
		}

	case "p":
		if !m.Scanning {
			return m, TogglePowerCmd(m.Client, !m.Adapter.Enabled)
		}

	case "enter":
		idx := m.Table.Cursor()
		if m.Scanning {
			if len(m.ActiveAPs) == 0 || idx >= len(m.ActiveAPs) {
				return m, nil
			}
			m.SelectedAP = m.ActiveAPs[idx]

			if m.SelectedAP.Security == "open" || IsProfileSaved(m.Saved, m.SelectedAP.SSID) {
				return m, ConnectToAccessPoint(m.Ctx, m.Client, m.SelectedAP, "")
			}
			m.UIState = StatePasswordInput
		} else {
			if len(m.Saved) > 0 && idx < len(m.Saved) {
				m.UIState = StateSavedActionsMenu
				m.SelectedSaved = m.Saved[idx]
				m.MenuCursor = 0
			}
		}
		//m.Table.SetHeight(config.ListHeightHalf)

	default:
		var cmd tea.Cmd
		m.Table, cmd = m.Table.Update(msg)
		m.Cursor = m.Table.Cursor()
		return m, cmd
	}
	return m, nil
}
