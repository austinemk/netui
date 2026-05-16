package bluetooth

import (
	"fmt"

	"netui/pkg/bluetooth"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SubTab int

const (
	ScanSubTab SubTab = iota
	SavedSubTab
)

type (
	DevicesLoadedMsg []bluetooth.Device
	ErrMsg           error
)

type Model struct {
	ActiveSubTab SubTab
	Devices      []bluetooth.Device
	Cursor       int
	Loading      bool
	Err          error
}

func New() Model {
	return Model{
		ActiveSubTab: ScanSubTab,
		Loading:      true,
	}
}

func FetchDevicesCmd() tea.Cmd {
	return func() tea.Msg {
		devs, err := bluetooth.ScanDevices()
		if err != nil {
			return ErrMsg(err)
		}
		return DevicesLoadedMsg(devs)
	}
}

func (m Model) Init() tea.Cmd {
	return FetchDevicesCmd()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case DevicesLoadedMsg:
		m.Devices = msg
		m.Loading = false
		return m, nil

	case ErrMsg:
		m.Err = msg
		m.Loading = false
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
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
			if m.Cursor < len(m.Devices)-1 {
				m.Cursor++
			}
		}
	}
	return m, nil
}

func (m Model) View() string {
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6")).Underline(true).Bold(true).Padding(0, 1)
	inactiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Padding(0, 1)

	var subtabs string
	if m.ActiveSubTab == ScanSubTab {
		subtabs = lipgloss.JoinHorizontal(lipgloss.Top, activeStyle.Render("• Scan Available"), inactiveStyle.Render("Paired Devices"))
	} else {
		subtabs = lipgloss.JoinHorizontal(lipgloss.Top, inactiveStyle.Render("Scan Available"), activeStyle.Render("• Paired Devices"))
	}

	var body string
	if m.Loading {
		body = "\n  🔄 Querying bluetooth controller interfaces..."
	} else {
		body = "\n"
		for i, dev := range m.Devices {
			// Basic filtering logic to show paired vs unpaired devices in separate tabs
			if m.ActiveSubTab == SavedSubTab && !dev.Paired {
				continue
			}

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
	}

	return lipgloss.JoinVertical(lipgloss.Left, subtabs, body)
}
