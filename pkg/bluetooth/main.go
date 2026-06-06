package bluetooth

import (
	"fmt"

	"github.com/austinemk/linktui/pkg/bus"
	"github.com/austinemk/linktui/pkg/config"

	"charm.land/bubbles/v2/table"
	tea "charm.land/bubbletea/v2"
	"github.com/godbus/dbus/v5"
)

func New() Model {
	columns := []table.Column{
		{Title: "", Width: config.ListWidthSixteenth},
		{Title: "", Width: config.ListWidthHalf},
		{Title: "", Width: (config.ListWidthQuarter + config.ListWidthSixteenth)},
	}

	t := table.New(table.WithColumns(columns))
	t.SetWidth(config.ListWidth)
	t.SetHeight(config.ListHeight)
	t.Focus()

	s := table.DefaultStyles()
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

var AgentIncomingMsgs = make(chan PasskeyRequestMsg)

func RegisterAgentCmd() tea.Cmd {
	return func() tea.Msg {
		conn := bus.Get()
		agent := &BluetoothAgent{MsgChan: AgentIncomingMsgs}

		err := conn.Export(agent, dbus.ObjectPath(agentPath), agentInterface)
		if err != nil {
			return ErrMsg(fmt.Errorf("failed to export Agent to DBus: %v", err))
		}

		amObj := conn.Object(bluezInterface, dbus.ObjectPath("/org/bluez"))
		call := amObj.Call("org.bluez.AgentManager1.RegisterAgent", 0, dbus.ObjectPath(agentPath), "KeyboardDisplay")
		if call.Err != nil {
			return ErrMsg(fmt.Errorf("failed to register Agent with BlueZ: %v", call.Err))
		}

		call = amObj.Call("org.bluez.AgentManager1.RequestDefaultAgent", 0, dbus.ObjectPath(agentPath))
		if call.Err != nil {
			return ErrMsg(fmt.Errorf("failed to request default Agent: %v", call.Err))
		}
		return nil
	}
}

func ListenForAgentRequests() tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-AgentIncomingMsgs
		if !ok {
			return nil
		}
		return msg
	}
}

func (m Model) Init() tea.Cmd {
	// We can use tea.Batch to run the Ping check and the Data Fetching concurrently
	return tea.Batch(
		// Command 1: Check if BlueZ is running
		func() tea.Msg {
			conn := bus.Get()
			obj := conn.Object("org.bluez", dbus.ObjectPath("/"))

			if err := obj.Call("org.freedesktop.DBus.Peer.Ping", 0).Err; err != nil {
				return BluezStatusMsg(false)
			}
			return BluezStatusMsg(true)
		},

		// Command 2: Load the device and adapter data
		func() tea.Msg {
			saved, err := LoadPairedDevices()
			if err != nil {
				return ErrMsg(err)
			}

			ad, err := FetchAdapterInfo()
			if err != nil {
				return ErrMsg(err)
			}

			// Just return the data message directly
			return InfoLoadedMsg{
				Adapter: ad,
				Devices: saved,
			}
		},

		// Command 3: Your background agent listener
		ListenForAgentRequests(),
	)
}

func (m Model) Clean() {
	conn := bus.Get()
	if conn == nil {
		return
	}

	amObj := conn.Object(bluezInterface, dbus.ObjectPath("/org/bluez"))
	_ = amObj.Call("org.bluez.AgentManager1.UnregisterAgent", 0, dbus.ObjectPath(agentPath))

	if m.Scanning {
		obj := conn.Object(bluezInterface, adapterPath)
		_ = obj.Call("org.bluez.Adapter1.StopDiscovery", 0)
	}

	// Do NOT close conn here — bus.Close() in run.go handles it
	close(AgentIncomingMsgs)
}
