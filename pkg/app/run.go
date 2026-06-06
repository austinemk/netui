package app

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/austinemk/linktui/pkg/bluetooth"
	"github.com/austinemk/linktui/pkg/bus"
	"github.com/austinemk/linktui/pkg/config"
	"github.com/austinemk/linktui/pkg/vpn"
	"github.com/austinemk/linktui/pkg/wifi"

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

	// 1. Initialize shared D-Bus connection (used by wifi, vpn, bluetooth)
	if err := bus.Init(); err != nil {
		fmt.Fprintf(os.Stderr, "D-Bus unavailable: %v\n", err)
		os.Exit(1)
	}
	defer bus.Close()

	// 2. Define and parse terminal flags
	tabFlag := flag.String("tab", "wifi", "Initial tab to open (wifi, bluetooth, vpn)")
	flag.Parse()

	// 3. Map string input to internal Tab type
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
		BusReady:   true, // bus.Init() already succeeded above
	}

	var finalModel tea.Model

	// 5. Defer cleanup safely
	defer func() {
		if finalModel != nil {
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
	p := tea.NewProgram(&initialAppModel)
	finalModel, err = p.Run()
	if err != nil {
		fmt.Printf("Error running netui application: %v\n", err)
		os.Exit(1)
	}
}
