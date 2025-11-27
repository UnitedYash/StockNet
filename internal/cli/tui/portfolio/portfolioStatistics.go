package portfolio

import (
	"context"
	"fmt"
	"sort"
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/database"
)

type StockStatistic struct {
	Symbol string
	COV    float64
	Beta   float64
}

type ViewPortfolioStatisticsModel struct {
	PortfolioID      string
	CurrentUserID    int
	TimeInterval     string
	IntervalIndex    int
	Statistics       []StockStatistic
	BackPressed      bool
	Loading          bool
	Error            string
	Selected         int
	ScrollOffset     int
	MatrixType       string // "correlation" or "covariance"
	ShowMatrix       bool
	CorrelationData  map[string]map[string]float64
	CovarianceData   map[string]map[string]float64
	MatrixLoading    bool
}

type statisticsLoadedMsg struct {
	statistics []StockStatistic
}

type statisticsErrorMsg struct {
	err error
}

type matrixLoadedMsg struct {
	matrixType string                        // "correlation" or "covariance"
	data       map[string]map[string]float64
}

type matrixErrorMsg struct {
	err error
}

var timeIntervals = []string{"week", "month", "quarter", "year", "5years"}

// NewViewPortfolioStatisticsPageWithPortfolioID creates a new portfolio statistics page
func NewViewPortfolioStatisticsPageWithPortfolioID(userID int, portfolioID string) *ViewPortfolioStatisticsModel {
	return &ViewPortfolioStatisticsModel{
		PortfolioID:   portfolioID,
		CurrentUserID: userID,
		TimeInterval:  "year",
		IntervalIndex: 3,
		Statistics:    []StockStatistic{},
		BackPressed:   false,
		Loading:       true,
		Error:         "",
		Selected:      0,
		ScrollOffset:  0,
		ShowMatrix:    false,
	}
}

// Init fetches the statistics data
func (m *ViewPortfolioStatisticsModel) Init() tea.Cmd {
	return func() tea.Msg {
		db := database.New()

		// Get portfolio holdings
		query := `SELECT symbol FROM hasStockFromPortfolio WHERE portfolio_id = $1 ORDER BY symbol`
		rows, err := db.GetDB().Query(query, m.PortfolioID)
		if err != nil {
			return statisticsErrorMsg{err: err}
		}
		defer rows.Close()

		var symbols []string
		for rows.Next() {
			var symbol string
			if err := rows.Scan(&symbol); err != nil {
				return statisticsErrorMsg{err: err}
			}
			symbols = append(symbols, symbol)
		}

		if len(symbols) == 0 {
			return statisticsErrorMsg{err: fmt.Errorf("portfolio has no holdings")}
		}

		// Calculate statistics for each stock
		var statistics []StockStatistic
		portfolioID := parsePortfolioID(m.PortfolioID)
		for _, symbol := range symbols {
			cov, _ := db.GetCOV(context.Background(), portfolioID, symbol, m.TimeInterval)
			beta, _ := db.GetBeta(context.Background(), portfolioID, symbol, m.TimeInterval)

			statistics = append(statistics, StockStatistic{
				Symbol: symbol,
				COV:    cov,
				Beta:   beta,
			})
		}

		// Sort by symbol
		sort.Slice(statistics, func(i, j int) bool {
			return statistics[i].Symbol < statistics[j].Symbol
		})

		return statisticsLoadedMsg{statistics: statistics}
	}
}

// Update handles incoming events
func (m *ViewPortfolioStatisticsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case statisticsLoadedMsg:
		m.Statistics = msg.statistics
		m.Loading = false
	case statisticsErrorMsg:
		m.Error = fmt.Sprintf("Error loading statistics: %v", msg.err)
		m.Loading = false
	case matrixLoadedMsg:
		if msg.matrixType == "correlation" {
			m.CorrelationData = msg.data
		} else {
			m.CovarianceData = msg.data
		}
		m.MatrixLoading = false
	case matrixErrorMsg:
		m.Error = fmt.Sprintf("Error loading matrix: %v", msg.err)
		m.MatrixLoading = false
	case tea.KeyMsg:
		if m.ShowMatrix {
			// Matrix view navigation
			switch msg.String() {
			case "up", "k":
				if m.Selected > 0 {
					m.Selected--
				}
			case "down", "j":
				if m.Selected < len(m.Statistics)-1 {
					m.Selected++
				}
			case "esc", "ctrl+b":
				m.ShowMatrix = false
				m.Selected = 0
			}
		} else {
			// Main statistics view navigation
			switch msg.String() {
			case "up", "k":
				if m.Selected > 0 {
					m.Selected--
				} else {
					m.Selected = len(m.Statistics) - 1
				}
			case "down", "j":
				if m.Selected < len(m.Statistics)-1 {
					m.Selected++
				} else {
					m.Selected = 0
				}
			case "left", "h":
				if m.IntervalIndex > 0 {
					m.IntervalIndex--
					m.TimeInterval = timeIntervals[m.IntervalIndex]
					m.Loading = true
					return m, m.Init()
				}
			case "right", "l":
				if m.IntervalIndex < len(timeIntervals)-1 {
					m.IntervalIndex++
					m.TimeInterval = timeIntervals[m.IntervalIndex]
					m.Loading = true
					return m, m.Init()
				}
			case "c":
				// View correlation matrix
				m.ShowMatrix = true
				m.MatrixType = "correlation"
				m.Selected = 0
				m.MatrixLoading = true
				return m, m.loadMatrix()
			case "o":
				// View covariance matrix
				m.ShowMatrix = true
				m.MatrixType = "covariance"
				m.Selected = 0
				m.MatrixLoading = true
				return m, m.loadMatrix()
			case "ctrl+b", "esc":
				m.BackPressed = true
			}
		}
	}
	return m, nil
}

