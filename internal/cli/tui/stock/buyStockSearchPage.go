package stock

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/database"
)

// BuyStockSearchModel - Search and select stock to buy
type BuyStockSearchModel struct {
	Stocks       []Stock
	Selected     int
	Loading      bool
	BackPressed  bool
	error        string
	currentUserID int
	PortfolioID  string
	ScrollOffset int
	CashAccount  float64
}

type buyStockSearchLoadedMsg struct {
	stocks []Stock
}

type buyStockSearchErrorMsg struct {
	err error
}

type StockSelectedForBuyMsg struct {
	Stock Stock
}

// returns initial buy stock search page model
func NewBuyStockSearchPageWithPortfolio(userID int, portfolioID string, cashAccount float64) *BuyStockSearchModel {
	return &BuyStockSearchModel{
		Stocks:       []Stock{},
		Selected:     0,
		Loading:      true,
		BackPressed:  false,
		currentUserID: userID,
		PortfolioID:  portfolioID,
		ScrollOffset: 0,
		CashAccount:  cashAccount,
	}
}

// returns initial command for the buy stock search page
func (m *BuyStockSearchModel) Init() tea.Cmd {
	return func() tea.Msg {
		// Query current stock prices from database
		db := database.New().GetDB()

		rows, err := db.Query("SELECT symbol, price, timestamp FROM CurrentPrices ORDER BY symbol")
		if err != nil {
			return buyStockSearchErrorMsg{err: err}
		}
		defer rows.Close()

		var stocks []Stock
		for rows.Next() {
			var s Stock
			if err := rows.Scan(&s.Symbol, &s.Price, &s.Timestamp); err != nil {
				return buyStockSearchErrorMsg{err: err}
			}
			stocks = append(stocks, s)
		}

		return buyStockSearchLoadedMsg{stocks: stocks}
	}
}

// Handles incoming events and updates the model
func (m *BuyStockSearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case buyStockSearchLoadedMsg:
		m.Stocks = msg.stocks
		m.Loading = false
		m.error = ""
	case buyStockSearchErrorMsg:
		m.Loading = false
		m.error = fmt.Sprintf("Error loading stocks: %v", msg.err)
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Selected > 0 {
				m.Selected--
			} else {
				m.Selected = len(m.Stocks) - 1
			}
		case "down", "j":
			if m.Selected < len(m.Stocks)-1 {
				m.Selected++
			} else {
				m.Selected = 0
			}
		case "enter":
			if len(m.Stocks) > 0 {
				selectedStock := m.Stocks[m.Selected]
				return m, func() tea.Msg {
					return StockSelectedForBuyMsg{Stock: selectedStock}
				}
			}
		case "ctrl+b", "esc":
			m.BackPressed = true
		}
	}
	return m, nil
}

// renders the buy stock search page
func (m *BuyStockSearchModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("📈 Select Stock to Buy") + "\n"
	s += styles.InputStyle.Render(fmt.Sprintf("Cash Available: $%.2f", m.CashAccount)) + "\n\n"

	if m.Loading {
		s += "Loading available stocks...\n"
	} else if m.error != "" {
		s += styles.ErrorStyle.Render(m.error) + "\n"
	} else if len(m.Stocks) == 0 {
		s += "No stocks available for purchase.\n"
	} else {
		// Display at most 10 stocks at a time, centered on selected
		const viewportHeight = 10

		// Adjust scrollOffset to keep selected item in view
		if m.Selected < m.ScrollOffset {
			m.ScrollOffset = m.Selected
		}
		if m.Selected >= m.ScrollOffset+viewportHeight {
			m.ScrollOffset = m.Selected - viewportHeight + 1
		}

		// Ensure scrollOffset doesn't go past the end
		maxScrollOffset := len(m.Stocks) - viewportHeight
		if maxScrollOffset < 0 {
			maxScrollOffset = 0
		}
		if m.ScrollOffset > maxScrollOffset {
			m.ScrollOffset = maxScrollOffset
		}

		// Display header
		s += styles.HeaderStyle.Render("Symbol          Price        Updated") + "\n"
		s += "────────────────────────────────────────────────────────\n"

		// Display visible stocks
		endIdx := m.ScrollOffset + viewportHeight
		if endIdx > len(m.Stocks) {
			endIdx = len(m.Stocks)
		}

		for i := m.ScrollOffset; i < endIdx; i++ {
			stock := m.Stocks[i]
			line := fmt.Sprintf("%-15s $%-10.2f %s", stock.Symbol, stock.Price, stock.Timestamp)
			if i == m.Selected {
				s += styles.SelectedStyle.Render("→ " + line) + "\n"
			} else {
				s += styles.UnselectedStyle.Render("  " + line) + "\n"
			}
		}

		// Show pagination info
		s += fmt.Sprintf("\n(%d of %d)\n", m.Selected+1, len(m.Stocks))
	}

	s += "\n"
	s += styles.FooterStyle.Render("↑/↓ or k/j to navigate • Press Enter to select stock • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}
