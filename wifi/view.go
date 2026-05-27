package wifi

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.Loading {
		return "\n Connecting to System Bus Interfaces..."
	}
	if m.Err != nil {
		return fmt.Sprintf("\n  ❌ Error: %v", m.Err)
	}

	var segments []string

	// 1. Status Banner
	bannerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6")).Bold(true).Padding(0, 1)
	if m.Scanning {
		segments = append(segments, bannerStyle.Render("⚡ [SCANNING] Active Airwaves (Refreshing every 5s | Press 's' to Stop)"))
	} else {
		segments = append(segments, bannerStyle.Render("🛰️  [OFFLINE MODE] Ready (Press 's' to scan | 'p' to toggle power)"))
	}

	// 2. Conditional Interface Block Rendering (Loops Removed!)
	if m.Scanning {
		// --- SCANNING ON: Let the table component render the nearby access points ---
		apBlock := "\n Nearby Access Points\n\n" + m.Table.View() + "\n"
		segments = append(segments, lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Render(apBlock))
	} else {
		// --- SCANNING OFF: Show Hardware & the table component for saved configurations ---
		adapterBlock := fmt.Sprintf(
			"\n🎛️  [Hardware Settings]\n  Interface:    %s\n  Link Status:  %s\n  Radio Power:  %s\n",
			m.Adapter.Interface, m.Adapter.State, map[bool]string{true: "Enabled [ON]", false: "Disabled [OFF]"}[m.Adapter.Enabled],
		)
		segments = append(segments, lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render(adapterBlock))

		savedBlock := "\n💾 [Saved Station Configuration Registry]\n\n" + m.Table.View() + "\n"
		segments = append(segments, lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B")).Render(savedBlock))
	}

	screen := lipgloss.JoinVertical(lipgloss.Left, segments...)

	// 3. Popup Overlay Processing
	if m.UIState == StatePasswordInput {
		hiddenPassword := strings.Repeat("*", len(m.PasswordInput))
		box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#EF4444")).Padding(1, 3).Margin(1, 2)
		popup := box.Render(fmt.Sprintf("Enter Password for: %s\n\n %s_", m.SelectedAP.SSID, hiddenPassword))
		return lipgloss.JoinVertical(lipgloss.Center, screen, popup)
	}

	if m.UIState == StateSavedActionsMenu {
		options := []string{"Toggle AutoConnect", "Forget Network/Delete"}
		var menuLines []string
		menuLines = append(menuLines, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#EF4444")).Render("── Saved Network Actions ──"))
		for i, opt := range options {
			if m.MenuCursor == i {
				menuLines = append(menuLines, fmt.Sprintf(" > \x1b[1m%s\x1b[0m", opt))
			} else {
				menuLines = append(menuLines, fmt.Sprintf("   %s", opt))
			}
		}
		box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#F59E0B")).Padding(1, 2).Margin(1, 2)
		return lipgloss.JoinVertical(lipgloss.Center, screen, box.Render(strings.Join(menuLines, "\n")))
	}

	return screen
}
