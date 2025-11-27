package stock

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/database"
)

// SearchStockForDataModel - Search and select stock to add data for
type SearchStockForDataModel struct {
	Stocks       []Stock
	Selected     int
	Loading      bool
	BackPressed  bool
	error        string
	ScrollOffset int
}

type searchStockForDataLoadedMsg struct {
	stocks []Stock
}

type searchStockForDataErrorMsg struct {
	err error
}

type StockSelectedForDataMsg struct {
	Stock Stock
}

// NewSearchStockForDataPage creates a new search stock for data page
func NewSearchStockForDataPage() *SearchStockForDataModel {
	return &SearchStockForDataModel{
		Stocks:       []Stock{},
		Selected:     0,
		Loading:      true,
		BackPressed:  false,
		ScrollOffset: 0,
	}
}

// Init returns initial command for the search stock for data page
func (m *SearchStockForDataModel) Init() tea.Cmd {
	return func() tea.Msg {
		// Query all distinct stock symbols from stocks table
		db := database.New().GetDB()

		rows, err := db.Query("SELECT DISTINCT symbol FROM stocks ORDER BY symbol")
		if err != nil {
			return searchStockForDataErrorMsg{err: err}
		}
		defer rows.Close()

		var stocks []Stock
		for rows.Next() {
			var symbol string
			if err := rows.Scan(&symbol); err != nil {
				return searchStockForDataErrorMsg{err: err}
			}
			stocks = append(stocks, Stock{
				Symbol:    symbol,
				Price:     0,
				Timestamp: "",
			})
		}

		return searchStockForDataLoadedMsg{stocks: stocks}
	}
}

// Update handles incoming events and updates the model
func (m *SearchStockForDataModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case searchStockForDataLoadedMsg:
		m.Stocks = msg.stocks
		m.Loading = false
		m.error = ""
	case searchStockForDataErrorMsg:
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
					return StockSelectedForDataMsg{Stock: selectedStock}
				}
			}
		case "ctrl+b", "esc":
			m.BackPressed = true
		}
	}
	return m, nil
}

// View renders the search stock for data page
func (m *SearchStockForDataModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("📊 Select Stock to Add Data") + "\n\n"

	if m.Loading {
		s += "Loading available stocks...\n"
	} else if m.error != "" {
		s += styles.ErrorStyle.Render(m.error) + "\n"
	} else if len(m.Stocks) == 0 {
		s += "No stocks available in the database.\n"
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
		for i := m.ScrollOffset; i < m.ScrollOffset+viewportHeight && i < len(m.Stocks); i++ {
			stock := m.Stocks[i]
			if i == m.Selected {
				s += styles.SelectedStyle.Render("→ " + stock.Symbol) + "\n"
			} else {
				s += styles.UnselectedStyle.Render("  " + stock.Symbol) + "\n"
			}
		}

		// Show pagination info
		s += fmt.Sprintf("\n(%d of %d)\n", m.Selected+1, len(m.Stocks))
	}

	s += "\n"
	s += styles.FooterStyle.Render("↑/↓ or k/j to navigate • Press Enter to select stock • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}
