package tui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
)

// Model for the stock analysis page
type StockAnalysisModel struct {
	backPressed bool
	options     []string  // menu options
	selected    int
	confirmed   bool
	stocks      []Stock   // loaded stocks
	loading     bool      // loading state
}

// returns initial stock analysis page model
func newStockAnalysisPage() *StockAnalysisModel {
	return &StockAnalysisModel{
		options:   []string{"View Current Stocks", "Search Stock"},
		selected:  0,
		confirmed: false,
		stocks:    []Stock{},
		loading:   false,
	}
}
// returns initial command for the stock analysis page to run (nothing)
func (m *StockAnalysisModel) Init() tea.Cmd {
	return nil
}
// Handles incoming events and updates the model accordingly
func (m *StockAnalysisModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
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
		case "ctrl+b", "esc":
			m.backPressed = true
		}
	}
	return m, nil
}

func (m *StockAnalysisModel) View() string {
	s := "\n"
	s += TitleStyle.Render("📊 Stock Data & Analysis") + "\n\n"

	// Display menu options
	for i, option := range m.options {
		if i == m.selected {
			s += fmt.Sprintf("%s\n", SelectedStyle.Render("→ "+option))
		} else {
			s += fmt.Sprintf("%s\n", UnselectedStyle.Render("  "+option))
		}
	}

	s += FooterStyle.Render("↑/↓ or k/j to navigate • Enter to select • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}








