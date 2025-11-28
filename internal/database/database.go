package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "github.com/joho/godotenv/autoload"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error

	// GetDB returns the underlying database connection.
	GetDB() *sql.DB

	// BuyStock executes a stock purchase transaction for a portfolio.
	BuyStock(ctx context.Context, portfolioID int, symbol string, quantity int, price float64) error

	// SellStock executes a stock sale transaction for a portfolio.
	SellStock(ctx context.Context, portfolioID int, symbol string, quantity int, price float64) error

	// DepositCash adds cash to a portfolio.
	DepositCash(ctx context.Context, portfolioID int, amount float64) error

	// WithdrawCash removes cash from a portfolio.
	WithdrawCash(ctx context.Context, portfolioID int, amount float64) error

	// GetHoldings retrieves all stock holdings for a specific portfolio.
	// It returns a slice of holdings with current price information.
	GetHoldings(ctx context.Context, portfolioID int) ([]interface{}, error)

	// AddStockData inserts stock price data into the stocks table and updates currentPrices if newer.
	AddStockData(ctx context.Context, symbol string, date string, open, high, low, close float64, volume int) error

	// Cache management methods
	GetStatisticCache(ctx context.Context, portfolioID int, statsType, interval, symbol string) (interface{}, error)
	SetStatisticCache(ctx context.Context, portfolioID int, statsType, interval, symbol string, value float64, matrixJSON *string, latestDate time.Time) error
	InvalidatePortfolioCache(ctx context.Context, portfolioID int) error
	InvalidatePortfolioStockCache(ctx context.Context, symbol string) error

	// Portfolio statistics methods
	// Returns coefficient of variation for a specific stock in a portfolio over a time interval
	GetCOV(ctx context.Context, portfolioID int, symbol, timeInterval string) (float64, error)
	// Returns beta (correlation) of a stock with the market (all other stocks) over a time interval
	GetBeta(ctx context.Context, portfolioID int, symbol, timeInterval string) (float64, error)
	// Returns correlation matrix for all stocks in portfolio over a time interval
	GetCorrelationMatrix(ctx context.Context, portfolioID int, timeInterval string) (map[string]map[string]float64, error)
	// Returns covariance matrix for all stocks in portfolio over a time interval
	GetCovarianceMatrix(ctx context.Context, portfolioID int, timeInterval string) (map[string]map[string]float64, error)
	// Returns predicted future prices for a stock using linear regression
	GetPricePredictions(ctx context.Context, symbol string, daysAhead int) ([]PricePrediction, error)

	// Stock list statistics methods
	// Returns coefficient of variation for each stock in a stock list over a time interval
	GetCOVForStockList(ctx context.Context, stockListID int, timeInterval string) (map[string]float64, error)
	// Returns beta for each stock in a stock list (correlation with other stocks in list)
	GetBetaForStockList(ctx context.Context, stockListID int, timeInterval string) (map[string]float64, error)
	// Returns correlation matrix for all stocks in a stock list over a time interval
	GetCorrelationMatrixForStockList(ctx context.Context, stockListID int, timeInterval string) (map[string]map[string]float64, error)
	// Returns covariance matrix for all stocks in a stock list over a time interval
	GetCovarianceMatrixForStockList(ctx context.Context, stockListID int, timeInterval string) (map[string]map[string]float64, error)
}

// PricePrediction represents a predicted price point
type PricePrediction struct {
	Date  time.Time
	Price float64
}

type service struct {
	db *sql.DB
}

var (
	database   = os.Getenv("BLUEPRINT_DB_DATABASE")
	password   = os.Getenv("BLUEPRINT_DB_PASSWORD")
	username   = os.Getenv("BLUEPRINT_DB_USERNAME")
	port       = os.Getenv("BLUEPRINT_DB_PORT")
	host       = os.Getenv("BLUEPRINT_DB_HOST")
	schema     = os.Getenv("BLUEPRINT_DB_SCHEMA")
	dbInstance *service
)

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", username, password, host, port, database, schema)
	db, err := sql.Open("pgx", connStr)
	if err != nil {
		log.Fatal(err)
	}
	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf("db down: %v", err) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", database)
	return s.db.Close()
}

// GetDB returns the underlying database connection.
func (s *service) GetDB() *sql.DB {
	return s.db
}


