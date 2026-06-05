package app

import (
	"fmt"

	"linktui/pkg/config"
	"linktui/pkg/vpn"
	"linktui/pkg/wifi"

	tea "charm.land/bubbletea/v2"
)

// Notice the switch to *AppModel so state mutations stick!
func (m *AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	if windowMsg, ok := msg.(tea.WindowSizeMsg); ok {
		if config.WindowWidth < 70 || config.WindowHeight < 25 {
			m.SizeError = fmt.Sprintf("⚠️  Configuration Error!\n\n  Configured sizes too small...\n  Absolute Minimum: 70x25")
			return m, nil // Returning pointer is fine here, Bubble Tea handles it
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
			switch keyMsg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			}
		}
		return m, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		oldTab := m.ActiveTab

		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit

		case "q":
			isWifiInput := m.ActiveTab == WifiTab && m.WifiView.UIState == wifi.StatePasswordInput
			isVpnInput := m.ActiveTab == VpnTab && m.VpnView.UIState == vpn.StateAddForm
			if !isWifiInput && !isVpnInput {
				return m, tea.Quit
			}

		case "tab", "pagedown", "pgdown":
			m.ActiveTab = (m.ActiveTab + 1) % 3

		case "shift+tab", "pageup", "pgup":
			m.ActiveTab = (m.ActiveTab - 1 + 3) % 3
		}

		if oldTab != m.ActiveTab {
			initCmd := m.lazyLoadTab(m.ActiveTab)
			if initCmd != nil {
				cmds = append(cmds, initCmd)
			}
			return m, tea.Batch(cmds...)
		}
	}

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
