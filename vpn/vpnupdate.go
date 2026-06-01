package vpn

import (
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

	return m.handleCoreLifecycle(msg)
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
