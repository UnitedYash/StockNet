package tui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
)

// BuyStockModel - Enter quantity and confirm purchase
type BuyStockModel struct {
	stock         Stock
	quantity      string // Input as string for editing
	cursor        int
	loading       bool
	backPressed   bool
	confirmed     bool
	error         string
	currentUserID int
	portfolioID   string
	cashAccount   float64
}

type buyStockConfirmedMsg struct {
	Stock    Stock
	Quantity int
}

type buyStockErrorMsg struct {
	err error
}

// returns initial buy stock page model
func newBuyStockPageWithStock(userID int, portfolioID string, stock Stock, cashAccount float64) *BuyStockModel {
	return &BuyStockModel{
		stock:         stock,
		quantity:      "",
		cursor:        0,
		loading:       false,
		backPressed:   false,
		confirmed:     false,
		currentUserID: userID,
		portfolioID:   portfolioID,
		cashAccount:   cashAccount,
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
		m.error = msg.err.Error()
		m.loading = false
	case tea.KeyMsg:
		switch msg.String() {
		case "backspace":
			if m.cursor > 0 {
				m.quantity = m.quantity[:m.cursor-1] + m.quantity[m.cursor:]
				m.cursor--
			}
		case "left":
			if m.cursor > 0 {
				m.cursor--
			}
		case "right":
			if m.cursor < len(m.quantity) {
				m.cursor++
			}
		case "home":
			m.cursor = 0
		case "end":
			m.cursor = len(m.quantity)
		case "enter":
			if len(m.quantity) > 0 && !m.confirmed {
				m.confirmed = true
				return m, nil
			}
		case "ctrl+b", "esc":
			m.backPressed = true
		default:
			// Add only numeric characters
			if len(msg.String()) == 1 {
				runes := []rune(msg.String())
				ch := runes[0]
				if ch >= '0' && ch <= '9' {
					m.quantity = m.quantity[:m.cursor] + string(ch) + m.quantity[m.cursor:]
					m.cursor++
				}
			}
		}
	}
	return m, nil
}

// renders the buy stock page
func (m *BuyStockModel) View() string {
	s := "\n"
	s += TitleStyle.Render(fmt.Sprintf("💰 Buy %s", m.stock.Symbol)) + "\n"
	s += InputStyle.Render(fmt.Sprintf("Cash Available: $%.2f", m.cashAccount)) + "\n\n"

	if m.error != "" {
		s += ErrorStyle.Render(m.error) + "\n\n"
	}

	// Display stock information
	s += InputStyle.Render(fmt.Sprintf("Current Price: $%.2f", m.stock.Price)) + "\n"
	s += InputStyle.Render(fmt.Sprintf("Last Updated: %s", m.stock.Timestamp)) + "\n\n"

	// Display quantity input
	s += "Enter quantity to buy:\n"

	// Display input field with cursor
	quantityDisplay := m.quantity
	if m.cursor <= len(quantityDisplay) {
		if m.cursor == len(quantityDisplay) {
			quantityDisplay += "|"
		} else {
			quantityDisplay = quantityDisplay[:m.cursor] + "|" + quantityDisplay[m.cursor:]
		}
	}

	s += InputStyle.Render(quantityDisplay) + "\n\n"

	// Calculate and display total cost if quantity is entered
	if len(m.quantity) > 0 {
		quantity := 0
		fmt.Sscanf(m.quantity, "%d", &quantity)
		if quantity > 0 {
			totalCost := m.stock.Price * float64(quantity)
			s += SuccessStyle.Render(fmt.Sprintf("Total Cost: $%.2f", totalCost)) + "\n\n"
			s += FooterStyle.Render("Press Enter to confirm purchase • 'Ctrl + b' or 'Esc' to cancel") + "\n\n"
		} else {
			s += FooterStyle.Render("Enter a valid quantity • 'Ctrl + b' or 'Esc' to cancel") + "\n\n"
		}
	} else {
		s += FooterStyle.Render("Type quantity and press Enter • 'Ctrl + b' or 'Esc' to cancel") + "\n\n"
	}

	return s
}

// GetQuantity returns the entered quantity as an integer
func (m *BuyStockModel) GetQuantity() int {
	quantity := 0
	fmt.Sscanf(m.quantity, "%d", &quantity)
	return quantity
}

// GetStock returns the selected stock
func (m *BuyStockModel) GetStock() Stock {
	return m.stock
}
