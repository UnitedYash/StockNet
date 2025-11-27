package stocklist

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/auth"
	"fmt"
)

// Model for the stock list page
type EditStockListModel struct {
	Options		[]string
	Selected  	int
	BackPressed bool
	Confirmed 	bool
	User        *auth.User
	StockList	StockList
}

// returns initial edit stock list model
func NewEditStockListPage(stockList StockList, user *auth.User) *EditStockListModel {
	return &EditStockListModel{
		User: user,
		Options: []string{
			"Add Stock",
			"Update Shares",
		},
		Selected: 	0,
		StockList:	stockList,
	}
}
// returns initial command for the edit stock list page to run (nothing)
func (m *EditStockListModel) Init() tea.Cmd {
	return nil
}
// Handles incoming events and updates the model accordingly
func (m *EditStockListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+b", "esc":
			m.BackPressed = true
		case "up", "k":
			if m.Selected > 0 {
				// go up an option
				m.Selected--
			} else {
				// at the top so wrap around to bottom
				m.Selected = len(m.Options) - 1
			}
		case "down", "j":
			if m.Selected < len(m.Options) - 1 {
				m.Selected++
			} else {
				// at last option so wrap around to the top
				m.Selected = 0
			}
		case "enter":
			m.Confirmed = true
		}
	}
	return m, nil
}

func (m *EditStockListModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("📝  Edit Stock List") + "\n\n"

	// highlight the selected option with a →
	for i, option := range m.Options {
		if i == m.Selected {
			s += fmt.Sprintf("%s\n", styles.SelectedStyle.Render("→ "+option))
		} else {
			s += fmt.Sprintf("%s\n", styles.UnselectedStyle.Render("  "+option))
		}
	}
	s += styles.FooterStyle.Render("Enter to select • ↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}

// returns (logged in) user
func (m *EditStockListModel) GetUser() *auth.User {
	return m.User
}