func (s *service) BuyStock(ctx context.Context, portfolioID int, symbol string, quantity int, price float64) error {
	// Start a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Calculate total cost
	totalCost := price * float64(quantity)

	// Check if portfolio has enough cash
	var cashAccount float64
	err = tx.QueryRowContext(ctx,
		"SELECT cash_account FROM Portfolio WHERE portfolio_id = $1",
		portfolioID,
	).Scan(&cashAccount)
	if err != nil {
		return fmt.Errorf("failed to fetch portfolio: %w", err)
	}

	if cashAccount < totalCost {
		return fmt.Errorf("insufficient funds: need $%.2f but only have $%.2f", totalCost, cashAccount)
	}

	// Deduct from cash account
	_, err = tx.ExecContext(ctx,
		"UPDATE Portfolio SET cash_account = cash_account - $1 WHERE portfolio_id = $2",
		totalCost, portfolioID,
	)
	if err != nil {
		return fmt.Errorf("failed to update portfolio cash: %w", err)
	}

	// Insert or update stock holding
	_, err = tx.ExecContext(ctx,
		`INSERT INTO hasStockFromPortfolio (symbol, portfolio_id, shares)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (symbol, portfolio_id)
		 DO UPDATE SET shares = hasStockFromPortfolio.shares + $3`,
		symbol, portfolioID, quantity,
	)
	if err != nil {
		return fmt.Errorf("failed to update stock holding: %w", err)
	}

	// Record transaction
	_, err = tx.ExecContext(ctx,
		`INSERT INTO transaction (portfolio_id, time, amount, buy_sell_price, shares_moved, type, stock_symbol)
		 VALUES ($1, NOW(), $2, $3, $4, 'BUY', $5)`,
		portfolioID, totalCost, price, quantity, symbol,
	)
	if err != nil {
		return fmt.Errorf("failed to record transaction: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Invalidate cache after successful purchase
	s.InvalidatePortfolioCache(ctx, portfolioID)

	return nil
}

func (s *service) SellStock(ctx context.Context, portfolioID int, symbol string, quantity int, price float64) error {
	// Start a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Calculate total proceeds
	totalProceeds := price * float64(quantity)

	// Check if portfolio has enough shares of this stock
	var shares int
	err = tx.QueryRowContext(ctx,
		"SELECT shares FROM hasStockFromPortfolio WHERE portfolio_id = $1 AND symbol = $2",
		portfolioID, symbol,
	).Scan(&shares)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("portfolio does not own this stock")
		}
		return fmt.Errorf("failed to fetch stock holding: %w", err)
	}

	if shares < quantity {
		return fmt.Errorf("insufficient shares: trying to sell %d but only have %d", quantity, shares)
	}

	// Add to cash account
	_, err = tx.ExecContext(ctx,
		"UPDATE Portfolio SET cash_account = cash_account + $1 WHERE portfolio_id = $2",
		totalProceeds, portfolioID,
	)
	if err != nil {
		return fmt.Errorf("failed to update portfolio cash: %w", err)
	}

	// Update or delete stock holding
	newShares := shares - quantity
	if newShares <= 0 {
		// Delete the holding if no shares remain
		_, err = tx.ExecContext(ctx,
			"DELETE FROM hasStockFromPortfolio WHERE portfolio_id = $1 AND symbol = $2",
			portfolioID, symbol,
		)
	} else {
		// Update the holding
		_, err = tx.ExecContext(ctx,
			"UPDATE hasStockFromPortfolio SET shares = $1 WHERE portfolio_id = $2 AND symbol = $3",
			newShares, portfolioID, symbol,
		)
	}
	if err != nil {
		return fmt.Errorf("failed to update stock holding: %w", err)
	}

	// Record transaction
	_, err = tx.ExecContext(ctx,
		`INSERT INTO transaction (portfolio_id, time, amount, buy_sell_price, shares_moved, type, stock_symbol)
		 VALUES ($1, NOW(), $2, $3, $4, 'SELL', $5)`,
		portfolioID, totalProceeds, price, quantity, symbol,
	)
	if err != nil {
		return fmt.Errorf("failed to record transaction: %w", err)
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Invalidate cache after successful sale
	s.InvalidatePortfolioCache(ctx, portfolioID)

	return nil
}

func (s *service) DepositCash(ctx context.Context, portfolioID int, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("deposit amount must be positive")
	}

	// Update portfolio cash account
	_, err := s.db.ExecContext(ctx,
		"UPDATE Portfolio SET cash_account = cash_account + $1 WHERE portfolio_id = $2",
		amount, portfolioID,
	)
	if err != nil {
		return fmt.Errorf("failed to deposit cash: %w", err)
	}

	return nil
}

