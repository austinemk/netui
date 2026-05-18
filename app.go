package main

import (
	"netui/bluetooth"
	"netui/components"
	"netui/vpn"
	"netui/wifi"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type FocusedWindow int

const (
	MainWindow FocusedWindow = iota
	PopupWindow
	OptionsPopupWindow // Added to map the new options popup safely
)

type Tab int

const (
	WifiTab Tab = iota
	BluetoothTab
	VpnTab
)

type AppModel struct {
	Focus        FocusedWindow
	ActiveTab    Tab
	WifiView     wifi.Model
	BtView       bluetooth.Model
	VpnView      vpn.Model
	OptionsPopup components.OptionsPopupModel // Custom context action popup
	InputPopup   components.InputPopupModel   // Custom maskable credential popup
	LogMessage   string                       // Contextual logs panel text
}

func (m AppModel) Init() tea.Cmd {
	// Gather and batch initial commands from all your independent tabs
	return tea.Batch(
		m.WifiView.Init(),
		m.BtView.Init(),
		m.VpnView.Init(),
	)
}

func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	// 1. INTERCEPT INPUT IF ANY POPUP IS VISIBLE
	if m.Focus == PopupWindow {
		m.InputPopup, cmd = m.InputPopup.Update(msg)
		if !m.InputPopup.Visible {
			m.Focus = MainWindow // Restore focus back to active list
		}
		return m, cmd
	}

	if m.Focus == OptionsPopupWindow {
		m.OptionsPopup, cmd = m.OptionsPopup.Update(msg)

		// Catch selection signal fired from the reusable options component
		if selMsg, ok := msg.(components.OptionSelectedMsg); ok {
			m.OptionsPopup.Active = false
			m.Focus = MainWindow
			m.LogMessage = "Selected action: " + selMsg.Option

			// Pass selection message down so the active sub-tab executes it
			msg = selMsg
		} else if !m.OptionsPopup.Active {
			m.Focus = MainWindow // Restores focus if user hits 'esc'
		}
		// Fall through or return if you want to skip standard list navigation while popup is open
		if m.Focus == MainWindow {
			return m, cmd
		}
	}

	// 2. GLOBAL SYSTEM EVENT KEYMAPS
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			//bluetooth.CleanupAll() // Instantly kills background bluetoothctl processes safely
			return m, tea.Quit
		case "1", "2", "3":
			m.ActiveTab = Tab(msg.String()[0] - '1')
			return m, nil
		}
	}

	// 3. ROUTE MESSAGES & KEYPRESSES TO THE ACTIVE TAB ONLY
	// This lets sub-views change their own internal m.Status variables independently.
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

func (m AppModel) View() string {
	// A. Render Global Navigation Tabs
	header := components.RenderHeader(int(m.ActiveTab))

	// B. Render Centralized Live Dashboard Grid
	// Pulls isolated sub-tab variables automatically!
	/*statusGrid := "\n  " + components.RenderStatusGrid(
		m.WifiView.Status, // e.g. components.StatusScanning
		m.BtView.Status,   // e.g. components.StatusConnected
		m.VpnView.Status,  // e.g. components.StatusIdle
	) + "\n"*/

	// C. Render Focused Sub-View Body
	var body string
	switch m.ActiveTab {
	case WifiTab:
		body = m.WifiView.View()
	case BluetoothTab:
		body = m.BtView.View()
	case VpnTab:
		body = m.VpnView.View()
	}

	// D. Render Bottom Logging Context Frame
	var logView string
	if m.LogMessage != "" {
		logView = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Render("\n[LOG] " + m.LogMessage)
	}

	isPopupOpen := m.Focus != MainWindow
	// E. Render Layout Actions Footer
	footer := components.RenderFooter(int(m.ActiveTab), isPopupOpen)

	// Stitch our interface frame layers vertically
	mainLayout := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		// statusGrid,
		body,
		logView,
		footer,
	)

	// F. PHYSICAL OVERLAY RENDERING FOR POPUPS
	if m.Focus == OptionsPopupWindow {
		return components.RenderOverlay(mainLayout, m.OptionsPopup.View())
	}
	if m.Focus == PopupWindow {
		return components.RenderOverlay(mainLayout, m.InputPopup.View())
	}

	return mainLayout
}
