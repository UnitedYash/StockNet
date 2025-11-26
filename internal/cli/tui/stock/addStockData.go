package stock

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
)

// AddStockDataModel - Input form for stock data (date, open, high, low, close, volume)
type AddStockDataModel struct {
	Symbol       string
	DateInput    string // YYYY-MM-DD format
	OpenInput    string
	HighInput    string
	LowInput     string
	CloseInput   string
	VolumeInput  string
	CurrentField int // 0=date, 1=open, 2=high, 3=low, 4=close, 5=volume
	Cursor       int
	BackPressed  bool
	Confirmed    bool
	Error        string
}

type AddStockDataCompletedMsg struct {
	Symbol   string
	Date     string
	Open     float64
	High     float64
	Low      float64
	Close    float64
	Volume   int
	Success  bool
	Message  string
}

// NewAddStockDataPageWithSymbol creates a new add stock data page
func NewAddStockDataPageWithSymbol(symbol string) *AddStockDataModel {
	return &AddStockDataModel{
		Symbol:       symbol,
		DateInput:    "",
		OpenInput:    "",
		HighInput:    "",
		LowInput:     "",
		CloseInput:   "",
		VolumeInput:  "",
		CurrentField: 0,
		Cursor:       0,
		BackPressed:  false,
		Confirmed:    false,
		Error:        "",
	}
}

// Init returns initial command
func (m *AddStockDataModel) Init() tea.Cmd {
	return nil
}

// Update handles incoming events
func (m *AddStockDataModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab":
			// Move to next field
			m.CurrentField = (m.CurrentField + 1) % 6
			m.Cursor = 0
		case "shift+tab":
			// Move to previous field
			m.CurrentField = (m.CurrentField - 1 + 6) % 6
			m.Cursor = 0
		case "backspace":
			if m.Cursor > 0 {
				current := m.getCurrentFieldString()
				m.updateCurrentField(current[:m.Cursor-1] + current[m.Cursor:])
				m.Cursor--
			}
		case "left":
			if m.Cursor > 0 {
				m.Cursor--
			}
		case "right":
			currentVal := m.getCurrentFieldString()
			if m.Cursor < len(currentVal) {
				m.Cursor++
			}
		case "home":
			m.Cursor = 0
		case "end":
			m.Cursor = len(m.getCurrentFieldString())
		case "enter":
			if m.validateAllFields() && !m.Confirmed {
				m.Confirmed = true
				return m, nil
			}
		case "ctrl+b", "esc":
			m.BackPressed = true
		default:
			// Add valid characters
			if len(msg.String()) == 1 {
				runes := []rune(msg.String())
				ch := runes[0]
				// Allow digits, dots, and hyphens (for dates)
				if (ch >= '0' && ch <= '9') || ch == '.' || ch == '-' {
					// Validate field-specific requirements
					if m.CurrentField == 0 { // Date field
						if m.validateDateInput(ch, m.getCurrentFieldString()) {
							m.addCharToCurrentField(ch)
						}
					} else if m.CurrentField == 6 { // Volume field (integer only)
						if ch >= '0' && ch <= '9' {
							m.addCharToCurrentField(ch)
						}
					} else { // Numeric fields (open, high, low, close)
						if m.validateNumericInput(ch, m.getCurrentFieldString()) {
							m.addCharToCurrentField(ch)
						}
					}
				}
			}
		}
	}
	return m, nil
}

// View renders the add stock data page
func (m *AddStockDataModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render(fmt.Sprintf("📝 Add Stock Data for %s", m.Symbol)) + "\n\n"

	if m.Error != "" {
		s += styles.ErrorStyle.Render(m.Error) + "\n\n"
	}

	// Field labels and inputs
	fields := []struct {
		label string
		value string
		hint  string
	}{
		{"Date (YYYY-MM-DD):", m.DateInput, "Enter date in YYYY-MM-DD format"},
		{"Open Price:", m.OpenInput, "Enter opening price"},
		{"High Price:", m.HighInput, "Enter highest price"},
		{"Low Price:", m.LowInput, "Enter lowest price"},
		{"Close Price:", m.CloseInput, "Enter closing price"},
		{"Volume:", m.VolumeInput, "Enter trading volume"},
	}

	for i, field := range fields {
		prefix := "  "
		if i == m.CurrentField {
			prefix = "→ "
		}

		s += prefix + field.label + "\n"

		// Display input with cursor
		displayVal := field.value
		if i == m.CurrentField {
			if m.Cursor <= len(displayVal) {
				if m.Cursor == len(displayVal) {
					displayVal += "|"
				} else {
					displayVal = displayVal[:m.Cursor] + "|" + displayVal[m.Cursor:]
				}
			}
			s += "    " + styles.SelectedStyle.Render(displayVal) + "\n"
		} else {
			if displayVal == "" {
				s += "    (empty)\n"
			} else {
				s += "    " + displayVal + "\n"
			}
		}
		s += "\n"
	}

	s += "\n"
	s += styles.FooterStyle.Render("Tab/Shift+Tab to move between fields • Enter to submit • 'Ctrl + b' or 'Esc' to cancel") + "\n\n"

	return s
}

