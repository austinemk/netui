package main

import (
	"fmt"
	"os"

	"netui/bluetooth"
	"netui/vpn"
	"netui/wifi"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	// Initialize the advanced AppModel from app.go instead of the old mainModel
	initialAppModel := AppModel{
		Focus:     MainWindow,
		ActiveTab: WifiTab,
		WifiView:  wifi.New(),
		BtView:    bluetooth.New(),
		VpnView:   vpn.New(),
		// OptionsPopup and InputPopup will initialize with their zero-values
		// or you can explicitly instantiate them here if they have New() constructor funcs.
	}

	p := tea.NewProgram(initialAppModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running netui application: %v", err)
		os.Exit(1)
	}
}
