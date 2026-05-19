package bluetooth

import (
	"fmt"
	"strings"

	"netui/config"

	"github.com/charmbracelet/lipgloss"
)

func (m Model) View() string {
	var header string
	var bodyLines []string

	// 1. Setup Active Dynamic Headers
	if m.Scanning {
		header = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Bold(true).Render("✨ Scanning Mode: Active Live Discovery Feed...")
	} else {
		header = lipgloss.NewStyle().Foreground(lipgloss.Color("#3B82F6")).Bold(true).Render("📦 Saved Devices Storage (Offline Configuration Manager)")
	}

	// SPACE SAVING MECHANISM: Hide general settings entirely while actively scanning
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

	// --- EXCEPTION INTERCEPT DISPLAY ---
	if m.Err != nil {
		errorStyle := config.Styles.Notice
		bodyLines = append(bodyLines, fmt.Sprintf("  ⚠️  Problem: %s", errorStyle.Render(m.Err.Error())))
	}

	visibleDevices := m.getFilteredDevices()
	deviceLineOffsets := make(map[int]int)

	if len(visibleDevices) == 0 {
		if m.Scanning {
			bodyLines = append(bodyLines, "  🔄 Listening for local broadcasts over interfaces...")
		} else {
			bodyLines = append(bodyLines, "    No saved devices found in storage cache. Press [s] to scan.")
		}
	} else {
		for i, dev := range visibleDevices {
			cursor := " "
			if m.Cursor == i {
				cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Render(">")
			}

			iconSymbol := IconGenericBluetooth.String()
			if dev.Icon != "" {
				iconSymbol = FromString(dev.Icon).String()
			}

			status := fmt.Sprintf(" %s ", iconSymbol)
			if dev.Connected {
				status = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Render("")
			}

			// Save the precise layout line coordinate array index for this entry
			deviceLineOffsets[i] = len(bodyLines)
			bodyLines = append(bodyLines, fmt.Sprintf("  %s%s%-25s \t[%s]", cursor, status, dev.Name, dev.MAC))
		}

		if m.Scanning {
			bodyLines = append(bodyLines, "", lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Render("  🔄 Scanning actively... Press [s] to lock listings and manage configurations offline."))
		}
	}

	// 2. IN-LINE POPUP RENDERING INTERCEPTOR
	if m.PopupMenu.Active {
		popupRaw := m.PopupMenu.View()
		popupLines := strings.Split(popupRaw, "\n")

		if targetLineIdx, ok := deviceLineOffsets[m.Cursor]; ok {
			startOverlayLine := targetLineIdx + 1

			for offset, pLine := range popupLines {
				destIdx := startOverlayLine + offset
				styledPopupLine := "      " + pLine

				if destIdx < len(bodyLines) {
					// Overwrite upcoming list rows without moving layout targets down
					bodyLines[destIdx] = styledPopupLine
				} else {
					bodyLines = append(bodyLines, styledPopupLine)
				}
			}
		}
	}

	return fmt.Sprintf("%s\n%s", header, strings.Join(bodyLines, "\n"))
}
