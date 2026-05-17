package bluetooth

import (
	"fmt"
	"time"

	"netui/components" // Ensure this points to your actual module path

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SubTab int

const (
	ScanSubTab SubTab = iota
	SavedSubTab
)

type (
	DevicesLoadedMsg []Device
	ScanStoppedMsg   struct{}
	TickMsg          time.Time
	ErrMsg           error
	ActionSuccessMsg string
)

type Model struct {
	ActiveSubTab SubTab
	Devices      []Device
	Cursor       int
	Status       components.OpStatus
	Err          error

	// Embedded context options menu
	PopupMenu   components.OptionsPopupModel
	SelectedMac string
}

func New() Model {
	return Model{
		ActiveSubTab: ScanSubTab,
		Status:       components.StatusIdle,
		PopupMenu:    components.NewOptionsPopup("", []string{}),
	}
}

func FetchDevicesCmd() tea.Cmd {
	return func() tea.Msg {
		devs, err := FetchCachedDevices()
		if err != nil {
			return ErrMsg(err)
		}
		return DevicesLoadedMsg(devs)
	}
}

func StartScanCmd() tea.Cmd {
	return func() tea.Msg {
		_ = ControlScan(true)
		return nil
	}
}

func StopScanCmd() tea.Cmd {
	return func() tea.Msg {
		_ = ControlScan(false)
		return ScanStoppedMsg{}
	}
}

// PollDevicesTicker regularly pulls live devices during an active scan session
func PollDevicesTicker() tea.Cmd {
	return tea.Tick(1*time.Second, func(t time.Time) tea.Msg {
		return TickMsg(t)
	})
}

func ExecuteActionCmd(action string, mac string) tea.Cmd {
	return func() tea.Msg {
		var err error
		switch action {
		case "Connect":
			_ = Trust(mac)
			err = Connect(mac)
		case "Disconnect":
			err = Disconnect(mac)
		case "Pair":
			err = Pair(mac)
		case "Trust":
			err = Trust(mac)
		case "Untrust":
			err = Distrust(mac)
		case "Remove / Unpair":
			err = Remove(mac)
		}

		if err != nil {
			return ErrMsg(err)
		}
		return ActionSuccessMsg(fmt.Sprintf("%s action successful", action))
	}
}

func (m Model) Init() tea.Cmd {
	return FetchDevicesCmd()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	if m.PopupMenu.Active {
		var cmd tea.Cmd
		m.PopupMenu, cmd = m.PopupMenu.Update(msg)

		if selectMsg, ok := msg.(components.OptionSelectedMsg); ok {
			m.PopupMenu.Active = false
			m.Status = components.StatusConnecting
			return m, ExecuteActionCmd(selectMsg.Option, m.SelectedMac)
		}
		return m, cmd
	}

	switch msg := msg.(type) {

	case DevicesLoadedMsg:
		m.Devices = msg
		if m.Status == components.StatusConnecting {
			m.Status = components.StatusIdle
		}
		return m, nil

	case ScanStoppedMsg:
		m.Status = components.StatusIdle
		return m, FetchDevicesCmd() // Final catch after stopping scan

	case ActionSuccessMsg:
		m.Status = components.StatusIdle
		return m, FetchDevicesCmd()

	case TickMsg:
		if m.Status == components.StatusScanning {
			// Pull items concurrently AND spin up the next tick cycle
			return m, tea.Batch(FetchDevicesCmd(), PollDevicesTicker())
		}
		return m, nil

	case ErrMsg:
		m.Err = msg
		m.Status = components.StatusError
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Critical: Turn off scanning on abrupt application shutdowns
			if m.Status == components.StatusScanning {
				_ = ControlScan(false)
			}
			return m, tea.Quit

		case "s":
			if m.Status == components.StatusScanning {
				// User manually requests to stop discovery stream
				return m, StopScanCmd()
			} else {
				// Start infinite scanning state loop until explicitly closed
				m.Status = components.StatusScanning
				m.Cursor = 0
				return m, tea.Batch(StartScanCmd(), PollDevicesTicker())
			}

		case "right", "tab", "left", "shift+tab":
			if m.ActiveSubTab == ScanSubTab {
				m.ActiveSubTab = SavedSubTab
			} else {
				m.ActiveSubTab = ScanSubTab
			}
			m.Cursor = 0
			return m, nil

		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			filteredCount := len(m.getFilteredDevices())
			if m.Cursor < filteredCount-1 {
				m.Cursor++
			}

		case "enter":
			visibleDevices := m.getFilteredDevices()
			if len(visibleDevices) == 0 || m.Cursor >= len(visibleDevices) {
				return m, nil
			}

			targetDev := visibleDevices[m.Cursor]
			m.SelectedMac = targetDev.MAC

			var opts []string
			if targetDev.Connected {
				opts = append(opts, "Disconnect")
			} else {
				opts = append(opts, "Connect")
			}

			if !targetDev.Paired {
				opts = append(opts, "Pair")
			}

			if !targetDev.Trusted {
				opts = append(opts, "Trust")
			} else {
				opts = append(opts, "Untrust")
			}

			if targetDev.Paired || targetDev.Trusted {
				opts = append(opts, "Remove / Unpair")
			}

			m.PopupMenu.Title = fmt.Sprintf("Actions: %s", targetDev.Name)
			m.PopupMenu.Options = opts
			m.PopupMenu.Cursor = 0
			m.PopupMenu.Active = true
		}
	}
	return m, nil
}

