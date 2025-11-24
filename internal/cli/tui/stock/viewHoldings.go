package stock

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/database"
)

// Holding represents a stock holding in a portfolio
type Holding struct {
	Symbol    string
	Shares    int
	Price     float64
	Timestamp string
}

type ViewHoldingsModel struct {
	PortfolioID  string
	Holdings     []Holding
	Selected     int
	Loading      bool
	BackPressed  bool
	ScrollOffset int
	Error        string
}

type holdingsLoadedMsg struct {
	holdings []Holding
}

type holdingsLoadError struct {
	err error
}

// NewViewHoldingsPageWithPortfolioID creates a new view holdings page for a portfolio
func NewViewHoldingsPageWithPortfolioID(portfolioID string) *ViewHoldingsModel {
	return &ViewHoldingsModel{
		PortfolioID: portfolioID,
		Holdings:    []Holding{},
		Selected:    0,
		Loading:     true,
		Error:       "",
	}
}

func getStocksFromPortfolio(portfolioID string) ([]Holding, error) {
	db := database.New().GetDB()
	rows, err := db.Query(
		"SELECT hsp.symbol, hsp.shares, cp.price, cp.timestamp FROM hasStockFromPortfolio hsp " +
			"JOIN CurrentPrices cp ON hsp.symbol = cp.symbol " +
			"WHERE hsp.portfolio_id = $1::INTEGER ORDER BY hsp.symbol",
		portfolioID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	
	var holdings []Holding
	for rows.Next() {
		var h Holding
		if err := rows.Scan(&h.Symbol, &h.Shares, &h.Price, &h.Timestamp); err != nil {
			return nil, err
		}
		holdings = append(holdings, h)
	}
	return holdings, nil
}
// Init fetches the holdings from the database
func (m *ViewHoldingsModel) Init() tea.Cmd {
		return func() tea.Msg {
		holdings, err := getStocksFromPortfolio(m.PortfolioID)
		if err != nil {
			return holdingsLoadError{err: err}
		}
		return holdingsLoadedMsg{holdings: holdings}
	}
}

// Update handles incoming events
func (m *ViewHoldingsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case holdingsLoadedMsg:
		m.Holdings = msg.holdings
		m.Loading = false
	case holdingsLoadError:
		m.Error = msg.err.Error()
		m.Loading = false
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Selected > 0 {
				m.Selected--
			} else {
				m.Selected = len(m.Holdings) - 1
			}
		case "down", "j":
			if m.Selected < len(m.Holdings)-1 {
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

// View renders the holdings page
func (m *ViewHoldingsModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("📊 Portfolio Holdings") + "\n\n"

	if m.Loading {
		s += "Loading holdings...\n"
	} else if m.Error != "" {
		s += styles.ErrorStyle.Render(m.Error) + "\n\n"
	} else {
		if len(m.Holdings) == 0 {
			s += "No holdings in this portfolio.\n"
		} else {
			// Display at most 10 holdings at a time
			const viewportHeight = 10

			// Adjust scrollOffset to keep selected item in view
			if m.Selected < m.ScrollOffset {
				m.ScrollOffset = m.Selected
			}
			if m.Selected >= m.ScrollOffset+viewportHeight {
				m.ScrollOffset = m.Selected - viewportHeight + 1
			}

			// Ensure scrollOffset doesn't go past the end
			maxScrollOffset := len(m.Holdings) - viewportHeight
			if maxScrollOffset < 0 {
				maxScrollOffset = 0
			}
			if m.ScrollOffset > maxScrollOffset {
				m.ScrollOffset = maxScrollOffset
			}

			// Display visible holdings
			endIdx := m.ScrollOffset + viewportHeight
			if endIdx > len(m.Holdings) {
				endIdx = len(m.Holdings)
			}

			s += styles.InputStyle.Render("Symbol    | Shares | Price  | Last Updated") + "\n"
			s += styles.InputStyle.Render("----------+--------+--------+------------------") + "\n"

			for i := m.ScrollOffset; i < endIdx; i++ {
				holding := m.Holdings[i]
				totalValue := float64(holding.Shares) * holding.Price
				line := fmt.Sprintf("%-9s | %6d | $%.2f | %s", holding.Symbol, holding.Shares, holding.Price, holding.Timestamp)
				if i == m.Selected {
					s += fmt.Sprintf("%s\n", styles.SelectedStyle.Render("→ "+line))
					s += fmt.Sprintf("  Total Value: $%.2f\n", totalValue)
				} else {
					s += fmt.Sprintf("%s\n", styles.UnselectedStyle.Render("  "+line))
				}
			}

			// Show pagination info
			s += fmt.Sprintf("\n(%d of %d holdings)\n", m.Selected+1, len(m.Holdings))
		}
	}

	s += styles.FooterStyle.Render("↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}

// GetSelectedHolding returns the currently selected holding
func (m *ViewHoldingsModel) GetSelectedHolding() Holding {
	if m.Selected >= 0 && m.Selected < len(m.Holdings) {
		return m.Holdings[m.Selected]
	}
	return Holding{}
}
