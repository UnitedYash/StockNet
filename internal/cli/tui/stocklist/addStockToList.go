package stocklist

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/auth"
	"StockNet/internal/database"
	"fmt"
	"strconv"

)

// Model for the add stock to stocklist page
type AddStockToListModel struct {
	BackPressed 	bool
	User        	*auth.User
	StockList		StockList
	Error      		string
	SuccessMessage  string
	SymbolName 		string		// user inputs a symbol
	Shares			string		// user inputs number of shares (converted to float after)
	Step 			int			// 0 = enter symbol, 1 = enter amount of shares

}

type stockAddedCreatedMsg struct {
	stockListID int
	symbol		string
}

type stockAddedErrorMsg struct {
	err error
}

// returns initial edit stock list model
func NewAddStockToListPage(stockList StockList, user *auth.User) *AddStockToListModel {
	return &AddStockToListModel{
		User: 			user,
		StockList:		stockList,
		Error:			"",
		SuccessMessage:	"",
		SymbolName:		"",
		Shares:			"",
		Step:			0,
	}
}
// returns initial command for the edit stock list page to run (nothing)
func (m *AddStockToListModel) Init() tea.Cmd {
	return nil
}
// Handles incoming events and updates the model accordingly
func (m *AddStockToListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case stockAddedCreatedMsg:
		m.SuccessMessage = fmt.Sprintf("Stock '%s' added successfully to %s (%d)", msg.symbol, m.StockList.Name, m.StockList.StockListID)
	case stockAddedErrorMsg:
		m.Error = fmt.Sprintf("Error adding Stock %s: %v. Try Again", m.SymbolName, msg.err)
		m.Step = 0
		m.SymbolName = ""
		m.Shares = ""
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+b", "esc":
			m.BackPressed = true
		case "enter":

			if m.Step == 0 && len(m.SymbolName) > 0 {
				m.Step = 1
			} else if m.Step == 1 && len(m.Shares) > 0 {
				return m, m.AddStockToStockListInDB()
			}
		default:
			// any other key would indicate user typing in the text field
			// also block any non-printable letters
			if len(msg.String()) == 1 {
				r := rune(msg.String()[0])
				if r >= 32 && r != 127 {
					if m.Step == 0 {
						m.SymbolName += msg.String()
					} else if m.Step == 1 {
						m.Shares += msg.String()
					}
				}
			}
		case "backspace":
			if m.Step == 0 && len(m.SymbolName) > 0 {
				m.SymbolName = m.SymbolName[:len(m.SymbolName)-1]
			} else if m.Step == 1 && len(m.Shares) > 0 {
				m.Shares = m.Shares[:len(m.Shares) - 1]
			}
			
		}
		
	}
	return m, nil
}

func (m *AddStockToListModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("📝  Add Stock to a Stock List") + "\n\n"

    s += "Enter Stock Symbol (up to 10 characters): \n"
    s += styles.InputStyle.Render(m.SymbolName) + "\n\n"
	if m.Step == 1 {
		s += "Enter number of shares: \n"
		s += styles.InputStyle.Render(m.Shares) + "\n\n"
	}

	if m.Error != "" {
    	s += styles.ErrorStyle.Render(m.Error) + "\n"
	}	
	if m.SuccessMessage != "" {
		s += styles.SuccessStyle.Render(m.SuccessMessage) + "\n"
	}

	s += styles.FooterStyle.Render("'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}

// AddStockToStockListInDB creates a new stocklist in the database
func (m *AddStockToListModel) AddStockToStockListInDB() tea.Cmd {
	return func() tea.Msg {

		sharesFloat, err := strconv.ParseFloat(m.Shares, 64)
		if err != nil || sharesFloat < 0 {
			return stockAddedErrorMsg{err: fmt.Errorf("invalid number of shares")}
		}
		stockListID := m.StockList.StockListID
		symbol := m.SymbolName

		db := database.New().GetDB()

		// first check that if the symbol exist in currentprices relation

		var exists bool
		err = db.QueryRow("SELECT EXISTS(SELECT 1 FROM currentprices WHERE symbol = $1)", symbol).Scan(&exists)
		if err != nil {
			return stockAddedErrorMsg{err: err}
		}
		if !exists {
			return stockAddedErrorMsg{err: fmt.Errorf("symbol '%s' does not exist", symbol)}
		}

		// the symbol exist so now insert stockListID, symbol and sharesFloat into hasstockfromstocklist
		// note we should only insert if this is a new symbol

        err = db.QueryRow(`
			SELECT EXISTS (
				SELECT 1 FROM hasstockfromstocklist
				WHERE stocklist_id=$1 AND symbol=$2
			)
		`, stockListID, symbol).Scan(&exists)

		if err != nil {
			return stockAddedErrorMsg{err: err}
		}
		if exists {
			return stockAddedErrorMsg{err: fmt.Errorf("stock already exists in this list")}
		}
		// greenflag to insert
		_, err = db.Exec(`
			INSERT INTO hasStockFromStockList (stocklist_id, symbol, shares)
			VALUES ($1, $2, $3)
		`, stockListID, symbol, sharesFloat)
		if err != nil {
			return stockAddedErrorMsg{err: err}
		}


		return stockAddedCreatedMsg{stockListID: stockListID, symbol: symbol}
	}
}