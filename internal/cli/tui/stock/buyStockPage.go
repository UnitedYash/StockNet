package stock

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
)

// BuyStockModel - Enter quantity and confirm purchase
type BuyStockModel struct {
	Stock         Stock
	Quantity      string // Input as string for editing
	Cursor        int
	Loading       bool
	BackPressed   bool
	Confirmed     bool
	Error         string
	CurrentUserID int
	PortfolioID   string
	CashAccount   float64
}

type buyStockConfirmedMsg struct {
	Stock    Stock
	Quantity int
}

type buyStockErrorMsg struct {
	err error
}

type BuyStockCompletedMsg struct {
	Success bool
	Message string
}

// returns initial buy stock page model
func NewBuyStockPageWithStock(userID int, portfolioID string, stock Stock, cashAccount float64) *BuyStockModel {
	return &BuyStockModel{
		Stock:         stock,
		Quantity:      "",
		Cursor:        0,
		Loading:       false,
		BackPressed:   false,
		Confirmed:     false,
		CurrentUserID: userID,
		PortfolioID:   portfolioID,
		CashAccount:   cashAccount,
	}
}

// returns initial command for the buy stock page
func (m *BuyStockModel) Init() tea.Cmd {
	return nil
}

// Handles incoming events and updates the model
func (m *BuyStockModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case buyStockErrorMsg:
		m.Error = msg.err.Error()
		m.Loading = false
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
				m.Confirmed = true
				return m, nil
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

// renders the buy stock page
func (m *BuyStockModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render(fmt.Sprintf("💰 Buy %s", m.Stock.Symbol)) + "\n"
	s += styles.InputStyle.Render(fmt.Sprintf("Cash Available: $%.2f", m.CashAccount)) + "\n\n"

	if m.Error != "" {
		s += styles.ErrorStyle.Render(m.Error) + "\n\n"
	}

	// Display stock information
	s += styles.InputStyle.Render(fmt.Sprintf("Current Price: $%.2f", m.Stock.Price)) + "\n"
	s += styles.InputStyle.Render(fmt.Sprintf("Last Updated: %s", m.Stock.Timestamp)) + "\n\n"

	// Display quantity input
	s += "Enter quantity to buy:\n"

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

	// Calculate and display total cost if quantity is entered
	if len(m.Quantity) > 0 {
		quantity := 0
		fmt.Sscanf(m.Quantity, "%d", &quantity)
		if quantity > 0 {
			totalCost := m.Stock.Price * float64(quantity)
			s += styles.SuccessStyle.Render(fmt.Sprintf("Total Cost: $%.2f", totalCost)) + "\n\n"
			s += styles.FooterStyle.Render("Press Enter to confirm purchase • 'Ctrl + b' or 'Esc' to cancel") + "\n\n"
		} else {
			s += styles.FooterStyle.Render("Enter a valid quantity • 'Ctrl + b' or 'Esc' to cancel") + "\n\n"
		}
	} else {
		s += styles.FooterStyle.Render("Type quantity and press Enter • 'Ctrl + b' or 'Esc' to cancel") + "\n\n"
	}

	return s
}

// GetQuantity returns the entered quantity as an integer
func (m *BuyStockModel) GetQuantity() int {
	quantity := 0
	fmt.Sscanf(m.Quantity, "%d", &quantity)
	return quantity
}

// GetStock returns the selected stock
func (m *BuyStockModel) GetStock() Stock {
	return m.Stock
}
