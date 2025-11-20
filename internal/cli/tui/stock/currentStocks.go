package stock

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/database"
)

// Struct for stock
type Stock struct {
	Symbol    string
	Price     float64
	Timestamp string
}

type CurrentStocksModel struct {
    BackPressed bool
    Stocks      []Stock
    Selected    int
    Loading     bool
    ScrollOffset int  // Track which stock to start displaying from
}

type stocksLoadedMsg struct {
	stocks []Stock
}

type stocksLoadError struct {
	err error
}

func NewCurrentStocksPage() *CurrentStocksModel {
	return &CurrentStocksModel{
		Stocks:   []Stock{},
		Selected: 0,
		Loading:  true,
	}
}

func (m *CurrentStocksModel) Init() tea.Cmd {
	return func() tea.Msg {
		// get database instance
		db := database.New().GetDB()

		//Query current prices
		rows, err := db.Query("SELECT symbol, price, timestamp FROM CurrentPrices ORDER BY symbol")
		if err != nil {
			fmt.Printf("Database error: %v\n", err)
			return stocksLoadError{err: err}
		}
		defer rows.Close()
		var stocks []Stock
		for rows.Next() {
			var s Stock
			if err := rows.Scan(&s.Symbol, &s.Price, &s.Timestamp); err != nil {
				fmt.Printf("Scan error: %v\n", err)
				return stocksLoadError{err: err}
			}
			stocks = append(stocks, s)
		}
		fmt.Printf("Loaded %d stocks\n", len(stocks))
		return stocksLoadedMsg{stocks: stocks}
	}
}

func (m *CurrentStocksModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case stocksLoadedMsg:
		m.Stocks = msg.stocks
		m.Loading = false
	case stocksLoadError:
		m.Loading = false
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
		case "ctrl+b", "esc":
			m.BackPressed = true
		}
	}
	return m, nil
}

func (m *CurrentStocksModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("📈 Current Stock Prices") + "\n\n"

	if m.Loading {
		s += "Loading current stock prices...\n"
	} else {
		if len(m.Stocks) == 0 {
			s += "No current stock prices available.\n"
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

			// Display visible stocks
			endIdx := m.ScrollOffset + viewportHeight
			if endIdx > len(m.Stocks) {
				endIdx = len(m.Stocks)
			}

			for i := m.ScrollOffset; i < endIdx; i++ {
				stock := m.Stocks[i]
				line := fmt.Sprintf("%s: $%.2f (as of %s)", stock.Symbol, stock.Price, stock.Timestamp)
				if i == m.Selected {
					s += fmt.Sprintf("%s\n", styles.SelectedStyle.Render("→ "+line))
				} else {
					s += fmt.Sprintf("%s\n", styles.UnselectedStyle.Render("  "+line))
				}
			}

			// Show pagination info
			s += fmt.Sprintf("\n(%d of %d)\n", m.Selected+1, len(m.Stocks))
		}
	}

	s += styles.FooterStyle.Render("↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}