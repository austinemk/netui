package bluetooth

import (
	"math"
	"time"

	"corntui/config"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
)

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	// 2. State-based Structural Intercepts
	switch m.UIState {
	case StateActionsMenu:
		return m.handleActionsMenu(msg)
	case StatePasskeyPrompt:
		return m.handlePasskeyPrompt(msg)
	}

	// 3. Normal State Core Navigation Loop
	switch msg := msg.(type) {
	case InfoLoadedMsg:
		return m.handleInfoLoaded(msg)

	case ScanFinishedMsg:
		return m.handleScanFinished(msg)

	case TickMsg:
		return m.handleTick()

	case AdapterToggledMsg:
		return m.handleAdapterOrActionSuccess()

	case PasskeyRequestMsg:
		m.UIState = StatePasskeyPrompt
		m.SelectedDev = msg.Device
		m.CurrentPasskey = msg.Passkey
		m.ActiveRespChan = msg.ResponseChan
		m.MenuCursor = 0                                               // Default to highlight 'Yes'
		m.Table.SetHeight(int(math.Floor(config.TabBodyHeight * 0.4))) // Shrink table layout to provide screen real estate
		return m, nil

	case ActionSuccessMsg:
		if m.Scanning {
			return m, tea.Batch(
				ContinueDiscoveryCmd(m.Client),
				FetchAdapterInfoCmd(m.Client),
			)
		}
		return m, tea.Batch(
			LoadPairedDevicesCmd(m.Client),
			FetchAdapterInfoCmd(m.Client),
		)
	case AdapterInfoLoadedMsg:
		m.Adapter = AdapterInfo(msg)
		return m, nil

	case ErrMsg:
		m.Err = msg
		m.LogID++
		return m, func() tea.Msg {
			time.Sleep(4 * time.Second) // Display duration before auto-removal
			return ClearLogMsg{ID: m.LogID}
		}
	case ClearLogMsg:
		if msg.ID == m.LogID {
			m.Err = nil
		}
		return m, nil

	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	}

	// Fallback only
	var cmd tea.Cmd
	m.Table, cmd = m.Table.Update(msg)
	return m, cmd
}

// syncTableRows maps data structures onto bubbletea table UI dimensions
// Note: Changed from pointer receiver to value receiver returning Model
func (m *Model) syncTableRows() {
	var rows []table.Row
	m.Table.SetRows(nil)
	// DO NOT clear or reset SetColumns inside here to prevent visual glitching

	for _, dev := range m.Devices {
		//fmt.Printf("devices: %s", dev.Icon)
		statusIcon := "󰂯"
		/*if dev.Icon != "" {
			statusIcon = dev.Icon
		}*/
		statusIcon = dev.Icon
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
	//fmt.Printf("table rows: %s", rows) // Return the mutated copy back
}
