package vpn

import (
	"time"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
)

// Update acts as the central router for incoming messages.
func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch m.UIState {
	case StateAddForm:
		return m.handleFormState(msg)
	case StateImportFile:
		return m.handleFilePickerState(msg)
	case StateActionsMenu:
		return m.handleActionsMenuState(msg)

	}
	switch msg := msg.(type) {
	case TunnelsLoadedMsg:
		m.Client = msg.Client
		m.Tunnels = msg.Tunnels

		// Map backend tunnels data cleanly to the UI table rows
		m.syncTableRows()
		return m, nil

	case IPInfoMsg:
		m.IPInfo = msg
		return m, nil

	case ActionSuccessMsg:
		return m, FetchTunnelsCmd(m.Client)

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
	var cmd tea.Cmd

	m.Table, cmd = m.Table.Update(msg)
	return m, cmd
}

// syncTableRows translates the internal Tunnels state into viewable table rows.
// Because it modifies the Model directly, we use a pointer receiver (*Model).
func (m *Model) syncTableRows() {
	var rows []table.Row

	for _, t := range m.Tunnels {
		status := "Inactive"
		if t.Active {
			status = "Active 󰌆"
		}
		rows = append(rows, table.Row{t.Name, t.Type, status})
	}

	m.Table.SetRows(rows)

	if m.Table.Cursor() >= len(rows) {
		m.Table.GotoTop()
		m.Cursor = m.Table.Cursor()
	}
}
