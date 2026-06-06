package vpn

import (
	"os"

	"github.com/austinemk/linktui/pkg/bus"
	"github.com/austinemk/linktui/pkg/config"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

func New() Model {
	columns := []table.Column{
		{Title: "", Width: config.ListWidthHalf},
		{Title: "", Width: config.ListWidthQuarter},
		{Title: "", Width: config.ListWidthQuarter},
	}

	t := table.New(table.WithColumns(columns))
	t.SetWidth(config.ListWidth)
	t.SetHeight(config.ListHeight)
	t.Focus()

	s := table.DefaultStyles()
	s.Header = lipgloss.NewStyle().Height(0).Padding(0, 0).MaxHeight(0)
	s.Selected = s.Selected.
		Foreground(config.Styles.HighlightText.GetForeground()).
		Background(config.Styles.HighlightText.GetBackground()).
		Bold(config.Styles.HighlightText.GetBold())
	t.SetStyles(s)

	fp := filepicker.New()
	fp.AllowedTypes = []string{".conf", ".wg"}
	fp.CurrentDirectory, _ = os.UserHomeDir()
	fp.AutoHeight = false
	fp.SetHeight(config.ListHeight)
	fp.Styles.File = fp.Styles.File.MaxWidth(30)
	fp.Styles.Directory = fp.Styles.Directory.MaxWidth(30)
	fp.Styles.Selected = fp.Styles.Selected.MaxWidth(30)
	fp.Styles.Cursor = fp.Styles.Selected.MaxWidth(20)
	fp.Styles.DisabledFile = fp.Styles.DisabledFile.MaxWidth(30)
	fp.ShowHidden = false

	return Model{
		Table:      t,
		FilePicker: fp,
		UIState:    StateNormal,
		FormInputs: make(map[FormField]string),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.FilePicker.Init(),

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
		// Command 2: Load VPN tunnel profiles
		func() tea.Msg {
			t, err := GetVPNConnections()
			if err != nil {
				return ErrMsg(err)
			}
			return TunnelsLoadedMsg(TunnelsLoadedData{Tunnels: t})
		},
	)
}
