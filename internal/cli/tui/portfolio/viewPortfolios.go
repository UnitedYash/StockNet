package portfolio

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/database"
)

// Portfolio represents a user's portfolio
type Portfolio struct {
	PortfolioID string
	Name        string
	CashAccount float64
}

// Model for viewing user's portfolios
type ViewPortfoliosModel struct {
	portfolios  []Portfolio
	Selected    int
	loading     bool
	BackPressed bool
	error       string
	currentUserID int
}

type PortfolioSelectedMsg struct {
	Portfolio Portfolio
}

type portfoliosLoadedMsg struct {
	portfolios []Portfolio
}

type portfoliosErrorMsg struct {
	err error
}

// NewViewPortfoliosPageWithUserID returns initial view portfolios page model
func NewViewPortfoliosPageWithUserID(userID int) *ViewPortfoliosModel {
	return &ViewPortfoliosModel{
		portfolios:    []Portfolio{},
		Selected:      0,
		loading:       true,
		BackPressed:   false,
		currentUserID: userID,
	}
}

// returns initial command for the view portfolios page
func (m *ViewPortfoliosModel) Init() tea.Cmd {
	return func() tea.Msg {
		// Query portfolios from database for current user
		db := database.New().GetDB()

		query := `SELECT portfolio_id, name, cash_account FROM Portfolio
		          WHERE user_id = $1
		          ORDER BY portfolio_id`

		rows, err := db.Query(query, m.currentUserID)
		if err != nil {
			return portfoliosErrorMsg{err: err}
		}
		defer rows.Close()

		var portfolios []Portfolio
		for rows.Next() {
			var portfolioID int
			var name string
			var cashAccount float64
			if err := rows.Scan(&portfolioID, &name, &cashAccount); err != nil {
				return portfoliosErrorMsg{err: err}
			}
			portfolios = append(portfolios, Portfolio{
				PortfolioID: fmt.Sprintf("%d", portfolioID),
				Name:        name,
				CashAccount: cashAccount,
			})
		}

		return portfoliosLoadedMsg{portfolios: portfolios}
	}
}

// Handles incoming events and updates the model
func (m *ViewPortfoliosModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case portfoliosLoadedMsg:
		m.portfolios = msg.portfolios
		m.loading = false
		m.error = ""
	case portfoliosErrorMsg:
		m.loading = false
		m.error = fmt.Sprintf("Error loading portfolios: %v", msg.err)
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Selected > 0 {
				m.Selected--
			}
		case "down", "j":
			if m.Selected < len(m.portfolios)-1 {
				m.Selected++
			}
		case "enter":
			if (len(m.portfolios) > 0) {
				selectedPortfolio := m.portfolios[m.Selected]
				return m, func() tea.Msg {
					return PortfolioSelectedMsg{Portfolio: selectedPortfolio}
				}
			}
		case "ctrl+b", "esc":
			m.BackPressed = true
		}
	}
	return m, nil
}

// renders the view portfolios page
func (m *ViewPortfoliosModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("📊 Your Portfolios") + "\n\n"

	if m.loading {
		s += "Loading portfolios...\n"
	} else if m.error != "" {
		s += styles.ErrorStyle.Render(m.error) + "\n"
	} else if len(m.portfolios) == 0 {
		s += "No portfolios found. Create one to get started!\n"
	} else {
		s += styles.HeaderStyle.Render("ID    Name                          Cash Account") + "\n"
		s += "────────────────────────────────────────────────────────\n"

		for i, p := range m.portfolios {
			line := fmt.Sprintf("%-5s %-30s $%.2f", p.PortfolioID, p.Name, p.CashAccount)
			if i == m.Selected {
				s += styles.SelectedStyle.Render(line) + "\n"
			} else {
				s += line + "\n"
			}
		}
	}

	s += "\n"
	s += styles.FooterStyle.Render("↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}
