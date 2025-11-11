package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/auth"
	"fmt"

)
// Model for the home page
type HomePageModel struct {
	options 	[]string
	selected	int
	confirmed 	bool
	backPressed bool
	user        *auth.User
	
}
// returns initial home page model
func NewHomePage(user *auth.User) *HomePageModel {
	return &HomePageModel{
		user: user,
		options: []string{
			"My Portfolios",
			"My Stock Lists",
			"Stock Data & Analysis",
			"Friends & Social",
		},
		selected: 0,
	}
}
// returns initial command for the home page to run (nothing)
func (m *HomePageModel) Init() tea.Cmd {
	return nil
}
// Handles incoming events and updates the model accordingly
func (m *HomePageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
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
		case "ctrl+l", "esc":
			m.backPressed = true
		}
	}
	return m, nil
}
// renders UI based on current version of homepage view model
func (m *HomePageModel) View() string {
	name := "Guest"
    if m.user != nil {
        name = m.user.Name
    }

	s := "\n"
	s += TitleStyle.Render("🏠 Home") + "\n\n"
	s += InputStyle.Render(fmt.Sprintf("Welcome, %s!", name)) + "\n\n"
	// highlight the selected option with a →
	for i, option := range m.options {
		if i == m.selected {
			s += fmt.Sprintf("%s\n", SelectedStyle.Render("→ "+option))
		} else {
			s += fmt.Sprintf("%s\n", UnselectedStyle.Render("  "+option))
		}
	}
	// footer
	s += FooterStyle.Render("↑/↓ or k/j to navigate • 'Ctrl + l' or 'Esc' to logout") + "\n\n"
	return s
}