// Helper functions

func (m *AddStockDataModel) getCurrentFieldString() string {
	switch m.CurrentField {
	case 0:
		return m.DateInput
	case 1:
		return m.OpenInput
	case 2:
		return m.HighInput
	case 3:
		return m.LowInput
	case 4:
		return m.CloseInput
	case 5:
		return m.VolumeInput
	default:
		return ""
	}
}

func (m *AddStockDataModel) getCurrentFieldValue() *string {
	switch m.CurrentField {
	case 0:
		return &m.DateInput
	case 1:
		return &m.OpenInput
	case 2:
		return &m.HighInput
	case 3:
		return &m.LowInput
	case 4:
		return &m.CloseInput
	case 5:
		return &m.VolumeInput
	default:
		return nil
	}
}

func (m *AddStockDataModel) updateCurrentField(value string) {
	switch m.CurrentField {
	case 0:
		m.DateInput = value
	case 1:
		m.OpenInput = value
	case 2:
		m.HighInput = value
	case 3:
		m.LowInput = value
	case 4:
		m.CloseInput = value
	case 5:
		m.VolumeInput = value
	}
}

func (m *AddStockDataModel) addCharToCurrentField(ch rune) {
	current := m.getCurrentFieldString()
	newVal := current[:m.Cursor] + string(ch) + current[m.Cursor:]
	m.updateCurrentField(newVal)
	m.Cursor++
}

func (m *AddStockDataModel) validateDateInput(ch rune, current string) bool {
	// Allow YYYY-MM-DD format
	if ch == '-' {
		// Allow only 2 hyphens (at positions 4 and 7)
		hyphens := 0
		for _, c := range current {
			if c == '-' {
				hyphens++
			}
		}
		return hyphens < 2 && len(current) <= 9
	}
	// Allow digits
	return ch >= '0' && ch <= '9' && len(current) < 10
}

func (m *AddStockDataModel) validateNumericInput(ch rune, current string) bool {
	// Allow one decimal point
	if ch == '.' {
		for _, c := range current {
			if c == '.' {
				return false
			}
		}
		return true
	}
	// Allow digits
	return ch >= '0' && ch <= '9'
}

func (m *AddStockDataModel) validateAllFields() bool {
	// Validate date format (basic check)
	if len(m.DateInput) != 10 || m.DateInput[4] != '-' || m.DateInput[7] != '-' {
		m.Error = "Invalid date format. Use YYYY-MM-DD"
		return false
	}

	// Validate numeric fields
	if m.OpenInput == "" || m.HighInput == "" || m.LowInput == "" || m.CloseInput == "" {
		m.Error = "All price fields are required"
		return false
	}

	if m.VolumeInput == "" {
		m.Error = "Volume is required"
		return false
	}

	m.Error = ""
	return true
}

// GetData returns parsed data as float64/int values
func (m *AddStockDataModel) GetData() (date string, open, high, low, close float64, volume int, err error) {
	date = m.DateInput
	_, err = fmt.Sscanf(m.OpenInput, "%f", &open)
	if err != nil {
		return
	}
	_, err = fmt.Sscanf(m.HighInput, "%f", &high)
	if err != nil {
		return
	}
	_, err = fmt.Sscanf(m.LowInput, "%f", &low)
	if err != nil {
		return
	}
	_, err = fmt.Sscanf(m.CloseInput, "%f", &close)
	if err != nil {
		return
	}
	_, err = fmt.Sscanf(m.VolumeInput, "%d", &volume)
	return
}