func (s *service) WithdrawCash(ctx context.Context, portfolioID int, amount float64) error {
	if amount <= 0 {
		return fmt.Errorf("withdrawal amount must be positive")
	}

	// Check if portfolio has enough cash
	var cashAccount float64
	err := s.db.QueryRowContext(ctx,
		"SELECT cash_account FROM Portfolio WHERE portfolio_id = $1",
		portfolioID,
	).Scan(&cashAccount)
	if err != nil {
		return fmt.Errorf("failed to fetch portfolio: %w", err)
	}

	if cashAccount < amount {
		return fmt.Errorf("insufficient funds: trying to withdraw $%.2f but only have $%.2f", amount, cashAccount)
	}

	// Update portfolio cash account
	_, err = s.db.ExecContext(ctx,
		"UPDATE Portfolio SET cash_account = cash_account - $1 WHERE portfolio_id = $2",
		amount, portfolioID,
	)
	if err != nil {
		return fmt.Errorf("failed to withdraw cash: %w", err)
	}

	return nil
}

func (s *service) GetHoldings(ctx context.Context, portfolioID int) ([]interface{}, error) {
	// placeholder now
	return nil, nil
}

func (s *service) AddStockData(ctx context.Context, symbol string, date string, open, high, low, close float64, volume int) error {
	// Start a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert into stocks table (historical data)
	_, err = tx.ExecContext(ctx,
		`INSERT INTO stocks (symbol, timestamp, open, high, low, close, volume)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (symbol, timestamp) DO UPDATE
		 SET open = $3, high = $4, low = $5, close = $6, volume = $7`,
		symbol, date, open, high, low, close, volume,
	)
	if err != nil {
		return fmt.Errorf("failed to insert stock data: %w", err)
	}

	// Check if this date is more recent than currentPrices
	var currentTimestamp string
	err = tx.QueryRowContext(ctx,
		"SELECT timestamp FROM CurrentPrices WHERE symbol = $1",
		symbol,
	).Scan(&currentTimestamp)

	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to fetch current price timestamp: %w", err)
	}

	// If no existing price or new date is more recent, update currentPrices
	if err == sql.ErrNoRows || date > currentTimestamp {
		_, err = tx.ExecContext(ctx,
			`INSERT INTO CurrentPrices (symbol, price, timestamp)
			 VALUES ($1, $2, $3)
			 ON CONFLICT (symbol) DO UPDATE
			 SET price = $2, timestamp = $3`,
			symbol, close, date,
		)
		if err != nil {
			return fmt.Errorf("failed to update current prices: %w", err)
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Invalidate cache for portfolios holding this stock
	s.InvalidatePortfolioStockCache(ctx, symbol)

	return nil
}

// GetStatisticCache retrieves a cached statistic from the database
func (s *service) GetStatisticCache(ctx context.Context, portfolioID int, statsType, interval, symbol string) (interface{}, error) {
	var value *float64
	var matrixJSON *string

	err := s.db.QueryRowContext(ctx,
		`SELECT value, matrix_json FROM portfolio_statistics_cache
		 WHERE portfolio_id = $1 AND statistic_type = $2 AND time_interval = $3 AND symbol = $4
		 ORDER BY computed_at DESC
		 LIMIT 1`,
		portfolioID, statsType, interval, symbol,
	).Scan(&value, &matrixJSON)

	if err == sql.ErrNoRows {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve cache: %w", err)
	}

	// Return matrix JSON if available, otherwise return scalar value
	if matrixJSON != nil {
		return *matrixJSON, nil
	}
	if value != nil {
		return *value, nil
	}

	return nil, nil
}

// SetStatisticCache stores a computed statistic in the cache
func (s *service) SetStatisticCache(ctx context.Context, portfolioID int, statsType, interval, symbol string, value float64, matrixJSON *string, latestDate time.Time) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO portfolio_statistics_cache
		 (portfolio_id, statistic_type, time_interval, symbol, value, matrix_json, latest_data_date)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 ON CONFLICT (portfolio_id, statistic_type, time_interval, symbol)
		 DO UPDATE SET value = $5, matrix_json = $6, computed_at = CURRENT_TIMESTAMP, latest_data_date = $7`,
		portfolioID, statsType, interval, symbol, value, matrixJSON, latestDate,
	)

	if err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	return nil
}

// InvalidatePortfolioCache removes all cached statistics for a portfolio
func (s *service) InvalidatePortfolioCache(ctx context.Context, portfolioID int) error {
	_, err := s.db.ExecContext(ctx,
		"DELETE FROM portfolio_statistics_cache WHERE portfolio_id = $1",
		portfolioID,
	)

	if err != nil {
		return fmt.Errorf("failed to invalidate cache: %w", err)
	}

	return nil
}

// InvalidatePortfolioStockCache removes cached statistics for portfolios holding a specific stock
func (s *service) InvalidatePortfolioStockCache(ctx context.Context, symbol string) error {
	// Delete cache entries where the symbol appears in stocks_in_portfolio
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM portfolio_statistics_cache
		 WHERE stocks_in_portfolio LIKE '%' || $1 || '%'`,
		symbol,
	)

	if err != nil {
		return fmt.Errorf("failed to invalidate stock cache: %w", err)
	}

	return nil
}

