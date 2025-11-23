package stocklist

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/auth"

)

// Model for the create stock list page
type CreateStockListModel struct {
	User        *auth.User
	BackPressed bool

}

// returns initial create stock list model
func NewCreateStockListPage(user *auth.User) *CreateStockListModel {
	return &CreateStockListModel{
		User: user,
	}
}
// returns initial command for the create stock list page to run (nothing)
func (m *CreateStockListModel) Init() tea.Cmd {
	return nil
}
// Handles incoming events and updates the model accordingly
func (m *CreateStockListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+b", "esc":
			m.BackPressed = true
		}
	}
	return m, nil
}

func (m *CreateStockListModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("🗂️  Create a Stock List") + "\n\n"

	s += styles.FooterStyle.Render("'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}