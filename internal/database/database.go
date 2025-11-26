package database

import (
	"context"
	"database/sql"
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

	return nil
}
