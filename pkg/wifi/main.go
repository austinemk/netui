package wifi

import (
	"context"

	"github.com/austinemk/linktui/pkg/bus"
	"github.com/austinemk/linktui/pkg/config"

	"charm.land/bubbles/v2/table"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func New() Model {
	columns := []table.Column{
		{Title: "Status", Width: config.ListWidthSixteenth},
		{Title: "Network Name (SSID)", Width: config.ListWidthHalf},
		{Title: "Signal", Width: config.ListWidthSixteenth},
		{Title: "Security", Width: config.ListWidthEigth},
	}

	t := table.New(table.WithColumns(columns))
	t.SetWidth(config.ListWidth)
	t.SetHeight(config.ListHeight)
	t.Focus()

	ti := textinput.New()
	ti.Placeholder = "Password"
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '*'
	ti.Focus()

	s := table.DefaultStyles()
	s.Header = lipgloss.NewStyle().Height(0).Padding(0, 0).MaxHeight(0)
	s.Selected = s.Selected.
		Foreground(config.Styles.HighlightText.GetForeground()).
		Background(config.Styles.HighlightText.GetBackground()).
		Bold(config.Styles.HighlightText.GetBold())
	t.SetStyles(s)

	ctx, cancel := context.WithCancel(context.Background())

	return Model{
		Scanning:    false,
		UIState:     StateNormal,
		Table:       t,
		PassInput:   ti,
		MenuOptions: []string{"autoconnect/off", "forget"},
		Ctx:         ctx,
		Cancel:      cancel,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		// Command 1: Simple confirmation if NetworkManager is running on the system bus
		func() tea.Msg {
			conn := bus.Get()
			if conn == nil {
				return NMStatusMsg(false)
			}

			// Check if the org.freedesktop.NetworkManager name is currently active on the bus
			var running bool
			err := conn.Object("org.freedesktop.DBus", "/org/freedesktop/DBus").
				Call("org.freedesktop.DBus.NameHasOwner", 0, "org.freedesktop.NetworkManager").
				Store(&running)

			// If there's an error or running is false, NetworkManager is not alive/installed
			if err != nil || !running {
				return NMStatusMsg(false)
			}

			return NMStatusMsg(true)
		},

		// Command 2: Load adapter, saved profiles, and access points
		func() tea.Msg {
			adapter, err := GetAdapterSettings()
			if err != nil {
				return ErrMsg(err)
			}

			saved, err := GetSavedProfiles()
			if err != nil {
				return ErrMsg(err)
			}

			aps, err := GetActiveAccessPoints()
			if err != nil {
				return ErrMsg(err)
			}

			return InfoLoadedMsg(InfoLoadedData{
				Adapter: adapter,
				Saved:   saved,
				APs:     aps,
			})
		},
	)
}

func (m Model) Clean() {
	if m.Cancel != nil {
		m.Cancel()
	}
	m.Scanning = false
	m.PassInput.Reset()
	m.Table.SetRows(nil)
}
