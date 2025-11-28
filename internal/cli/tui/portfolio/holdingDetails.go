package portfolio

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/database"
	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/linechart"
	"github.com/charmbracelet/lipgloss"
)

// HoldingHistoricalPrice represents a single historical price point
type HoldingHistoricalPrice struct {
	Date   time.Time
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume int
}

// Model for holding details page with historical chart
type HoldingDetailsModel struct {
	PortfolioID      string
	Symbol           string
	Shares           int
	CurrentPrice     float64
	TimeRange        string                // week, month, quarter, year, 5years
	SelectedRange    int                   // Index of selected time range
	BackPressed      bool
	ViewForecastPressed bool
	loading          bool
	prices           []HoldingHistoricalPrice
	error            string
}

type holdingHistoricalPricesLoadedMsg struct {
	prices []HoldingHistoricalPrice
}

type holdingHistoricalPricesErrorMsg struct {
	err error
}

// NewHoldingDetailsPage creates a new holding details page
func NewHoldingDetailsPage(portfolioID string, symbol string, shares int, currentPrice float64) *HoldingDetailsModel {
	return &HoldingDetailsModel{
		PortfolioID:   portfolioID,
		Symbol:        symbol,
		Shares:        shares,
		CurrentPrice:  currentPrice,
		SelectedRange: 0,
		loading:       true,
		prices:        []HoldingHistoricalPrice{},
	}
}

// GetTimeRanges returns available time range options
func (m *HoldingDetailsModel) GetTimeRanges() []string {
	return []string{"1 Week", "1 Month", "3 Months", "1 Year", "5 Years"}
}

// GetDaysForRange returns the number of days for the selected range
func (m *HoldingDetailsModel) GetDaysForRange() int {
	ranges := []int{7, 30, 90, 365, 1825}
	if m.SelectedRange < len(ranges) {
		return ranges[m.SelectedRange]
	}
	return 30
}

// Init returns initial command for the holding details page
func (m *HoldingDetailsModel) Init() tea.Cmd {
	return func() tea.Msg {
		// Query historical prices from database
		db := database.New().GetDB()
		days := m.GetDaysForRange()

		// First, get the date range for this symbol
		var maxDate *time.Time
		err := db.QueryRow(`SELECT MAX(timestamp) FROM Stocks WHERE symbol = $1`, m.Symbol).Scan(&maxDate)
		if err != nil {
			return holdingHistoricalPricesErrorMsg{err: err}
		}

		// Check if the symbol exists in the database
		if maxDate == nil {
			return holdingHistoricalPricesErrorMsg{err: fmt.Errorf("no data found for symbol: %s", m.Symbol)}
		}

		// Calculate the start date based on the max date minus days
		startDate := maxDate.AddDate(0, 0, -days)

		query := `SELECT timestamp, close
		          FROM Stocks
		          WHERE symbol = $1 AND timestamp >= $2 AND close IS NOT NULL
		          ORDER BY timestamp ASC`

		rows, err := db.Query(query, m.Symbol, startDate)
		if err != nil {
			return holdingHistoricalPricesErrorMsg{err: err}
		}
		defer rows.Close()

		var prices []HoldingHistoricalPrice
		for rows.Next() {
			var p HoldingHistoricalPrice
			if err := rows.Scan(&p.Date, &p.Close); err != nil {
				return holdingHistoricalPricesErrorMsg{err: err}
			}
			prices = append(prices, p)
		}

		return holdingHistoricalPricesLoadedMsg{prices: prices}
	}
}

// Update handles incoming events and updates the model
func (m *HoldingDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case holdingHistoricalPricesLoadedMsg:
		m.prices = msg.prices
		m.loading = false
		m.error = ""
	case holdingHistoricalPricesErrorMsg:
		m.loading = false
		m.error = fmt.Sprintf("Error loading prices: %v", msg.err)
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if m.SelectedRange > 0 {
				m.SelectedRange--
				m.loading = true
				return m, m.Init()
			}
		case "right", "l":
			ranges := m.GetTimeRanges()
			if m.SelectedRange < len(ranges)-1 {
				m.SelectedRange++
				m.loading = true
				return m, m.Init()
			}
		case "f":
			// View forecast
			m.ViewForecastPressed = true
		case "ctrl+b", "esc":
			m.BackPressed = true
		}
	}
	return m, nil
}

