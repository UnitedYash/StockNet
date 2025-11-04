package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"StockNet/internal/database"
	"StockNet/internal/models"
	"StockNet/internal/repository"
)

func main() {
	fmt.Println("=== StockNet User Operations Example ===\n")

	// Initialize database connection
	dbService := database.New()
	defer dbService.Close()

	db := dbService.GetDB()

	// Create the users table
	fmt.Println("1. Creating users table...")
	if err := createUsersTable(db); err != nil {
		log.Fatalf("Failed to create users table: %v", err)
	}
	fmt.Println("✓ Users table created successfully\n")

	// Initialize user repository
	userRepo := repository.NewUserRepository(db)
	ctx := context.Background()

	// Create users
	fmt.Println("2. Inserting users...")
	users := []models.CreateUserRequest{
		{Username: "john_doe", Email: "john@example.com", FullName: "John Doe"},
		{Username: "jane_smith", Email: "jane@example.com", FullName: "Jane Smith"},
		{Username: "bob_wilson", Email: "bob@example.com", FullName: "Bob Wilson"},
		{Username: "alice_johnson", Email: "alice@example.com", FullName: "Alice Johnson"},
		{Username: "charlie_brown", Email: "charlie@example.com", FullName: "Charlie Brown"},
	}

	for _, userReq := range users {
		user, err := userRepo.Create(ctx, userReq)
		if err != nil {
			log.Printf("Error creating user %s: %v", userReq.Username, err)
			continue
		}
		fmt.Printf("✓ Created user: ID=%d, Username=%s, Email=%s\n", user.ID, user.Username, user.Email)
	}
	fmt.Println()

	// Get all users
	fmt.Println("3. Fetching all users...")
	allUsers, err := userRepo.GetAll(ctx)
	if err != nil {
		log.Fatalf("Failed to get all users: %v", err)
	}
	fmt.Printf("✓ Total users in database: %d\n\n", len(allUsers))

	// Display all users
	fmt.Println("4. Displaying all users:")
	fmt.Println("---------------------------------------------------")
	for _, user := range allUsers {
		fmt.Printf("ID: %d | Username: %-15s | Email: %-25s | Name: %s\n",
			user.ID, user.Username, user.Email, user.FullName)
	}
	fmt.Println("---------------------------------------------------\n")

	// Get user by ID
	if len(allUsers) > 0 {
		firstUserID := allUsers[0].ID
		fmt.Printf("5. Fetching user by ID (%d)...\n", firstUserID)
		user, err := userRepo.GetByID(ctx, firstUserID)
		if err != nil {
			log.Printf("Error getting user by ID: %v", err)
		} else {
			fmt.Printf("✓ Found user: %s (%s)\n\n", user.Username, user.Email)
		}
	}

	// Get user by email
	fmt.Println("6. Fetching user by email (john@example.com)...")
	user, err := userRepo.GetByEmail(ctx, "john@example.com")
	if err != nil {
		log.Printf("Error getting user by email: %v", err)
	} else {
		fmt.Printf("✓ Found user: %s (ID: %d)\n\n", user.Username, user.ID)
	}

	// Update user
	if len(allUsers) > 0 {
		firstUserID := allUsers[0].ID
		fmt.Printf("7. Updating user with ID %d...\n", firstUserID)
		updateReq := models.CreateUserRequest{
			Username: "updated_user",
			Email:    "updated@example.com",
			FullName: "Updated User Name",
		}
		updatedUser, err := userRepo.Update(ctx, firstUserID, updateReq)
		if err != nil {
			log.Printf("Error updating user: %v", err)
		} else {
			fmt.Printf("✓ Updated user: %s -> %s\n\n", allUsers[0].Username, updatedUser.Username)
		}
	}

	// Delete user
	if len(allUsers) > 1 {
		lastUserID := allUsers[len(allUsers)-1].ID
		fmt.Printf("8. Deleting user with ID %d...\n", lastUserID)
		err := userRepo.Delete(ctx, lastUserID)
		if err != nil {
			log.Printf("Error deleting user: %v", err)
		} else {
			fmt.Printf("✓ User deleted successfully\n\n")
		}
	}

	// Get final count
	fmt.Println("9. Final user count...")
	finalUsers, err := userRepo.GetAll(ctx)
	if err != nil {
		log.Fatalf("Failed to get final user count: %v", err)
	}
	fmt.Printf("✓ Total users remaining: %d\n\n", len(finalUsers))

	fmt.Println("=== Example completed successfully! ===")
}

func createUsersTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			username VARCHAR(100) NOT NULL UNIQUE,
			email VARCHAR(255) NOT NULL UNIQUE,
			full_name VARCHAR(255),
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		);
		CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
	`
	_, err := db.Exec(query)
	return err
}