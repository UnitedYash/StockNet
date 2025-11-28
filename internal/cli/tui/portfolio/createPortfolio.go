package portfolio

import (
	"fmt"
	"strconv"
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/database"
)

// Model for creating a new portfolio
type CreatePortfolioModel struct {
	step            int     // 0 = input name, 1 = input cash, 2 = confirm
	nameInput       string  // User input for portfolio name
	cashInput       string  // User input for cash amount
	cursor          int     // Cursor position in current input
	portfolioName   string  // Parsed portfolio name
	cashAmount      float64 // Parsed cash amount
	Confirmed       bool    // User confirmed creation
	BackPressed     bool
	loading         bool
	error           string
	successMessage  string
	portfolioID     int
	currentUserID   int // Current logged-in user's ID
}

type portfolioCreatedMsg struct {
	portfolioID int
}

type portfolioCreationErrorMsg struct {
	err error
}

// NewCreatePortfolioPageWithUserID returns initial create portfolio page model
func NewCreatePortfolioPageWithUserID(userID int) *CreatePortfolioModel {
	return &CreatePortfolioModel{
		step:          0, // Start with name input
		nameInput:     "",
		cashInput:     "",
		cursor:        0,
		Confirmed:     false,
		BackPressed:   false,
		loading:       false,
		currentUserID: userID,
	}
}

// returns initial command for the create portfolio page
func (m *CreatePortfolioModel) Init() tea.Cmd {
	return nil
}

// Handles incoming events and updates the model
func (m *CreatePortfolioModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case portfolioCreatedMsg:
		m.loading = false
		m.successMessage = fmt.Sprintf("Portfolio '%s' created successfully! ID: %d", m.portfolioName, msg.portfolioID)
		m.portfolioID = msg.portfolioID
	case portfolioCreationErrorMsg:
		m.loading = false
		m.error = fmt.Sprintf("Error creating portfolio: %v", msg.err)
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+b", "esc":
			m.BackPressed = true
		case "enter":
			if m.step == 0 && len(m.nameInput) > 0 {
				// Move to cash input step
				m.portfolioName = m.nameInput
				m.step = 1
				m.cursor = 0
				m.error = ""
			} else if m.step == 1 && len(m.cashInput) > 0 {
				// Parse and validate cash amount
				amount, err := strconv.ParseFloat(m.cashInput, 64)
				if err != nil || amount < 0 {
					m.error = "Invalid cash amount. Please enter a valid positive number."
					return m, nil
				}
				m.cashAmount = amount
				m.step = 2
				m.Confirmed = true
				m.loading = true
				return m, m.createPortfolioInDB()
			}
		case "backspace":
			if m.step == 0 && m.cursor > 0 {
				m.nameInput = m.nameInput[:m.cursor-1] + m.nameInput[m.cursor:]
				m.cursor--
				m.error = ""
			} else if m.step == 1 && m.cursor > 0 {
				m.cashInput = m.cashInput[:m.cursor-1] + m.cashInput[m.cursor:]
				m.cursor--
				m.error = ""
			}
		case "left":
			if m.cursor > 0 {
				m.cursor--
			}
		case "right":
			if m.step == 0 && m.cursor < len(m.nameInput) {
				m.cursor++
			} else if m.step == 1 && m.cursor < len(m.cashInput) {
				m.cursor++
			}
		default:
			// Add characters to current input
			if len(msg.String()) == 1 {
				runes := []rune(msg.String())
				ch := runes[0]

				// Step 0: Portfolio name (alphanumeric, spaces, some special chars)
				if m.step == 0 {
					if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == ' ' || ch == '-' || ch == '_' {
						m.nameInput = m.nameInput[:m.cursor] + string(ch) + m.nameInput[m.cursor:]
						m.cursor++
						m.error = ""
					}
				} else if m.step == 1 {
					// Step 1: Cash amount (digits and decimal point)
					if (ch >= '0' && ch <= '9') || ch == '.' {
						m.cashInput = m.cashInput[:m.cursor] + string(ch) + m.cashInput[m.cursor:]
						m.cursor++
						m.error = ""
					}
				}
			}
		}
	}
	return m, nil
}

// createPortfolioInDB creates a new portfolio in the database
func (m *CreatePortfolioModel) createPortfolioInDB() tea.Cmd {
	return func() tea.Msg {
		db := database.New().GetDB()

		// Insert into Portfolio table (portfolio_id is auto-generated SERIAL)
		var portfolioID int
		err := db.QueryRow(
			`INSERT INTO Portfolio (user_id, name, cash_account) VALUES ($1, $2, $3) RETURNING portfolio_id`,
			m.currentUserID, m.portfolioName, m.cashAmount,
		).Scan(&portfolioID)

		if err != nil {
			return portfolioCreationErrorMsg{err: err}
		}

		return portfolioCreatedMsg{portfolioID: portfolioID}
	}
}

// renders the create portfolio page
func (m *CreatePortfolioModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("➕ Create New Portfolio") + "\n\n"

	if m.loading {
		s += "Creating portfolio...\n"
	} else if m.successMessage != "" {
		s += styles.SuccessStyle.Render(m.successMessage) + "\n"
		s += styles.FooterStyle.Render("Press 'Ctrl + b' or 'Esc' to go back to portfolio menu") + "\n\n"
	} else if m.step == 0 {
		// Input portfolio name
		s += "Enter portfolio name (up to 40 characters): \n\n"

		inputDisplay := m.nameInput
		if m.cursor <= len(inputDisplay) {
			if m.cursor == len(inputDisplay) {
				inputDisplay += "|"
			} else {
				inputDisplay = inputDisplay[:m.cursor] + "|" + inputDisplay[m.cursor:]
			}
		}

		s += styles.InputStyle.Render(inputDisplay) + "\n\n"

		if m.error != "" {
			s += styles.ErrorStyle.Render(m.error) + "\n\n"
		}

		if len(m.nameInput) > 0 {
			s += styles.SuccessStyle.Render(fmt.Sprintf("Portfolio Name: %s", m.nameInput)) + "\n"
		}

		s += styles.FooterStyle.Render("Type name and press Enter • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	} else if m.step == 1 {
		// Input cash amount
		s += fmt.Sprintf("Portfolio: %s\n", m.portfolioName)
		s += "Enter initial cash amount: \n\n"

		inputDisplay := m.cashInput
		if m.cursor <= len(inputDisplay) {
			if m.cursor == len(inputDisplay) {
				inputDisplay += "|"
			} else {
				inputDisplay = inputDisplay[:m.cursor] + "|" + inputDisplay[m.cursor:]
			}
		}

		s += styles.InputStyle.Render(inputDisplay) + "\n\n"

		if m.error != "" {
			s += styles.ErrorStyle.Render(m.error) + "\n\n"
		}

		if len(m.cashInput) > 0 {
			s += styles.SuccessStyle.Render(fmt.Sprintf("Initial Cash: $%s", m.cashInput)) + "\n"
		}

		s += styles.FooterStyle.Render("Type amount and press Enter • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	}

	return s
}
