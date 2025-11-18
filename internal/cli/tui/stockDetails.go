package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/database"
	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/linechart"
	"github.com/charmbracelet/lipgloss"
)

// HistoricalPrice represents a single historical price point
type HistoricalPrice struct {
	Date  time.Time
	Open  float64
	High  float64
	Low   float64
	Close float64
	Volume int
}

// Model for stock details page with historical chart
type StockDetailsModel struct {
	symbol        string
	timeRange     string                // week, month, quarter, year, 5years
	selectedRange int                   // Index of selected time range
	confirmed     bool
	backPressed   bool
	loading       bool
	prices        []HistoricalPrice
	error         string
}

type historicalPricesLoadedMsg struct {
	prices []HistoricalPrice
}

type historicalPricesErrorMsg struct {
	err error
}

// returns initial stock details page model
func newStockDetailsPage(symbol string) *StockDetailsModel {
	return &StockDetailsModel{
		symbol:        symbol,
		selectedRange: 0,
		loading:       true,
		prices:        []HistoricalPrice{},
	}
}

// GetTimeRanges returns available time range options
func (m *StockDetailsModel) GetTimeRanges() []string {
	return []string{"1 Week", "1 Month", "3 Months", "1 Year", "5 Years"}
}

// GetDaysForRange returns the number of days for the selected range
func (m *StockDetailsModel) GetDaysForRange() int {
	ranges := []int{7, 30, 90, 365, 1825}
	if m.selectedRange < len(ranges) {
		return ranges[m.selectedRange]
	}
	return 30
}

// returns initial command for the stock details page
func (m *StockDetailsModel) Init() tea.Cmd {
	return func() tea.Msg {
		// Query historical prices from database
		db := database.New().GetDB()
		days := m.GetDaysForRange()

		// First, get the date range for this symbol
		var maxDate *time.Time
		err := db.QueryRow(`SELECT MAX(timestamp) FROM Stocks WHERE symbol = $1`, m.symbol).Scan(&maxDate)
		if err != nil {
			return historicalPricesErrorMsg{err: err}
		}

		// Check if the symbol exists in the database
		if maxDate == nil {
			return historicalPricesErrorMsg{err: fmt.Errorf("no data found for symbol: %s", m.symbol)}
		}

		// Calculate the start date based on the max date minus days
		startDate := maxDate.AddDate(0, 0, -days)

		query := `SELECT timestamp, open, high, low, close, volume
		          FROM Stocks
		          WHERE symbol = $1 AND timestamp >= $2
		          ORDER BY timestamp ASC`

		rows, err := db.Query(query, m.symbol, startDate)
		if err != nil {
			return historicalPricesErrorMsg{err: err}
		}
		defer rows.Close()

		var prices []HistoricalPrice
		for rows.Next() {
			var p HistoricalPrice
			if err := rows.Scan(&p.Date, &p.Open, &p.High, &p.Low, &p.Close, &p.Volume); err != nil {
				return historicalPricesErrorMsg{err: err}
			}
			prices = append(prices, p)
		}

		return historicalPricesLoadedMsg{prices: prices}
	}
}

// Handles incoming events and updates the model
func (m *StockDetailsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case historicalPricesLoadedMsg:
		m.prices = msg.prices
		m.loading = false
		m.error = ""
	case historicalPricesErrorMsg:
		m.loading = false
		m.error = fmt.Sprintf("Error loading prices: %v", msg.err)
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if m.selectedRange > 0 {
				m.selectedRange--
				m.loading = true
				return m, m.Init()
			}
		case "right", "l":
			ranges := m.GetTimeRanges()
			if m.selectedRange < len(ranges)-1 {
				m.selectedRange++
				m.loading = true
				return m, m.Init()
			}
		case "ctrl+b", "esc":
			m.backPressed = true
		}
	}
	return m, nil
}

// renders ASCII chart of historical prices
func (m *StockDetailsModel) View() string {
	s := "\n"
	s += TitleStyle.Render(fmt.Sprintf("📊 %s Historical Prices", m.symbol)) + "\n\n"

	ranges := m.GetTimeRanges()
	s += "Select Time Range:\n"
	for i, r := range ranges {
		if i == m.selectedRange {
			s += SelectedStyle.Render("  → " + r) + "\n"
		} else {
			s += UnselectedStyle.Render("    " + r) + "\n"
		}
	}
	s += "\n"

	if m.loading {
		s += "Loading historical prices...\n"
	} else if m.error != "" {
		s += ErrorStyle.Render(m.error) + "\n"
	} else if len(m.prices) == 0 {
		s += "No historical data available for this stock.\n"
	} else {
		// Generate ASCII chart
		chart := m.generateChart()
		s += chart
		s += "\n"

		// Display summary statistics
		minPrice, maxPrice, currentPrice := m.calculateStats()
		change := ((currentPrice - minPrice) / minPrice) * 100

		s += fmt.Sprintf("Low:    $%.2f\n", minPrice)
		s += fmt.Sprintf("High:   $%.2f\n", maxPrice)
		s += fmt.Sprintf("Latest: $%.2f\n", currentPrice)
		s += SuccessStyle.Render(fmt.Sprintf("Change: %.2f%%\n", change))
	}

	s += "\n"
	s += FooterStyle.Render("← / → or h/l to change range • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}

// generateChart creates a line chart using ntcharts library
func (m *StockDetailsModel) generateChart() string {
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
func (m *StockDetailsModel) calculateStats() (float64, float64, float64) {
	if len(m.prices) == 0 {
		return 0, 0, 0
	}

	minPrice := m.prices[0].Close
	maxPrice := m.prices[0].Close
	currentPrice := m.prices[len(m.prices)-1].Close

	for _, p := range m.prices {
		if p.Close < minPrice {
			minPrice = p.Close
		}
		if p.Close > maxPrice {
			maxPrice = p.Close
		}
	}

	return minPrice, maxPrice, currentPrice
}
