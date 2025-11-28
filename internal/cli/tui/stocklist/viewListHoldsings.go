package stocklist

import (
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/auth"
	"fmt"
)

// Model for the delete stock from a stocklist
type ViewListHoldingsModel struct {
	BackPressed 	bool
	ViewStatsPressed bool
	User        	*auth.User
	StockList		StockList
	Error      		string
	StockHoldings	[]StockListHolding
	Selected		int
	Loading			bool
}

// returns initial edit stock list model
func NewViewListHoldingsPage(stockList StockList, user *auth.User) *ViewListHoldingsModel {
	return &ViewListHoldingsModel{
		User: user,
		Selected: 		0,
		StockList:		stockList,
		Error:			"",
		StockHoldings:	[]StockListHolding{},
		Loading:		true,
	}
}
// return the initial command to run, which is get all the holdings of the currently viewed stocklist
func (m *ViewListHoldingsModel) Init() tea.Cmd {
	return func() tea.Msg {
        holdings, err := getStockListHoldings(m.StockList.StockListID)
        if err != nil {
            return stockListHoldingsErrorMsg{err: err}
        }
        return stockListHoldingsLoadedMsg{stockHoldings: holdings}
    }
}

func (m *ViewListHoldingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		case "s":
			m.ViewStatsPressed = true
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
		}
	}
	return m, nil
}

func (m *ViewListHoldingsModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("👀  View holdings") + "\n\n"

	if m.Loading {
		s += "Loading Stock Holdings...\n"
	} else if m.Error != "" {
		s += styles.ErrorStyle.Render(m.Error) + "\n"
	} else if len(m.StockHoldings) == 0 {
		s += "No Holdings in this Stock List! Try adding some."
	} else {

		s += styles.HeaderStyle.Render("Symbol    Shares        Price         Net Worth") + "\n"
		s += "─────────────────────────────────────────────────────────────────────────\n"

		var totalNetWorth float64
		for i, holding := range m.StockHoldings {
			netWorth := holding.Shares * holding.CurrentPrice
			totalNetWorth += netWorth
			line := fmt.Sprintf("%-8s %-12.2f %-12.2f %-12.2f", holding.Symbol, holding.Shares, holding.CurrentPrice, netWorth)
			if i == m.Selected {
				s += fmt.Sprintf("%s\n", styles.SelectedStyle.Render("→ "+line))
			} else {
				s += fmt.Sprintf("%s\n", styles.UnselectedStyle.Render("  "+line))
			}
		}

		s += "─────────────────────────────────────────────────────────────────────────\n"
		s += fmt.Sprintf("%-8s %-12s %-12s %-12.2f\n", "", "", "Total:", totalNetWorth)
	}
	s += styles.FooterStyle.Render("↑/↓ or k/j to navigate • 's' for statistics • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}