// Helper function to get date range based on time interval
func getDateRangeForInterval(interval string) string {
	switch interval {
	case "week":
		return "7 days"
	case "month":
		return "30 days"
	case "quarter":
		return "90 days"
	case "year":
		return "365 days"
	case "5years":
		return "1825 days"
	default:
		return "365 days" // default to 1 year
	}
}

// Helper function to get the latest date from portfolio stocks
func (s *service) getLatestDateForPortfolio(ctx context.Context, portfolioID int) (time.Time, error) {
	query := `
	SELECT MAX(timestamp) as latest_date
	FROM Stocks
	WHERE symbol IN (SELECT DISTINCT symbol FROM hasStockFromPortfolio WHERE portfolio_id = $1)`

	var latestDate *time.Time
	err := s.db.QueryRowContext(ctx, query, portfolioID).Scan(&latestDate)
	if err != nil && err != sql.ErrNoRows {
		return time.Time{}, fmt.Errorf("failed to get latest date: %w", err)
	}

	// If no data found, return zero time
	if latestDate == nil {
		return time.Time{}, nil
	}

	return *latestDate, nil
}

// getAlignedPricesQuery returns a SQL query template for calculating pairwise metrics
// metric should be the aggregation function like "CORR(c1, c2) as correlation" or "COVAR_POP(c1, c2) as covariance"
func getAlignedPricesQuery(metric string, dateRange string) string {
	return `
	WITH prices1 AS (
	  SELECT timestamp, close FROM Stocks
	  WHERE symbol = $1 AND timestamp >= $3::timestamp - INTERVAL '` + dateRange + `' AND timestamp <= $3::timestamp
	),
	prices2 AS (
	  SELECT timestamp, close FROM Stocks
	  WHERE symbol = $2 AND timestamp >= $3::timestamp - INTERVAL '` + dateRange + `' AND timestamp <= $3::timestamp
	),
	aligned AS (
	  SELECT p1.close as c1, p2.close as c2
	  FROM prices1 p1
	  INNER JOIN prices2 p2 ON p1.timestamp = p2.timestamp
	)
	SELECT ` + metric + `
	FROM aligned`
}

// calculatePairwiseMatrix calculates a pairwise matrix between all pairs of symbols
// queryBuilder is a function that returns the SQL query for calculating a metric between two symbols
func (s *service) calculatePairwiseMatrix(ctx context.Context, symbols []string, latestDate time.Time, queryBuilder func(s1, s2 string) string) (map[string]map[string]float64, error) {
	matrix := make(map[string]map[string]float64)
	for _, s1 := range symbols {
		matrix[s1] = make(map[string]float64)
		for _, s2 := range symbols {
			query := queryBuilder(s1, s2)
			var value float64
			err := s.db.QueryRowContext(ctx, query, s1, s2, latestDate).Scan(&value)
			if err != nil && err != sql.ErrNoRows {
				value = 0
			}
			matrix[s1][s2] = value
		}
	}
	return matrix, nil
}

// GetCOV calculates coefficient of variation for a stock over a time interval
func (s *service) GetCOV(ctx context.Context, portfolioID int, symbol, timeInterval string) (float64, error) {
	// Try to get from cache first
	cached, err := s.GetStatisticCache(ctx, portfolioID, "cov", timeInterval, symbol)
	if err == nil && cached != nil {
		return cached.(float64), nil
	}

	// Get the latest date from portfolio data
	latestDate, err := s.getLatestDateForPortfolio(ctx, portfolioID)
	if err != nil {
		return 0, err
	}

	// If no data available, return 0
	if latestDate.IsZero() {
		return 0, nil
	}

	dateRange := getDateRangeForInterval(timeInterval)

	// Query time-windowed data from the latest date
	query := `
	SELECT CASE
	  WHEN AVG(close) = 0 THEN 0
	  WHEN COUNT(*) < 2 THEN 0
	  ELSE STDDEV_POP(close) / AVG(close)
	END as cov
	FROM Stocks
	WHERE symbol = $1 AND timestamp >= $2::timestamp - INTERVAL '` + dateRange + `' AND timestamp <= $2::timestamp`

	var cov float64
	err = s.db.QueryRowContext(ctx, query, symbol, latestDate).Scan(&cov)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to calculate COV: %w", err)
	}

	// Cache the result
	s.SetStatisticCache(ctx, portfolioID, "cov", timeInterval, symbol, cov, nil, latestDate)

	return cov, nil
}

