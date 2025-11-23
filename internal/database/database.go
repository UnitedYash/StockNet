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

	// GetHoldings retrieves all stock holdings for a specific portfolio.
	// It returns a slice of holdings with current price information.
	GetHoldings(ctx context.Context, portfolioID int) ([]interface{}, error)
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

// BuyStock executes a stock purchase transaction for a portfolio.
// It:
// 1. Validates the portfolio has enough cash
// 2. Deducts the cost from the portfolio's cash account
// 3. Adds/updates the stock holding in hasStockFromPortfolio
// 4. Records the transaction in the transaction table
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
		`INSERT INTO transaction (portfolio_id, time, amount, buy_sell_price, shares_moved, type)
		 VALUES ($1, NOW(), $2, $3, $4, 'BUY')`,
		portfolioID, totalCost, price, quantity,
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

// GetHoldings retrieves all stock holdings for a specific portfolio.
// It queries the hasStockFromPortfolio table and joins with CurrentPrices
// to get current price information for each holding.
//
// TODO: Implement this function to:
// 1. Query the hasStockFromPortfolio table for the portfolio
// 2. Join with CurrentPrices to get current price and timestamp
// 3. Return holdings ordered by symbol
//
// Expected return format: Slice of holdings with fields: symbol, shares, price, timestamp
func (s *service) GetHoldings(ctx context.Context, portfolioID int) ([]interface{}, error) {
	// TODO: Implement GetHoldings
	// You may need to define a Holding type in the stock package
	// Query structure:
	// SELECT hsp.symbol, hsp.shares, cp.price, cp.timestamp
	// FROM hasStockFromPortfolio hsp
	// JOIN CurrentPrices cp ON hsp.symbol = cp.symbol
	// WHERE hsp.portfolio_id = $1
	// ORDER BY hsp.symbol
	return nil, fmt.Errorf("GetHoldings not yet implemented")
}
