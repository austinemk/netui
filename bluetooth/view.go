package bluetooth

import (
	"fmt"
	"strings"

	"netui/config"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	var bodyLines []string

	// --- SECTION 1: HEADER & CONFIGURATION CONTROLS ---
	if !m.Scanning {
		powStatus := lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render("Off")
		if m.Powered {
			powStatus = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Render("On")
		}
		discStatus := lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render("Off")
		if m.Discoverable {
			discStatus = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Render("On")
		}
		pairStatus := lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Render("Off")
		if m.Pairable {
			pairStatus = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Render("On")
		}

		bodyLines = append(bodyLines, "", "  Adapter Profiles (Press hotkey to change configuration):")
		bodyLines = append(bodyLines, fmt.Sprintf("    [p] Power: %s", powStatus))
		bodyLines = append(bodyLines, fmt.Sprintf("    [d] Discoverable: %s", discStatus))
		bodyLines = append(bodyLines, fmt.Sprintf("    [b] Pairable: %s", pairStatus))
		bodyLines = append(bodyLines, "", "  Known Saved Storage Devices:")
	} else {
		bodyLines = append(bodyLines, "", "  Discovered Devices In Range:")
	}

	// --- SECTION 2: EXCEPTION INTERCEPT DISPLAY ---
	if m.Err != nil {
		errorStyle := config.Styles.Notice
		bodyLines = append(bodyLines, fmt.Sprintf("  ⚠️  Problem: %s", errorStyle.Render(m.Err.Error())))
	}

	// --- SECTION 3: DATA GRID MATRIX VIEW ---
	visibleDevices := m.getFilteredDevices()

	if len(visibleDevices) == 0 {
		if m.Scanning {
			bodyLines = append(bodyLines, " "+config.Styles.Notice.Render("listening for local broadcasts over interfaces..."))
		} else {
			bodyLines = append(bodyLines, "  No saved devices found in storage cache. Press [s] to scan.")
		}
	} else {
		// Output the active interactive table layout instead of plain arrays
		bodyLines = append(bodyLines, m.Table.View())

		if m.Scanning {
			bodyLines = append(bodyLines, "", lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render("  🔄 Scanning actively... Press [s] to lock listings and manage configurations offline."))
		}
	}

	// --- SECTION 4: MODAL OVERLAY INJECTION ---
	if m.PopupMenu.Active {
		bodyLines = append(bodyLines, "", "  ─── Options Menu Active ─────────────────────────────")
		popupRaw := m.PopupMenu.View()
		for _, pLine := range strings.Split(popupRaw, "\n") {
			bodyLines = append(bodyLines, "      "+pLine)
		}
		bodyLines = append(bodyLines, "  ─────────────────────────────────────────────────────")
	}

	// --- SECTION 5: CONTAINER PACKAGING ---
	fullContent := strings.Join(bodyLines, "\n")

	// Update the content on the local copy and immediately render it
	m.Viewport.SetContent(fullContent)
	return m.Viewport.View()
}