// GetBeta calculates correlation of a stock with all other stocks in stocks table
func (s *service) GetBeta(ctx context.Context, portfolioID int, symbol, timeInterval string) (float64, error) {
	// Try to get from cache first
	cached, err := s.GetStatisticCache(ctx, portfolioID, "beta", timeInterval, symbol)
	if err == nil && cached != nil {
		return cached.(float64), nil
	}

	// Get the latest date from portfolio data
	latestDate, err := s.getLatestDateForPortfolio(ctx, portfolioID)
	if err != nil {
		return 0, err
	}

	// If no data available, return 0
	if latestDate.IsZero() {
		return 0, nil
	}

	dateRange := getDateRangeForInterval(timeInterval)

	// Get correlation with average of all other stocks in portfolio
	query := `
	WITH portfolio_stocks AS (
	  SELECT DISTINCT symbol FROM hasStockFromPortfolio WHERE portfolio_id = $1
	),
	stock_data AS (
	  SELECT symbol, timestamp, close
	  FROM Stocks
	  WHERE symbol IN (SELECT symbol FROM portfolio_stocks)
	    AND timestamp >= $3::timestamp - INTERVAL '` + dateRange + `' AND timestamp <= $3::timestamp
	),
	target_prices AS (
	  SELECT timestamp, close FROM stock_data WHERE symbol = $2
	),
	other_prices AS (
	  SELECT timestamp, AVG(close) as market_close FROM stock_data
	  WHERE symbol != $2
	  GROUP BY timestamp
	),
	aligned_data AS (
	  SELECT
	    t.close as stock_close,
	    o.market_close
	  FROM target_prices t
	  INNER JOIN other_prices o ON t.timestamp = o.timestamp
	)
	SELECT CASE
	  WHEN STDDEV_POP(market_close) = 0 THEN 0
	  ELSE COVAR_POP(stock_close, market_close) / VARIANCE(market_close)
	END as beta
	FROM aligned_data`

	var beta float64
	err = s.db.QueryRowContext(ctx, query, portfolioID, symbol, latestDate).Scan(&beta)
	if err != nil && err != sql.ErrNoRows {
		return 0, fmt.Errorf("failed to calculate Beta: %w", err)
	}

	// Cache the result
	s.SetStatisticCache(ctx, portfolioID, "beta", timeInterval, symbol, beta, nil, latestDate)

	return beta, nil
}

// GetCorrelationMatrix calculates correlation between all pairs of stocks
func (s *service) GetCorrelationMatrix(ctx context.Context, portfolioID int, timeInterval string) (map[string]map[string]float64, error) {
	// Try to get from cache first
	cached, err := s.GetStatisticCache(ctx, portfolioID, "correlation_matrix", timeInterval, "")
	if err == nil && cached != nil {
		// Cache returns JSON string, unmarshal it
		if jsonStr, ok := cached.(string); ok {
			var matrix map[string]map[string]float64
			if err := json.Unmarshal([]byte(jsonStr), &matrix); err == nil {
				return matrix, nil
			}
		}
	}

	// Get the latest date from portfolio data
	latestDate, err := s.getLatestDateForPortfolio(ctx, portfolioID)
	if err != nil {
		return nil, err
	}

	// If no data available, return empty matrix
	if latestDate.IsZero() {
		return make(map[string]map[string]float64), nil
	}

	dateRange := getDateRangeForInterval(timeInterval)

	// Get all stock symbols in this portfolio
	query := `SELECT symbol FROM hasStockFromPortfolio WHERE portfolio_id = $1`
	rows, err := s.db.QueryContext(ctx, query, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio symbols: %w", err)
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			return nil, err
		}
		symbols = append(symbols, symbol)
	}

	if len(symbols) == 0 {
		return make(map[string]map[string]float64), nil
	}

	// Calculate pairwise correlations using helper
	queryBuilder := func(s1, s2 string) string {
		metric := `CASE
		  WHEN STDDEV_POP(c1) = 0 OR STDDEV_POP(c2) = 0 THEN 0
		  WHEN COUNT(*) < 2 THEN 0
		  ELSE COVAR_POP(c1, c2) / (STDDEV_POP(c1) * STDDEV_POP(c2))
		END as correlation`
		return getAlignedPricesQuery(metric, dateRange)
	}

	matrix, err := s.calculatePairwiseMatrix(ctx, symbols, latestDate, queryBuilder)
	if err != nil {
		return nil, err
	}

	// Set diagonal to 1 (correlation with self is always 1)
	for _, symbol := range symbols {
		matrix[symbol][symbol] = 1.0
	}

	// Convert matrix to JSON for caching
	matrixJSON, _ := json.Marshal(matrix)
	matrixJSONStr := string(matrixJSON)
	s.SetStatisticCache(ctx, portfolioID, "correlation_matrix", timeInterval, "", 0, &matrixJSONStr, latestDate)

	return matrix, nil
}

