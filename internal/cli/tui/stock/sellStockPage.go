package stock

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
)

// SellStockModel - Enter quantity and confirm sale
type SellStockModel struct {
	Holding       Holding
	Quantity      string // Input as string for editing
	Cursor        int
	Loading       bool
	BackPressed   bool
	Confirmed     bool
	Error         string
	CurrentUserID int
	PortfolioID   string
}

type SellStockCompletedMsg struct {
	Success bool
	Message string
}

// NewSellStockPageWithHolding creates a sell stock page with a holding
func NewSellStockPageWithHolding(userID int, portfolioID string, holding Holding) *SellStockModel {
	return &SellStockModel{
		Holding:       holding,
		Quantity:      "",
		Cursor:        0,
		Loading:       false,
		BackPressed:   false,
		Confirmed:     false,
		CurrentUserID: userID,
		PortfolioID:   portfolioID,
	}
}

// Init returns initial command for the sell stock page
func (m *SellStockModel) Init() tea.Cmd {
	return nil
}

// Update handles incoming events and updates the model
func (m *SellStockModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "backspace":
			if m.Cursor > 0 {
				m.Quantity = m.Quantity[:m.Cursor-1] + m.Quantity[m.Cursor:]
				m.Cursor--
			}
		case "left":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "right":
			if m.Cursor < len(m.Quantity) {
				m.Cursor++
			}
		case "home":
			m.Cursor = 0
		case "end":
			m.Cursor = len(m.Quantity)
		case "enter":
			if len(m.Quantity) > 0 && !m.Confirmed {
				// Validate quantity doesn't exceed available shares
				quantity := 0
				fmt.Sscanf(m.Quantity, "%d", &quantity)
				if quantity > 0 && quantity <= m.Holding.Shares {
					m.Confirmed = true
					return m, nil
				}
				m.Error = fmt.Sprintf("Invalid quantity. You have %d shares.", m.Holding.Shares)
			}
		case "ctrl+b", "esc":
			m.BackPressed = true
		default:
			// Add only numeric characters
			if len(msg.String()) == 1 {
				runes := []rune(msg.String())
				ch := runes[0]
				if ch >= '0' && ch <= '9' {
					m.Quantity = m.Quantity[:m.Cursor] + string(ch) + m.Quantity[m.Cursor:]
					m.Cursor++
				}
			}
		}
	}
	return m, nil
}

// View renders the sell stock page
func (m *SellStockModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render(fmt.Sprintf("💸 Sell %s", m.Holding.Symbol)) + "\n"
	s += styles.InputStyle.Render(fmt.Sprintf("Available: %d shares at $%.2f each", m.Holding.Shares, m.Holding.Price)) + "\n\n"

	if m.Error != "" {
		s += styles.ErrorStyle.Render(m.Error) + "\n\n"
	}

	// Display stock information
	s += styles.InputStyle.Render(fmt.Sprintf("Current Price: $%.2f", m.Holding.Price)) + "\n"
	s += styles.InputStyle.Render(fmt.Sprintf("Last Updated: %s", m.Holding.Timestamp)) + "\n\n"

	// Display quantity input
	s += "Enter quantity to sell:\n"

	// Display input field with cursor
	quantityDisplay := m.Quantity
	if m.Cursor <= len(quantityDisplay) {
		if m.Cursor == len(quantityDisplay) {
			quantityDisplay += "|"
		} else {
			quantityDisplay = quantityDisplay[:m.Cursor] + "|" + quantityDisplay[m.Cursor:]
		}
	}

	s += styles.InputStyle.Render(quantityDisplay) + "\n\n"

	// Calculate and display total proceeds if quantity is entered
	if len(m.Quantity) > 0 {
		quantity := 0
		fmt.Sscanf(m.Quantity, "%d", &quantity)
		if quantity > 0 && quantity <= m.Holding.Shares {
			totalProceeds := m.Holding.Price * float64(quantity)
			s += styles.SuccessStyle.Render(fmt.Sprintf("Total Proceeds: $%.2f", totalProceeds)) + "\n\n"
			s += styles.FooterStyle.Render("Press Enter to confirm sale • 'Ctrl + b' or 'Esc' to cancel") + "\n\n"
		} else if quantity > m.Holding.Shares {
			s += styles.ErrorStyle.Render(fmt.Sprintf("Cannot sell %d shares (only have %d)", quantity, m.Holding.Shares)) + "\n\n"
			s += styles.FooterStyle.Render("Enter a valid quantity • 'Ctrl + b' or 'Esc' to cancel") + "\n\n"
		} else {
			s += styles.FooterStyle.Render("Enter a valid quantity • 'Ctrl + b' or 'Esc' to cancel") + "\n\n"
		}
	} else {
		s += styles.FooterStyle.Render("Type quantity and press Enter • 'Ctrl + b' or 'Esc' to cancel") + "\n\n"
	}

	return s
}

// GetQuantity returns the entered quantity as an integer
func (m *SellStockModel) GetQuantity() int {
	quantity := 0
	fmt.Sscanf(m.Quantity, "%d", &quantity)
	return quantity
}

// GetHolding returns the holding being sold
func (m *SellStockModel) GetHolding() Holding {
	return m.Holding
}
