package config

import (
	"strings"

	"charm.land/lipgloss/v2"
)

func DividerBorder() string {
	divider := ""
	for i := 1; i < WindowWidth; i++ {
		divider = divider + "-"
	}

	return lipgloss.NewStyle().Foreground(lipgloss.Color(ColorBorder)).Render(divider)
}

func LogBlock(content string) string {
	truncated := Truncate(content, WindowWidth-6)

	return Styles.LogBox.Render(truncated)
}

// PlaceOverlay renders `popup` centered on top of `bg`.
// bg should be the full rendered background string.
func PlaceOverlay(bg, popup string) string {
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
}
