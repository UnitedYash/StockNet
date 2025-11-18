package tui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/database"
)

// Struct for stock
type Stock struct {
	Symbol    string
	Price     float64
	Timestamp string
}

type CurrentStocksModel struct {
    backPressed bool
    stocks      []Stock
    selected    int
    loading     bool
    scrollOffset int  // Track which stock to start displaying from
}

type stocksLoadedMsg struct {
	stocks []Stock
}

type stocksLoadError struct {
	err error
}

func newCurrentStocksPage() *CurrentStocksModel {
	return &CurrentStocksModel{
		stocks:   []Stock{},
		selected: 0,
		loading:  true,
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
		m.stocks = msg.stocks
		m.loading = false
	case stocksLoadError:
		m.loading = false
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
		case "ctrl+b", "esc":
			m.backPressed = true
		}
	}
	return m, nil
}

func (m *CurrentStocksModel) View() string {
	s := "\n"
	s += TitleStyle.Render("📈 Current Stock Prices") + "\n\n"

	if m.loading {
		s += "Loading current stock prices...\n"
	} else {
		if len(m.stocks) == 0 {
			s += "No current stock prices available.\n"
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

			// Display visible stocks
			endIdx := m.scrollOffset + viewportHeight
			if endIdx > len(m.stocks) {
				endIdx = len(m.stocks)
			}

			for i := m.scrollOffset; i < endIdx; i++ {
				stock := m.stocks[i]
				line := fmt.Sprintf("%s: $%.2f (as of %s)", stock.Symbol, stock.Price, stock.Timestamp)
				if i == m.selected {
					s += fmt.Sprintf("%s\n", SelectedStyle.Render("→ "+line))
				} else {
					s += fmt.Sprintf("%s\n", UnselectedStyle.Render("  "+line))
				}
			}

			// Show pagination info
			s += fmt.Sprintf("\n(%d of %d)\n", m.selected+1, len(m.stocks))
		}
	}

	s += FooterStyle.Render("↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}