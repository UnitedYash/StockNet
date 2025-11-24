package stocklist

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/auth"
	"fmt"
	"StockNet/internal/database"
)

// Model for the create stock list page
type CreateStockListModel struct {
	User        	*auth.User
	BackPressed 	bool
	error           string
	successMessage  string
	stockListID     int
	stockListName 	string
}

type stockListCreatedMsg struct {
	stockListID int
}

type stockListCreationErrorMsg struct {
	err error
}

// returns initial create stock list model
func NewCreateStockListPage(user *auth.User) *CreateStockListModel {
	return &CreateStockListModel{
		User: 			user,
		stockListName:  "",
	}
}
// returns initial command for the create stock list page to run (nothing)
func (m *CreateStockListModel) Init() tea.Cmd {
	return nil
}
// Handles incoming events and updates the model accordingly
func (m *CreateStockListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case stockListCreatedMsg:
		m.successMessage = fmt.Sprintf("Stock List '%s' created successfully! ID: %d", m.stockListName, msg.stockListID)
		m.stockListID = msg.stockListID
	case stockListCreationErrorMsg:
		m.error = fmt.Sprintf("Error creating Stock List: %v", msg.err)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+b", "esc":
			m.BackPressed = true
		case "enter":
			if len(m.stockListName) == 0 {
				break
			}
			return m, m.createStockListInDB()
		default:
			// any other key would indicate user typing in the text field
			// also block any non-printable letters
			if len(msg.String()) == 1 {
				r := rune(msg.String()[0])
				if r >= 32 && r != 127 {
					m.stockListName += msg.String()
				}
			}
		case "backspace":
			if len(m.stockListName) > 0 {
				m.stockListName = m.stockListName[:len(m.stockListName)-1]
			}
		}
		
	}
	return m, nil
}

func (m *CreateStockListModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("🗂️  Create a Stock List") + "\n\n"

	s += "Enter Stock list name (up to 40 characters): \n\n"

	s += styles.InputStyle.Render(m.stockListName) + "\n\n"

	if m.successMessage != "" {
		s += styles.SuccessStyle.Render(m.successMessage) + "\n"
	}


	s += styles.FooterStyle.Render("'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}

// createStockListInDB creates a new stocklist in the database
func (m *CreateStockListModel) createStockListInDB() tea.Cmd {
	return func() tea.Msg {
		db := database.New().GetDB()

		// Insert into stocklist table (stocklist_id is auto-generated SERIAL)
		var stockListID int

		err := db.QueryRow(
			`INSERT INTO StockList (user_id, name, visibility) VALUES ($1, $2, $3) RETURNING stocklist_id`,
			m.User.UserID, m.stockListName, "private",
		).Scan(&stockListID)
		

		if err != nil {
			return stockListCreationErrorMsg{err: err}
		}

		return stockListCreatedMsg{stockListID: stockListID}
	}
}