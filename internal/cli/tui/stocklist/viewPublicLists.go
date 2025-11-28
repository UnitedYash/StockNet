package stocklist


import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/database"
	"StockNet/internal/auth"
)

// Model for viewing public stock lists
type ViewPublicListsModel struct {
    StockLists  []StockList
    Selected    int
    Loading     bool
    BackPressed bool
    Error       string
	User        *auth.User
	Confirmed	bool
}

type publicListsLoadedMsg struct { stockLists []StockList }
type publiclistSelectedMsg struct { StockList StockList }
type publicListsErrorMsg struct { err error }

// returns initial view pulbic stock lists model
func NewViewPublicLists(user *auth.User) *ViewPublicListsModel {
	return &ViewPublicListsModel{
		StockLists:    	[]StockList{},
		Selected:      	0,
		Loading:       	true,
		BackPressed:   	false,
		User: 			user,
		Confirmed:		false,
	}
}

// returns initial command for the view public stocklists page
func (m *ViewPublicListsModel) Init() tea.Cmd {
	return func() tea.Msg {
		// Query public stocklists from database for current user
		db := database.New().GetDB()

		query := `SELECT stocklist_id, user_id, name, visibility 
					FROM StockList 
					WHERE visibility = 'public';`

		rows, err := db.Query(query, m.User.UserID)
		if err != nil {
			return publicListsErrorMsg{err: err}
		}
		defer rows.Close()

		var stockLists []StockList
		for rows.Next() {
			var stockListID int
			var userID uint32
			var name string
			var visibility string
			if err := rows.Scan(&stockListID, &userID, &name, &visibility); err != nil {
				return publicListsErrorMsg{err: err}
			}
			stockLists = append(stockLists, StockList{
				StockListID:	stockListID,
				UserID:			userID
				Name:       	name,
				Visibility: 	visibility,
			})
		}

		return publicListsLoadedMsg{stockLists: stockLists}
	}
}

// Handles incoming events and updates the model
func (m *ViewPublicListsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case publicListsLoadedMsg:
		m.StockLists = msg.stockLists
		m.Loading = false
		m.Error = ""
	case publicListsErrorMsg:
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
				m.Confirmed = true
			}
		case "ctrl+b", "esc":
			m.BackPressed = true
		}
	}
	return m, nil
}

// renders the view stock list page
func (m *ViewPublicListsModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("🤼 Public Stock Lists") + "\n\n"

	if m.Loading {
		s += "Loading Public Stock lists...\n"
	} else if m.Error != "" {
		s += styles.ErrorStyle.Render(m.Error) + "\n"
	} else if len(m.StockLists) == 0 {
		s += "No public stock lists found. Tough Luck.\n"
	} else {
		s += styles.HeaderStyle.Render("ID    Name                          Owner ID") + "\n"
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
	s += styles.FooterStyle.Render("Press enter to select • ↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}

