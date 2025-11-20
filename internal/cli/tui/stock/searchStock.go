package stock

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
)

// Model for stock symbol search/input page
type SearchStockModel struct {
	input       string   // User input for stock symbol
	Confirmed   bool     // User pressed enter
	BackPressed bool
	cursor      int      // Cursor position in input
}

// returns initial search stock page model
func NewSearchStockPage() *SearchStockModel {
	return &SearchStockModel{
		input:    "",
		cursor:   0,
		Confirmed: false,
	}
}

// returns initial command for the search stock page
func (m *SearchStockModel) Init() tea.Cmd {
	return nil
}

// Handles incoming events and updates the model accordingly
func (m *SearchStockModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "backspace":
			if m.cursor > 0 {
				m.input = m.input[:m.cursor-1] + m.input[m.cursor:]
				m.cursor--
			}
		case "left":
			if m.cursor > 0 {
				m.cursor--
			}
		case "right":
			if m.cursor < len(m.input) {
				m.cursor++
			}
		case "home":
			m.cursor = 0
		case "end":
			m.cursor = len(m.input)
		case "enter":
			if len(m.input) > 0 && !m.Confirmed {
				m.Confirmed = true
				return m, nil
			}
		case "ctrl+b", "esc":
			m.BackPressed = true
		default:
			// Add printable characters to input
			if len(msg.String()) == 1 {
				runes := []rune(msg.String())
				ch := runes[0]
				// Only allow alphanumeric characters
				if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') {
					m.input = m.input[:m.cursor] + string(ch) + m.input[m.cursor:]
					m.cursor++
				}
			}
		}
	}
	return m, nil
}

// renders UI for stock symbol search
func (m *SearchStockModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("🔍 Search Stock") + "\n\n"

	s += "Enter stock symbol (e.g., AAPL, GOOGL): \n"

	// Display input field with cursor
	inputDisplay := m.input
	if m.cursor <= len(inputDisplay) {
		// Insert cursor at the right position
		if m.cursor == len(inputDisplay) {
			inputDisplay += "|"
		} else {
			inputDisplay = inputDisplay[:m.cursor] + "|" + inputDisplay[m.cursor:]
		}
	}

	s += styles.InputStyle.Render(inputDisplay) + "\n\n"

	if len(m.input) == 0 {
		s += styles.FooterStyle.Render("Type symbol and press Enter • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	} else {
		s += styles.SuccessStyle.Render(fmt.Sprintf("Symbol: %s", m.input)) + "\n"
		s += styles.FooterStyle.Render("Press Enter to search • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	}

	return s
}

// GetSymbol returns the entered stock symbol
func (m *SearchStockModel) GetSymbol() string {
	return m.input
}
