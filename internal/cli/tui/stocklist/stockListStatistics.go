package stocklist

import (
	"context"
	"fmt"
	"sort"
	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/database"
)

type StockListStatistic struct {
	Symbol string
	COV    float64
	Beta   float64
}

type ViewStockListStatisticsModel struct {
	StockListID      int
	TimeInterval     string
	IntervalIndex    int
	Statistics       []StockListStatistic
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

type stockListStatisticsLoadedMsg struct {
	statistics []StockListStatistic
}

type stockListStatisticsErrorMsg struct {
	err error
}

type stockListMatrixLoadedMsg struct {
	matrixType string                        // "correlation" or "covariance"
	data       map[string]map[string]float64
}

type stockListMatrixErrorMsg struct {
	err error
}

var timeIntervals = []string{"week", "month", "quarter", "year", "5years"}

// NewViewStockListStatisticsPage creates a new stock list statistics page
func NewViewStockListStatisticsPage(stockListID int) *ViewStockListStatisticsModel {
	return &ViewStockListStatisticsModel{
		StockListID:   stockListID,
		TimeInterval:  "year",
		IntervalIndex: 3,
		Statistics:    []StockListStatistic{},
		BackPressed:   false,
		Loading:       true,
		Error:         "",
		Selected:      0,
		ScrollOffset:  0,
		ShowMatrix:    false,
	}
}

// Init fetches the statistics data
func (m *ViewStockListStatisticsModel) Init() tea.Cmd {
	return func() tea.Msg {
		db := database.New()
		ctx := context.Background()

		// Get COV values
		covValues, err := db.GetCOVForStockList(ctx, m.StockListID, m.TimeInterval)
		if err != nil {
			return stockListStatisticsErrorMsg{err: err}
		}

		// Get Beta values
		betaValues, err := db.GetBetaForStockList(ctx, m.StockListID, m.TimeInterval)
		if err != nil {
			return stockListStatisticsErrorMsg{err: err}
		}

		// Get all symbols from the stock list
		query := `SELECT symbol FROM hasstockfromstocklist WHERE stocklist_id = $1 ORDER BY symbol`
		rows, err := db.GetDB().Query(query, m.StockListID)
		if err != nil {
			return stockListStatisticsErrorMsg{err: err}
		}
		defer rows.Close()

		var symbols []string
		for rows.Next() {
			var symbol string
			if err := rows.Scan(&symbol); err != nil {
				return stockListStatisticsErrorMsg{err: err}
			}
			symbols = append(symbols, symbol)
		}

		if len(symbols) == 0 {
			return stockListStatisticsErrorMsg{err: fmt.Errorf("stock list has no holdings")}
		}

		// Calculate statistics for each stock
		var statistics []StockListStatistic
		for _, symbol := range symbols {
			statistics = append(statistics, StockListStatistic{
				Symbol: symbol,
				COV:    covValues[symbol],
				Beta:   betaValues[symbol],
			})
		}

		// Sort by symbol
		sort.Slice(statistics, func(i, j int) bool {
			return statistics[i].Symbol < statistics[j].Symbol
		})

		return stockListStatisticsLoadedMsg{statistics: statistics}
	}
}

// Update handles incoming events and updates the model
func (m *ViewStockListStatisticsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case stockListStatisticsLoadedMsg:
		m.Statistics = msg.statistics
		m.Loading = false

	case stockListStatisticsErrorMsg:
		m.Loading = false
		m.Error = fmt.Sprintf("Error loading statistics: %v", msg.err)

	case stockListMatrixLoadedMsg:
		if msg.matrixType == "correlation" {
			m.CorrelationData = msg.data
		} else {
			m.CovarianceData = msg.data
		}
		m.MatrixLoading = false

	case stockListMatrixErrorMsg:
		m.MatrixLoading = false
		m.Error = fmt.Sprintf("Error loading matrix: %v", msg.err)

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+b", "esc":
			m.BackPressed = true

		case "up", "k":
			if !m.ShowMatrix && m.Selected > 0 {
				m.Selected--
			}

		case "down", "j":
			if !m.ShowMatrix && m.Selected < len(m.Statistics)-1 {
				m.Selected++
			}

		case "left", "h":
			if m.IntervalIndex > 0 {
				m.IntervalIndex--
				m.TimeInterval = timeIntervals[m.IntervalIndex]
				m.Loading = true
				m.Statistics = []StockListStatistic{}
				m.ShowMatrix = false
				m.CorrelationData = nil
				m.CovarianceData = nil
				return m, m.Init()
			}

		case "right", "l":
			if m.IntervalIndex < len(timeIntervals)-1 {
				m.IntervalIndex++
				m.TimeInterval = timeIntervals[m.IntervalIndex]
				m.Loading = true
				m.Statistics = []StockListStatistic{}
				m.ShowMatrix = false
				m.CorrelationData = nil
				m.CovarianceData = nil
				return m, m.Init()
			}

		case "c":
			// Toggle correlation matrix view
			if m.ShowMatrix && m.MatrixType == "correlation" {
				m.ShowMatrix = false
			} else if m.CorrelationData != nil && len(m.CorrelationData) > 0 {
				m.ShowMatrix = true
				m.MatrixType = "correlation"
			} else {
				// Load correlation matrix
				m.MatrixType = "correlation"
				m.MatrixLoading = true
				return m, m.loadCorrelationMatrix()
			}

		case "v":
			// Toggle covariance matrix view
			if m.ShowMatrix && m.MatrixType == "covariance" {
				m.ShowMatrix = false
			} else if m.CovarianceData != nil && len(m.CovarianceData) > 0 {
				m.ShowMatrix = true
				m.MatrixType = "covariance"
			} else {
				// Load covariance matrix
				m.MatrixType = "covariance"
				m.MatrixLoading = true
				return m, m.loadCovarianceMatrix()
			}
		}
	}
	return m, nil
}

