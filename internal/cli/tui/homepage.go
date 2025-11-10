package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)
// Model for the home page
type HomePageModel struct {
	backPressed bool
}
// returns initial home page model
func NewHomePage() *HomePageModel {
	return &HomePageModel{}
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
	s := "\n"
	s += TitleStyle.Render("📊 Home") + "\n\n"
	s += InputStyle.Render("Welcome to StockNet!") + "\n\n"
	s += FooterStyle.Render("Press 'Ctrl + l' or 'Esc' to logout") + "\n\n"
	return s
}
