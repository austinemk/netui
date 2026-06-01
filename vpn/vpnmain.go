package vpn

import (
	"math"
	"os"

	"corntui/config"

	"charm.land/bubbles/v2/filepicker"
	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/Wifx/gonetworkmanager/v3"
)

func New() Model {
	// 1. Initialize table view layout components
	columns := []table.Column{
		{Title: "Name", Width: 25},
		{Title: "Type", Width: 15},
		{Title: "Status", Width: 12},
	}

	t := table.New(
		table.WithColumns(columns),
	)
	t.SetWidth(int(config.TabBodyWidth))
	t.SetHeight(int(math.Floor(config.TabBodyHeight * 0.7)))
	t.Focus()

	// Apply theme defaults
	s := table.DefaultStyles()
	s.Header = lipgloss.NewStyle().Height(0).Padding(0, 0).MaxHeight(0)

	s.Selected = s.Selected.
		Foreground(config.Styles.HighlightText.GetForeground()).
		Background(config.Styles.HighlightText.GetBackground()).
		Bold(config.Styles.HighlightText.GetBold())
	t.SetStyles(s)

	// 2. Instantiate the file picker
	fp := filepicker.New()
	fp.AllowedTypes = []string{".conf", ".wg"}

	// Start at home directory; fall back to cwd if home is unavailabe
	fp.CurrentDirectory, _ = os.UserHomeDir()
	fp.AutoHeight = false
	fp.SetHeight(int(math.Floor(config.TabBodyHeight * 0.4)))

	// Truncate file/dir names to max 30 chars
	fp.Styles.File = fp.Styles.File.MaxWidth(30)
	fp.Styles.Directory = fp.Styles.Directory.MaxWidth(30)
	fp.Styles.Selected = fp.Styles.Selected.MaxWidth(30)
	fp.Styles.Cursor = fp.Styles.Selected.MaxWidth(20)
	fp.Styles.DisabledFile = fp.Styles.DisabledFile.MaxWidth(30)

	return Model{
		Client:     &DBusClient{NM: nil},
		Table:      t,
		FilePicker: fp,
		Loading:    true,
		UIState:    StateNormal,
		FormInputs: make(map[FormField]string),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.FilePicker.Init(), // Provisions directory lookup internal commands cleanly
		func() tea.Msg {
			nm, err := gonetworkmanager.NewNetworkManager()
			if err != nil {
				return ErrMsg(err)
			}

			tempClient := &DBusClient{NM: nm}
			t, err := GetVPNConnections(tempClient)
			if err != nil {
				return ErrMsg(err)
			}

			return TunnelsLoadedMsg(t)
		},
	)
}
