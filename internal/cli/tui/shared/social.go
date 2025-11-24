package shared

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/auth"
	"StockNet/internal/cli/tui/styles"
	"fmt"
)

// Model for the social page
type SocialModel struct {
	Options		[]string
	Selected  	int
	BackPressed bool
	Confirmed 	bool
	User        *auth.User
}

// returns initial social page model
func NewSocialPage(user *auth.User) *SocialModel {
	return &SocialModel{
		User: user,
		Options: []string{
			"Manage Friends",
			"View Friends Requests",
			"Send Friend Request",
			"View Shared Stock Lists",
		},
		Selected: 0,
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
			m.BackPressed = true
		case "up", "k":
			if m.Selected > 0 {
				// go up an option
				m.Selected--
			} else {
				// at the top so wrap around to bottom
				m.Selected = len(m.Options) - 1
			}
		case "down", "j":
			if m.Selected < len(m.Options) - 1 {
				m.Selected++
			} else {
				// at last option so wrap around to the top
				m.Selected = 0
			}
		case "enter":
			m.Confirmed = true
		}
	}
	return m, nil
}

func (m *SocialModel) View() string {

	s := "\n"
	s += styles.TitleStyle.Render("💬 Friends & Social") + "\n\n"

	// highlight the selected option with a →
	for i, option := range m.Options {
		if i == m.Selected {
			s += fmt.Sprintf("%s\n", styles.SelectedStyle.Render("→ "+option))
		} else {
			s += fmt.Sprintf("%s\n", styles.UnselectedStyle.Render("  "+option))
		}
	}
	s += styles.FooterStyle.Render("↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}

// returns (logged in) user
func (m *SocialModel) GetUser() *auth.User {
	return m.User
}







