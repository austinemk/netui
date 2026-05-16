package ui

import (
	"netui/ui/components"
	"netui/ui/views/bluetooth"
	"netui/ui/views/vpn"
	"netui/ui/views/wifi"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FocusedWindow int

const (
	MainWindow FocusedWindow = iota
	PopupWindow
)

type Tab int

const (
	WifiTab Tab = iota
	BluetoothTab
	VpnTab
)

type AppModel struct {
	Focus      FocusedWindow
	ActiveTab  Tab
	WifiView   wifi.Model
	BtView     bluetooth.Model
	VpnView    vpn.Model
	Popup      components.PopupModel // Shared overlay input
	LogMessage string                // Empty if no status, shows up if active
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	// 1. INTERCEPT IF POPUP IS ACTIVE
	if m.Focus == PopupWindow {
		m.Popup, cmd = m.Popup.Update(msg)
		// Check here if user submitted password or closed popup, then toggle focus back
		return m, cmd
	}

	// 2. GLOBAL KEYS (App event keys)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "1", "2", "3": // Direct Tab switching
			m.ActiveTab = Tab(msg.String()[0] - '1')
			return m, nil
		}
	}

	// 3. ROUTE KEYPRESS TO THE ACTIVE TAB ONLY
	switch m.ActiveTab {
	case WifiTab:
		m.WifiView, cmd = m.WifiView.Update(msg)
	case BluetoothTab:
		m.BtView, cmd = m.BtView.Update(msg)
	case VpnTab:
		m.VpnView, cmd = m.VpnView.Update(msg)
	}

	return m, cmd
}

func (m AppModel) View() string {
	// Render Header
	header := components.RenderHeader(m.ActiveTab)

	// Render Active Body Module
	var body string
	switch m.ActiveTab {
	case WifiTab:
		body = m.WifiView.View()
	case BluetoothTab:
		body = m.BtView.View()
	case VpnTab:
		body = m.VpnView.View()
	}

	// Render Log Frame conditionally
	var logView string
	if m.LogMessage != "" {
		logView = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Render("\n[LOG] " + m.LogMessage)
	}

	// Render dynamic hints based on state
	footer := components.RenderFooter(m.ActiveTab, m.Focus)

	// Stack vertically
	mainLayout := lipgloss.JoinVertical(lipgloss.Left, header, body, logView, footer)

	// If popup is active, overlay it onto main layout
	if m.Focus == PopupWindow {
		return components.RenderOverlay(mainLayout, m.Popup.View())
	}

	return mainLayout
}
