package portfolio

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
)

// DepositCashModel - Enter amount and confirm deposit
type DepositCashModel struct {
	PortfolioID   string
	Amount        string // Input as string for editing
	Cursor        int
	BackPressed   bool
	Confirmed     bool
	Error         string
	CurrentUserID int
}

type DepositCashCompletedMsg struct {
	Success bool
	Message string
}

// NewDepositCashPageWithPortfolioID creates a new deposit cash page
func NewDepositCashPageWithPortfolioID(userID int, portfolioID string) *DepositCashModel {
	return &DepositCashModel{
		PortfolioID:   portfolioID,
		Amount:        "",
		Cursor:        0,
		BackPressed:   false,
		Confirmed:     false,
		Error:         "",
		CurrentUserID: userID,
	}
}

// Init returns initial command
func (m *DepositCashModel) Init() tea.Cmd {
	return nil
}

// Update handles incoming events
func (m *DepositCashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "backspace":
			if m.Cursor > 0 {
				m.Amount = m.Amount[:m.Cursor-1] + m.Amount[m.Cursor:]
				m.Cursor--
			}
		case "left":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "right":
			if m.Cursor < len(m.Amount) {
				m.Cursor++
			}
		case "home":
			m.Cursor = 0
		case "end":
			m.Cursor = len(m.Amount)
		case "enter":
			if len(m.Amount) > 0 && !m.Confirmed {
				m.Confirmed = true
				return m, nil
			}
		case "ctrl+b", "esc":
			m.BackPressed = true
		default:
			// Add numeric characters and decimal point
			if len(msg.String()) == 1 {
				runes := []rune(msg.String())
				ch := runes[0]
				if (ch >= '0' && ch <= '9') || ch == '.' {
					// Allow only one decimal point
					if ch == '.' && containsDecimal(m.Amount) {
						return m, nil
					}
					m.Amount = m.Amount[:m.Cursor] + string(ch) + m.Amount[m.Cursor:]
					m.Cursor++
				}
			}
		}
	}
	return m, nil
}

// View renders the deposit cash page
func (m *DepositCashModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("💵 Deposit Cash") + "\n\n"

	if m.Error != "" {
		s += styles.ErrorStyle.Render(m.Error) + "\n\n"
	}

	// Display input prompt
	s += "Enter amount to deposit:\n"

	// Display input field with cursor
	amountDisplay := m.Amount
	if m.Cursor <= len(amountDisplay) {
		if m.Cursor == len(amountDisplay) {
			amountDisplay += "|"
		} else {
			amountDisplay = amountDisplay[:m.Cursor] + "|" + amountDisplay[m.Cursor:]
		}
	}

	s += styles.InputStyle.Render("$" + amountDisplay) + "\n\n"

	// Parse and display the amount
	if len(m.Amount) > 0 {
		amount := 0.0
		fmt.Sscanf(m.Amount, "%f", &amount)
		if amount > 0 {
			s += styles.SuccessStyle.Render(fmt.Sprintf("You will deposit: $%.2f", amount)) + "\n\n"
			s += styles.FooterStyle.Render("Press Enter to confirm deposit • 'Ctrl + b' or 'Esc' to cancel") + "\n\n"
		} else {
			s += styles.FooterStyle.Render("Enter a valid amount • 'Ctrl + b' or 'Esc' to cancel") + "\n\n"
		}
	} else {
		s += styles.FooterStyle.Render("Type amount and press Enter • 'Ctrl + b' or 'Esc' to cancel") + "\n\n"
	}

	return s
}

// GetAmount returns the entered amount as a float
func (m *DepositCashModel) GetAmount() float64 {
	amount := 0.0
	fmt.Sscanf(m.Amount, "%f", &amount)
	return amount
}

// Helper function to check if string contains decimal point
func containsDecimal(s string) bool {
	for _, ch := range s {
		if ch == '.' {
			return true
		}
	}
	return false
}
