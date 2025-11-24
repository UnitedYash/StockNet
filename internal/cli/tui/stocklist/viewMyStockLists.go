package stocklist

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/database"
	"StockNet/internal/auth"
)

// StockList represents a user's stockList
type StockList struct {
    StockListID int
    Name        string
    Visibility  string
}

// Model for viewing user's stocklists
type ViewStockListsModel struct {
    StockLists  []StockList
    Selected    int
    Loading     bool
    BackPressed bool
    Error       string
	User        *auth.User
}

type stockListsLoadedMsg struct { stockLists []StockList }
type stockListSelectedMsg struct { StockList StockList }
type stockListsErrorMsg struct { err error }

// returns initial view stock list model
func NewViewStockLists(user *auth.User) *ViewStockListsModel {
	return &ViewStockListsModel{
		StockLists:    	[]StockList{},
		Selected:      	0,
		Loading:       	true,
		BackPressed:   	false,
		User: 			user,
	}
}

// returns initial command for the view stocklists page
func (m *ViewStockListsModel) Init() tea.Cmd {
	return func() tea.Msg {
		// Query stocklists from database for current user
		db := database.New().GetDB()

		query := `SELECT stocklist_id, name, visibility 
					FROM StockList 
					WHERE user_id = $1;`

		rows, err := db.Query(query, m.User.UserID)
		if err != nil {
			return stockListsErrorMsg{err: err}
		}
		defer rows.Close()

		var stockLists []StockList
		for rows.Next() {
			var stockListID int
			var name string
			var visibility string
			if err := rows.Scan(&stockListID, &name, &visibility); err != nil {
				return stockListsErrorMsg{err: err}
			}
			stockLists = append(stockLists, StockList{
				StockListID:	stockListID,
				Name:       	name,
				Visibility: 	visibility,
			})
		}

		return stockListsLoadedMsg{stockLists: stockLists}
	}
}

// Handles incoming events and updates the model
func (m *ViewStockListsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case stockListsLoadedMsg:
		m.StockLists = msg.stockLists
		m.Loading = false
		m.Error = ""
	case stockListsErrorMsg:
		m.Loading = false
		m.Error = fmt.Sprintf("Error loading stock lists: %v", msg.err)
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Selected > 0 {
				// go up an option
				m.Selected--
			} else {
				// at the top so wrap around to bottom
				m.Selected = len(m.StockLists) - 1
			}
		case "down", "j":
			if m.Selected < len(m.StockLists) - 1 {
				m.Selected++
			} else {
				// at last option so wrap around to the top
				m.Selected = 0
			}
		case "enter":
			if (len(m.StockLists) > 0) {
				selectedStockList := m.StockLists[m.Selected]
				return m, func() tea.Msg {
					return stockListSelectedMsg{StockList: selectedStockList}
				}
			}
		case "ctrl+b", "esc":
			m.BackPressed = true
		}
	}
	return m, nil
}

// renders the view stock list page
func (m *ViewStockListsModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("📃 Your Stock Lists") + "\n\n"

	if m.Loading {
		s += "Loading Stock lists...\n"
	} else if m.Error != "" {
		s += styles.ErrorStyle.Render(m.Error) + "\n"
	} else if len(m.StockLists) == 0 {
		s += "No stock lists found. Create one to get started!\n"
	} else {
		s += styles.HeaderStyle.Render("ID    Name                          Visibility") + "\n"
		s += "────────────────────────────────────────────────────────\n"

		for i, stock := range m.StockLists {
			line := fmt.Sprintf("%-5d %-30s %-10s", stock.StockListID, stock.Name, stock.Visibility)
			if i == m.Selected {
				s += styles.SelectedStyle.Render(line) + "\n"
			} else {
				s += line + "\n"
			}
		}
	}

	s += "\n"
	s += styles.FooterStyle.Render("↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}
