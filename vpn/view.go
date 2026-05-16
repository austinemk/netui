package vpn

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type SubTab int

const (
	ListSubTab SubTab = iota
	AddSubTab
)

type (
	VpnsLoadedMsg []Connection
	ErrMsg        error
)

type Model struct {
	ActiveSubTab SubTab
	VPNS         []Connection
	Cursor       int
	Loading      bool
	Err          error
}

func New() Model {
	return Model{
		ActiveSubTab: ListSubTab,
		Loading:      true,
	}
}

func FetchVpnsCmd() tea.Cmd {
	return func() tea.Msg {
		conns, err := GetVPNConnections()
		if err != nil {
			return ErrMsg(err)
		}
		return VpnsLoadedMsg(conns)
	}
}

func (m Model) Init() tea.Cmd {
	return FetchVpnsCmd()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case VpnsLoadedMsg:
		m.VPNS = msg
		m.Loading = false
		return m, nil

	case ErrMsg:
		m.Err = msg
		m.Loading = false
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "right", "tab", "left", "shift+tab":
			if m.ActiveSubTab == ListSubTab {
				m.ActiveSubTab = AddSubTab
			} else {
				m.ActiveSubTab = ListSubTab
			}
			m.Cursor = 0
			return m, nil

		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			if m.Cursor < len(m.VPNS)-1 {
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
	if m.ActiveSubTab == ListSubTab {
		subtabs = lipgloss.JoinHorizontal(lipgloss.Top, activeStyle.Render("• Saved List"), inactiveStyle.Render("Add New Tunnel"))
	} else {
		subtabs = lipgloss.JoinHorizontal(lipgloss.Top, inactiveStyle.Render("Saved List"), activeStyle.Render("• Add New Tunnel"))
	}

	var body string
	if m.Loading {
		body = "\n  🔄 Interrogating active network tunnels..."
	} else if m.ActiveSubTab == AddSubTab {
		body = "\n  🔒 [WireGuard / OpenVPN Configurations]\n     Importing endpoint profiles via TUI forms is pending setup."
	} else {
		body = "\n"
		for i, v := range m.VPNS {
			cursor := " "
			if m.Cursor == i {
				cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6")).Render(">")
			}

			lockIcon := " 🔒 "
			if v.Active {
				lockIcon = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Render(" 🛡️  ")
			}

			body += fmt.Sprintf("  %s%s%-25s \t(%s)\n", cursor, lockIcon, v.Name, v.Type)
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, subtabs, body)
}
