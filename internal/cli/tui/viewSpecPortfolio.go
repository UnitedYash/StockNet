package tui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
)

type specificPortfolio struct {
	PortfolioID string
	Name 	  string
	CashAccount float64
}

type viewSpecPortfolioModel struct {
	portfolio    specificPortfolio
	portfolioID string
	backPressed  bool
	error        string
	currentUserID int
	options      []string
	selected     int
	optionSelected string // Track which option was selected
}


type specPortfolioLoadedMsg struct {
	portfolio specificPortfolio
}

type specPortfolioErrorMsg struct {
	err error
}

// returns initial view specific portfolio page model
func newViewSpecPortfolioPageWithUserID(userID int, portfolioID string) *viewSpecPortfolioModel {
	return &viewSpecPortfolioModel{
		portfolio:    specificPortfolio{},
		portfolioID: portfolioID,
		backPressed:  false,
		currentUserID: userID,
		options:      []string{"View Holdings", "View Transactions", "Buy Stock", "Sell Stock" ,"Back"},
		selected:     0,
	}
}

func (m *viewSpecPortfolioModel) Init() tea.Cmd {
	return nil
}

func (m *viewSpecPortfolioModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case portfolioSelectedMsg:
		m.portfolio = specificPortfolio{
			PortfolioID: msg.Portfolio.PortfolioID,
			Name:        msg.Portfolio.Name,
			CashAccount: msg.Portfolio.CashAccount,
		}
	case specPortfolioLoadedMsg:
		m.portfolio = specificPortfolio{
			PortfolioID: msg.portfolio.PortfolioID,
			Name:        msg.portfolio.Name,
			CashAccount: msg.portfolio.CashAccount,
		}
	case specPortfolioErrorMsg:
		m.error = msg.err.Error()
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selected > 0 {
				m.selected--
			} else {
				m.selected = len(m.options) - 1
			}
		case "down", "j":
			if m.selected < len(m.options) - 1 {
				m.selected++
			} else {
				m.selected = 0
			}
		case "enter":
			if m.selected >= 0 && m.selected < len(m.options) {
				m.optionSelected = m.options[m.selected]
			}
		case "ctrl+b", "esc":
			m.backPressed = true
		}
	}
	return m, nil
}

func (m *viewSpecPortfolioModel) View() string {
	if m.error != "" {
		return fmt.Sprintf("Error: %s", m.error)
	}
	portfolioName := m.portfolio.Name
	cashAccount := m.portfolio.CashAccount
	
	s := "\n"
	s += TitleStyle.Render(fmt.Sprintf("Portfolio: %s", portfolioName)) + "\n"
	s += InputStyle.Render(fmt.Sprintf("Cash Available: $%.2f", cashAccount)) + "\n\n"
	s += "Options:\n"

	for i, option := range m.options {
		if i == m.selected {
			s += fmt.Sprintf("%s\n", SelectedStyle.Render("→ "+option))
		} else {
			s += fmt.Sprintf("%s\n", UnselectedStyle.Render("  "+option))
		}
	}
	// footer
	s += FooterStyle.Render("↑/↓ or k/j to navigate 'Enter to select option' • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}




