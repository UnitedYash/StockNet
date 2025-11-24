package portfolio

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/database"
)

type ViewTransactionsModel struct {
	PortfolioID   string
	Transactions  []transactionRecord
	BackPressed   bool
	error         string
	ScrollOffset  int
	Selected      int
}

type transactionRecord struct {
	TransactionID   int
	PortfolioID     int
	Time            string
	Amount          float64
	StockSymbol     *string
	Buy_SellPrice   float64
	SharesMoved     int
	Type            string
}

type transactionsLoadedMsg struct {
	transactions []transactionRecord
}

type transactionsErrorMsg struct {
	err error
}

func NewViewTransactionsPage(portfolioID string) *ViewTransactionsModel {
	return &ViewTransactionsModel{
		PortfolioID:  portfolioID,
		Transactions: []transactionRecord{},
		BackPressed:  false,
	}
}

func (m *ViewTransactionsModel) Init() tea.Cmd {
	return m.RefreshTransactions()
}

// RefreshTransactions reloads the transactions data from the database
func (m *ViewTransactionsModel) RefreshTransactions() tea.Cmd {
	return func() tea.Msg {
		db := database.New().GetDB()

		query := `SELECT transaction_id, portfolio_id, time, amount, buy_sell_price, shares_moved, type, stock_symbol
		          FROM transaction
		          WHERE portfolio_id = $1::INTEGER
		          ORDER BY time DESC`

		rows, err := db.Query(query, m.PortfolioID)
		if err != nil {
			return transactionsErrorMsg{err: err}
		}
		defer rows.Close()

		var transactions []transactionRecord
		for rows.Next() {
			var t transactionRecord
			if err := rows.Scan(&t.TransactionID, &t.PortfolioID, &t.Time, &t.Amount, &t.Buy_SellPrice, &t.SharesMoved, &t.Type, &t.StockSymbol); err != nil {
				return transactionsErrorMsg{err: err}
			}
			transactions = append(transactions, t)
		}

		return transactionsLoadedMsg{transactions: transactions}
	}
}

// Update handles incoming events
func (m *ViewTransactionsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case transactionsLoadedMsg:
		m.Transactions = msg.transactions
	case transactionsErrorMsg:
		m.error = msg.err.Error()
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Selected > 0 {
				m.Selected--
			} else {
				m.Selected = len(m.Transactions) - 1
			}
		case "down", "j":
			if m.Selected < len(m.Transactions)-1 {
				m.Selected++
			} else {
				m.Selected = 0
			}
		case "ctrl+b", "esc":
			m.BackPressed = true
		}
	}
	return m, nil
}

// View renders the transactions page
func (m *ViewTransactionsModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("📋 Transaction History") + "\n\n"

	if m.error != "" {
		s += styles.ErrorStyle.Render(m.error) + "\n\n"
	} else if len(m.Transactions) == 0 {
		s += "No transactions in this portfolio.\n"
	} else {
		s += styles.InputStyle.Render("Date/Time          | Type | Symbol | Shares | Price   | Total") + "\n"
		s += styles.InputStyle.Render("-------------------+------+--------+--------+---------+--------") + "\n"

		const viewportHeight = 10

		// Adjust scrollOffset to keep selected item in view
		if m.Selected < m.ScrollOffset {
			m.ScrollOffset = m.Selected
		}
		if m.Selected >= m.ScrollOffset+viewportHeight {
			m.ScrollOffset = m.Selected - viewportHeight + 1
		}

		// Ensure scrollOffset doesn't go past the end
		maxScrollOffset := len(m.Transactions) - viewportHeight
		if maxScrollOffset < 0 {
			maxScrollOffset = 0
		}
		if m.ScrollOffset > maxScrollOffset {
			m.ScrollOffset = maxScrollOffset
		}

		// Display visible transactions
		endIdx := m.ScrollOffset + viewportHeight
		if endIdx > len(m.Transactions) {
			endIdx = len(m.Transactions)
		}

		for i := m.ScrollOffset; i < endIdx; i++ {
			tx := m.Transactions[i]
			symbol := ""
			if tx.StockSymbol != nil {
				symbol = *tx.StockSymbol
			}
			line := fmt.Sprintf("%-19s | %-4s | %-6s | %6d | $%.2f | $%.2f",
				tx.Time[:19], tx.Type, symbol, tx.SharesMoved, tx.Buy_SellPrice, tx.Amount)
			if i == m.Selected {
				s += fmt.Sprintf("%s\n", styles.SelectedStyle.Render("→ "+line))
			} else {
				s += fmt.Sprintf("%s\n", styles.UnselectedStyle.Render("  "+line))
			}
		}

		// Show pagination info
		s += fmt.Sprintf("\n(%d of %d transactions)\n", m.Selected+1, len(m.Transactions))
	}

	s += styles.FooterStyle.Render("↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}
