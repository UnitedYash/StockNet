package stocklist

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/auth"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/database"
	"fmt"
)

// Model for the share stock list page
type ShareStockListModel struct {
	StockList		StockList
	Error       	string
	BackPressed 	bool
	SuccessMessage	string
	Confirmed		bool
	User        	*auth.User
	UserInput		string

}


// returns initial share stock list model
func NewShareStockListPage(stockList StockList, user *auth.User) *ShareStockListModel {
	return &ShareStockListModel{
		User: 			user,
		StockList:		stockList,
		Error:			"",
		SuccessMessage:	"",
		UserInput:		"",
	}
}

// returns initial command for sharing a stock list page to run (nothing)
func (m *ShareStockListModel) Init() tea.Cmd {
	return nil
}

// Handles incoming events and updates the model accordingly
func (m *ShareStockListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+b", "esc":
			m.BackPressed = true
		case "enter":
			if m.UserInput == "" {
				break
			}
			if m.UserInput == m.User.Email {
				m.Error = "You cannot share a stocklist with yourself. Goofy aaah"
				m.SuccessMessage = ""
				break
			}
			err := ShareStockList(m.StockList.StockListID, m.UserInput)
			if err != nil {
				m.Error = err.Error()
				m.SuccessMessage = ""
			} else {
				m.SuccessMessage = "Stock list shared successfully!"
				m.Error = ""
			}

		case "backspace":
			if len(m.UserInput) > 0 {
				m.UserInput = m.UserInput[:len(m.UserInput)-1]
			}
		default:
			// any other key would indicate user typing in the text field
			// also block any non-printable letters
			if len(msg.String()) == 1 {
				r := rune(msg.String()[0])
				if r >= 32 && r != 127 {
					m.UserInput += msg.String()
				}
			}
		}
		
	}
	return m, nil

}

func (m *ShareStockListModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("📨 Share Stock List") + "\n\n"
	s += styles.InputStyle.Render("Enter email: "+ m.UserInput) + "\n"

	if m.Error != "" {
		s += styles.ErrorStyle.Render(m.Error) + "\n"
	} else if m.SuccessMessage != "" {
		s += styles.SuccessStyle.Render(m.SuccessMessage) + "\n"
	}
	s += styles.FooterStyle.Render("'Ctrl + b' or 'Esc' to go back") + "\n\n"


	return s
}
// share a stocklist given the stocklist id and the email
func ShareStockList(stocklistID int, email string) error {
	db := database.New().GetDB()

	var found bool

	// we need to check if that user input is a real email in DB
	err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM accounts WHERE email = $1)`, email).Scan(&found)
	if err != nil {
		return fmt.Errorf("failure in accounts: %v", err)
	}
	if !found {
		return fmt.Errorf("no account found with email: %s", email)
	}
	// also check if already shared 
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM sharedstocklists
			WHERE stocklist_id = $1 AND email = $2
		)
	`, stocklistID, email).Scan(&found)
	if err != nil {
		return fmt.Errorf("failed to check in sharedstocklists: %v", err)
	}
	if found {
		return fmt.Errorf("stocklist already shared with %s", email)
	}
	
	// now we can finally insert
	_, err = db.Exec(`
		INSERT INTO sharedstocklists (stocklist_id, email)
		VALUES ($1, $2)
	`, stocklistID, email)

	if err != nil {
		return fmt.Errorf("Failed to share stocklist: %v", err)
	}

	return nil



	

	if err != nil {
		return fmt.Errorf("Failure to share: %v", err)
	}
	return nil
}