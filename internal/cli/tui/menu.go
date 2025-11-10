package tui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)
// Model for application main page
type MainMenuModel struct {
	options   []string
	selected  int
	confirmed bool
}
// returns application initial model
func NewMainMenu() *MainMenuModel {
	return &MainMenuModel{
		options: []string{
			"Login",
			"Register",
			"Configure",
			"Quit",
		},
		selected: 0,
	}
}
// returns intial command for the application to run
func (m *MainMenuModel) Init() tea.Cmd {
	// Note: For now, we don't have any initial I/O commands to do. nil = "no command"
	return nil
}

// Handles incoming events and updates the model accordingly
func (m *MainMenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	// Check if a key pressed, if so do the corresponding update to model
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			} else {
				m.selected = len(m.options) - 1
			}
		case "down", "j":
			if m.selected < len(m.options)-1 {
				m.selected++
			} else {
				m.selected = 0
			}
		case "enter":
			m.confirmed = true
		}
	}
	return m, nil
}
// renders UI, looks at model at current state, return string s which is the UI 
func (m *MainMenuModel) View() string {
	s := "\n"
	s += TitleStyle.Render("🚀 StockNet") + "\n\n"
	// highlight the selected option with a →
	for i, option := range m.options {
		if i == m.selected {
			s += fmt.Sprintf("%s\n", SelectedStyle.Render("→ "+option))
		} else {
			s += fmt.Sprintf("%s\n", UnselectedStyle.Render("  "+option))
		}
	}
	// The footer
	s += FooterStyle.Render("↑/↓ or k/j to navigate • Enter to select • Ctrl + c to quit") + "\n\n"
	
	return s
}
