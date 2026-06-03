package wifi

import (
	"context"

	"corntui/config"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Wifx/gonetworkmanager/v3"
)

func New() Model {
	// Initialize default columns structure
	columns := []table.Column{
		{Title: "Status", Width: config.ListWidthSixteenth},
		{Title: "Network Name (SSID)", Width: config.ListWidthHalf}, // Cast explicitly for safety
		{Title: "Signal", Width: config.ListWidthSixteenth},
		{Title: "Security", Width: config.ListWidthEigth},
	}

	t := table.New(
		table.WithColumns(columns),
	)
	// V2: Width and Height use explicit setter functions instead of direct structural fields
	t.SetWidth(config.ListWidth)
	t.SetHeight(config.ListHeight)
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
		UIState:     StateNormal,
		Table:       t,
		PassInput:   ti,
		MenuOptions: []string{"autoconnect/off", "forget"},
		Ctx:         ctx,
		Cancel:      cancel,
	}
}

// Init that initializes the package
func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		nm, err := gonetworkmanager.NewNetworkManager()
		if err != nil {
			return ErrMsg(err)
		}

		adapter, err := GetAdapterSettings(nm)
		if err != nil {
			return ErrMsg(err)
		}

		saved, err := GetSavedProfiles(nm)
		if err != nil {
			return ErrMsg(err)
		}

		aps, err := GetActiveAccessPoints(nm)
		if err != nil {
			return ErrMsg(err)
		}

		return InfoLoadedMsg(InfoLoadedData{
			Client:  nm, // <-- Pass it along here
			Adapter: adapter,
			Saved:   saved,
			APs:     aps,
		})
	}
}

// Clean gracefully halts background procedures, context loops, and stops leaks
func (m Model) Clean() {
	// 1. Cancel the background context to instantly terminate the monitor goroutine
	if m.Cancel != nil {
		m.Cancel()
	}

	// 2. Explicitly stop the hardware scanning loop state flag
	m.Scanning = false

	// 3. Clean up inner component state instances
	m.PassInput.Reset()
	m.Table.SetRows(nil)
}
