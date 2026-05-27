package wifi

import (
	"fmt"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// 1. State-based Structural Intercepts
	switch m.UIState {
	case StatePasswordInput:
		return m.handlePasswordInput(msg)
	case StateSavedActionsMenu:
		return m.handleSavedActionsMenu(msg)
	}

	// 2. Normal State Core Navigation Loop
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	case InfoLoadedMsg:
		return m.handleInfoLoaded(msg)

	case ScanFinishedMsg:
		return m.handleScanFinished(msg)

	case TickMsg:
		return m.handleTick()

	case AdapterToggledMsg, ActionSuccessMsg:
		return m.handleAdapterOrActionSuccess()

	case ErrMsg:
		m.Err = msg
		m.Scanning = false
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyInput(msg)
	}

	// 3. Fallback to sub-component updates
	var cmd tea.Cmd
	var cmds []tea.Cmd

	m.Table, cmd = m.Table.Update(msg)
	cmds = append(cmds, cmd)

	m.Viewport, cmd = m.Viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// --- Dedicated Handler Functions ---

func (m Model) handlePasswordInput(msg tea.Msg) (Model, tea.Cmd) {
	keyMsg, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}

	switch keyMsg.String() {
	case "esc":
		m.UIState = StateNormal
		m.PasswordInput = ""
	case "backspace":
		if len(m.PasswordInput) > 0 {
			m.PasswordInput = m.PasswordInput[:len(m.PasswordInput)-1]
		}
	case "enter":
		cmd := ConnectToAccessPoint(m.Client, m.SelectedAP, m.PasswordInput)
		m.UIState = StateNormal
		m.PasswordInput = ""
		return m, cmd
	default:
		if len(keyMsg.String()) == 1 {
			m.PasswordInput += keyMsg.String()
		}
	}
	return m, nil
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

func (m Model) handleWindowSize(msg tea.WindowSizeMsg) (Model, tea.Cmd) {
	m.Viewport.Width = msg.Width
	m.Viewport.Height = msg.Height

	m.Table.SetWidth(msg.Width - 4)
	tableHeight := msg.Height - 14
	if tableHeight < 4 {
		tableHeight = 4
	}
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
		m.Viewport.GotoTop()
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
				m.MenuCursor = 0
			}
		}

	default:
		// FIX 1: Forward navigation keys (up/down/j/k) directly to the table component
		var cmd tea.Cmd
		m.Table, cmd = m.Table.Update(msg)

		// FIX 2: Keep your custom m.Cursor synced with the table's internal cursor position
		m.Cursor = m.Table.Cursor()
		return m, cmd
	}
	return m, nil
}

// --- Table UI Logic ---

func (m *Model) syncTableRows() {
	var rows []table.Row

	if m.Scanning {
		// Clear the old rows cache before defining the new columns!
		m.Table.SetRows(nil)

		m.Table.SetColumns([]table.Column{
			{Title: "Status", Width: 8},
			{Title: "Network Name (SSID)", Width: 35},
			{Title: "Signal", Width: 10},
			{Title: "Security", Width: 15},
		})

		for _, ap := range m.ActiveAPs {
			activeMark := " "
			if ap.IsActive {
				activeMark = " ✔"
			}
			rows = append(rows, table.Row{
				activeMark,
				ap.SSID,
				fmt.Sprintf("%d%%", ap.Strength),
				ap.Security,
			})
		}
	} else {
		// Clear the old rows cache before defining the new columns!
		m.Table.SetRows(nil)

		m.Table.SetColumns([]table.Column{
			{Title: "Profile Name (SSID)", Width: 35},
			{Title: "Connection Mode", Width: 18},
			{Title: "UUID Fingerprint", Width: 15},
		})

		for _, prof := range m.Saved {
			autoStr := "Manual Only"
			if prof.AutoConnect {
				autoStr = "AutoConnect"
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

	// Bonus Safeguard: Reset the table cursor if it's suddenly out of bounds
	// for the new data array size, otherwise clicking "enter" later might crash.
	if m.Table.Cursor() >= len(rows) {
		m.Table.GotoTop()
		m.Cursor = m.Table.Cursor()
	}
}
