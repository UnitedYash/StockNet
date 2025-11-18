package tui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
)

// Model for stock symbol search/input page
type SearchStockModel struct {
	input       string   // User input for stock symbol
	confirmed   bool     // User pressed enter
	backPressed bool
	cursor      int      // Cursor position in input
}

// returns initial search stock page model
func newSearchStockPage() *SearchStockModel {
	return &SearchStockModel{
		input:    "",
		cursor:   0,
		confirmed: false,
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
			if len(m.input) > 0 && !m.confirmed {
				m.confirmed = true
				return m, nil
			}
		case "ctrl+b", "esc":
			m.backPressed = true
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
	s += TitleStyle.Render("🔍 Search Stock") + "\n\n"

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

	s += InputStyle.Render(inputDisplay) + "\n\n"

	if len(m.input) == 0 {
		s += FooterStyle.Render("Type symbol and press Enter • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	} else {
		s += SuccessStyle.Render(fmt.Sprintf("Symbol: %s", m.input)) + "\n"
		s += FooterStyle.Render("Press Enter to search • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	}

	return s
}

// GetSymbol returns the entered stock symbol
func (m *SearchStockModel) GetSymbol() string {
	return m.input
}
