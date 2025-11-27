package portfolio

import (
	"context"
	"fmt"
	"math"

	tea "github.com/charmbracelet/bubbletea"
	"StockNet/internal/cli/tui/styles"
	"StockNet/internal/database"
	"github.com/NimbleMarkets/ntcharts/canvas"
	"github.com/NimbleMarkets/ntcharts/linechart"
	"github.com/charmbracelet/lipgloss"
)

// HoldingForecastModel displays future price predictions
type HoldingForecastModel struct {
	Symbol           string
	CurrentPrice     float64
	DaysAhead        int           // Number of days to predict (7, 14, 30)
	SelectedRange    int           // Index of selected days ahead option
	BackPressed      bool
	Loading          bool
	Error            string
	Predictions      []database.PricePrediction
	HistoricalPrice  []HoldingHistoricalPrice
}

type forecastLoadedMsg struct {
	predictions []database.PricePrediction
}

type forecastErrorMsg struct {
	err error
}

// NewHoldingForecastPage creates a new holding forecast page
func NewHoldingForecastPage(symbol string, currentPrice float64) *HoldingForecastModel {
	return &HoldingForecastModel{
		Symbol:        symbol,
		CurrentPrice:  currentPrice,
		SelectedRange: 1, // Default to 14 days
		DaysAhead:     14,
		Loading:       true,
		Predictions:   []database.PricePrediction{},
	}
}

// GetForecastRanges returns available forecast range options
func (m *HoldingForecastModel) GetForecastRanges() []string {
	return []string{"7 Days", "14 Days", "30 Days"}
}

// GetDaysForRange returns the number of days for the selected range
func (m *HoldingForecastModel) GetDaysForRange() int {
	ranges := []int{7, 14, 30}
	if m.SelectedRange < len(ranges) {
		return ranges[m.SelectedRange]
	}
	return 14
}

// Init returns initial command for the holding forecast page
func (m *HoldingForecastModel) Init() tea.Cmd {
	return func() tea.Msg {
		db := database.New()
		daysAhead := m.GetDaysForRange()

		ctx := context.Background()
		predictions, err := db.GetPricePredictions(ctx, m.Symbol, daysAhead)
		if err != nil {
			return forecastErrorMsg{err: err}
		}

		return forecastLoadedMsg{predictions: predictions}
	}
}

// Update handles incoming events and updates the model
func (m *HoldingForecastModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case forecastLoadedMsg:
		m.Predictions = msg.predictions
		m.Loading = false
		m.Error = ""
	case forecastErrorMsg:
		m.Loading = false
		m.Error = fmt.Sprintf("Error loading predictions: %v", msg.err)
	case tea.KeyMsg:
		switch msg.String() {
		case "left", "h":
			if m.SelectedRange > 0 {
				m.SelectedRange--
				m.Loading = true
				return m, m.Init()
			}
		case "right", "l":
			ranges := m.GetForecastRanges()
			if m.SelectedRange < len(ranges)-1 {
				m.SelectedRange++
				m.Loading = true
				return m, m.Init()
			}
		case "ctrl+b", "esc":
			m.BackPressed = true
		}
	}
	return m, nil
}

// View renders the holding forecast page
func (m *HoldingForecastModel) View() string {
	s := "\n"
	s += styles.TitleStyle.Render(fmt.Sprintf("🔮 %s Price Forecast", m.Symbol)) + "\n\n"

	// Display current information
	s += fmt.Sprintf("Current Price: $%.2f\n\n", m.CurrentPrice)

	ranges := m.GetForecastRanges()
	s += "Select Forecast Range:\n"
	for i, r := range ranges {
		if i == m.SelectedRange {
			s += styles.SelectedStyle.Render("  → " + r) + "\n"
		} else {
			s += styles.UnselectedStyle.Render("    " + r) + "\n"
		}
	}
	s += "\n"

	if m.Loading {
		s += "Loading price predictions...\n"
	} else if m.Error != "" {
		s += styles.ErrorStyle.Render(m.Error) + "\n"
	} else if len(m.Predictions) == 0 {
		s += "No predictions available for this stock.\n"
	} else {
		// Generate ASCII chart with predictions
		chart := m.generateForecastChart()
		s += chart
		s += "\n"

		// Display prediction statistics
		lastPrice := m.Predictions[len(m.Predictions)-1].Price
		change := ((lastPrice - m.CurrentPrice) / m.CurrentPrice) * 100

		s += fmt.Sprintf("Current: $%.2f\n", m.CurrentPrice)
		s += fmt.Sprintf("Predicted: $%.2f\n", lastPrice)
		if change >= 0 {
			s += styles.SuccessStyle.Render(fmt.Sprintf("Change: +%.2f%%\n", change))
		} else {
			s += styles.ErrorStyle.Render(fmt.Sprintf("Change: %.2f%%\n", change))
		}
	}

	s += "\n"
	s += styles.FooterStyle.Render("← / → or h/l to change range • 'Ctrl + b' or 'Esc' to go back") + "\n\n"
	return s
}

// generateForecastChart creates a line chart with predictions
func (m *HoldingForecastModel) generateForecastChart() string {
	if len(m.Predictions) == 0 {
		return "No data to display"
	}

	// Find min and max prices from predictions
	minPrice := m.Predictions[0].Price
	maxPrice := m.Predictions[0].Price
	for _, p := range m.Predictions {
		if p.Price < minPrice {
			minPrice = p.Price
		}
		if p.Price > maxPrice {
			maxPrice = p.Price
		}
	}

	// Add padding to y-axis range (20% padding for better visibility)
	padding := (maxPrice - minPrice) * 0.2
	if padding == 0 {
		// If all predictions are the same price, use 10% of that price as padding
		padding = math.Abs(minPrice) * 0.1
		if padding == 0 {
			padding = 1
		}
	}
	minPrice -= padding
	maxPrice += padding

	// Create line chart (width: 80, height: 14)
	axisStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	lineStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // Yellow for predictions

	lc := linechart.New(
		80, 14,
		0.0, float64(len(m.Predictions)),
		minPrice, maxPrice,
		linechart.WithXYSteps(5, 2),
		linechart.WithStyles(axisStyle, labelStyle, lineStyle),
	)

	// Draw axes and labels
	lc.DrawXYAxisAndLabel()

	// Draw the prediction line using braille characters
	if len(m.Predictions) > 1 {
		for i := 0; i < len(m.Predictions)-1; i++ {
			point1 := canvas.Float64Point{X: float64(i), Y: m.Predictions[i].Price}
			point2 := canvas.Float64Point{X: float64(i + 1), Y: m.Predictions[i+1].Price}
			lc.DrawBrailleLineWithStyle(point1, point2, lineStyle)
		}
	} else {
		// Single point case
		point := canvas.Float64Point{X: 0, Y: m.Predictions[0].Price}
		lc.DrawBrailleLineWithStyle(point, point, lineStyle)
	}

	// Get date range and render
	startDate := m.Predictions[0].Date.Format("2006-01-02")
	endDate := m.Predictions[len(m.Predictions)-1].Date.Format("2006-01-02")

	output := lc.View()
	output += fmt.Sprintf("Forecast Range: %s to %s\n", startDate, endDate)

	return output
}

