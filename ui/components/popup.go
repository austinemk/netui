package components

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type PopupModel struct {
	TextInput  textinput.Model
	Visible    bool
	TargetSSID string
}

func NewPopup() PopupModel {
	ti := textinput.New()
	ti.Placeholder = "Enter Network Password..."
	ti.EchoMode = textinput.EchoPassword
	ti.EchoCharacter = '•'
	ti.Focus()

	return PopupModel{
		TextInput: ti,
		Visible:   false,
	}
}

func (p PopupModel) Update(msg tea.Msg) (PopupModel, tea.Cmd) {
	var cmd tea.Cmd
	p.TextInput, cmd = p.TextInput.Update(msg)
	return p, cmd
}

func (p PopupModel) View() string {
	popupStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#EF4444")).
		Padding(1, 2).
		Width(50)

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F59E0B"))

	body := lipgloss.JoinVertical(
		lipgloss.Left,
		titleStyle.Render("🔒 Authentication Required"),
		"Connecting to: "+p.TargetSSID+"\n",
		p.TextInput.View(),
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