func (m Model) getFilteredDevices() []Device {
	var filtered []Device
	for _, dev := range m.Devices {
		if m.ActiveSubTab == SavedSubTab && !dev.Paired {
			continue
		}
		filtered = append(filtered, dev)
	}
	return filtered
}

func (m Model) View() string {
	if m.Err != nil {
		return fmt.Sprintf("\n  ❌ Error: %v\n", m.Err)
	}

	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6")).Underline(true).Bold(true).Padding(0, 1)
	inactiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Padding(0, 1)

	var subtabs string
	if m.ActiveSubTab == ScanSubTab {
		subtabs = lipgloss.JoinHorizontal(lipgloss.Top, activeStyle.Render("• Scan Available"), inactiveStyle.Render("Paired Devices"))
	} else {
		subtabs = lipgloss.JoinHorizontal(lipgloss.Top, inactiveStyle.Render("Scan Available"), activeStyle.Render("• Paired Devices"))
	}

	var body string

	statusIndicator := lipgloss.NewStyle().Foreground(lipgloss.Color("#A78BFA")).Render(fmt.Sprintf(" [State: %s]", m.Status.String()))
	subtabs = lipgloss.JoinHorizontal(lipgloss.Center, subtabs, statusIndicator)

	visibleDevices := m.getFilteredDevices()

	if len(visibleDevices) == 0 {
		if m.Status == components.StatusScanning {
			body = "\n  🔄 Querying bluetooth interfaces (Scanning active...)\n"
		} else {
			body = "\n  No devices found. Press [s] to start scanning.\n"
		}
	} else {
		body = "\n"
		for i, dev := range visibleDevices {
			cursor := " "
			if m.Cursor == i {
				cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6")).Render(">")
			}

			status := " 🎧 "
			if dev.Connected {
				status = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Render(" 🟢 ")
			}

			body += fmt.Sprintf("  %s%s%-25s \t[%s]\n", cursor, status, dev.Name, dev.MAC)
		}

		if m.Status == components.StatusScanning {
			body += lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render("\n  🔄 Scanning active... Press [s] to stop and lock listings.\n")
		} else {
			body += lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render("\n  Press [s] to trigger a new wireless discovery scan.\n")
		}
	}

	if m.PopupMenu.Active {
		popupView := m.PopupMenu.View()
		return lipgloss.JoinVertical(lipgloss.Left, subtabs, body, "\n", popupView)
	}

	return lipgloss.JoinVertical(lipgloss.Left, subtabs, body)
}
