package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"corntui/bluetooth"
	"corntui/vpn"
	"corntui/wifi"

	tea "charm.land/bubbletea/v2"
)

func main() {
	// 1. Define and parse terminal flags
	tabFlag := flag.String("tab", "wifi", "Initial tab to open (wifi, bluetooth, vpn)")
	flag.Parse()

	// 2. Map string input to our internal Tab type
	var initialTab Tab
	switch strings.ToLower(*tabFlag) {
	case "bluetooth", "bt", "2":
		initialTab = BluetoothTab
	case "vpn", "3":
		initialTab = VpnTab
	default:
		initialTab = WifiTab // Fallback / default
	}

	// 3. Setup a pointer to hold the BlueZ client if needed right away
	// var btClient *bluetooth.BlueZClient
	var err error

	// ONLY connect to BlueZ immediately if the user requested the Bluetooth tab on startup
	/*if initialTab == BluetoothTab {
		btClient, err = bluetooth.NewBlueZClient()
	}

	btView := bluetooth.New()
	if err != nil {
		btView.Err = err
	} else if btClient != nil {
		btView.Client = btClient
	}*/

	// 4. Initialize AppModel with tracked loaded states
	initialAppModel := AppModel{
		ActiveTab: initialTab,
		WifiView:  wifi.New(),
		BtView:    bluetooth.New(),
		VpnView:   vpn.New(),

		// Mark which view is explicitly active and loaded right now
		LoadedTabs: map[Tab]bool{initialTab: true},
	}

	var finalModel tea.Model

	// 5. Defer cleanup safely
	defer func() {
		if finalModel != nil {
			if app, ok := finalModel.(AppModel); ok {
				// Only clean up views if they were actually ever initialized
				if app.LoadedTabs[WifiTab] {
					app.WifiView.Clean()
				}
				if app.LoadedTabs[BluetoothTab] {
					app.BtView.Clean()
				}
			}
		} else {
			if initialAppModel.LoadedTabs[WifiTab] {
				initialAppModel.WifiView.Clean()
			}
			if initialAppModel.LoadedTabs[BluetoothTab] {
				initialAppModel.BtView.Clean()
			}
		}
	}()

	// 6. Run the Bubble Tea program
	p := tea.NewProgram(initialAppModel)
	finalModel, err = p.Run()
	if err != nil {
		fmt.Printf("Error running netui application: %v\n", err)
		os.Exit(1)
	}
}
