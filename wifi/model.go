package wifi

import (
	"context"
	"math"

	"netui/config"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Wifx/gonetworkmanager/v3"
)

type Model struct {
	Client    gonetworkmanager.NetworkManager
	Adapter   AdapterInfo
	Saved     []SavedProfile
	ActiveAPs []AccessPoint

	// Context for graceful cleanup
	Ctx    context.Context
	Cancel context.CancelFunc // <-- Add this to track background tasks

	// Dynamic Layout Elements
	Table     table.Model
	PassInput textinput.Model

	// Navigation & Component UI states
	Cursor     int // Kept for backend array mapping compatibility
	MenuCursor int
	UIState    UIState
	Scanning   bool
	Loading    bool
	Err        error

	// Password handling for secured lines
	SelectedAP    AccessPoint
	SelectedSaved SavedProfile
	PasswordInput string
	MenuOptions   []string
}

func New() Model {
	// Initialize default columns structure
	columns := []table.Column{
		{Title: "Status", Width: int(math.Floor(config.TabBodyWidth * 0.05))},
		{Title: "Network Name (SSID)", Width: int(math.Floor(config.TabBodyWidth * 0.5))}, // Cast explicitly for safety
		{Title: "Signal", Width: int(math.Floor(config.TabBodyWidth * 0.2))},
		{Title: "Security", Width: int(math.Floor(config.TabBodyWidth * 0.24))},
	}

	t := table.New(
		table.WithColumns(columns),
	)
	// V2: Width and Height use explicit setter functions instead of direct structural fields
	t.SetWidth(int(math.Floor(config.TabBodyWidth)))
	t.SetHeight(int(math.Floor(config.TabBodyHeight * 0.8)))
	t.Focus()

	ti := textinput.New()
	ti.Placeholder = "Password"
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '*'
	ti.Focus()

	// Apply theme defaults
	s := table.DefaultStyles()
	s.Header = lipgloss.NewStyle().Height(0).Padding(0, 0).MaxHeight(0)

	s.Selected = s.Selected.
		Foreground(config.Styles.HighlightText.GetForeground()).
		Background(config.Styles.HighlightText.GetBackground()).
		Bold(config.Styles.HighlightText.GetBold())
	t.SetStyles(s)

	// Create a cancellable context for background tasks
	ctx, cancel := context.WithCancel(context.Background())

	return Model{
		Client:      nil, // Will be loaded dynamically inside Init()
		Scanning:    false,
		Loading:     true,
		UIState:     StateNormal,
		Table:       t,
		PassInput:   ti,
		MenuOptions: []string{"autoconnect/off", "forget"},
		Ctx:         ctx,
		Cancel:      cancel,
	}
}

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
		m.Scanning = false
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
