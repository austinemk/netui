package main

import (
	"fmt"

	"netui/bluetooth"
	"netui/config"
	"netui/vpn"
	"netui/wifi"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type Tab int

const (
	WifiTab Tab = iota
	BluetoothTab
	VpnTab
)

type AppModel struct {
	ActiveTab  Tab
	WifiView   wifi.Model
	BtView     bluetooth.Model
	VpnView    vpn.Model
	LogMessage string
	SizeError  string

	// Keeps track of which tabs have been initialized
	LoadedTabs map[Tab]bool
}

func (m AppModel) Init() tea.Cmd {
	// Don't call lazyLoadTab here — LoadedTabs is already seeded in main.go
	// Just fire the appropriate Init cmd directly
	switch m.ActiveTab {
	case WifiTab:
		return m.WifiView.Init()
	case BluetoothTab:
		return m.BtView.Init()
	case VpnTab:
		if m.VpnView.Client == nil {
			m.VpnView.Client = &vpn.DBusClient{}
		}
		return m.VpnView.Init()
	}
	return nil
}

// Helper method to conditionally load a tab and return its Init command
func (m *AppModel) lazyLoadTab(tab Tab) tea.Cmd {
	if m.LoadedTabs == nil {
		m.LoadedTabs = make(map[Tab]bool)
	}

	if m.LoadedTabs[tab] {
		return nil
	}

	m.LoadedTabs[tab] = true

	switch tab {
	case WifiTab:
		// WiFi is ready to issue its initialization command sequence
		return m.WifiView.Init()

	case BluetoothTab:
		if m.BtView.Client == nil && m.BtView.Err == nil {
			btClient, err := bluetooth.NewBlueZClient()
			if err == nil {
				m.BtView.Client = btClient
			}
		}
		return m.BtView.Init()

	case VpnTab:
		// Make sure the Client struct wrapper pointer exists BEFORE calling Init()
		if m.VpnView.Client == nil {
			m.VpnView.Client = &vpn.DBusClient{}
		}
		return m.VpnView.Init()
	}
	return nil
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// Window Resizing Logic (Kept intact)
	if windowMsg, ok := msg.(tea.WindowSizeMsg); ok {
		if config.WindowWidth < 70 || config.WindowHeight < 25 {
			m.SizeError = fmt.Sprintf("⚠️  Configuration Error!\n\n  Configured sizes too small...\n  Absolute Minimum: 70x25")
			return m, nil
		}
		if config.WindowWidth > windowMsg.Width || config.WindowHeight > windowMsg.Height {
			m.SizeError = fmt.Sprintf("⚠️  Terminal screen too small!\n\n  Please resize your terminal window.")
			return m, nil
		}
		m.SizeError = ""
		windowMsg.Width = config.WindowWidth
		windowMsg.Height = config.WindowHeight
		msg = windowMsg
	}

	if m.SizeError != "" {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "q" || keyMsg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		}
		return m, nil
	}

	// 1. GLOBAL SYSTEM EVENT KEYMAPS
	switch msg := msg.(type) {
	case tea.KeyMsg:
		oldTab := m.ActiveTab

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "1", "2", "3":
			m.ActiveTab = Tab(msg.String()[0] - '1')
		case "tab", "pagedown", "pgdown":
			m.ActiveTab = (m.ActiveTab + 1) % 3
		case "shift+tab", "pageup", "pgup":
			m.ActiveTab = (m.ActiveTab - 1 + 3) % 3
		}

		// If the tab changed, check if we need to load the new one!
		if oldTab != m.ActiveTab {
			initCmd := m.lazyLoadTab(m.ActiveTab)
			if initCmd != nil {
				cmds = append(cmds, initCmd)
			}
			return m, tea.Batch(cmds...)
		}
	}

	// 2. ROUTE MESSAGES ONLY TO THE ACTIVE, LOADED TAB
	switch m.ActiveTab {
	case WifiTab:
		m.WifiView, cmd = m.WifiView.Update(msg)
		cmds = append(cmds, cmd)
	case BluetoothTab:
		m.BtView, cmd = m.BtView.Update(msg)
		cmds = append(cmds, cmd)
	case VpnTab:
		m.VpnView, cmd = m.VpnView.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m AppModel) View() tea.View {
	if m.SizeError != "" {
		boxStyle := lipgloss.NewStyle().Border(lipgloss.DoubleBorder()).BorderForeground(lipgloss.Color("#EF4444")).Padding(2, 4)
		v := tea.NewView(boxStyle.Render(m.SizeError))
		v.AltScreen = true
		return v
	}

	header := RenderHeader(int(m.ActiveTab))

	// Render Body (Only renders if initialized, otherwise displays loading state)
	var body string
	if !m.LoadedTabs[m.ActiveTab] {
		body = "\n  Loading interface details..."
	} else {
		switch m.ActiveTab {
		case WifiTab:
			body = m.WifiView.View()
		case BluetoothTab:
			body = m.BtView.View()
		case VpnTab:
			body = m.VpnView.View()
		}
	}

	var logView string
	if m.LogMessage != "" {
		logView = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Render("\n[LOG] " + m.LogMessage)
	}

	footer := RenderFooter(int(m.ActiveTab), false)

	mainLayout := lipgloss.JoinVertical(lipgloss.Left, header, body, logView, footer)
	appBorderStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("8")).Padding(0, 1)
	mainLayout = appBorderStyle.Render(mainLayout)

	v := tea.NewView(mainLayout)
	v.AltScreen = true
	return v
}
