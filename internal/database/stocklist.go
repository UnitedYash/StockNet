package database

import (
	"context"
	"database/sql"
	"fmt"
)

// StockListHolding represents a stock holding in a stock list
type StockListHolding struct {
	Symbol string
	Shares int
	Price  float64
}

// StockListInfo represents basic stock list information
type StockListInfo struct {
	StockListID int
	UserID      int
	Name        string
	Visibility  string
}

// GetUserStockLists retrieves all stock lists for a specific user
func (s *service) GetUserStockLists(ctx context.Context, userID int) ([]StockListInfo, error) {
	query := `SELECT stocklist_id, user_id, name, visibility FROM StockList WHERE user_id = $1 ORDER BY name`
	rows, err := s.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stock lists: %w", err)
	}
	defer rows.Close()

	var lists []StockListInfo
	for rows.Next() {
		var list StockListInfo
		if err := rows.Scan(&list.StockListID, &list.UserID, &list.Name, &list.Visibility); err != nil {
			return nil, fmt.Errorf("failed to scan stock list: %w", err)
		}
		lists = append(lists, list)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stock lists: %w", err)
	}

	return lists, nil
}

// GetPublicStockLists retrieves all public stock lists
func (s *service) GetPublicStockLists(ctx context.Context) ([]StockListInfo, error) {
	query := `SELECT stocklist_id, user_id, name, visibility FROM StockList WHERE visibility = 'public' ORDER BY name`
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get public stock lists: %w", err)
	}
	defer rows.Close()

	var lists []StockListInfo
	for rows.Next() {
		var list StockListInfo
		if err := rows.Scan(&list.StockListID, &list.UserID, &list.Name, &list.Visibility); err != nil {
			return nil, fmt.Errorf("failed to scan stock list: %w", err)
		}
		lists = append(lists, list)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stock lists: %w", err)
	}

	return lists, nil
}

// GetSharedStockLists retrieves all stock lists shared with a specific email
func (s *service) GetSharedStockLists(ctx context.Context, email string) ([]StockListInfo, error) {
	query := `
	SELECT s.stocklist_id, s.user_id, s.name, s.visibility
	FROM StockList s
	INNER JOIN SharedStockLists ss ON s.stocklist_id = ss.stocklist_id
	WHERE ss.email = $1
	ORDER BY s.name`
	rows, err := s.db.QueryContext(ctx, query, email)
	if err != nil {
		return nil, fmt.Errorf("failed to get shared stock lists: %w", err)
	}
	defer rows.Close()

	var lists []StockListInfo
	for rows.Next() {
		var list StockListInfo
		if err := rows.Scan(&list.StockListID, &list.UserID, &list.Name, &list.Visibility); err != nil {
			return nil, fmt.Errorf("failed to scan stock list: %w", err)
		}
		lists = append(lists, list)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating stock lists: %w", err)
	}

	return lists, nil
}

// CreateStockList creates a new stock list
func (s *service) CreateStockList(ctx context.Context, userID int, name string, visibility string) (int, error) {
	var stockListID int
	err := s.db.QueryRowContext(ctx,
		`INSERT INTO StockList (user_id, name, visibility) VALUES ($1, $2, $3) RETURNING stocklist_id`,
		userID, name, visibility,
	).Scan(&stockListID)

	if err != nil {
		return 0, fmt.Errorf("failed to create stock list: %w", err)
	}

	return stockListID, nil
}

// DeleteStockList deletes a stock list and all associated data
func (s *service) DeleteStockList(ctx context.Context, stockListID int) error {
	// This cascades to delete holdings due to foreign key constraints
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM StockList WHERE stocklist_id = $1`,
		stockListID,
	)

	if err != nil {
		return fmt.Errorf("failed to delete stock list: %w", err)
	}

	return nil
}

// UpdateStockListVisibility updates the visibility of a stock list (private/public)
func (s *service) UpdateStockListVisibility(ctx context.Context, stockListID int, visibility string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE StockList SET visibility = $1 WHERE stocklist_id = $2`,
		visibility, stockListID,
	)

	if err != nil {
		return fmt.Errorf("failed to update stock list visibility: %w", err)
	}

	return nil
}

// AddStockToStockList adds a stock to a stock list or updates existing quantity
func (s *service) AddStockToStockList(ctx context.Context, stockListID int, symbol string, shares int) error {
	// Check if symbol exists in currentprices
	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM CurrentPrices WHERE symbol = $1)`,
		symbol,
	).Scan(&exists)

	if err != nil {
		return fmt.Errorf("failed to check if symbol exists: %w", err)
	}

	if !exists {
		return fmt.Errorf("symbol %s does not exist in database", symbol)
	}

	// Check if stock already in this stock list
	var alreadyExists bool
	err = s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM hasStockFromStockList WHERE stocklist_id = $1 AND symbol = $2)`,
		stockListID, symbol,
	).Scan(&alreadyExists)

	if err != nil {
		return fmt.Errorf("failed to check existing stock: %w", err)
	}

	if alreadyExists {
		// Update existing
		_, err := s.db.ExecContext(ctx,
			`UPDATE hasStockFromStockList SET shares = $1 WHERE stocklist_id = $2 AND symbol = $3`,
			shares, stockListID, symbol,
		)
		if err != nil {
			return fmt.Errorf("failed to update stock in list: %w", err)
		}
	} else {
		// Insert new
		_, err := s.db.ExecContext(ctx,
			`INSERT INTO hasStockFromStockList (stocklist_id, symbol, shares) VALUES ($1, $2, $3)`,
			stockListID, symbol, shares,
		)
		if err != nil {
			return fmt.Errorf("failed to add stock to list: %w", err)
		}
	}

	return nil
}

