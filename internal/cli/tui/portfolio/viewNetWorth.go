package portfolio

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/database"
)

type PortfolioNetWorth struct {
	PortfolioID         string
	CashAccount         float64
	TotalStockValue     float64
	TotalPortfolioValue float64
}

type ViewNetWorthModel struct {
	PortfolioID   string
	NetWorth      PortfolioNetWorth
	BackPressed   bool
	Loading       bool
	Error         string
	CurrentUserID int
}

type netWorthLoadedMsg struct {
	netWorth PortfolioNetWorth
}

type netWorthErrorMsg struct {
	err error
}

// NewViewNetWorthPageWithPortfolioID creates a new view net worth page
func NewViewNetWorthPageWithPortfolioID(userID int, portfolioID string) *ViewNetWorthModel {
	return &ViewNetWorthModel{
		PortfolioID:   portfolioID,
		NetWorth:      PortfolioNetWorth{},
		BackPressed:   false,
		Loading:       true,
		Error:         "",
		CurrentUserID: userID,
	}
}

// Init fetches the net worth data from the database
func (m *ViewNetWorthModel) Init() tea.Cmd {
	return func() tea.Msg {
		db := database.New().GetDB()

		// Get portfolio cash account
		query := `SELECT portfolio_id, cash_account FROM Portfolio WHERE portfolio_id = $1`

		var portfolioID int
		var cashAccount float64

		err := db.QueryRow(query, m.PortfolioID).Scan(&portfolioID, &cashAccount)
		if err != nil {
			return netWorthErrorMsg{err: err}
		}

		// Calculate total stock value from holdings
		holdingsQuery := `SELECT COALESCE(SUM(hsp.shares * cp.price), 0) as total_stock_value
		                  FROM hasStockFromPortfolio hsp
		                  JOIN CurrentPrices cp ON hsp.symbol = cp.symbol
		                  WHERE hsp.portfolio_id = $1`

		var totalStockValue float64
		err = db.QueryRow(holdingsQuery, m.PortfolioID).Scan(&totalStockValue)
		if err != nil {
			return netWorthErrorMsg{err: err}
		}

		totalPortfolioValue := cashAccount + totalStockValue

		return netWorthLoadedMsg{
			netWorth: PortfolioNetWorth{
				PortfolioID:         fmt.Sprintf("%d", portfolioID),
				CashAccount:         cashAccount,
				TotalStockValue:     totalStockValue,
				TotalPortfolioValue: totalPortfolioValue,
			},
		}
	}
}

// Update handles incoming events
func (m *ViewNetWorthModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case netWorthLoadedMsg:
		m.NetWorth = msg.netWorth
		m.Loading = false
	case netWorthErrorMsg:
		m.Error = msg.err.Error()
		m.Loading = false
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+b", "esc":
			m.BackPressed = true
		}
	}
	return m, nil
}

// View renders the net worth page
func (m *ViewNetWorthModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("💰 Portfolio Net Worth") + "\n\n"

	if m.Loading {
		s += "Loading net worth data...\n"
	} else if m.Error != "" {
		s += styles.ErrorStyle.Render(m.Error) + "\n\n"
	} else {
		// Display portfolio breakdown
		s += styles.InputStyle.Render("═══════════════════════════════════") + "\n"
		s += styles.InputStyle.Render(fmt.Sprintf("Cash Available:     $%10.2f", m.NetWorth.CashAccount)) + "\n"
		s += styles.InputStyle.Render(fmt.Sprintf("Stock Holdings:     $%10.2f", m.NetWorth.TotalStockValue)) + "\n"
		s += styles.InputStyle.Render("───────────────────────────────────") + "\n"
		s += styles.SuccessStyle.Render(fmt.Sprintf("Total Net Worth:    $%10.2f", m.NetWorth.TotalPortfolioValue)) + "\n"
		s += styles.InputStyle.Render("═══════════════════════════════════") + "\n\n"
	}

	s += styles.FooterStyle.Render("'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}
