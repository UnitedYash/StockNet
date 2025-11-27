package stocklist

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/auth"
	"fmt"
	"StockNet/internal/database"
)

type StockListHolding struct {
    Symbol	string
    Shares  float64
}

// Model for the delete stock from a stocklist
type DeleteStockFromList struct {
	BackPressed 	bool
	User        	*auth.User
	StockList		StockList
	Error      		string
	SuccessMessage  string
	StockHoldings	[]StockListHolding
	Selected		int
	Loading			bool
}

// holds stocklists and error message for update function 
type stockListHoldingsLoadedMsg struct { stockHoldings []StockListHolding }
type stockListHoldingsErrorMsg struct { err error }


// returns initial edit stock list model
func NewDeleteStockFromListPage(stockList StockList, user *auth.User) *DeleteStockFromList {
	return &DeleteStockFromList{
		User: user,
		Selected: 		0,
		StockList:		stockList,
		Error:			"",
		SuccessMessage:	"",
		StockHoldings:	[]StockListHolding{},
		Loading:		true,
	}
}
// return the initial command to run, which is get all the holdings of the currently viewed stocklist
func (m *DeleteStockFromList) Init() tea.Cmd {
	return func() tea.Msg {
        holdings, err := getStockListHoldings(m.StockList.StockListID)
        if err != nil {
            return stockListHoldingsErrorMsg{err: err}
        }
        return stockListHoldingsLoadedMsg{stockHoldings: holdings}
    }

}

func (m *DeleteStockFromList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case stockListHoldingsLoadedMsg:
		m.StockHoldings = msg.stockHoldings
		m.Loading = false
		m.Error = ""
	
	case stockListHoldingsErrorMsg:
		m.Loading = false
		m.Error = fmt.Sprintf("Error loading holdings: %v", msg.err)
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
				m.Selected = len(m.StockHoldings) - 1
			}
		case "down", "j":
			if m.Selected < len(m.StockHoldings) - 1 {
				m.Selected++
			} else {
				// at last option so wrap around to the top
				m.Selected = 0
			}
		case "r":
			// remove the holding for the stocklist

			holding := m.StockHoldings[m.Selected]

			db := database.New().GetDB()

			_, err := db.Exec(`
				DELETE FROM hasstockfromstocklist
				WHERE stocklist_id=$1 AND symbol=$2
			`, m.StockList.StockListID, holding.Symbol)

			if err != nil {
				m.Error = fmt.Sprintf("Failed to remove the stock %s: %v", holding.Symbol, err)
				return m, nil
			}
			// remove holding from UI holdings list
			m.StockHoldings = append(m.StockHoldings[:m.Selected], m.StockHoldings[m.Selected+1:]...)

			// fix the indexing
			if m.Selected >= len(m.StockHoldings) && len(m.StockHoldings) > 0 {
				m.Selected = len(m.StockHoldings) - 1
			}
		}
	}
	return m, nil
}

func (m *DeleteStockFromList) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("💔  Delete a Stock") + "\n\n"

	if m.Loading {
		s += "Loading Stock Holdings...\n"
	} else if m.Error != "" {
		s += styles.ErrorStyle.Render(m.Error) + "\n"
	} else if len(m.StockHoldings) == 0 {
		s += "No Holdings in this Stock List! Try adding some."
	} else {

		s += styles.HeaderStyle.Render("Symbol    Share") + "\n"
		s += "────────────────────────────────────────────────────────\n"
		for i, holding := range m.StockHoldings {
			line := fmt.Sprintf("%-5s %-30f", holding.Symbol, holding.Shares)
			if i == m.Selected {
			s += fmt.Sprintf("%s\n", styles.SelectedStyle.Render("→ "+ line))
			} else {
				s += fmt.Sprintf("%s\n", styles.UnselectedStyle.Render("  "+ line))
			}
		}
	}
	s += styles.FooterStyle.Render("'r' to remove holding • ↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}


// get all the holdings of a stocklist, (fundamental)
func getStockListHoldings(stockListID int) ([]StockListHolding, error){
	db := database.New().GetDB()

	rows, err := db.Query(`
		SELECT symbol, shares
		FROM hasstockfromstocklist
		WHERE stocklist_id = $1
	
	`, stockListID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var holdings []StockListHolding
	for rows.Next() {
		var holding StockListHolding
		err := rows.Scan(&holding.Symbol, &holding.Shares)
		if err != nil {
			return nil, err
		}
		holdings = append(holdings, holding)
	}
	return holdings, rows.Err()

}

