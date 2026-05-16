package main

import (
	"fmt"
	"os"

	"netui/ui/components"
	"netui/ui/views/bluetooth"
	"netui/ui/views/vpn"
	"netui/ui/views/wifi"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type mainModel struct {
	activeTab int // 0: Wifi, 1: Bluetooth, 2: VPN
	wifiView  wifi.Model
	btView    bluetooth.Model
	vpnView   vpn.Model
}

func initialModel() mainModel {
	return mainModel{
		activeTab: 0,
		wifiView:  wifi.New(),
		btView:    bluetooth.New(),
		vpnView:   vpn.New(),
	}
}

func (m mainModel) Init() tea.Cmd {
	// Trigger setup sequences for all background structures on start
	return tea.Batch(
		m.wifiView.Init(),
		m.btView.Init(),
		m.vpnView.Init(),
	)
}

func (m mainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "1":
			m.activeTab = 0
			return m, nil
		case "2":
			m.activeTab = 1
			return m, nil
		case "3":
			m.activeTab = 2
			return m, nil
		}
	}

	// Route events to the dedicated module currently displayed on screen
	switch m.activeTab {
	case 0:
		m.wifiView, cmd = m.wifiView.Update(msg)
	case 1:
		m.btView, cmd = m.btView.Update(msg)
	case 2:
		m.vpnView, cmd = m.vpnView.Update(msg)
	}

	return m, cmd
}

func (m mainModel) View() string {
	containerStyle := lipgloss.NewStyle().Margin(1, 2).Width(74).Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#4B5563"))

	// Draw components/header
	headerView := components.RenderHeader(m.activeTab)

	// Pull active frame string
	var coreBody string
	switch m.activeTab {
	case 0:
		coreBody = m.wifiView.View()
	case 1:
		coreBody = m.btView.View()
	case 2:
		coreBody = m.vpnView.View()
	}

	// Draw components/footer
	footerView := components.RenderFooter(m.activeTab, false)

	return containerStyle.Render(headerView + coreBody + footerView)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running netui application: %v", err)
		os.Exit(1)
	}
}
