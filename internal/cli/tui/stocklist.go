package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Model for the stock list page
type StockListModel struct {
	backPressed bool
}

// returns initial stock list model
func newStockListPage() *StockListModel {
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
			m.backPressed = true
		}
	}
	return m, nil
}

func (m *StockListModel) View() string {
	s := "\n"
	s += TitleStyle.Render("🗂️  Stock List Dashboard") + "\n\n"
	s += FooterStyle.Render("↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}