// GetCovarianceMatrix calculates covariance between all pairs of stocks
func (s *service) GetCovarianceMatrix(ctx context.Context, portfolioID int, timeInterval string) (map[string]map[string]float64, error) {
	// Try to get from cache first
	cached, err := s.GetStatisticCache(ctx, portfolioID, "covariance_matrix", timeInterval, "")
	if err == nil && cached != nil {
		// Cache returns JSON string, unmarshal it
		if jsonStr, ok := cached.(string); ok {
			var matrix map[string]map[string]float64
			if err := json.Unmarshal([]byte(jsonStr), &matrix); err == nil {
				return matrix, nil
			}
		}
	}

	// Get the latest date from portfolio data
	latestDate, err := s.getLatestDateForPortfolio(ctx, portfolioID)
	if err != nil {
		return nil, err
	}

	// If no data available, return empty matrix
	if latestDate.IsZero() {
		return make(map[string]map[string]float64), nil
	}

	dateRange := getDateRangeForInterval(timeInterval)

	// Get all stock symbols in this portfolio
	query := `SELECT symbol FROM hasStockFromPortfolio WHERE portfolio_id = $1`
	rows, err := s.db.QueryContext(ctx, query, portfolioID)
	if err != nil {
		return nil, fmt.Errorf("failed to get portfolio symbols: %w", err)
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			return nil, err
		}
		symbols = append(symbols, symbol)
	}

	if len(symbols) == 0 {
		return make(map[string]map[string]float64), nil
	}

	// Calculate pairwise covariances using helper
	queryBuilder := func(s1, s2 string) string {
		metric := "COVAR_POP(c1, c2) as covariance"
		return getAlignedPricesQuery(metric, dateRange)
	}

	matrix, err := s.calculatePairwiseMatrix(ctx, symbols, latestDate, queryBuilder)
	if err != nil {
		return nil, err
	}

	// Convert matrix to JSON for caching
	matrixJSON, _ := json.Marshal(matrix)
	matrixJSONStr := string(matrixJSON)
	s.SetStatisticCache(ctx, portfolioID, "covariance_matrix", timeInterval, "", 0, &matrixJSONStr, latestDate)

	return matrix, nil
}

// GetPricePredictions uses linear regression to predict future prices
func (s *service) GetPricePredictions(ctx context.Context, symbol string, daysAhead int) ([]PricePrediction, error) {
	// Get the latest date for the symbol
	var latestDate *time.Time
	err := s.db.QueryRowContext(ctx, "SELECT MAX(timestamp) FROM Stocks WHERE symbol = $1", symbol).Scan(&latestDate)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get latest date: %w", err)
	}
	if latestDate == nil {
		return nil, fmt.Errorf("no data found for symbol: %s", symbol)
	}

	// Get historical prices, using last 60 days of data
	startDate := latestDate.AddDate(0, 0, -60)
	query := `
	SELECT timestamp, close FROM Stocks
	WHERE symbol = $1 AND timestamp >= $2 AND timestamp <= $3
	ORDER BY timestamp ASC`

	rows, err := s.db.QueryContext(ctx, query, symbol, startDate, latestDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch historical prices: %w", err)
	}
	defer rows.Close()

	var historicalData []struct {
		date  time.Time
		price float64
	}

	for rows.Next() {
		var date time.Time
		var price float64
		if err := rows.Scan(&date, &price); err != nil {
			return nil, err
		}
		historicalData = append(historicalData, struct {
			date  time.Time
			price float64
		}{date, price})
	}

	// check at least 2 points for lin regression
	if len(historicalData) < 2 {
		return nil, fmt.Errorf("insufficient historical data for prediction")
	}

	// Simple linear regression: y = a + bx
	// Where x is day index, y is price
	n := float64(len(historicalData))
	var sumX, sumY, sumXY, sumX2 float64

	for i, point := range historicalData {
		x := float64(i)
		y := point.price
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}

	// Calculate slope (b) and intercept (a)
	denominator := n*sumX2 - sumX*sumX
	if denominator == 0 {
		return nil, fmt.Errorf("cannot calculate regression: insufficient price variance")
	}
	slope := (n*sumXY - sumX*sumY) / denominator
	intercept := (sumY - slope*sumX) / n

	// Generate predictions for future days
	var predictions []PricePrediction
	lastDate := historicalData[len(historicalData)-1].date
	lastIndex := float64(len(historicalData) - 1)

	for i := 1; i <= daysAhead; i++ {
		futureX := lastIndex + float64(i)
		predictedPrice := intercept + slope*futureX
		futureDate := lastDate.AddDate(0, 0, i)

		predictions = append(predictions, PricePrediction{
			Date:  futureDate,
			Price: predictedPrice,
		})
	}

	return predictions, nil
}

