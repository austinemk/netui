package components

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type OptionSelectedMsg struct {
	Option string
}

type OptionsPopupModel struct {
	Title   string
	Options []string
	Cursor  int
	Active  bool
	Width   int
}

func NewOptionsPopup(title string, options []string) OptionsPopupModel {
	return OptionsPopupModel{
		Title:   title,
		Options: options,
		Active:  false,
		Width:   40,
	}
}

func (p OptionsPopupModel) Update(msg tea.Msg) (OptionsPopupModel, tea.Cmd) {
	if !p.Active {
		return p, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if p.Cursor > 0 {
				p.Cursor--
			}
		case "down", "j":
			if p.Cursor < len(p.Options)-1 {
				p.Cursor++
			}
		case "enter":
			return p, func() tea.Msg {
				return OptionSelectedMsg{Option: p.Options[p.Cursor]}
			}
		case "esc":
			p.Active = false
		}
	}
	return p, nil
}

func (p OptionsPopupModel) View() string {
	if !p.Active {
		return ""
	}

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3B82F6")).
		Padding(1, 2).
		Width(p.Width)

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#F3F4F6")).
		MarginBottom(1)

	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FFFFFF")).
		Background(lipgloss.Color("#3B82F6")).
		Bold(true)

	unselectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9CA3AF"))

	hintStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#6B7280")).
		Italic(true)

	var content strings.Builder
	content.WriteString(titleStyle.Render(p.Title) + "\n")

	for i, opt := range p.Options {
		if p.Cursor == i {
			content.WriteString(selectedStyle.Render("> "+opt) + "\n")
		} else {
			content.WriteString(unselectedStyle.Render("  "+opt) + "\n")
		}
	}

	// Bottom explicit hints block
	content.WriteString("\n" + hintStyle.Render("↑/↓: navigate • enter: select • esc: cancel"))

	return borderStyle.Render(content.String())
}
