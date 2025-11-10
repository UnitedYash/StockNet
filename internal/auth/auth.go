package auth

import (
	"fmt"
	"StockNet/internal/database"
)

// First function to login to database VM
type User struct {
	UserID uint32 // I made this uint32 as user IDs are typically non-negative
	Email string
	Password string
	Name string
}

// Login function to authenticate user's credentials
func Login(email string, password string) (*User, error) {
	// first we connect to the database
	dbService := database.New()
	db := dbService.GetDB()
	
	// Prepare the SQL query to fetch user details
	query := "SELECT user_id, email, password, name FROM accounts WHERE email=$1 AND password=$2"
	row := db.QueryRow(query, email, password)

	// Create a User instance to hold the fetched data
	var user User
	err := row.Scan(&user.UserID, &user.Email, &user.Password, &user.Name)
	if err != nil {
		return nil, fmt.Errorf("login failed: Check your email and password")
	}
	
	return &user, nil
}

func Register(email string, password string, name string) (*User, error) {
	// Connect to the database
	dbService := database.New()
	db := dbService.GetDB()

	// First going to make sure that email is not already registered
	var existingUserID uint32
	checkQuery := "SELECT user_id FROM accounts WHERE email=$1"
	err := db.QueryRow(checkQuery, email).Scan(&existingUserID)
	if err == nil {
		return nil, fmt.Errorf("registration failed: email already registered")
	}

	// now insert the newly registered user
	insertQuery := "INSERT INTO accounts (email, password, name) VALUES ($1, $2, $3) RETURNING user_id"
	var newUserID uint32
	err = db.QueryRow(insertQuery, email, password, name).Scan(&newUserID)
	if err != nil {
		return nil, fmt.Errorf("registration failed: %v", err)
	}

	// create instance of a user
	newUser := &User{
		UserID: newUserID,
		Email:   email,
		Password: password,
		Name:    name,
	}

	return newUser, nil
}