// GetCOVForStockList calculates COV (coefficient of variation) for each stock in a stock list
func (s *service) GetCOVForStockList(ctx context.Context, stockListID int, timeInterval string) (map[string]float64, error) {
	// Get the latest date where all stocks in the stock list have data
	var latestDate *time.Time
	dateQuery := `
	SELECT MIN(max_date) FROM (
		SELECT MAX(s.timestamp) as max_date FROM Stocks s
		WHERE s.symbol IN (SELECT symbol FROM hasstockfromstocklist WHERE stocklist_id = $1)
		GROUP BY s.symbol
	) as stock_dates`
	err := s.db.QueryRowContext(ctx, dateQuery, stockListID).Scan(&latestDate)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get latest date: %w", err)
	}
	if latestDate == nil {
		return nil, fmt.Errorf("no stock data available for this stock list")
	}

	dateRange := getDateRangeForInterval(timeInterval)

	// Get all stock symbols in this stock list
	query := `SELECT symbol FROM hasstockfromstocklist WHERE stocklist_id = $1`
	rows, err := s.db.QueryContext(ctx, query, stockListID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock list symbols: %w", err)
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			return nil, err
		}
		symbols = append(symbols, symbol)
	}

	if len(symbols) == 0 {
		return make(map[string]float64), nil
	}

	// Calculate COV for each symbol using selected time interval from latest date
	covResults := make(map[string]float64)
	for _, symbol := range symbols {
		covQuery := `
		WITH prices AS (
		  SELECT close FROM Stocks
		  WHERE symbol = $1 AND timestamp >= $2::timestamp - INTERVAL '` + dateRange + `' AND timestamp <= $2::timestamp AND close IS NOT NULL
		)
		SELECT
		  STDDEV_POP(close) / AVG(close) as cov
		FROM prices`

		var cov float64
		err := s.db.QueryRowContext(ctx, covQuery, symbol, latestDate).Scan(&cov)
		if err != nil && err != sql.ErrNoRows {
			cov = 0
		}
		covResults[symbol] = cov
	}

	return covResults, nil
}

// GetBetaForStockList calculates Beta for each stock in a stock list (correlation with market)
func (s *service) GetBetaForStockList(ctx context.Context, stockListID int, timeInterval string) (map[string]float64, error) {
	// Get the latest date where all stocks in the stock list have data
	var latestDate *time.Time
	dateQuery := `
	SELECT MIN(max_date) FROM (
		SELECT MAX(s.timestamp) as max_date FROM Stocks s
		WHERE s.symbol IN (SELECT symbol FROM hasstockfromstocklist WHERE stocklist_id = $1)
		GROUP BY s.symbol
	) as stock_dates`
	err := s.db.QueryRowContext(ctx, dateQuery, stockListID).Scan(&latestDate)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get latest date: %w", err)
	}
	if latestDate == nil {
		return nil, fmt.Errorf("no stock data available for this stock list")
	}

	dateRange := getDateRangeForInterval(timeInterval)

	// Get all stock symbols in this stock list
	query := `SELECT symbol FROM hasstockfromstocklist WHERE stocklist_id = $1`
	rows, err := s.db.QueryContext(ctx, query, stockListID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock list symbols: %w", err)
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			return nil, err
		}
		symbols = append(symbols, symbol)
	}

	if len(symbols) == 0 {
		return make(map[string]float64), nil
	}

	betaResults := make(map[string]float64)
	for _, s1 := range symbols {
		// Beta = Cov(stock, market) / Var(market)
		// Market is the average of all stocks in the Stocks table, only for dates within the selected time interval
		betaQuery := `
		WITH target_prices AS (
		  SELECT timestamp, close FROM Stocks
		  WHERE symbol = $1 AND timestamp >= $2::timestamp - INTERVAL '` + dateRange + `' AND timestamp <= $2::timestamp
		),
		market_prices AS (
		  SELECT timestamp, AVG(close) as market_close
		  FROM Stocks
		  WHERE timestamp IN (SELECT timestamp FROM target_prices)
		  GROUP BY timestamp
		),
		aligned_data AS (
		  SELECT
		    t.close as stock_close,
		    m.market_close
		  FROM target_prices t
		  INNER JOIN market_prices m ON t.timestamp = m.timestamp
		)
		SELECT CASE
		  WHEN COUNT(*) = 0 THEN 0
		  WHEN VARIANCE(market_close) = 0 THEN 0
		  ELSE COVAR_POP(stock_close, market_close) / VARIANCE(market_close)
		END as beta
		FROM aligned_data`

		var beta float64
		err := s.db.QueryRowContext(ctx, betaQuery, s1, latestDate).Scan(&beta)
		if err != nil && err != sql.ErrNoRows {
			beta = 0
		}
		betaResults[s1] = beta
	}

	return betaResults, nil
}