// DeleteStockFromStockList removes a stock from a stock list
func (s *service) DeleteStockFromStockList(ctx context.Context, stockListID int, symbol string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM hasStockFromStockList WHERE stocklist_id = $1 AND symbol = $2`,
		stockListID, symbol,
	)

	if err != nil {
		return fmt.Errorf("failed to delete stock from list: %w", err)
	}

	return nil
}

// GetStockListHoldings retrieves all stock holdings for a stock list with current prices
func (s *service) GetStockListHoldings(ctx context.Context, stockListID int) ([]StockListHolding, error) {
	query := `
	SELECT h.symbol, h.shares, COALESCE(cp.price, 0)
	FROM hasStockFromStockList h
	LEFT JOIN CurrentPrices cp ON h.symbol = cp.symbol
	WHERE h.stocklist_id = $1
	ORDER BY h.symbol`

	rows, err := s.db.QueryContext(ctx, query, stockListID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock list holdings: %w", err)
	}
	defer rows.Close()

	var holdings []StockListHolding
	for rows.Next() {
		var holding StockListHolding
		if err := rows.Scan(&holding.Symbol, &holding.Shares, &holding.Price); err != nil {
			return nil, fmt.Errorf("failed to scan holding: %w", err)
		}
		holdings = append(holdings, holding)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating holdings: %w", err)
	}

	return holdings, nil
}

// GetReview retrieves a user's review for a stock list
func (s *service) GetReview(ctx context.Context, userID int, stockListID int) (string, error) {
	var reviewText string
	err := s.db.QueryRowContext(ctx,
		`SELECT text FROM Reviews WHERE user_id = $1 AND stocklist_id = $2`,
		userID, stockListID,
	).Scan(&reviewText)

	if err == sql.ErrNoRows {
		return "", nil // No review exists
	}

	if err != nil {
		return "", fmt.Errorf("failed to get review: %w", err)
	}

	return reviewText, nil
}

// UpsertReview inserts or updates a user's review for a stock list
func (s *service) UpsertReview(ctx context.Context, userID int, stockListID int, text string) error {
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO Reviews (user_id, stocklist_id, text)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (user_id, stocklist_id)
		 DO UPDATE SET text = $3`,
		userID, stockListID, text,
	)

	if err != nil {
		return fmt.Errorf("failed to upsert review: %w", err)
	}

	return nil
}

// ShareStockList shares a stock list with a user by email
func (s *service) ShareStockList(ctx context.Context, stockListID int, email string) error {
	// Check if email exists
	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM Accounts WHERE email = $1)`,
		email,
	).Scan(&exists)

	if err != nil {
		return fmt.Errorf("failed to check email existence: %w", err)
	}

	if !exists {
		return fmt.Errorf("email %s does not exist in database", email)
	}

	// Check if already shared
	var alreadyShared bool
	err = s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM SharedStockLists WHERE stocklist_id = $1 AND email = $2)`,
		stockListID, email,
	).Scan(&alreadyShared)

	if err != nil {
		return fmt.Errorf("failed to check share status: %w", err)
	}

	if alreadyShared {
		return fmt.Errorf("stock list is already shared with %s", email)
	}

	// Insert share relationship
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO SharedStockLists (stocklist_id, email) VALUES ($1, $2)`,
		stockListID, email,
	)

	if err != nil {
		return fmt.Errorf("failed to share stock list: %w", err)
	}

	return nil
}

// UnshareStockList removes a stock list share with a user by email
func (s *service) UnshareStockList(ctx context.Context, stockListID int, email string) error {
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM SharedStockLists WHERE stocklist_id = $1 AND email = $2`,
		stockListID, email,
	)

	if err != nil {
		return fmt.Errorf("failed to unshare stock list: %w", err)
	}

	return nil
}

// CheckEmailExists checks if an email exists in the accounts table
func (s *service) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	var exists bool
	err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM Accounts WHERE email = $1)`,
		email,
	).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("failed to check email: %w", err)
	}

	return exists, nil
}

// GetStockListSymbols retrieves all stock symbols in a stock list
func (s *service) GetStockListSymbols(ctx context.Context, stockListID int) ([]string, error) {
	query := `SELECT symbol FROM hasStockFromStockList WHERE stocklist_id = $1 ORDER BY symbol`
	rows, err := s.db.QueryContext(ctx, query, stockListID)
	if err != nil {
		return nil, fmt.Errorf("failed to get stock list symbols: %w", err)
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
