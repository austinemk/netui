package app

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"linktui/pkg/bluetooth"
	"linktui/pkg/config"
	"linktui/pkg/vpn"
	"linktui/pkg/wifi"

	tea "charm.land/bubbletea/v2"
)

// RunApp handles the flags, config loading, initialization, and cleanup lifecycle.
func RunApp() {
	// 0. Locate and attempt to load config
	userConfigDir, err := os.UserConfigDir()
	if err == nil {
		configPath := filepath.Join(userConfigDir, "linktui", "config.toml")
		_ = config.LoadConfig(configPath)
	} else {
		_ = config.LoadConfig("")
	}

	// 1. Define and parse terminal flags
	tabFlag := flag.String("tab", "wifi", "Initial tab to open (wifi, bluetooth, vpn)")
	flag.Parse()

	// 2. Map string input to internal Tab type
	var initialTab Tab
	switch strings.ToLower(*tabFlag) {
	case "bluetooth", "bt", "2":
		initialTab = BluetoothTab
	case "vpn", "3":
		initialTab = VpnTab
	default:
		initialTab = WifiTab
	}

	// 4. Initialize AppModel
	initialAppModel := AppModel{
		ActiveTab:  initialTab,
		WifiView:   wifi.New(),
		BtView:     bluetooth.New(),
		VpnView:    vpn.New(),
		LoadedTabs: map[Tab]bool{initialTab: true},
	}

	var finalModel tea.Model

	// 5. Defer cleanup safely
	defer func() {
		if finalModel != nil {
			// FIXED: Assert to *AppModel (pointer) instead of AppModel (value)
			if app, ok := finalModel.(*AppModel); ok {
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
	// FIXED: Pass &initialAppModel (the pointer) so it satisfies tea.Model
	p := tea.NewProgram(&initialAppModel)
	finalModel, err = p.Run()
	if err != nil {
		fmt.Printf("Error running netui application: %v\n", err)
		os.Exit(1)
	}
}
