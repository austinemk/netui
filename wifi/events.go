package wifi

import (
	"math"

	"netui/components"
	"netui/config"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

// --- Dedicated Handler Functions ---

func (m Model) handlePasswordInput(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		// Still allow system messages to pass through to the input if needed
		var cmd tea.Cmd
		m.PassInput, cmd = m.PassInput.Update(msg)
		return m, cmd
	}

	switch keyMsg.String() {
	case "esc":
		m.UIState = StateNormal
		m.PassInput.Reset() // Clear the input field completely
		return m, nil

	case "enter":
		// Grab the clean text value directly from the component
		passwordValue := m.PassInput.Value()
		cmd := ConnectToAccessPoint(m.Client, m.SelectedAP, passwordValue)

		m.UIState = StateNormal
		m.PassInput.Reset()
		return m, cmd

	default:
		// Forward all typing events directly to the textinput bubble
		var cmd tea.Cmd
		m.PassInput, cmd = m.PassInput.Update(msg)
		return m, cmd
	}
}

func (m Model) handleSavedActionsMenu(msg tea.Msg) (Model, tea.Cmd) {
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
				cmd = ToggleAutoConnectCmd(m.Client, prof.UUID, !prof.AutoConnect)
			} else {
				cmd = ForgetProfileCmd(m.Client, prof.UUID)
			}
		}
		m.UIState = StateNormal
		return m, cmd
	}
	return m, nil
}

func (m Model) handleTableSize() (Model, tea.Cmd) {
	m.Table.SetWidth(config.TabBodyWidth)

	// Base height layout allocation
	tableHeight := int(math.Floor(config.TabBodyHeight * 0.8))

	// If offline/not scanning, shrink the saved table down
	// further to cleanly allocate room for hardware settings text blocks
	if m.UIState == StateSavedActionsMenu || m.UIState == StatePasswordInput {
		tableHeight = int(math.Floor(config.TabBodyHeight * 0.5))
	}

	tableHeight = max(tableHeight, config.MinTableHeight)
	m.Table.SetHeight(tableHeight)
	return m, nil
}

func (m Model) handleInfoLoaded(msg InfoLoadedMsg) (Model, tea.Cmd) {
	m.Adapter = msg.Adapter
	m.Saved = msg.Saved
	m.ActiveAPs = msg.APs
	m.Loading = false
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
		return InfoLoadedMsg(InfoLoadedData{Adapter: a, Saved: s, APs: aps})
	}
}

func (m Model) handleKeyInput(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch msg.String() {
	case "s":
		m.Scanning = !m.Scanning
		m.Table.GotoTop()

		m.handleTableSize()

		// Force trigger a programmatic resize window sequence to re-adjust
		// table heights layout based on scanning active state constraints

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
				return m, ConnectToAccessPoint(m.Client, m.SelectedAP, "")
			}
			m.UIState = StatePasswordInput
		} else {
			if len(m.Saved) > 0 && idx < len(m.Saved) {
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

// --- Table UI Logic ---

func (m *Model) syncTableRows() {
	var rows []table.Row

	wdth := config.WindowWidth - 4

	if m.Scanning {
		m.Table.SetRows(nil)

		m.Table.SetColumns([]table.Column{
			{Width: wdth / 8},
			{Title: "ssid", Width: (wdth * 2) / 5},
			{Width: wdth / 4},
			{Width: wdth / 8},
		})

		for _, ap := range m.ActiveAPs {
			activeMark := " "
			if ap.IsActive {
				activeMark = ""
			}
			rows = append(rows, table.Row{
				components.RenderSignal(ap.Strength, ap.Security),
				ap.SSID,
				ap.Security,
				activeMark,
			})
		}
	} else {
		m.Table.SetRows(nil)

		m.Table.SetColumns([]table.Column{
			{Title: "SSID", Width: (wdth * 2) / 5},
			{Title: "Auto", Width: wdth / 10},
			{Title: "UUID", Width: (wdth * 2) / 5},
		})

		for _, prof := range m.Saved {
			autoStr := " "
			if prof.AutoConnect {
				autoStr = "󰁪"
			}
			uuidShort := ""
			if len(prof.UUID) >= 8 {
				uuidShort = prof.UUID[:8]
			}
			rows = append(rows, table.Row{
				prof.Name,
				autoStr,
				uuidShort,
			})
		}
	}

	m.Table.SetRows(rows)

	if m.Table.Cursor() >= len(rows) {
		m.Table.GotoTop()
		m.Cursor = m.Table.Cursor()
	}
}
