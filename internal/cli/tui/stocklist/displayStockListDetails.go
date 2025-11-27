package stocklist

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/auth"

)

// Model for a displaying stocklist
type DisplayStockListModel struct {
	StockList		StockList
    Selected    	int
    BackPressed 	bool
    Error       	string
	User        	*auth.User	
	OwnerUserID 	int
	Options			[]string
}

// returns initial displaying  list model
func NewDisplayStockListPage(stockList StockList, user *auth.User, OwnerID) *DisplayStockListModel {
	model := &DisplayStockListModel{
		StockList:		stockList,
		Selected:		0,
		BackPressed:	false,
		Error:			"",
		User:			user,
		OwnerUserID:	OwnerID,
	}
	// if the current user is the owner of the stocklist, give edit and delete options
	if user != nil && user.UserID == OwnerUserID {
		model.Options = []string{"View Stocks", "Edit List", "View Statistics", "Reviews", "Share", "Delete List"}
	} else {
		model.Options = []string{"View Stocks", "View Statistics", "Reviews", "Share"}
	}

	return model
}

// returns initial command for displaying stock list page to run (nothing)
func (m *DisplayStockListModel) Init() tea.Cmd {
	return nil
}

// Handles incoming events and updates the model accordingly
func (m *DisplayStockListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		}
	}
	return m, nil
}

func (m *SocialModel) View() string {

	s := "\n"
	s += styles.TitleStyle.Render("📃 " + m.StockList.Name + " Stock List") + "\n\n"

	// Stock list details
	s += fmt.Sprintf("Visibility: %s\n", m.StockList.Visibility)
	if m.User != nil && m.User.UserID == m.OwnerUserID {
		s += "Owner: You\n"
	} else {
		s += fmt.Sprintf("Owner UserID: %d\n", m.OwnerUserID)
	}
	s += "\n"

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