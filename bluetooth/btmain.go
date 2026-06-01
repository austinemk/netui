package bluetooth

import (
	"math"

	"corntui/config"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"github.com/godbus/dbus/v5"
)

func NewBlueZClient() (*BlueZClient, error) {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return nil, err
	}
	return &BlueZClient{Conn: conn}, nil
}

func New() Model {
	columns := []table.Column{
		{Title: "", Width: int(math.Floor(config.TabBodyWidth * 0.05))},
		{Title: "", Width: int(math.Floor(config.TabBodyWidth * 0.5))},
		{Title: "", Width: int(math.Floor(config.TabBodyWidth * 0.44))},
	}

	t := table.New(
		table.WithColumns(columns),
		// table.WithHeight(int(math.Floor(config.TabBodyHeight*0.8))),
	)
	t.SetWidth(int(math.Floor(config.TabBodyWidth)))
	t.SetHeight(int(math.Floor(config.TabBodyHeight * 0.8)))
	t.Focus()

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

func (m Model) Init() tea.Cmd {
	return func() tea.Msg {
		bluez, err := NewBlueZClient()
		if err != nil {
			return ErrMsg(err)
		}

		saved, err := LoadPairedDevices(bluez)
		if err != nil {
			return ErrMsg(err)
		}

		ad, err := FetchAdapterInfo(bluez)
		if err != nil {
			return ErrMsg(err)
		}
		return InfoLoadedMsg(InfoLoadedData{
			Client:  bluez,
			Adapter: ad,
			Devices: saved,
		})
	}
}

// Clean gracefully stops any hardware discovery and closes the system bus connection to prevent memory leaks.
func (m Model) Clean() {
	if m.Client == nil || m.Client.Conn == nil {
		return
	}

	// 1. If the hardware is actively discovering devices, tell BlueZ to stop immediately
	if m.Scanning {
		obj := m.Client.Conn.Object(bluezInterface, adapterPath)
		// Send a direct synchronous DBus call to ensure it hits the OS before the binary exits
		_ = obj.Call("org.bluez.Adapter1.StopDiscovery", 0)
	}

	// 2. Close the D-Bus connection completely to clear system RAM and file descriptors
	_ = m.Client.Conn.Close()
}
