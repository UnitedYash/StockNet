package shared

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/auth"
	"StockNet/internal/cli/tui/styles"
	"fmt"

)
// Model for the home page
type HomePageModel struct {
	Options 	[]string
	Selected	int
	Confirmed 	bool
	BackPressed bool
	User        *auth.User

}
// returns initial home page model
func NewHomePage(user *auth.User) *HomePageModel {
	return &HomePageModel{
		User: user,
		Options: []string{
			"My Portfolios",
			"My Stock Lists",
			"Stock Data & Analysis",
			"Friends & Social",
		},
		Selected: 0,
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
		case "ctrl+l", "esc":
			m.BackPressed = true
		}
	}
	return m, nil
}
// renders UI based on current version of homepage view model
func (m *HomePageModel) View() string {
	name := "Guest"
    if m.User != nil {
        name = m.User.Name
    }

	s := "\n"
	s += styles.TitleStyle.Render("🏠 Home") + "\n\n"
	s += styles.InputStyle.Render(fmt.Sprintf("Welcome, %s!", name)) + "\n\n"
	// highlight the selected option with a →
	for i, option := range m.Options {
		if i == m.Selected {
			s += fmt.Sprintf("%s\n", styles.SelectedStyle.Render("→ "+option))
		} else {
			s += fmt.Sprintf("%s\n", styles.UnselectedStyle.Render("  "+option))
		}
	}
	// footer
	s += styles.FooterStyle.Render("↑/↓ or k/j to navigate • 'Ctrl + l' or 'Esc' to logout") + "\n\n"
	return s
}

// returns (logged in) user
func (m *HomePageModel) GetUser() *auth.User {
	return m.User
}