// GetCorrelationMatrixForStockList calculates correlation between all pairs of stocks in a stock list
func (s *service) GetCorrelationMatrixForStockList(ctx context.Context, stockListID int, timeInterval string) (map[string]map[string]float64, error) {
	// Get the latest date where all stocks in the stock list have data
	var latestDate *time.Time
	dateQuery := `
	SELECT MIN(max_date) FROM (
		SELECT MAX(s.timestamp) as max_date FROM Stocks s
		WHERE s.symbol IN (SELECT symbol FROM hasstockfromstocklist WHERE stocklist_id = $1)
		GROUP BY s.symbol
	) as stock_dates`
	err := s.db.QueryRowContext(ctx, dateQuery, stockListID).Scan(&latestDate)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get latest date: %w", err)
	}
	if latestDate == nil {
		return nil, fmt.Errorf("no stock data available for this stock list")
	}

	dateRange := getDateRangeForInterval(timeInterval)

	// Get all stock symbols in this stock list
	query := `SELECT symbol FROM hasstockfromstocklist WHERE stocklist_id = $1`
	rows, err := s.db.QueryContext(ctx, query, stockListID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock list symbols: %w", err)
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			return nil, err
		}
		symbols = append(symbols, symbol)
	}

	if len(symbols) == 0 {
		return make(map[string]map[string]float64), nil
	}

	// Calculate pairwise correlations using helper
	queryBuilder := func(s1, s2 string) string {
		metric := "CORR(c1, c2) as correlation"
		return getAlignedPricesQuery(metric, dateRange)
	}

	return s.calculatePairwiseMatrix(ctx, symbols, *latestDate, queryBuilder)
}

// GetCovarianceMatrixForStockList calculates covariance between all pairs of stocks in a stock list
func (s *service) GetCovarianceMatrixForStockList(ctx context.Context, stockListID int, timeInterval string) (map[string]map[string]float64, error) {
	// Get the latest date where all stocks in the stock list have data
	var latestDate *time.Time
	dateQuery := `
	SELECT MIN(max_date) FROM (
		SELECT MAX(s.timestamp) as max_date FROM Stocks s
		WHERE s.symbol IN (SELECT symbol FROM hasstockfromstocklist WHERE stocklist_id = $1)
		GROUP BY s.symbol
	) as stock_dates`
	err := s.db.QueryRowContext(ctx, dateQuery, stockListID).Scan(&latestDate)
	if err != nil && err != sql.ErrNoRows {
		return nil, fmt.Errorf("failed to get latest date: %w", err)
	}
	if latestDate == nil {
		return nil, fmt.Errorf("no stock data available for this stock list")
	}

	dateRange := getDateRangeForInterval(timeInterval)

	// Get all stock symbols in this stock list
	query := `SELECT symbol FROM hasstockfromstocklist WHERE stocklist_id = $1`
	rows, err := s.db.QueryContext(ctx, query, stockListID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock list symbols: %w", err)
	}
	defer rows.Close()

	var symbols []string
	for rows.Next() {
		var symbol string
		if err := rows.Scan(&symbol); err != nil {
			return nil, err
		}
		symbols = append(symbols, symbol)
	}

	if len(symbols) == 0 {
		return make(map[string]map[string]float64), nil
	}

	// Calculate pairwise covariances using helper
	queryBuilder := func(s1, s2 string) string {
		metric := "COVAR_POP(c1, c2) as covariance"
		return getAlignedPricesQuery(metric, dateRange)
	}

	return s.calculatePairwiseMatrix(ctx, symbols, *latestDate, queryBuilder)
}
