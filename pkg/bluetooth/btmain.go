package bluetooth

import (
	"fmt"

	"linktui/pkg/config"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"github.com/godbus/dbus/v5"
)

func NewBlueZClient() (*BlueZClient, error) {
	// 1. Check if the D-Bus system bus is even accessible.
	// If the system doesn't have D-Bus or the user lacks permissions, this fails immediately.
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		return nil, fmt.Errorf("system D-Bus is unavailable (are you running on Linux with proper permissions?): %w", err)
	}

	// 2. Explicitly ping the BlueZ daemon via D-Bus.
	// This proves that the 'bluez' package is installed, running, and actively listening.
	// We call a standard D-Bus peer Ping on the BlueZ service destination.
	obj := conn.Object("org.bluez", dbus.ObjectPath("/"))
	err = obj.Call("org.freedesktop.DBus.Peer.Ping", 0).Err
	if err != nil {
		// If D-Bus is working but BlueZ doesn't reply, it means the package/service is missing or stopped.
		conn.Close() // Clean up the connection before leaving
		return nil, fmt.Errorf("bluez service is not responding. Ensure 'bluez' is installed and the bluetooth service is running: %w", err)
	}

	return &BlueZClient{Conn: conn}, nil
}

func New() Model {
	columns := []table.Column{
		{Title: "", Width: config.ListWidthSixteenth},
		{Title: "", Width: config.ListWidthHalf},
		{Title: "", Width: (config.ListWidthQuarter + config.ListWidthSixteenth)},
	}

	t := table.New(
		table.WithColumns(columns),
		// table.WithHeight(int(math.Floor(config.TabBodyHeight*0.8))),
	)
	t.SetWidth(config.ListWidth)
	t.SetHeight(config.ListHeight)
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

// We create a global or Model-bound channel to pipe asynchronous D-Bus events into Bubble Tea
var AgentIncomingMsgs = make(chan PasskeyRequestMsg)

func RegisterAgentCmd(client *BlueZClient) tea.Cmd {
	return func() tea.Msg {
		agent := &BluetoothAgent{MsgChan: AgentIncomingMsgs}

		// 1. Export our Agent object path to the System Bus
		err := client.Conn.Export(agent, dbus.ObjectPath(agentPath), agentInterface)
		if err != nil {
			return ErrMsg(fmt.Errorf("failed to export Agent to DBus: %v", err))
		}

		// 2. Call BlueZ AgentManager1 to register our agent path globally
		amObj := client.Conn.Object(bluezInterface, dbus.ObjectPath("/org/bluez"))
		call := amObj.Call("org.bluez.AgentManager1.RegisterAgent", 0, dbus.ObjectPath(agentPath), "KeyboardDisplay")
		if call.Err != nil {
			return ErrMsg(fmt.Errorf("failed to register Agent with BlueZ: %v", call.Err))
		}

		// 3. Request BlueZ to make this the default agent for handling pairing requests
		call = amObj.Call("org.bluez.AgentManager1.RequestDefaultAgent", 0, dbus.ObjectPath(agentPath))
		if call.Err != nil {
			return ErrMsg(fmt.Errorf("failed to request default Agent: %v", call.Err))
		}

		logToFile("🛡️ BlueZ Agent registered successfully at path: %s", agentPath)
		return nil // Success
	}
}

// Background worker that listens to our Agent channel and maps it directly to bubbletea messages
func ListenForAgentRequests() tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-AgentIncomingMsgs
		if !ok {
			return nil // Channel closed, exit cleanly
		}
		return msg
	}
}

// Init for package initialization
func (m Model) Init() tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
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

			// 👇 WE MUST BATCH REGISTRATION AND INITIAL DATALOAD TOGETHER
			return tea.Batch(
				RegisterAgentCmd(bluez),
				func() tea.Msg {
					return InfoLoadedMsg(InfoLoadedData{
						Client:  bluez,
						Adapter: ad,
						Devices: saved,
					})
				},
			)()
		},
		ListenForAgentRequests(), // Start monitoring the channel immediately
	)
}

// Clean gracefully stops any hardware discovery and closes the system bus connection to prevent memory leaks.
func (m Model) Clean() {
	if m.Client == nil || m.Client.Conn == nil {
		return
	}

	// Unregister Agent from BlueZ
	amObj := m.Client.Conn.Object(bluezInterface, dbus.ObjectPath("/org/bluez"))
	_ = amObj.Call("org.bluez.AgentManager1.UnregisterAgent", 0, dbus.ObjectPath(agentPath))

	if m.Scanning {
		obj := m.Client.Conn.Object(bluezInterface, adapterPath)
		_ = obj.Call("org.bluez.Adapter1.StopDiscovery", 0)
	}

	_ = m.Client.Conn.Close()

	// ✅ Unblock the ListenForAgentRequests goroutine so it can exit
	close(AgentIncomingMsgs)
}
