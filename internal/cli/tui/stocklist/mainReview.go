package stocklist

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/auth"
	"fmt"

)

// Model for the main review page
type MainReviewModel struct {
	Options		[]string
	Selected  	int
	BackPressed bool
	Confirmed 	bool
	User        *auth.User
	StockList	StockList

}

// returns initial main reivew page model
func NewMainReviewPage(stockList StockList, user *auth.User) *MainReviewModel {
	return &MainReviewModel{
		User: 		user,
		StockList:	stockList,
		Options: 	[]string{
				"Write/Edit Review",
				"View Reviews",
		},
		Selected: 	0,
	}
}
// returns initial command for the main review page to run (nothing)
func (m *MainReviewModel) Init() tea.Cmd {
	return nil
}
// Handles incoming events and updates the model accordingly
func (m *MainReviewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

func (m *MainReviewModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("⭐  Reviews") + "\n\n"

	// highlight the selected option with a →
	for i, option := range m.Options {
		if i == m.Selected {
			s += fmt.Sprintf("%s\n", styles.SelectedStyle.Render("→ "+option))
		} else {
			s += fmt.Sprintf("%s\n", styles.UnselectedStyle.Render("  "+option))
		}
	}
	s += styles.FooterStyle.Render("Enter to Select • ↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}

