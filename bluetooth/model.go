package bluetooth

import (
	"netui/config"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Devices      []Device
	Powered      bool
	Discoverable bool
	Pairable     bool

	// Dynamic Layout Elements (Viewport removed since table is self-contained)
	Table table.Model

	// Navigation & Component UI states
	Cursor     int
	MenuCursor int
	UIState    UIState
	Scanning   bool
	Err        error

	// Shared State targeting values
	SelectedMac string
	SelectedDev Device
	MenuOptions []string
}

func New() Model {
	columns := []table.Column{
		{Title: "Status", Width: 8},
		{Title: "Device Name", Width: 26},
		{Title: "MAC Address", Width: 18},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)

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
	// 1. State-based Structural Intercepts
	switch m.UIState {
	case StateActionsMenu:
		return m.handleActionsMenu(msg)
	}

	// 2. Normal State Core Navigation Loop
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg)

	case DevicesLoadedMsg:
		return m.handleDevicesLoaded(msg)

	case ScanStartedMsg:
		m.Scanning = true
		return m, nil

	case ScanStoppedMsg:
		return m.handleScanStopped()

	case TickMsg:
		return m.handleTick()

	case AdapterToggledMsg, ActionSuccessMsg:
		return m, tea.Batch(FetchDevicesCmd(), FetchAdapterInfoCmd())

	case AdapterInfoLoadedMsg:
		m.handleAdapterInfoLoaded(msg)
		return m, nil

	case ErrMsg:
		m.Err = msg
		m.Scanning = false
		return m, nil

	case tea.KeyMsg:
		return m.handleKeyInput(msg)
	}

	// 3. Fallback to sub-component updates
	var cmd tea.Cmd
	m.Table, cmd = m.Table.Update(msg)
	m.Cursor = m.Table.Cursor()

	return m, cmd
}
