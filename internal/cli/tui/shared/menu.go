package shared

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
)
// Model for main page view
type MainMenuModel struct {
	Options   []string
	Selected  int
	Confirmed bool
}
// returns main page initial model
func NewMainMenu() *MainMenuModel {
	return &MainMenuModel{
		Options: []string{
			"Login",
			"Register",
			"Quit",
		},
		Selected: 0,
	}
}
// returns intial command for the menu page to run
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
			if m.Selected > 0 {
				m.Selected--
			} else {
				m.Selected = len(m.Options) - 1
			}
		case "down", "j":
			if m.Selected < len(m.Options)-1 {
				m.Selected++
			} else {
				m.Selected = 0
			}
		case "enter":
			m.Confirmed = true
		}
	}
	return m, nil
}
// renders UI, looks at model at current state, return string s which is the UI
func (m *MainMenuModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("🚀 StockNet") + "\n\n"
	// highlight the selected option with a →
	for i, option := range m.Options {
		if i == m.Selected {
			s += fmt.Sprintf("%s\n", styles.SelectedStyle.Render("→ "+option))
		} else {
			s += fmt.Sprintf("%s\n", styles.UnselectedStyle.Render("  "+option))
		}
	}
	// The footer
	s += styles.FooterStyle.Render("↑/↓ or k/j to navigate • Enter to select • Ctrl + c to quit") + "\n\n"

	return s
}
