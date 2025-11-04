# User Operations Examples

This directory contains examples demonstrating how to work with PostgreSQL tables in StockNet.

## What's Included

### 1. SQL Migration File
Location: `migrations/001_create_users_table.sql`

This file contains:
- Table schema definition (users table)
- Index creation for better query performance
- Sample data insertion

### 2. User Model
Location: `internal/models/user.go`

Defines the User struct and CreateUserRequest struct for type-safe operations.

### 3. User Repository
Location: `internal/repository/user_repository.go`

Provides full CRUD operations:
- `Create()` - Insert new users
- `GetByID()` - Fetch user by ID
- `GetByEmail()` - Fetch user by email
- `GetByUsername()` - Fetch user by username
- `GetAll()` - Fetch all users
- `Update()` - Update existing user
- `Delete()` - Remove user

### 4. Integration Tests
Location: `internal/repository/user_repository_test.go`

Comprehensive test suite with:
- Table creation
- Multiple user insertion
- CRUD operations testing
- Constraint validation (unique email/username)
- Error handling tests

## Running the Examples

### Option 1: Run the Example Script

1. Start PostgreSQL:
```bash
make docker-run
```

2. Run the example:
```bash
go run examples/user_operations.go
```

This will:
- Create the users table
- Insert 5 sample users
- Demonstrate all CRUD operations
- Show query results

### Option 2: Run Tests

Run all integration tests:
```bash
make test
```

Run only repository tests:
```bash
go test -v ./internal/repository/
```

The tests use testcontainers to automatically spin up a PostgreSQL instance, so you don't need to manually start Docker.

### Option 3: Manual SQL Migration

1. Start PostgreSQL:
```bash
make docker-run
```

2. Connect to the database:
```bash
docker exec -it stocknet-postgres psql -U melkey -d blueprint
```

3. Run the migration:
```sql
\i migrations/001_create_users_table.sql
```

4. Query the data:
```sql
SELECT * FROM users;
```

## Example Output

When you run `user_operations.go`, you'll see:

```
=== StockNet User Operations Example ===

1. Creating users table...
✓ Users table created successfully

2. Inserting users...
✓ Created user: ID=1, Username=john_doe, Email=john@example.com
✓ Created user: ID=2, Username=jane_smith, Email=jane@example.com
✓ Created user: ID=3, Username=bob_wilson, Email=bob@example.com
✓ Created user: ID=4, Username=alice_johnson, Email=alice@example.com
✓ Created user: ID=5, Username=charlie_brown, Email=charlie@example.com

3. Fetching all users...
✓ Total users in database: 5

4. Displaying all users:
---------------------------------------------------
ID: 1 | Username: john_doe        | Email: john@example.com           | Name: John Doe
ID: 2 | Username: jane_smith      | Email: jane@example.com           | Name: Jane Smith
...
---------------------------------------------------

=== Example completed successfully! ===
```

## Database Schema

```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    full_name VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

## Using in Your Code

```go
package main

import (
    "context"
    "StockNet/internal/database"
    "StockNet/internal/models"
    "StockNet/internal/repository"
)

func main() {
    // Initialize database
    dbService := database.New()
    defer dbService.Close()

    // Create repository
    userRepo := repository.NewUserRepository(dbService.GetDB())
    ctx := context.Background()

    // Create user
    user, err := userRepo.Create(ctx, models.CreateUserRequest{
        Username: "newuser",
        Email:    "newuser@example.com",
        FullName: "New User",
    })

    // Get user by ID
    user, err = userRepo.GetByID(ctx, 1)

    // Update user
    user, err = userRepo.Update(ctx, 1, models.CreateUserRequest{
        Username: "updateduser",
        Email:    "updated@example.com",
        FullName: "Updated Name",
    })

    // Delete user
    err = userRepo.Delete(ctx, 1)
}
```

## Next Steps

1. Extend the User model with additional fields (e.g., password hash, role, etc.)
2. Create additional tables for your application
3. Add more complex queries and relationships
4. Implement pagination for GetAll()
5. Add filtering and sorting capabilities