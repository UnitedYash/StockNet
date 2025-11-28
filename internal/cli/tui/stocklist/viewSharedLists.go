package stocklist


import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/database"
	"StockNet/internal/auth"
	"fmt"
	"StockNet/internal/cli/tui/styles"
)

// Model for viewing shared stock lists
type ViewSharedListsModel struct {
    StockLists  []StockList
    Selected    int
    Loading     bool
    BackPressed bool
    Error       string
	User        *auth.User
	Confirmed	bool
}

type sharedListsLoadedMsg struct { stockLists []StockList }
type sharedlistSelectedMsg struct { StockList StockList }
type sharedListsErrorMsg struct { err error }

// returns initial view shared stock lists model
func NewViewSharedListsPage(user *auth.User) *ViewSharedListsModel {
	return &ViewSharedListsModel{
		StockLists:    	[]StockList{},
		Selected:      	0,
		Loading:       	true,
		BackPressed:   	false,
		User: 			user,
		Confirmed:		false,
	}
}

// returns initial command for the shared stocklists page
func (m *ViewSharedListsModel) Init() tea.Cmd {
	return func() tea.Msg {
		// query shared stocklists from database for current user
		db := database.New().GetDB()
		// get all stocklists and its information 
		rows, err := db.Query(`
			SELECT s.stocklist_id, s.user_id, s.name, s.visibility
			FROM stocklist s
			JOIN sharedstocklists ss
				ON s.stocklist_id = ss.stocklist_id
			WHERE ss.email = $1
		`, m.User.Email)
		if err != nil {
			return sharedListsErrorMsg{err: err}
		}
		defer rows.Close()

		var stockLists []StockList
		for rows.Next() {
			var stockListID int
			var userID uint32
			var name string
			var visibility string
			if err := rows.Scan(&stockListID, &userID, &name, &visibility); err != nil {
				return sharedListsErrorMsg{err: err}
			}
			stockLists = append(stockLists, StockList{
				StockListID:	stockListID,
				UserID:			userID,
				Name:       	name,
				Visibility: 	visibility,
			})
		}

		return sharedListsLoadedMsg{stockLists: stockLists}
	}
}

// Handles incoming events and updates the model
func (m *ViewSharedListsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case sharedListsLoadedMsg:
		m.StockLists = msg.stockLists
		m.Loading = false
		m.Error = ""
	case sharedListsErrorMsg:
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
func (m *ViewSharedListsModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("🥹 Shared Stock Lists") + "\n\n"

	if m.Loading {
		s += "Loading Shared Stock lists...\n"
	} else if m.Error != "" {
		s += styles.ErrorStyle.Render(m.Error) + "\n"
	} else if len(m.StockLists) == 0 {
		s += "No shared stock lists found. Tough Luck.\n"
	} else {
		s += styles.HeaderStyle.Render("ID    Name                          Owner ID") + "\n"
		s += "────────────────────────────────────────────────────────\n"

		for i, stock := range m.StockLists {
			line := fmt.Sprintf("%-5d %-30s %-10d", stock.StockListID, stock.Name, stock.UserID)
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

