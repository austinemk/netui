package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type InputPopupModel struct {
	TextInput  textinput.Model
	Visible    bool
	TargetSSID string
	IsMasked   bool // State tracking value visibility status
}

// NewInputPopup sets initial visibility mask via parameters
func NewInputPopup(hideValuesByDefault bool) InputPopupModel {
	ti := textinput.New()
	ti.Placeholder = "Enter Network Password..."

	if hideValuesByDefault {
		ti.EchoMode = textinput.EchoPassword
		ti.EchoCharacter = '•'
	} else {
		ti.EchoMode = textinput.EchoNormal
	}
	ti.Focus()

	return InputPopupModel{
		TextInput: ti,
		Visible:   false,
		IsMasked:  hideValuesByDefault,
	}
}

func (p InputPopupModel) Update(msg tea.Msg) (InputPopupModel, tea.Cmd) {
	if !p.Visible {
		return p, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		// Keybind shortcut combo to toggle password input visibility state on-the-fly
		case "ctrl+v":
			p.IsMasked = !p.IsMasked
			if p.IsMasked {
				p.TextInput.EchoMode = textinput.EchoPassword
				p.TextInput.EchoCharacter = '•'
			} else {
				p.TextInput.EchoMode = textinput.EchoNormal
			}
			return p, nil
		case "esc":
			p.Visible = false
			p.TextInput.Blur()
			return p, nil
		}
	}

	var cmd tea.Cmd
	p.TextInput, cmd = p.TextInput.Update(msg)
	return p, cmd
}

func (p InputPopupModel) View() string {
	if !p.Visible {
		return ""
	}

	popupStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#EF4444")).
		Padding(1, 2).
		Width(52)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F59E0B"))
	hintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280")).Italic(true)
	iconStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).MarginLeft(1)

	// Determine matching status eye indicator icon
	var eyeIcon string
	if p.IsMasked {
		eyeIcon = iconStyle.Render("🙈 Hidden")
	} else {
		eyeIcon = iconStyle.Render("👁️  Visible")
	}

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("🔒 Authentication Required"),
		"Connecting to: "+p.TargetSSID+"\n",
		lipgloss.JoinHorizontal(lipgloss.Center, p.TextInput.View(), eyeIcon),
		"\n"+hintStyle.Render("ctrl+v: toggle visibility • enter: submit • esc: close"),
	)

	return popupStyle.Render(body)
}

// RenderOverlay physically places the popup text directly on top of the main UI frame layout
func RenderOverlay(mainView string, popupView string) string {
	return lipgloss.Place(
		74, 15,
		lipgloss.Center, lipgloss.Center,
		popupView,
		lipgloss.WithWhitespaceChars(mainView),
		lipgloss.WithWhitespaceBackground(lipgloss.NoColor{}),
	)
}
