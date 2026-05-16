package wifi

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SubTab state
type SubTab int

const (
	AvailableSubTab SubTab = iota
	SavedSubTab
)

// Messages for background threads to talk back to Bubble Tea
type (
	NetworksLoadedMsg []Network
	ProfilesLoadedMsg []SavedProfile
	ErrMsg            error
)

type Model struct {
	ActiveSubTab SubTab
	Available    []Network
	Saved        []SavedProfile
	Cursor       int
	Loading      bool
	Err          error
}

func New() Model {
	return Model{
		ActiveSubTab: AvailableSubTab,
		Loading:      true,
	}
}

// Commands to trigger background processing
func FetchNetworksCmd() tea.Cmd {
	return func() tea.Msg {
		nets, err := ScanNetworks()
		if err != nil {
			return ErrMsg(err)
		}
		return NetworksLoadedMsg(nets)
	}
}

func FetchProfilesCmd() tea.Cmd {
	return func() tea.Msg {
		profs, err := GetSavedProfiles()
		if err != nil {
			return ErrMsg(err)
		}
		return ProfilesLoadedMsg(profs)
	}
}

func (m Model) Init() tea.Cmd {
	return FetchNetworksCmd()
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case NetworksLoadedMsg:
		m.Available = msg
		m.Loading = false
		return m, nil

	case ProfilesLoadedMsg:
		m.Saved = msg
		m.Loading = false
		return m, nil

	case ErrMsg:
		m.Err = msg
		m.Loading = false
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		// Switch Subtabs with left/right arrows
		case "right", "tab":
			if m.ActiveSubTab == AvailableSubTab {
				m.ActiveSubTab = SavedSubTab
				m.Loading = true
				m.Cursor = 0
				return m, FetchProfilesCmd()
			} else {
				m.ActiveSubTab = AvailableSubTab
				m.Loading = true
				m.Cursor = 0
				return m, FetchNetworksCmd()
			}

		case "left", "shift+tab":
			if m.ActiveSubTab == SavedSubTab {
				m.ActiveSubTab = AvailableSubTab
				m.Loading = true
				m.Cursor = 0
				return m, FetchNetworksCmd()
			} else {
				m.ActiveSubTab = SavedSubTab
				m.Loading = true
				m.Cursor = 0
				return m, FetchProfilesCmd()
			}

		// List Navigation
		case "up", "k":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "down", "j":
			max := 0
			if m.ActiveSubTab == AvailableSubTab {
				max = len(m.Available) - 1
			} else {
				max = len(m.Saved) - 1
			}
			if m.Cursor < max {
				m.Cursor++
			}
		}
	}

	return m, cmd
}

func (m Model) View() string {
	// Styling layouts
	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6")).Underline(true).Bold(true).Padding(0, 1)
	inactiveStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Padding(0, 1)

	// Subtab headers
	var subtabs string
	if m.ActiveSubTab == AvailableSubTab {
		subtabs = lipgloss.JoinHorizontal(lipgloss.Top, activeStyle.Render("• Available Networks"), inactiveStyle.Render("Manage Saved"))
	} else {
		subtabs = lipgloss.JoinHorizontal(lipgloss.Top, inactiveStyle.Render("Available Networks"), activeStyle.Render("• Manage Saved"))
	}

	// Content body
	var body string
	if m.Loading {
		body = "\n  🔄 Scanning system interfaces..."
	} else if m.Err != nil {
		body = fmt.Sprintf("\n  ❌ Error: %v", m.Err)
	} else {
		body = "\n"
		if m.ActiveSubTab == AvailableSubTab {
			for i, net := range m.Available {
				cursor := " "
				if m.Cursor == i {
					cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6")).Render(">")
				}
				activeIndicator := "  "
				if net.IsActive {
					activeIndicator = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Render("✔ ")
				}

				body += fmt.Sprintf("  %s %s%-25s \t%s \t%s\n", cursor, activeIndicator, net.SSID, net.Signal, net.Security)
			}
		} else {
			for i, prof := range m.Saved {
				cursor := " "
				if m.Cursor == i {
					cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6")).Render(">")
				}
				body += fmt.Sprintf("  %s 💾 %-25s \t(UUID: %s...)\n", cursor, prof.Name, prof.UUID[:8])
			}
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, subtabs, body)
}
