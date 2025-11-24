package stock

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
)

// Model for the stock list page
type StockListModel struct {
	BackPressed bool
}

// returns initial stock list model
func NewStockListPage() *StockListModel {
	return &StockListModel{

	}
}
// returns initial command for the stock list page to run (nothing)
func (m *StockListModel) Init() tea.Cmd {
	return nil
}
// Handles incoming events and updates the model accordingly
func (m *StockListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+b", "esc":
			m.BackPressed = true
		}
	}
	return m, nil
}

func (m *StockListModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("🗂️  Stock List Dashboard") + "\n\n"
	s += styles.FooterStyle.Render("↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}
