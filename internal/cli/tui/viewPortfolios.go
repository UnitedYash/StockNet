package tui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
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
	selected    int
	loading     bool
	backPressed bool
	error       string
	currentUserID int
}

type portfolioSelectedMsg struct {
	Portfolio Portfolio
}

type portfoliosLoadedMsg struct {
	portfolios []Portfolio
}

type portfoliosErrorMsg struct {
	err error
}

// returns initial view portfolios page model
func newViewPortfoliosPageWithUserID(userID int) *ViewPortfoliosModel {
	return &ViewPortfoliosModel{
		portfolios:    []Portfolio{},
		selected:      0,
		loading:       true,
		backPressed:   false,
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
			if m.selected > 0 {
				m.selected--
			}
		case "down", "j":
			if m.selected < len(m.portfolios)-1 {
				m.selected++
			}
		case "enter":
			if (len(m.portfolios) > 0) {
				selectedPortfolio := m.portfolios[m.selected]
				return m, func() tea.Msg {
					return portfolioSelectedMsg{Portfolio: selectedPortfolio}
				}
			}
		case "ctrl+b", "esc":
			m.backPressed = true
		}
	}
	return m, nil
}

// renders the view portfolios page
func (m *ViewPortfoliosModel) View() string {
	s := "\n"
	s += TitleStyle.Render("📊 Your Portfolios") + "\n\n"

	if m.loading {
		s += "Loading portfolios...\n"
	} else if m.error != "" {
		s += ErrorStyle.Render(m.error) + "\n"
	} else if len(m.portfolios) == 0 {
		s += "No portfolios found. Create one to get started!\n"
	} else {
		s += HeaderStyle.Render("ID    Name                          Cash Account") + "\n"
		s += "────────────────────────────────────────────────────────\n"

		for i, p := range m.portfolios {
			line := fmt.Sprintf("%-5s %-30s $%.2f", p.PortfolioID, p.Name, p.CashAccount)
			if i == m.selected {
				s += SelectedStyle.Render(line) + "\n"
			} else {
				s += line + "\n"
			}
		}
	}

	s += "\n"
	s += FooterStyle.Render("↑/↓ or k/j to navigate • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}
