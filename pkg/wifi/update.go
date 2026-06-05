package wifi

import (
	"time"

	"linktui/pkg/config"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
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

	// V2 Change: KeyMsg is now KeyPressMsg
	case tea.KeyPressMsg:
		return m.handleKeyInput(msg)
	}

	// 3. Fallback to sub-component updates
	var cmd tea.Cmd
	m.Table, cmd = m.Table.Update(msg)

	return m, cmd
}

func (m *Model) syncTableRows() {
	var rows []table.Row

	if m.Scanning {
		m.Table.SetRows(nil)
		// V2 Fix: Explicitly mapped table.Column keys
		m.Table.SetColumns([]table.Column{
			{Title: "", Width: config.ListWidthSixteenth},
			{Title: "", Width: config.ListWidthHalf},
			{Title: "", Width: config.ListWidthEigth},
			{Title: "", Width: config.ListWidthSixteenth},
		})

		for _, ap := range m.ActiveAPs {
			activeMark := " "
			if ap.IsActive {
				activeMark = ""
			}
			rows = append(rows, table.Row{
				RenderSignal(ap.Strength, ap.Security),
				ap.SSID,
				ap.Security,
				activeMark,
			})
		}
	} else {
		m.Table.SetRows(nil)
		// V2 Fix: Explicitly mapped table.Column keys
		m.Table.SetColumns([]table.Column{
			{Title: "", Width: config.ListWidthHalf},
			{Title: "", Width: config.ListWidthSixteenth},
			{Title: "", Width: (config.ListWidthHalf - config.ListWidthSixteenth)},
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
