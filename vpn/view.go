package vpn

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	if m.Loading {
		return "\n  🔄 Querying Active secure interfaces via System Bus..."
	}
	if m.Err != nil {
		return fmt.Sprintf("\n  ❌ System Failure Hook: %v", m.Err)
	}

	// FIXED: Direct View-State Routing. If in the input state, render ONLY the form.
	if m.UIState == StateAddForm {
		var fLines []string
		fLines = append(fLines, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#10B981")).Render("── Create WireGuard Interface Profile ──\n"))

		fields := []struct {
			id    FormField
			label string
		}{
			{FieldProfileName, "Connection Name : "},
			{FieldInterfaceName, "Interface (wg0) : "},
			{FieldPrivateKey, "Private Key     : "},
			{FieldPeerEndpoint, "Peer Endpoint   : "},
			{FieldPeerPublicKey, "Peer Public Key : "},
		}

		for _, f := range fields {
			cursorPrefix := "  "
			if m.ActiveField == f.id {
				cursorPrefix = "> "
			}
			fLines = append(fLines, fmt.Sprintf("%s%s%s", cursorPrefix, f.label, m.FormInputs[f.id]))
		}

		doneRow := "  [ SUBMIT AND REGISTER ]"
		if m.ActiveField == FieldDone {
			doneRow = "> \x1b[1m[ SUBMIT AND REGISTER ]\x1b[0m"
		}
		fLines = append(fLines, "\n"+doneRow)
		fLines = append(fLines, lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render("\n(Press 'Esc' to abandon form)"))

		box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#10B981")).Padding(1, 4).Margin(1, 2)
		return box.Render(strings.Join(fLines, "\n"))
	}

	// --- OTHERWISE: Render Saved Profiles Dashboard ---
	var sections []string

	// 1. Status Monitoring Bar
	hasActive := false
	for _, t := range m.Tunnels {
		if t.Active {
			hasActive = true
			break
		}
	}

	bannerStyle := lipgloss.NewStyle().Bold(true).Padding(0, 1)
	if hasActive {
		sections = append(sections, bannerStyle.Foreground(lipgloss.Color("#10B981")).Render("🛡️  TUNNEL SECURITY: ACTIVE (Overlay active)"))
	} else {
		sections = append(sections, bannerStyle.Foreground(lipgloss.Color("#F59E0B")).Render("🔓 TUNNEL SECURITY: UNPROTECTED (Press 'a' to add new wireguard profile)"))
	}

	// 2. Saved Profile Registry List
	listBlock := "\n🔒 [Configured Secure Overlays & Routing Tunnels]\n"
	if len(m.Tunnels) == 0 {
		listBlock += "  No endpoints registered in system databases.\n"
	}
	for i, t := range m.Tunnels {
		cursor := " "
		if m.Cursor == i && m.UIState == StateNormal {
			cursor = ">"
		}

		stateStr := lipgloss.NewStyle().Foreground(lipgloss.Color("#9CA3AF")).Render("Inactive")
		if t.Active {
			stateStr = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Bold(true).Render("Active 󰌆")
		}
		listBlock += fmt.Sprintf("  %s %-25s \t%-15s \t[%s]\n", cursor, t.Name, t.Type, stateStr)
	}
	sections = append(sections, lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6")).Render(listBlock))

	screen := lipgloss.JoinVertical(lipgloss.Left, sections...)

	// 3. Action Context Dialog Overlay (Only displays over the dashboard registry list)
	if m.UIState == StateActionsMenu {
		target := m.Tunnels[m.Cursor]
		actLabel := "Activate Link"
		if target.Active {
			actLabel = "Deactivate Link"
		}
		options := []string{actLabel, "Cancel Operation"}

		var menuLines []string
		menuLines = append(menuLines, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#EF4444")).Render(fmt.Sprintf("── Actions: %s ──", target.Name)))
		for i, opt := range options {
			if m.MenuCursor == i {
				menuLines = append(menuLines, fmt.Sprintf(" > \x1b[1m%s\x1b[0m", opt))
			} else {
				menuLines = append(menuLines, fmt.Sprintf("   %s", opt))
			}
		}
		box := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("#EF4444")).Padding(1, 2).Margin(1, 2)
		return lipgloss.JoinVertical(lipgloss.Center, screen, box.Render(strings.Join(menuLines, "\n")))
	}

	return screen
}
