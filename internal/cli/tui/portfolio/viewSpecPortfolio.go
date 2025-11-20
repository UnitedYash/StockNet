package portfolio

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
)

type specificPortfolio struct {
	PortfolioID string
	Name 	  string
	CashAccount float64
}

type ViewSpecPortfolioModel struct {
	Portfolio    specificPortfolio
	PortfolioID string
	BackPressed  bool
	error        string
	currentUserID int
	Options      []string
	Selected     int
	OptionSelected string // Track which option was selected
}


type specPortfolioLoadedMsg struct {
	portfolio specificPortfolio
}

type specPortfolioErrorMsg struct {
	err error
}

// NewViewSpecPortfolioPageWithUserID returns initial view specific portfolio page model
func NewViewSpecPortfolioPageWithUserID(userID int, portfolioID string) *ViewSpecPortfolioModel {
	return &ViewSpecPortfolioModel{
		Portfolio:    specificPortfolio{},
		PortfolioID: portfolioID,
		BackPressed:  false,
		currentUserID: userID,
		Options:      []string{"View Holdings", "View Transactions", "Buy Stock", "Sell Stock" ,"Back"},
		Selected:     0,
	}
}

func (m *ViewSpecPortfolioModel) Init() tea.Cmd {
	return nil
}

func (m *ViewSpecPortfolioModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case PortfolioSelectedMsg:
		m.Portfolio = specificPortfolio{
			PortfolioID: msg.Portfolio.PortfolioID,
			Name:        msg.Portfolio.Name,
			CashAccount: msg.Portfolio.CashAccount,
		}
	case specPortfolioLoadedMsg:
		m.Portfolio = specificPortfolio{
			PortfolioID: msg.portfolio.PortfolioID,
			Name:        msg.portfolio.Name,
			CashAccount: msg.portfolio.CashAccount,
		}
	case specPortfolioErrorMsg:
		m.error = msg.err.Error()
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.Selected > 0 {
				m.Selected--
			} else {
				m.Selected = len(m.Options) - 1
			}
		case "down", "j":
			if m.Selected < len(m.Options) - 1 {
				m.Selected++
			} else {
				m.Selected = 0
			}
		case "enter":
			if m.Selected >= 0 && m.Selected < len(m.Options) {
				m.OptionSelected = m.Options[m.Selected]
			}
		case "ctrl+b", "esc":
			m.BackPressed = true
		}
	}
	return m, nil
}

func (m *ViewSpecPortfolioModel) View() string {
	if m.error != "" {
		return fmt.Sprintf("Error: %s", m.error)
	}
	portfolioName := m.Portfolio.Name
	cashAccount := m.Portfolio.CashAccount

	s := "\n"
	s += styles.TitleStyle.Render(fmt.Sprintf("Portfolio: %s", portfolioName)) + "\n"
	s += styles.InputStyle.Render(fmt.Sprintf("Cash Available: $%.2f", cashAccount)) + "\n\n"
	s += "Options:\n"

	for i, option := range m.Options {
		if i == m.Selected {
			s += fmt.Sprintf("%s\n", styles.SelectedStyle.Render("→ "+option))
		} else {
			s += fmt.Sprintf("%s\n", styles.UnselectedStyle.Render("  "+option))
		}
	}
	// footer
	s += styles.FooterStyle.Render("↑/↓ or k/j to navigate 'Enter to select option' • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}




