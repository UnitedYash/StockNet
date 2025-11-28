package database

import (
	"context"
	"fmt"
)

// StockPrice represents a stock's price data
type StockPrice struct {
	Symbol    string
	Price     float64
	Timestamp string
}

// StockHistoricalData represents historical OHLCV data for a stock
type StockHistoricalData struct {
	Timestamp string
	Open      float64
	High      float64
	Low       float64
	Close     float64
	Volume    int
}

// Holding represents a user's stock holding in a portfolio
type Holding struct {
	Symbol    string
	Shares    int
	Price     float64
	Timestamp string
}

// GetCurrentPrices retrieves all current stock prices, ordered by symbol
func (s *service) GetCurrentPrices(ctx context.Context) ([]StockPrice, error) {
	query := `SELECT symbol, price, timestamp FROM CurrentPrices ORDER BY symbol`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get current prices: %w", err)
	}
	defer rows.Close()

	var prices []StockPrice
	for rows.Next() {
		var price StockPrice
		if err := rows.Scan(&price.Symbol, &price.Price, &price.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan stock price: %w", err)
		}
		prices = append(prices, price)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stock prices: %w", err)
	}

	return prices, nil
}

// GetStockHistoricalData retrieves historical OHLCV data for a stock from a start date onwards
func (s *service) GetStockHistoricalData(ctx context.Context, symbol string, startDate string) ([]StockHistoricalData, error) {
	query := `
	SELECT timestamp, open, high, low, close, volume
	FROM Stocks
	WHERE symbol = $1 AND timestamp >= $2
	ORDER BY timestamp ASC`

	rows, err := s.db.QueryContext(ctx, query, symbol, startDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock historical data: %w", err)
	}
	defer rows.Close()

	var historicalData []StockHistoricalData
	for rows.Next() {
		var data StockHistoricalData
		if err := rows.Scan(&data.Timestamp, &data.Open, &data.High, &data.Low, &data.Close, &data.Volume); err != nil {
			return nil, fmt.Errorf("failed to scan historical data: %w", err)
		}
		historicalData = append(historicalData, data)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating historical data: %w", err)
	}

	return historicalData, nil
}

// GetLatestDateForStock retrieves the most recent trading date for a stock
func (s *service) GetLatestDateForStock(ctx context.Context, symbol string) (string, error) {
	var latestDate string
	err := s.db.QueryRowContext(ctx,
		`SELECT MAX(timestamp) FROM Stocks WHERE symbol = $1`,
		symbol,
	).Scan(&latestDate)

	if err != nil {
		return "", fmt.Errorf("failed to get latest date for stock: %w", err)
	}

	return latestDate, nil
}

// GetStocksFromPortfolio retrieves all stock holdings in a portfolio with current prices
func (s *service) GetStocksFromPortfolio(ctx context.Context, portfolioID int) ([]Holding, error) {
	query := `
	SELECT hsp.symbol, hsp.shares, cp.price, cp.timestamp
	FROM hasStockFromPortfolio hsp
	LEFT JOIN CurrentPrices cp ON hsp.symbol = cp.symbol
	WHERE hsp.portfolio_id = $1
	ORDER BY hsp.symbol`

	rows, err := s.db.QueryContext(ctx, query, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stocks from portfolio: %w", err)
	}
	defer rows.Close()

	var holdings []Holding
	for rows.Next() {
		var holding Holding
		if err := rows.Scan(&holding.Symbol, &holding.Shares, &holding.Price, &holding.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan holding: %w", err)
		}
		holdings = append(holdings, holding)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating holdings: %w", err)
	}

	return holdings, nil
}

// GetAllStockSymbols retrieves all distinct stock symbols in the database
func (s *service) GetAllStockSymbols(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT symbol FROM Stocks ORDER BY symbol`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock symbols: %w", err)
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			return nil, fmt.Errorf("failed to scan symbol: %w", err)
		}
		symbols = append(symbols, symbol)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating symbols: %w", err)
	}

	return symbols, nil
}
