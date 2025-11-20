package tui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/database"
)

// BuyStockSearchModel - Search and select stock to buy
type BuyStockSearchModel struct {
	stocks       []Stock
	selected     int
	loading      bool
	backPressed  bool
	error        string
	currentUserID int
	portfolioID  string
	scrollOffset int
	cashAccount  float64
}

type buyStockSearchLoadedMsg struct {
	stocks []Stock
}

type buyStockSearchErrorMsg struct {
	err error
}

type stockSelectedForBuyMsg struct {
	Stock Stock
}

// returns initial buy stock search page model
func newBuyStockSearchPageWithPortfolio(userID int, portfolioID string, cashAccount float64) *BuyStockSearchModel {
	return &BuyStockSearchModel{
		stocks:       []Stock{},
		selected:     0,
		loading:      true,
		backPressed:  false,
		currentUserID: userID,
		portfolioID:  portfolioID,
		scrollOffset: 0,
		cashAccount:  cashAccount,
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
		m.stocks = msg.stocks
		m.loading = false
		m.error = ""
	case buyStockSearchErrorMsg:
		m.loading = false
		m.error = fmt.Sprintf("Error loading stocks: %v", msg.err)
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			} else {
				m.selected = len(m.stocks) - 1
			}
		case "down", "j":
			if m.selected < len(m.stocks)-1 {
				m.selected++
			} else {
				m.selected = 0
			}
		case "enter":
			if len(m.stocks) > 0 {
				selectedStock := m.stocks[m.selected]
				return m, func() tea.Msg {
					return stockSelectedForBuyMsg{Stock: selectedStock}
				}
			}
		case "ctrl+b", "esc":
			m.backPressed = true
		}
	}
	return m, nil
}

// renders the buy stock search page
func (m *BuyStockSearchModel) View() string {
	s := "\n"
	s += TitleStyle.Render("📈 Select Stock to Buy") + "\n"
	s += InputStyle.Render(fmt.Sprintf("Cash Available: $%.2f", m.cashAccount)) + "\n\n"

	if m.loading {
		s += "Loading available stocks...\n"
	} else if m.error != "" {
		s += ErrorStyle.Render(m.error) + "\n"
	} else if len(m.stocks) == 0 {
		s += "No stocks available for purchase.\n"
	} else {
		// Display at most 10 stocks at a time, centered on selected
		const viewportHeight = 10

		// Adjust scrollOffset to keep selected item in view
		if m.selected < m.scrollOffset {
			m.scrollOffset = m.selected
		}
		if m.selected >= m.scrollOffset+viewportHeight {
			m.scrollOffset = m.selected - viewportHeight + 1
		}

		// Ensure scrollOffset doesn't go past the end
		maxScrollOffset := len(m.stocks) - viewportHeight
		if maxScrollOffset < 0 {
			maxScrollOffset = 0
		}
		if m.scrollOffset > maxScrollOffset {
			m.scrollOffset = maxScrollOffset
		}

		// Display header
		s += HeaderStyle.Render("Symbol          Price        Updated") + "\n"
		s += "────────────────────────────────────────────────────────\n"

		// Display visible stocks
		endIdx := m.scrollOffset + viewportHeight
		if endIdx > len(m.stocks) {
			endIdx = len(m.stocks)
		}

		for i := m.scrollOffset; i < endIdx; i++ {
			stock := m.stocks[i]
			line := fmt.Sprintf("%-15s $%-10.2f %s", stock.Symbol, stock.Price, stock.Timestamp)
			if i == m.selected {
				s += SelectedStyle.Render("→ " + line) + "\n"
			} else {
				s += UnselectedStyle.Render("  " + line) + "\n"
			}
		}

		// Show pagination info
		s += fmt.Sprintf("\n(%d of %d)\n", m.selected+1, len(m.stocks))
	}

	s += "\n"
	s += FooterStyle.Render("↑/↓ or k/j to navigate • Press Enter to select stock • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}
