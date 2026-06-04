package config

import (
	"strings"

	"charm.land/lipgloss/v2"
)

/*func DividerWithTopSpace() string {
	colorb := "8"
	if ColorBorder != "" {
		colorb = ColorBorder
	}

	return lipgloss.NewStyle().
		Border(lipgloss.Border{
			Bottom: "-",
		}, false, false, true, false). // Top border only
		BorderForeground(lipgloss.Color(colorb)).
		Width(WindowWidth - 4).
		Render("")
}

func DividerWithBottomSpace() string {
	colorb := "8"
	if ColorBorder != "" {
		colorb = ColorBorder
	}

	return lipgloss.NewStyle().
		Border(lipgloss.Border{
			Top: "-",
		}, true, false, false, false). // Top border only
		BorderForeground(lipgloss.Color(colorb)).
		Width(WindowWidth - 4).
		Render("")
}*/

func DividerBorder() string {
	borderColor := "8"
	if ColorBorder != "" {
		borderColor = ColorBorder
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color(borderColor)).Render(strings.Repeat("-", WindowWidth-2))
}

func LogBlock(content string) string {
	truncated := Truncate(content, WindowWidth-6)

	return Styles.LogBox.Render(truncated)
}

// PlaceOverlay renders `popup` centered on top of `bg`.
// bg should be the full rendered background string.
/*func PlaceOverlay(bg, popup string) string {
	bgLines := strings.Split(bg, "\n")
	popupLines := strings.Split(popup, "\n")

	popupH := len(popupLines)
	popupW := lipgloss.Width(popup)
	bgH := len(bgLines)
	bgW := WindowWidth

	// Center the popup
	startY := (bgH - popupH) / 2
	startX := (bgW - popupW) / 2
	if startY < 0 {
		startY = 0
	}
	if startX < 0 {
		startX = 0
	}

	// Pad bg if too short
	for len(bgLines) < startY+popupH {
		bgLines = append(bgLines, "")
	}

	for i, pLine := range popupLines {
		bgIdx := startY + i
		bgLine := bgLines[bgIdx]

		// Strip ANSI from bg line for safe rune slicing
		plain := stripANSI(bgLine)
		runes := []rune(plain)

		// Pad to startX if line is shorter
		for len(runes) < startX {
			runes = append(runes, ' ')
		}

		// Rebuild: left of popup | popup line | right of popup
		left := string(runes[:startX])
		endX := startX + lipgloss.Width(pLine)
		right := ""
		if endX < len(runes) {
			right = string(runes[endX:])
		}

		bgLines[bgIdx] = left + pLine + right
	}

	return strings.Join(bgLines, "\n")
}

// stripANSI removes ANSI escape codes for safe rune-level slicing
func stripANSI(s string) string {
	var b strings.Builder
	inEsc := false
	for _, r := range s {
		if r == '\x1b' {
			inEsc = true
			continue
		}
		if inEsc {
			if r == 'm' {
				inEsc = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}*/

// PlaceOverlay renders `popup` centered on top of `bg` using Lip Gloss v2.
func PlaceOverlay(bg, popup string) string {
	bgW, bgH := lipgloss.Width(bg), lipgloss.Height(bg)
	popupW, popupH := lipgloss.Width(popup), lipgloss.Height(popup)

	// Center math
	startX := (bgW - popupW) / 2
	startY := (bgH - popupH) / 2
	if startX < 0 {
		startX = 0
	}
	if startY < 0 {
		startY = 0
	}

	// 1. Create the popup layer first
	popupLayer := lipgloss.NewLayer(popup).
		X(startX).
		Y(startY).
		Z(1)

	// 2. Create the background layer standalone (pass nil or leave children empty)
	bgLayer := lipgloss.NewLayer(bg).X(0).Y(0)

	// 3. Explicitly attach the child layer to avoid the variadic unpacking panic
	bgLayer.AddLayers(popupLayer)

	// 4. Render using the root background container
	comp := lipgloss.NewCompositor(bgLayer)

	return comp.Render()
}
