package repository

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"StockNet/internal/models"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	// Set up PostgreSQL container
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second)),
	)
	if err != nil {
		panic(fmt.Sprintf("failed to start container: %s", err))
	}

	// Clean up container after tests
	defer func() {
		if err := testcontainers.TerminateContainer(pgContainer); err != nil {
			panic(fmt.Sprintf("failed to terminate container: %s", err))
		}
	}()

	// Get connection string
	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(fmt.Sprintf("failed to get connection string: %s", err))
	}

	// Connect to database
	testDB, err = sql.Open("pgx", connStr)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to database: %s", err))
	}
	defer testDB.Close()

	// Create the users table
	if err := createUsersTable(testDB); err != nil {
		panic(fmt.Sprintf("failed to create table: %s", err))
	}

	// Run tests
	code := m.Run()
	os.Exit(code)
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

func clearUsersTable(t *testing.T) {
	_, err := testDB.Exec("DELETE FROM users")
	if err != nil {
		t.Fatalf("failed to clear users table: %v", err)
	}
}

func TestCreateUser(t *testing.T) {
	clearUsersTable(t)
	repo := NewUserRepository(testDB)
	ctx := context.Background()

	req := models.CreateUserRequest{
		Username: "testuser",
		Email:    "test@example.com",
		FullName: "Test User",
	}

	user, err := repo.Create(ctx, req)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	if user.ID == 0 {
		t.Error("expected user ID to be set")
	}
	if user.Username != req.Username {
		t.Errorf("expected username %s, got %s", req.Username, user.Username)
	}
	if user.Email != req.Email {
		t.Errorf("expected email %s, got %s", req.Email, user.Email)
	}
	if user.FullName != req.FullName {
		t.Errorf("expected full name %s, got %s", req.FullName, user.FullName)
	}
	if user.CreatedAt.IsZero() {
		t.Error("expected created_at to be set")
	}
	if user.UpdatedAt.IsZero() {
		t.Error("expected updated_at to be set")
	}
}

func TestCreateMultipleUsers(t *testing.T) {
	clearUsersTable(t)
	repo := NewUserRepository(testDB)
	ctx := context.Background()

	users := []models.CreateUserRequest{
		{Username: "user1", Email: "user1@example.com", FullName: "User One"},
		{Username: "user2", Email: "user2@example.com", FullName: "User Two"},
		{Username: "user3", Email: "user3@example.com", FullName: "User Three"},
		{Username: "user4", Email: "user4@example.com", FullName: "User Four"},
		{Username: "user5", Email: "user5@example.com", FullName: "User Five"},
	}

	createdUsers := make([]*models.User, 0, len(users))
	for _, req := range users {
		user, err := repo.Create(ctx, req)
		if err != nil {
			t.Fatalf("failed to create user %s: %v", req.Username, err)
		}
		createdUsers = append(createdUsers, user)
	}

	if len(createdUsers) != 5 {
		t.Errorf("expected 5 users, got %d", len(createdUsers))
	}

	// Verify each user was created correctly
	for i, user := range createdUsers {
		if user.Username != users[i].Username {
			t.Errorf("user %d: expected username %s, got %s", i, users[i].Username, user.Username)
		}
		if user.Email != users[i].Email {
			t.Errorf("user %d: expected email %s, got %s", i, users[i].Email, user.Email)
		}
	}
}

func TestGetUserByID(t *testing.T) {
	clearUsersTable(t)
	repo := NewUserRepository(testDB)
	ctx := context.Background()

	// Create a user
	created, err := repo.Create(ctx, models.CreateUserRequest{
		Username: "findme",
		Email:    "findme@example.com",
		FullName: "Find Me",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Retrieve the user by ID
	found, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to get user by ID: %v", err)
	}

	if found.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, found.ID)
	}
	if found.Username != created.Username {
		t.Errorf("expected username %s, got %s", created.Username, found.Username)
	}
}

func TestGetUserByEmail(t *testing.T) {
	clearUsersTable(t)
	repo := NewUserRepository(testDB)
	ctx := context.Background()

	// Create a user
	created, err := repo.Create(ctx, models.CreateUserRequest{
		Username: "emailuser",
		Email:    "unique@example.com",
		FullName: "Email User",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Retrieve the user by email
	found, err := repo.GetByEmail(ctx, "unique@example.com")
	if err != nil {
		t.Fatalf("failed to get user by email: %v", err)
	}

	if found.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, found.ID)
	}
	if found.Email != created.Email {
		t.Errorf("expected email %s, got %s", created.Email, found.Email)
	}
}

func TestGetUserByUsername(t *testing.T) {
	clearUsersTable(t)
	repo := NewUserRepository(testDB)
	ctx := context.Background()

	// Create a user
	created, err := repo.Create(ctx, models.CreateUserRequest{
		Username: "uniqueuser",
		Email:    "uniqueuser@example.com",
		FullName: "Unique User",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Retrieve the user by username
	found, err := repo.GetByUsername(ctx, "uniqueuser")
	if err != nil {
		t.Fatalf("failed to get user by username: %v", err)
	}

	if found.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, found.ID)
	}
	if found.Username != created.Username {
		t.Errorf("expected username %s, got %s", created.Username, found.Username)
	}
}

