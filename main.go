package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"corntui/bluetooth"
	"corntui/config" // Imported your local config package
	"corntui/vpn"
	"corntui/wifi"

	tea "charm.land/bubbletea/v2"
)

func main() {
	// 0. Locate and attempt to load the ~/.config/corntui/config.toml file
	userConfigDir, err := os.UserConfigDir()
	if err == nil {
		// Constructs ~/.config/corntui/config.toml cleanly across OS targets
		configPath := filepath.Join(userConfigDir, "corntui", "config.toml")

		// If it errors or doesn't exist, LoadConfig handles it internally
		// and safely falls back to your built-in defaults.
		_ = config.LoadConfig(configPath)
	} else {
		// Fallback to defaults if we can't fetch the user config directory profile
		_ = config.LoadConfig("")
	}

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