// View renders the holding details page with ASCII chart
func (m *HoldingDetailsModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render(fmt.Sprintf("📊 %s Holdings", m.Symbol)) + "\n\n"

	// Display holding information
	holdingValue := float64(m.Shares) * m.CurrentPrice
	s += fmt.Sprintf("Shares:        %d\n", m.Shares)
	s += fmt.Sprintf("Current Price: $%.2f\n", m.CurrentPrice)
	s += styles.SuccessStyle.Render(fmt.Sprintf("Holding Value: $%.2f\n", holdingValue)) + "\n"

	ranges := m.GetTimeRanges()
	s += "Select Time Range:\n"
	for i, r := range ranges {
		if i == m.SelectedRange {
			s += styles.SelectedStyle.Render("  → " + r) + "\n"
		} else {
			s += styles.UnselectedStyle.Render("    " + r) + "\n"
		}
	}
	s += "\n"

	if m.loading {
		s += "Loading historical prices...\n"
	} else if m.error != "" {
		s += styles.ErrorStyle.Render(m.error) + "\n"
	} else if len(m.prices) == 0 {
		s += "No historical data available for this stock.\n"
	} else {
		// Generate ASCII chart
		chart := m.generateChart()
		s += chart
		s += "\n"

		// Display summary statistics
		minPrice, maxPrice, latestPrice := m.calculateStats()
		change := ((latestPrice - minPrice) / minPrice) * 100

		s += fmt.Sprintf("Low:    $%.2f\n", minPrice)
		s += fmt.Sprintf("High:   $%.2f\n", maxPrice)
		s += fmt.Sprintf("Latest: $%.2f\n", latestPrice)
		if change >= 0 {
			s += styles.SuccessStyle.Render(fmt.Sprintf("Change: +%.2f%%\n", change))
		} else {
			s += styles.ErrorStyle.Render(fmt.Sprintf("Change: %.2f%%\n", change))
		}
	}

	s += "\n"
	s += styles.FooterStyle.Render("← / → or h/l to change range • 'f' for forecast • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}

// generateChart creates a line chart using ntcharts library
func (m *HoldingDetailsModel) generateChart() string {
	if len(m.prices) == 0 {
		return "No data to display"
	}

	// Find min and max prices for scaling
	minPrice := m.prices[0].Close
	maxPrice := m.prices[0].Close
	for _, p := range m.prices {
		if p.Close < minPrice {
			minPrice = p.Close
		}
		if p.Close > maxPrice {
			maxPrice = p.Close
		}
	}

	// Add padding to y-axis range
	padding := (maxPrice - minPrice) * 0.1
	if padding == 0 {
		padding = 1
	}
	minPrice -= padding
	maxPrice += padding

	// Create line chart (width: 80, height: 14)
	axisStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	lineStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("4"))

	lc := linechart.New(
		80, 14,
		0.0, float64(len(m.prices)-1),
		minPrice, maxPrice,
		linechart.WithXYSteps(5, 2),
		linechart.WithStyles(axisStyle, labelStyle, lineStyle),
	)

	// Draw axes and labels
	lc.DrawXYAxisAndLabel()

	// Draw the price line using braille characters for smooth visualization
	if len(m.prices) > 1 {
		for i := 0; i < len(m.prices)-1; i++ {
			point1 := canvas.Float64Point{X: float64(i), Y: m.prices[i].Close}
			point2 := canvas.Float64Point{X: float64(i + 1), Y: m.prices[i+1].Close}
			lc.DrawBrailleLineWithStyle(point1, point2, lineStyle)
		}
	} else {
		// Single point case
		point := canvas.Float64Point{X: 0, Y: m.prices[0].Close}
		lc.DrawBrailleLineWithStyle(point, point, lineStyle)
	}

	// Get date range and render
	startDate := m.prices[0].Date.Format("2006-01-02")
	endDate := m.prices[len(m.prices)-1].Date.Format("2006-01-02")

	output := lc.View()
	output += fmt.Sprintf("Date Range: %s to %s\n", startDate, endDate)

	return output
}

// calculateStats returns min, max, and latest prices
func (m *HoldingDetailsModel) calculateStats() (float64, float64, float64) {
	if len(m.prices) == 0 {
		return 0, 0, 0
	}

	minPrice := m.prices[0].Close
	maxPrice := m.prices[0].Close
	for _, p := range m.prices {
		if p.Close < minPrice {
			minPrice = p.Close
		}
		if p.Close > maxPrice {
			maxPrice = p.Close
		}
	}

	return minPrice, maxPrice, m.prices[len(m.prices)-1].Close
}
