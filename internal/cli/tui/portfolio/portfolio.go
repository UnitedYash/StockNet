package portfolio

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
)

// Model for the portfolio page
type PortfolioModel struct {
	selectedOption int // 0 = View Portfolios, 1 = Create Portfolio
	BackPressed    bool
}

// NewPortfolioPage returns initial portfolio page model
func NewPortfolioPage() *PortfolioModel {
	return &PortfolioModel{
		selectedOption: 0,
		BackPressed:    false,
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
		case "up", "k":
			if m.selectedOption > 0 {
				m.selectedOption--
			}
		case "down", "j":
			if m.selectedOption < 1 {
				m.selectedOption++
			}
		case "enter":
			// Handle selection - will be processed by app.go
			return m, nil
		case "ctrl+b", "esc":
			m.BackPressed = true
		}
	}
	return m, nil
}

func (m *PortfolioModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("💼 Portfolio Dashboard") + "\n\n"

	options := []string{
		"View Your Portfolios",
		"Create New Portfolio",
	}

	for i, option := range options {
		if i == m.selectedOption {
			s += styles.SelectedStyle.Render("  → " + option) + "\n"
		} else {
			s += styles.UnselectedStyle.Render("    " + option) + "\n"
		}
	}

	s += "\n"
	s += styles.FooterStyle.Render("↑/↓ or k/j to navigate • Enter to select • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}

// GetSelectedOption returns the currently selected option (0 = View, 1 = Create)
func (m *PortfolioModel) GetSelectedOption() int {
	return m.selectedOption
}




