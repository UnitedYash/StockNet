package stock

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
)

type SellStockSearchModel struct {
	Holdings       []Holding
	Selected       int
	Loading        bool
	BackPressed    bool
	PortfolioID    string
	CurrentUserID  int
}

type StockSelectedForSellMsg struct {
	Holding Holding
}

// NewSellStockSearchPageWithPortfolio creates a sell stock search page for a portfolio
func NewSellStockSearchPageWithPortfolio(userID int, portfolioID string) *SellStockSearchModel {
	return &SellStockSearchModel{
		Holdings:      []Holding{},
		Selected:      0,
		Loading:       true,
		BackPressed:   false,
		PortfolioID:   portfolioID,
		CurrentUserID: userID,
	}
}

// Init fetches the holdings from the database
func (m *SellStockSearchModel) Init() tea.Cmd {
	return func() tea.Msg {
		holdings, err := getStocksFromPortfolio(m.PortfolioID)
		if err != nil {
			return holdingsLoadError{err: err}
		}
		return holdingsLoadedMsg{holdings: holdings}
	}
}

// Update handles incoming events
func (m *SellStockSearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case holdingsLoadedMsg:
		m.Holdings = msg.holdings
		m.Loading = false
	case holdingsLoadError:
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
		case "enter":
			if m.Selected >= 0 && m.Selected < len(m.Holdings) {
				return m, func() tea.Msg {
					return StockSelectedForSellMsg{Holding: m.Holdings[m.Selected]}
				}
			}
		case "ctrl+b", "esc":
			m.BackPressed = true
		}
	}
	return m, nil
}

// View renders the sell stock search page
func (m *SellStockSearchModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("📊 Select Stock to Sell") + "\n\n"

	if m.Loading {
		s += "Loading your holdings...\n"
	} else if len(m.Holdings) == 0 {
		s += "You have no stocks to sell.\n"
	} else {
		s += styles.InputStyle.Render("Symbol | Shares | Price") + "\n"
		s += styles.InputStyle.Render("--------+--------+----------") + "\n"

		for i, holding := range m.Holdings {
			line := fmt.Sprintf("%-6s | %6d | $%.2f",
				holding.Symbol, holding.Shares, holding.Price)
			if i == m.Selected {
				s += fmt.Sprintf("%s\n", styles.SelectedStyle.Render("→ "+line))
			} else {
				s += fmt.Sprintf("%s\n", styles.UnselectedStyle.Render("  "+line))
			}
		}
	}

	s += styles.FooterStyle.Render("↑/↓ or k/j to navigate • Enter to select • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}
