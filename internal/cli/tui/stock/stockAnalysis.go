package stock

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
)

// Model for the stock analysis page
type StockAnalysisModel struct {
	BackPressed bool
	Options     []string  // menu options
	Selected    int
	Confirmed   bool
	stocks      []Stock   // loaded stocks
	loading     bool      // loading state
}

// returns initial stock analysis page model
func NewStockAnalysisPage() *StockAnalysisModel {
	return &StockAnalysisModel{
		Options:   []string{"View Current Stocks", "Search Stock"},
		Selected:  0,
		Confirmed: false,
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
		case "ctrl+b", "esc":
			m.BackPressed = true
		}
	}
	return m, nil
}

func (m *StockAnalysisModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("📊 Stock Data & Analysis") + "\n\n"

	// Display menu options
	for i, option := range m.Options {
		if i == m.Selected {
			s += fmt.Sprintf("%s\n", styles.SelectedStyle.Render("→ "+option))
		} else {
			s += fmt.Sprintf("%s\n", styles.UnselectedStyle.Render("  "+option))
		}
	}

	s += styles.FooterStyle.Render("↑/↓ or k/j to navigate • Enter to select • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}








