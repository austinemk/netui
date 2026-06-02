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
		{Title: "", Width: int(math.Floor(config.TabBodyWidth * 0.4))},
		{Title: "", Width: int(math.Floor(config.TabBodyWidth * 0.2))},
		{Title: "", Width: int(math.Floor(config.TabBodyWidth * 0.25))},
	}

	t := table.New(
		table.WithColumns(columns),
	)
	t.SetWidth(int(config.TabBodyWidth))
	t.SetHeight(int(math.Floor(config.TabBodyHeight * 0.85)))
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
		Client:     nil,
		Table:      t,
		FilePicker: fp,
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

			t, err := GetVPNConnections(nm)
			if err != nil {
				return ErrMsg(err)
			}

			return TunnelsLoadedMsg(TunnelsLoadedData{
				Client:  nm,
				Tunnels: t,
			})
		},
	)
}