func TestGetAllUsers(t *testing.T) {
	clearUsersTable(t)
	repo := NewUserRepository(testDB)
	ctx := context.Background()

	// Create multiple users
	userCount := 10
	for i := 1; i <= userCount; i++ {
		_, err := repo.Create(ctx, models.CreateUserRequest{
			Username: fmt.Sprintf("user%d", i),
			Email:    fmt.Sprintf("user%d@example.com", i),
			FullName: fmt.Sprintf("User %d", i),
		})
		if err != nil {
			t.Fatalf("failed to create user %d: %v", i, err)
		}
	}

	// Get all users
	users, err := repo.GetAll(ctx)
	if err != nil {
		t.Fatalf("failed to get all users: %v", err)
	}

	if len(users) != userCount {
		t.Errorf("expected %d users, got %d", userCount, len(users))
	}
}

func TestUpdateUser(t *testing.T) {
	clearUsersTable(t)
	repo := NewUserRepository(testDB)
	ctx := context.Background()

	// Create a user
	created, err := repo.Create(ctx, models.CreateUserRequest{
		Username: "oldname",
		Email:    "old@example.com",
		FullName: "Old Name",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Wait a bit to ensure updated_at will be different
	time.Sleep(10 * time.Millisecond)

	// Update the user
	updateReq := models.CreateUserRequest{
		Username: "newname",
		Email:    "new@example.com",
		FullName: "New Name",
	}
	updated, err := repo.Update(ctx, created.ID, updateReq)
	if err != nil {
		t.Fatalf("failed to update user: %v", err)
	}

	if updated.Username != "newname" {
		t.Errorf("expected username 'newname', got %s", updated.Username)
	}
	if updated.Email != "new@example.com" {
		t.Errorf("expected email 'new@example.com', got %s", updated.Email)
	}
	if updated.FullName != "New Name" {
		t.Errorf("expected full name 'New Name', got %s", updated.FullName)
	}
	if !updated.UpdatedAt.After(created.UpdatedAt) {
		t.Error("expected updated_at to be after created_at")
	}
}

func TestDeleteUser(t *testing.T) {
	clearUsersTable(t)
	repo := NewUserRepository(testDB)
	ctx := context.Background()

	// Create a user
	created, err := repo.Create(ctx, models.CreateUserRequest{
		Username: "deleteme",
		Email:    "deleteme@example.com",
		FullName: "Delete Me",
	})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	// Delete the user
	err = repo.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}

	// Verify the user is deleted
	_, err = repo.GetByID(ctx, created.ID)
	if err == nil {
		t.Error("expected error when getting deleted user, got nil")
	}
}

func TestCreateUserWithDuplicateEmail(t *testing.T) {
	clearUsersTable(t)
	repo := NewUserRepository(testDB)
	ctx := context.Background()

	// Create first user
	_, err := repo.Create(ctx, models.CreateUserRequest{
		Username: "user1",
		Email:    "duplicate@example.com",
		FullName: "User One",
	})
	if err != nil {
		t.Fatalf("failed to create first user: %v", err)
	}

	// Try to create second user with same email
	_, err = repo.Create(ctx, models.CreateUserRequest{
		Username: "user2",
		Email:    "duplicate@example.com",
		FullName: "User Two",
	})
	if err == nil {
		t.Error("expected error when creating user with duplicate email, got nil")
	}
}

func TestCreateUserWithDuplicateUsername(t *testing.T) {
	clearUsersTable(t)
	repo := NewUserRepository(testDB)
	ctx := context.Background()

	// Create first user
	_, err := repo.Create(ctx, models.CreateUserRequest{
		Username: "sameusername",
		Email:    "email1@example.com",
		FullName: "User One",
	})
	if err != nil {
		t.Fatalf("failed to create first user: %v", err)
	}

	// Try to create second user with same username
	_, err = repo.Create(ctx, models.CreateUserRequest{
		Username: "sameusername",
		Email:    "email2@example.com",
		FullName: "User Two",
	})
	if err == nil {
		t.Error("expected error when creating user with duplicate username, got nil")
	}
}

func TestGetNonExistentUser(t *testing.T) {
	clearUsersTable(t)
	repo := NewUserRepository(testDB)
	ctx := context.Background()

	_, err := repo.GetByID(ctx, 99999)
	if err == nil {
		t.Error("expected error when getting non-existent user, got nil")
	}
}

func TestDeleteNonExistentUser(t *testing.T) {
	clearUsersTable(t)
	repo := NewUserRepository(testDB)
	ctx := context.Background()

	err := repo.Delete(ctx, 99999)
	if err == nil {
		t.Error("expected error when deleting non-existent user, got nil")
	}
}