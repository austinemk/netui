// Package app the pkg entry
package app

import (
	"github.com/austinemk/linktui/pkg/bluetooth"
	"github.com/austinemk/linktui/pkg/vpn"
	"github.com/austinemk/linktui/pkg/wifi"

	tea "charm.land/bubbletea/v2"
)

type Tab int

const (
	WifiTab Tab = iota
	BluetoothTab
	VpnTab
)

type AppModel struct {
	ActiveTab  Tab
	WifiView   wifi.Model
	BtView     bluetooth.Model
	VpnView    vpn.Model
	LogMessage string
	SizeError  string
	LoadedTabs map[Tab]bool
	BusReady   bool
}

func (m AppModel) Init() tea.Cmd {
	switch m.ActiveTab {
	case WifiTab:
		return m.WifiView.Init()
	case BluetoothTab:
		return m.BtView.Init()
	case VpnTab:
		return m.VpnView.Init()
	}
	return nil
}

func (m *AppModel) lazyLoadTab(tab Tab) tea.Cmd {
	if m.LoadedTabs == nil {
		m.LoadedTabs = make(map[Tab]bool)
	}

	if m.LoadedTabs[tab] {
		return nil
	}

	m.LoadedTabs[tab] = true

	switch tab {
	case WifiTab:
		return m.WifiView.Init()
	case BluetoothTab:
		return m.BtView.Init()
	case VpnTab:
		return m.VpnView.Init()
	}
	return nil
}