// View renders the statistics page
func (m *ViewPortfolioStatisticsModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("📊 Portfolio Statistics") + "\n\n"

	if m.ShowMatrix {
		return m.viewMatrix()
	}

	if m.Loading {
		s += "Loading statistics...\n"
	} else if m.Error != "" {
		s += styles.ErrorStyle.Render(m.Error) + "\n"
	} else if len(m.Statistics) == 0 {
		s += "No holdings in this portfolio.\n"
	} else {
		// Show time interval selector
		s += "Time Interval: "
		for i, interval := range timeIntervals {
			if i == m.IntervalIndex {
				s += styles.SelectedStyle.Render("["+ interval + "]") + " "
			} else {
				s += styles.UnselectedStyle.Render("[" + interval + "]") + " "
			}
		}
		s += "\n\n"

		// Show statistics table
		s += styles.HeaderStyle.Render("Symbol        COV         Beta") + "\n"
		s += "──────────────────────────────────────────────────\n"

		const viewportHeight = 8
		if m.Selected < m.ScrollOffset {
			m.ScrollOffset = m.Selected
		}
		if m.Selected >= m.ScrollOffset+viewportHeight {
			m.ScrollOffset = m.Selected - viewportHeight + 1
		}

		endIdx := m.ScrollOffset + viewportHeight
		if endIdx > len(m.Statistics) {
			endIdx = len(m.Statistics)
		}

		for i := m.ScrollOffset; i < endIdx; i++ {
			stat := m.Statistics[i]
			line := fmt.Sprintf("%-13s %-11.4f %-11.4f", stat.Symbol, stat.COV, stat.Beta)
			if i == m.Selected {
				s += styles.SelectedStyle.Render("→ " + line) + "\n"
			} else {
				s += styles.UnselectedStyle.Render("  " + line) + "\n"
			}
		}

		s += fmt.Sprintf("\n(%d of %d)\n", m.Selected+1, len(m.Statistics))
	}

	s += "\n"
	s += styles.FooterStyle.Render("← / → to change interval • c: correlation matrix • o: covariance matrix • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}

// loadMatrix fetches correlation or covariance matrix from database
func (m *ViewPortfolioStatisticsModel) loadMatrix() tea.Cmd {
	return func() tea.Msg {
		db := database.New()
		portfolioID := parsePortfolioID(m.PortfolioID)

		var data map[string]map[string]float64
		var err error

		if m.MatrixType == "correlation" {
			data, err = db.GetCorrelationMatrix(context.Background(), portfolioID, m.TimeInterval)
		} else {
			data, err = db.GetCovarianceMatrix(context.Background(), portfolioID, m.TimeInterval)
		}

		if err != nil {
			return matrixErrorMsg{err: err}
		}

		return matrixLoadedMsg{matrixType: m.MatrixType, data: data}
	}
}

// viewMatrix displays the correlation or covariance matrix for selected stock pair
func (m *ViewPortfolioStatisticsModel) viewMatrix() string {
	s := "\n"
	s += styles.TitleStyle.Render(fmt.Sprintf("📐 %s Matrix (%s)", m.formatMatrixType(), m.TimeInterval)) + "\n\n"

	if m.MatrixLoading {
		s += "Loading matrix...\n"
	} else if m.Selected >= len(m.Statistics) {
		s += "No stock selected.\n"
	} else {
		selectedSymbol := m.Statistics[m.Selected].Symbol
		label := "Correlations with:"
		if m.MatrixType == "covariance" {
			label = "Covariances with:"
		}
		s += fmt.Sprintf("%s %s\n\n", label, styles.SelectedStyle.Render(selectedSymbol))

		s += styles.HeaderStyle.Render("Stock       Value") + "\n"
		s += "─────────────────────────────\n"

		const viewportHeight = 10

		// Get the matrix data
		var matrix map[string]map[string]float64
		if m.MatrixType == "correlation" {
			matrix = m.CorrelationData
		} else {
			matrix = m.CovarianceData
		}

		// Display values for selected stock
		if matrix != nil && matrix[selectedSymbol] != nil {
			for i, stat := range m.Statistics {
				value := matrix[selectedSymbol][stat.Symbol]
				line := fmt.Sprintf("%-11s %.4f", stat.Symbol, value)
				s += styles.UnselectedStyle.Render("  " + line) + "\n"

				if i >= viewportHeight {
					break
				}
			}
		} else {
			s += "No matrix data available.\n"
		}
	}

	s += "\n"
	s += styles.FooterStyle.Render("↑/↓ to select stock • 'Esc' to go back") + "\n\n"
	return s
}

func (m *ViewPortfolioStatisticsModel) formatMatrixType() string {
	if m.MatrixType == "correlation" {
		return "Correlation"
	}
	return "Covariance"
}

func parsePortfolioID(portfolioIDStr string) int {
	var portfolioID int
	fmt.Sscanf(portfolioIDStr, "%d", &portfolioID)
	return portfolioID
}
