package vpn

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
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
			doneRow = "> [ SUBMIT AND REGISTER ]"
		}
		fLines = append(fLines, "\n"+doneRow)
		return lipgloss.JoinVertical(lipgloss.Left, fLines...)
	}

	var sections []string
	sections = append(sections, lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6")).Render("🔒 Register Link tunnels (WireGuard / OpenVPN)\n"))

	listBlock := ""
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
				menuLines = append(menuLines, fmt.Sprintf(" > %s", opt))
			} else {
				menuLines = append(menuLines, fmt.Sprintf("   %s", opt))
			}
		}

		popupDialog := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("9")).
			Padding(1, 3).
			Render(strings.Join(menuLines, "\n"))

		// Render multi-layered stacked screen blocks together using style method positions
		return lipgloss.JoinVertical(lipgloss.Left, screen, "\n", popupDialog)
	}

	return screen
}
