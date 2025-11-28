package database

import (
	"context"
	"fmt"
)

// PortfolioInfo represents portfolio information
type PortfolioInfo struct {
	PortfolioID int
	UserID      int
	Name        string
	CashAccount float64
}

// PortfolioSummary represents portfolio summary with net worth
type PortfolioSummary struct {
	PortfolioID    int
	Name           string
	CashAccount    float64
	StockValue     float64
	TotalNetWorth  float64
}

// Transaction represents a buy/sell transaction
type TransactionRecord struct {
	TransactionID int
	PortfolioID   int
	Time          string
	Amount        float64
	BuySellPrice  float64
	SharesMoved   int
	Type          string // BUY or SELL
	StockSymbol   string
}

// GetPortfolio retrieves a single portfolio by ID
func (s *service) GetPortfolio(ctx context.Context, portfolioID int) (*PortfolioInfo, error) {
	var portfolio PortfolioInfo
	err := s.db.QueryRowContext(ctx,
		`SELECT portfolio_id, user_id, name, cash_account FROM Portfolio WHERE portfolio_id = $1`,
		portfolioID,
	).Scan(&portfolio.PortfolioID, &portfolio.UserID, &portfolio.Name, &portfolio.CashAccount)

	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio: %w", err)
	}

	return &portfolio, nil
}

// GetUserPortfolios retrieves all portfolios for a user, ordered by ID
func (s *service) GetUserPortfolios(ctx context.Context, userID int) ([]PortfolioInfo, error) {
	query := `SELECT portfolio_id, user_id, name, cash_account FROM Portfolio WHERE user_id = $1 ORDER BY portfolio_id`
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user portfolios: %w", err)
	}
	defer rows.Close()

	var portfolios []PortfolioInfo
	for rows.Next() {
		var portfolio PortfolioInfo
		if err := rows.Scan(&portfolio.PortfolioID, &portfolio.UserID, &portfolio.Name, &portfolio.CashAccount); err != nil {
			return nil, fmt.Errorf("failed to scan portfolio: %w", err)
		}
		portfolios = append(portfolios, portfolio)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating portfolios: %w", err)
	}

	return portfolios, nil
}

// CreatePortfolio creates a new portfolio for a user
func (s *service) CreatePortfolio(ctx context.Context, userID int, name string, initialCash float64) (int, error) {
	var portfolioID int
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO Portfolio (user_id, name, cash_account) VALUES ($1, $2, $3) RETURNING portfolio_id`,
		userID, name, initialCash,
	).Scan(&portfolioID)

	if err != nil {
		return 0, fmt.Errorf("failed to create portfolio: %w", err)
	}

	return portfolioID, nil
}

// GetPortfolioNetWorth retrieves portfolio summary including cash and stock value
func (s *service) GetPortfolioNetWorth(ctx context.Context, portfolioID int) (*PortfolioSummary, error) {
	var summary PortfolioSummary
	var stockValue float64

	// Get cash account
	err := s.db.QueryRowContext(ctx,
		`SELECT portfolio_id, name, cash_account FROM Portfolio WHERE portfolio_id = $1`,
		portfolioID,
	).Scan(&summary.PortfolioID, &summary.Name, &summary.CashAccount)

	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio cash: %w", err)
	}

	// Get total stock value
	err = s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(hsp.shares * cp.price), 0) as total_stock_value
		 FROM hasStockFromPortfolio hsp
		 LEFT JOIN CurrentPrices cp ON hsp.symbol = cp.symbol
		 WHERE hsp.portfolio_id = $1`,
		portfolioID,
	).Scan(&stockValue)

	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio stock value: %w", err)
	}

	summary.StockValue = stockValue
	summary.TotalNetWorth = summary.CashAccount + stockValue

	return &summary, nil
}

// GetPortfolioTransactions retrieves all transactions for a portfolio, ordered by time descending
func (s *service) GetPortfolioTransactions(ctx context.Context, portfolioID int) ([]TransactionRecord, error) {
	query := `
	SELECT transaction_id, portfolio_id, time, amount, buy_sell_price, shares_moved, type, stock_symbol
	FROM transaction
	WHERE portfolio_id = $1
	ORDER BY time DESC`

	rows, err := s.db.QueryContext(ctx, query, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio transactions: %w", err)
	}
	defer rows.Close()

	var transactions []TransactionRecord
	for rows.Next() {
		var tx TransactionRecord
		if err := rows.Scan(&tx.TransactionID, &tx.PortfolioID, &tx.Time, &tx.Amount, &tx.BuySellPrice, &tx.SharesMoved, &tx.Type, &tx.StockSymbol); err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}
		transactions = append(transactions, tx)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating transactions: %w", err)
	}

	return transactions, nil
}

// GetPortfolioHoldings retrieves all stock holdings in a portfolio with current prices
func (s *service) GetPortfolioHoldings(ctx context.Context, portfolioID int) ([]Holding, error) {
	query := `
	SELECT hsp.symbol, hsp.shares, cp.price, cp.timestamp
	FROM hasStockFromPortfolio hsp
	LEFT JOIN CurrentPrices cp ON hsp.symbol = cp.symbol
	WHERE hsp.portfolio_id = $1
	ORDER BY hsp.symbol`

	rows, err := s.db.QueryContext(ctx, query, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio holdings: %w", err)
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
