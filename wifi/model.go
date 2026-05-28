package wifi

import (
	"math"

	"netui/config"

	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	Client    *DBusClient
	Adapter   AdapterInfo
	Saved     []SavedProfile
	ActiveAPs []AccessPoint

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
		{Title: "Network Name (SSID)", Width: config.TabBodyWidth * 0.5},
		{Title: "Signal", Width: int(math.Floor(config.TabBodyWidth * 0.2))},
		{Title: "Security", Width: int(math.Floor(config.TabBodyWidth * 0.24))},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithWidth(int(math.Floor(config.TabBodyWidth))),
		table.WithHeight(int(math.Floor(config.TabBodyHeight*0.8))),
		table.WithFocused(true),
	)

	// ─── INITIALIZE THE TEXT INPUT HERE ──────────────────────────────
	ti := textinput.New()
	ti.Placeholder = "Password"
	ti.EchoMode = textinput.EchoPassword // Automatically handles hiding text with asterisks/dots
	ti.EchoCharacter = '*'
	ti.Focus() // Start with the input focused

	// Apply beautiful theme defaults
	s := table.DefaultStyles()
	s.Header = lipgloss.NewStyle().Height(0).Padding(0, 0).MaxHeight(0)

	s.Selected = s.Selected.
		Foreground(config.Styles.HighlightText.GetForeground()).
		Background(config.Styles.HighlightText.GetBackground()). // Uses your blue color from Styles.CursorColor
		Bold(config.Styles.HighlightText.GetBold())
	t.SetStyles(s)

	return Model{
		Client:      &DBusClient{Conn: nil},
		Scanning:    false,
		Loading:     true,
		UIState:     StateNormal,
		Table:       t,
		PassInput:   ti,
		MenuOptions: []string{"autoconnect/off", "forget"},
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

	case tea.KeyMsg:
		return m.handleKeyInput(msg)
	}

	// 3. Fallback to sub-component updates
	var cmd tea.Cmd
	m.Table, cmd = m.Table.Update(msg)

	return m, cmd
}
