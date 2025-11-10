package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/auth"
	"fmt"

)
// Model for the home page
type HomePageModel struct {
	backPressed bool
	user        *auth.User	
}
// returns initial home page model
func NewHomePage(user *auth.User) *HomePageModel {
	return &HomePageModel{
		user: user,
	}
}
// returns initial command 
func (m *HomePageModel) Init() tea.Cmd {
	return nil
}
// Handles incoming events and updates the model accordingly
func (m *HomePageModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
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
	s += FooterStyle.Render("Press 'Ctrl + l' or 'Esc' to logout") + "\n\n"
	return s
}
