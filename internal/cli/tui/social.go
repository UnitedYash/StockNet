package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Model for the social page
type SocialModel struct {
	backPressed bool
}

// returns initial social page model
func newSocialPage() *SocialModel {
	return &SocialModel{

	}
}
// returns initial command for the social page to run (nothing)
func (m *SocialModel) Init() tea.Cmd {
	return nil
}
// Handles incoming events and updates the model accordingly
func (m *SocialModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+b", "esc":
			m.backPressed = true
		}
	}
	return m, nil
}

func (m *SocialModel) View() string {
	s := "\n"
	s += TitleStyle.Render("💬 Friends & Social") + "\n\n"
	s += FooterStyle.Render("↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}








