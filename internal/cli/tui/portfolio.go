package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Model for the portfolio page
type PortfolioModel struct {
	backPressed bool
}

// returns initial portfolio page model
func newPortfolioPage() *PortfolioModel {
	return &PortfolioModel{

	}
}
// returns initial command for the portfolio page to run (nothing)
func (m *PortfolioModel) Init() tea.Cmd {
	return nil
}
// Handles incoming events and updates the model accordingly
func (m *PortfolioModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+b", "esc":
			m.backPressed = true
		}
	}
	return m, nil
}

func (m *PortfolioModel) View() string {
	s := "\n"
	s += TitleStyle.Render("💼 Portfolio Dashboard") + "\n\n"
	s += FooterStyle.Render("↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}