func (m *ViewStockListStatisticsModel) loadCorrelationMatrix() tea.Cmd {
	return func() tea.Msg {
		db := database.New()
		ctx := context.Background()

		data, err := db.GetCorrelationMatrixForStockList(ctx, m.StockListID, m.TimeInterval)
		if err != nil {
			return stockListMatrixErrorMsg{err: err}
		}

		return stockListMatrixLoadedMsg{matrixType: "correlation", data: data}
	}
}

func (m *ViewStockListStatisticsModel) loadCovarianceMatrix() tea.Cmd {
	return func() tea.Msg {
		db := database.New()
		ctx := context.Background()

		data, err := db.GetCovarianceMatrixForStockList(ctx, m.StockListID, m.TimeInterval)
		if err != nil {
			return stockListMatrixErrorMsg{err: err}
		}

		return stockListMatrixLoadedMsg{matrixType: "covariance", data: data}
	}
}

// View renders the statistics page
func (m *ViewStockListStatisticsModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render("📊 Stock List Statistics") + "\n\n"

	// Display time interval selector
	s += "Select Time Interval:\n"
	for i, interval := range timeIntervals {
		if i == m.IntervalIndex {
			s += styles.SelectedStyle.Render("  → " + interval) + "\n"
		} else {
			s += styles.UnselectedStyle.Render("    " + interval) + "\n"
		}
	}
	s += "\n"

	if m.Loading {
		s += "Loading statistics...\n"
	} else if m.Error != "" {
		s += styles.ErrorStyle.Render(m.Error) + "\n"
	} else if m.ShowMatrix {
		// Display matrix
		s += m.renderMatrix()
	} else if len(m.Statistics) == 0 {
		s += "No statistics available for this stock list.\n"
	} else {
		// Display statistics table
		s += styles.HeaderStyle.Render(fmt.Sprintf("%-10s %-15s %-15s", "Symbol", "COV", "Beta")) + "\n"
		s += "────────────────────────────────────────────────────\n"

		// Adjust scrollOffset to keep selected item in view
		const viewportHeight = 8
		if m.Selected < m.ScrollOffset {
			m.ScrollOffset = m.Selected
		}
		if m.Selected >= m.ScrollOffset+viewportHeight {
			m.ScrollOffset = m.Selected - viewportHeight + 1
		}

		// Ensure scrollOffset doesn't go past the end
		maxScrollOffset := len(m.Statistics) - viewportHeight
		if maxScrollOffset < 0 {
			maxScrollOffset = 0
		}
		if m.ScrollOffset > maxScrollOffset {
			m.ScrollOffset = maxScrollOffset
		}

		// Display visible statistics
		endIdx := m.ScrollOffset + viewportHeight
		if endIdx > len(m.Statistics) {
			endIdx = len(m.Statistics)
		}

		for i := m.ScrollOffset; i < endIdx; i++ {
			stat := m.Statistics[i]
			line := fmt.Sprintf("%-10s %-15.2f %-15.2f", stat.Symbol, stat.COV, stat.Beta)
			if i == m.Selected {
				s += fmt.Sprintf("%s\n", styles.SelectedStyle.Render("→ "+line))
			} else {
				s += fmt.Sprintf("%s\n", styles.UnselectedStyle.Render("  "+line))
			}
		}
	}

	s += "\n"
	footer := "← / → or h/l to change interval • 'c' for correlation matrix • 'v' for covariance matrix • 'Ctrl + b' or 'Esc' to go back"
	s += styles.FooterStyle.Render(footer) + "\n\n"
	return s
}

func (m *ViewStockListStatisticsModel) renderMatrix() string {
	var data map[string]map[string]float64
	matrixName := ""

	if m.MatrixType == "correlation" {
		data = m.CorrelationData
		matrixName = "Correlation Matrix"
	} else {
		data = m.CovarianceData
		matrixName = "Covariance Matrix"
	}

	s := fmt.Sprintf("📋 %s (%s)\n\n", matrixName, m.TimeInterval)

	if m.MatrixLoading {
		s += "Loading matrix...\n"
		return s
	}

	if len(data) == 0 {
		s += "No data available.\n"
		return s
	}

	// Get symbols and sort them
	var symbols []string
	for symbol := range data {
		symbols = append(symbols, symbol)
	}
	sort.Strings(symbols)

	// Calculate column widths
	colWidth := 10
	headerLine := fmt.Sprintf("%-12s", "")
	for _, symbol := range symbols {
		headerLine += fmt.Sprintf("%-*s", colWidth, symbol)
	}
	s += headerLine + "\n"
	s += "────────────────────────────────────────────────────────────────\n"

	// Display matrix
	for _, s1 := range symbols {
		line := fmt.Sprintf("%-12s", s1)
		for _, s2 := range symbols {
			value := data[s1][s2]
			line += fmt.Sprintf("%-*.2f", colWidth, value)
		}
		s += line + "\n"
	}

	return s
}
