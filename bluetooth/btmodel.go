package bluetooth

import (
	"math"

	"netui/config"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
)

type Model struct {
	Client       *BlueZClient
	Devices      []Device
	Powered      bool
	Discoverable bool
	Pairable     bool
	Table        table.Model
	Cursor       int
	MenuCursor   int
	UIState      UIState
	Scanning     bool
	Err          error
	SelectedMac  string
	SelectedDev  Device
	MenuOptions  []string
}

func New() Model {
	columns := []table.Column{
		{Title: "", Width: int(math.Floor(config.TabBodyWidth * 0.05))},
		{Title: "", Width: int(math.Floor(config.TabBodyWidth * 0.5))},
		{Title: "", Width: int(math.Floor(config.TabBodyWidth * 0.44))},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
		// table.WithHeight(int(math.Floor(config.TabBodyHeight*0.8))),
	)

	s := table.DefaultStyles()
	/*s.Header = s.Header.
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240")).
	BorderBottom(true).
	Bold(true)*/

	s.Selected = s.Selected.
		Foreground(config.Styles.HighlightText.GetForeground()).
		Background(config.Styles.HighlightText.GetBackground()).
		Bold(config.Styles.HighlightText.GetBold())
	t.SetStyles(s)

	return Model{
		Scanning: false,
		UIState:  StateNormal,
		Table:    t,
	}
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var tableCmd tea.Cmd

	// 1. CRITICAL: Let the table consume the message first
	// This captures scroll tracking, terminal updates, and geometry
	m.Table, tableCmd = m.Table.Update(msg)
	m.Cursor = m.Table.Cursor()

	// 2. State-based Structural Intercepts
	switch m.UIState {
	case StateActionsMenu:
		m, menuCmd := m.handleActionsMenu(msg)
		return m, tea.Batch(tableCmd, menuCmd)
	}

	// 3. Normal State Core Navigation Loop
	switch msg := msg.(type) {
	case PairedDevicesLoadedMsg:
		m, cmd := m.handleDevicesLoaded(msg)
		return m, tea.Batch(tableCmd, cmd)

	case DiscoveredDevicesLoadedMsg:
		m, cmd := m.handleDiscoveredLoaded(msg)
		return m, tea.Batch(tableCmd, cmd)

	case AdapterInfoLoadedMsg:
		m, cmd := m.handleAdapterInfoLoaded(AdapterInfo(msg))
		return m, tea.Batch(tableCmd, cmd)

	case ScanStartedMsg:
		m.Scanning = true
		return m, tea.Batch(tableCmd, PollBluetoothTicker())

	case ScanStoppedMsg:
		m, cmd := m.handleScanStopped()
		return m, tea.Batch(tableCmd, cmd)

	case TickMsg:
		m, cmd := m.handleTick()
		return m, tea.Batch(tableCmd, cmd)

	case AdapterToggledMsg:
		return m, tea.Batch(tableCmd, LoadPairedDevicesCmd(m.Client), FetchAdapterInfoCmd(m.Client))

	case ActionSuccessMsg:
		if m.Scanning {
			return m, tea.Batch(tableCmd, DiscoverDevicesCmd(m.Client), FetchAdapterInfoCmd(m.Client))
		}
		return m, tea.Batch(tableCmd, LoadPairedDevicesCmd(m.Client), FetchAdapterInfoCmd(m.Client))

	case ErrMsg:
		m.Err = msg
		m.Scanning = false
		return m, tableCmd

	case tea.KeyPressMsg:
		m, cmd := m.handleNormalStateNavigation(msg)
		return m, tea.Batch(tableCmd, cmd)
	}

	return m, tableCmd
}
