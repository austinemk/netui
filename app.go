package main

import (
	"fmt"

	"netui/bluetooth"
	"netui/components"
	"netui/config"
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
	// Track if the current screen size is invalid
	SizeError string
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

	// Intercept Window Resizing and enforce strict sizing rules
	if windowMsg, ok := msg.(tea.WindowSizeMsg); ok {

		// 1. Check if config values are less than the hardcoded absolute minimums
		if config.WindowWidth < 70 || config.WindowHeight < 25 {
			m.SizeError = fmt.Sprintf(
				"⚠️  Configuration Error!\n\n  Configured sizes are too small for layout items.\n  Config: %dx%d\n  Absolute Minimum: 70x25",
				config.WindowWidth, config.WindowHeight,
			)
			return m, nil
		}

		// 2. Check if config values are larger than the actual physical window size
		if config.WindowWidth > windowMsg.Width || config.WindowHeight > windowMsg.Height {
			m.SizeError = fmt.Sprintf(
				"⚠️  Terminal screen is too small!\n\n  Current Window: %dx%d\n  Config Demands: %dx%d\n\n Please resize your terminal window.",
				windowMsg.Width, windowMsg.Height, config.WindowWidth, config.WindowHeight,
			)
			return m, nil
		}

		// If everything passes, clean up any previous sizing errors
		m.SizeError = ""

		// Enforce and clamp the dimensions to the strict config boundaries
		windowMsg.Width = config.WindowWidth
		windowMsg.Height = config.WindowHeight

		// Overwrite the message payload with our clamped dimensions
		msg = windowMsg
	}

	// If a sizing error exists, completely block any other user inputs/updates
	if m.SizeError != "" {
		// Allow 'q' or 'ctrl+c' to still function so users can exit if they want to
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			if keyMsg.String() == "q" || keyMsg.String() == "ctrl+c" {
				return m, tea.Quit
			}
		}
		return m, nil
	}

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
			return m, tea.Quit
		case "1", "2", "3":
			m.ActiveTab = Tab(msg.String()[0] - '1')
			return m, nil
		}
	}

	// 3. ROUTE MESSAGES & KEYPRESSES TO THE ACTIVE TAB ONLY
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
	// If terminal dimensions or config setups are broken, display a clear overlay message
	if m.SizeError != "" {
		boxStyle := lipgloss.NewStyle().
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#EF4444")).
			Padding(2, 4).
			Margin(2, 2)
		return boxStyle.Render(m.SizeError)
	}

	// A. Render Global Navigation Tabs
	header := components.RenderHeader(int(m.ActiveTab))

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

	// 1. Stitch our interface frame layers vertically first
	mainLayout := lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		body,
		logView,
		footer,
	)

	// === Wrap the entire stitched layout in a global window border ===
	appBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Padding(0, 1)

	mainLayout = appBorderStyle.Render(mainLayout)

	// F. PHYSICAL OVERLAY RENDERING FOR POPUPS
	if m.Focus == OptionsPopupWindow {
		return components.RenderOverlay(mainLayout, m.OptionsPopup.View())
	}
	if m.Focus == PopupWindow {
		return components.RenderOverlay(mainLayout, m.InputPopup.View())
	}

	return mainLayout
}
