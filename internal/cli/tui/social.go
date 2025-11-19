package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/auth"
	"fmt"
)

// Model for the social page
type SocialModel struct {
	options		[]string
	selected  	int
	backPressed bool
	confirmed 	bool
	user        *auth.User
}

// returns initial social page model
func newSocialPage(user *auth.User) *SocialModel {
	return &SocialModel{
		user: user,
		options: []string{
			"View Friends",
			"View Friends Requests",
			"Send Friend Request",
			"Remove Friend",
			"View Shared Stock Lists",
		},
		selected: 0,
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
		case "up", "k":
			if m.selected > 0 {
				// go up an option
				m.selected--
			} else {
				// at the top so wrap around to bottom
				m.selected = len(m.options) - 1
			}
		case "down", "j":
			if m.selected < len(m.options) - 1 {
				m.selected++
			} else {
				// at last option so wrap around to the top
				m.selected = 0
			}
		case "enter":
			m.confirmed = true
		}
	}
	return m, nil
}

func (m *SocialModel) View() string {

	s := "\n"
	s += TitleStyle.Render("💬 Friends & Social") + "\n\n"

	// highlight the selected option with a →
	for i, option := range m.options {
		if i == m.selected {
			s += fmt.Sprintf("%s\n", SelectedStyle.Render("→ "+option))
		} else {
			s += fmt.Sprintf("%s\n", UnselectedStyle.Render("  "+option))
		}
	}
	s += FooterStyle.Render("↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}








