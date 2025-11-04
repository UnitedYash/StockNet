# Setup Guide for New Collaborators

Welcome to the StockNet project! This guide will help you get set up and running after cloning the repository.

---

## Prerequisites

Before you start, make sure you have the following installed:

### Required Software
- **Go 1.25.3 or later** - [Download here](https://golang.org/dl/)
- **Git** - [Download here](https://git-scm.com/downloads)
- **psql** (PostgreSQL client) - For testing database connections
  - macOS: `brew install libpq` then add to PATH
  - Linux: `sudo apt install postgresql-client`
  - Windows: Download from [PostgreSQL website](https://www.postgresql.org/download/)

### Optional (for local development)
- **Docker Desktop** - For running PostgreSQL locally (optional)

---

## Step 1: Clone the Repository

```bash
git clone <repository-url>
cd StockNet
```

---

## Step 2: Install Go Dependencies

```bash
go mod download
```

This downloads all the required packages including:
- `github.com/jackc/pgx/v5` - PostgreSQL driver
- `github.com/joho/godotenv` - Environment variable management
- `github.com/testcontainers/testcontainers-go` - For integration tests

---

## Step 3: Configure Environment Variables

### Create Your `.env` File

The `.env` file contains sensitive configuration (database credentials, etc.) and is **not committed to git**.

**Copy the example and fill in the values:**

```bash
# Create .env file
cat > .env << 'EOF'
# GCP VM PostgreSQL Configuration
BLUEPRINT_DB_HOST=YOUR_VM_EXTERNAL_IP
BLUEPRINT_DB_PORT=5432
BLUEPRINT_DB_DATABASE=postgres
BLUEPRINT_DB_USERNAME=postgres
BLUEPRINT_DB_PASSWORD=YOUR_PASSWORD
BLUEPRINT_DB_SCHEMA=public

# Application Configuration
PORT=8080
APP_ENV=development
EOF
```

### Get the Configuration Values

**Ask your teammate for:**
1. **GCP VM External IP** (e.g., `0000.22`)
2. **PostgreSQL password** (set during GCP setup)

**Update your `.env` file with these values:**
```bash
BLUEPRINT_DB_HOST=000       # ← Replace with actual IP
BLUEPRINT_DB_PASSWORD=pass          # ← Replace with actual password
```

---

## Step 4: Test Database Connection

Before running the application, verify you can connect to the database:

### Test with psql

```bash
psql -h YOUR_VM_EXTERNAL_IP -U postgres -d postgres
```

When prompted, enter the password.

**Expected output:**
```
psql (18.0)
Type "help" for help.

postgres=#
```

If you see this, you're connected! Type `\q` to exit.

### Test with Go

```bash
# Make the run script executable
chmod +x run.sh

# Run the example program
./run.sh examples/user_operations.go
```

**Expected output:**
```
=== StockNet User Operations Example ===

1. Creating users table...
✓ Users table created successfully

2. Inserting users...
✓ Created user: ID=1, Username=john_doe, Email=john@example.com
...
=== Example completed successfully! ===
```

---

## Step 5: Run the Application

### Start the API Server

```bash
./run.sh cmd/api/main.go
```

**Expected output:**
```
2025/11/04 12:00:00 Server listening on :8080
```

### Test the Health Endpoint

Open a new terminal and run:

```bash
curl http://localhost:8080/health
```

**Expected response:**
```json
{
  "status": "up",
  "message": "It's healthy",
  "open_connections": "1",
  "in_use": "0",
  "idle": "1"
}
```

If you see this, **your setup is complete!** 

---

## Step 6: Run Tests

The project includes comprehensive integration tests.

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests for a specific package
go test -v ./internal/repository/
```

**Note:** Tests use `testcontainers` to automatically spin up a temporary PostgreSQL instance, so they don't affect the production database.

---

## Project Structure Overview

Here's what you need to know:

```
StockNet/
├── cmd/api/main.go              # Application entry point - START HERE
├── internal/
│   ├── database/                # Database connection management
│   ├── server/                  # HTTP server & routes
│   ├── models/                  # Data structures (User, etc.)
│   └── repository/              # Database queries (CRUD operations)
├── migrations/                  # SQL schema files
├── examples/                    # Example code showing how to use the codebase
├── .env                        # Your local configuration (NOT in git)
├── run.sh                      # Helper script to run with environment variables
├── Makefile                    # Build automation
└── README.md                   # Project documentation
```

**Key files to read:**
1. **[PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md)** - Detailed architecture explanation

---

## Common Tasks

### Running the Application

```bash
# Using the helper script (recommended)
./run.sh cmd/api/main.go

# Or manually with environment variables
export $(cat .env | grep -v '^#' | xargs)
go run cmd/api/main.go
```

### Running Tests

```bash
make test              # Run all tests
make itest            # Run integration tests only
```

### Building the Application

```bash
make build            # Creates binary in current directory
./api                 # Run the built binary
```

### Viewing Database

```bash
# Connect to PostgreSQL
psql -h YOUR_VM_IP -U postgres -d postgres

# Common psql commands:
\dt                   # List all tables
\d users             # Describe users table
SELECT * FROM users; # Query users
\q                   # Quit
```

---

## Troubleshooting

### Issue: "Connection refused" when running the app

**Possible causes:**
1. PostgreSQL is not running on the GCP VM
2. Wrong IP address in `.env`
3. Firewall blocking port 5432

**Solutions:**
```bash
# 1. Check if VM is running (ask teammate or check GCP Console)

# 2. Verify IP in .env matches VM external IP

# 3. Test connection manually:
psql -h YOUR_VM_IP -U postgres -d postgres
```

### Issue: "Password authentication failed"

**Solution:** Check that your `.env` file has the correct password.

### Issue: `.env` not loading / using old values

**Solution:** Use the `run.sh` script instead of `go run` directly:
```bash
./run.sh cmd/api/main.go
```

### Issue: "Database does not exist"

**Solution:** Make sure you're connecting to the `postgres` database, not a custom one:
```bash
BLUEPRINT_DB_DATABASE=postgres  # In your .env file
```

### Issue: Tests failing

**Possible causes:**
1. Docker not running (tests use testcontainers)
2. Port 5432 already in use

**Solutions:**
```bash
# Make sure Docker Desktop is running
docker ps

# If port 5432 is in use, stop local PostgreSQL:
# macOS:
brew services stop postgresql

# Linux:
sudo systemctl stop postgresql
```

---

## Development Workflow

### Daily Workflow

1. **Pull latest changes:**
   ```bash
   git pull origin main
   ```

2. **Create a feature branch:**
   ```bash
   git checkout -b feature/your-feature-name
   ```

3. **Make changes and test:**
   ```bash
   # Run tests
   go test ./...

   # Run the application
   ./run.sh cmd/api/main.go
   ```

4. **Commit and push:**
   ```bash
   git add .
   git commit -m "Add your feature description"
   git push origin feature/your-feature-name
   ```

5. **Create Pull Request** on GitHub/GitLab

### Before Pushing Code

Always run these checks:
```bash
# 1. Format code
go fmt ./...

# 2. Run tests
go test ./...

# 3. Build to check for errors
go build cmd/api/main.go
```

---

## Adding New Features

### Example: Adding a New Table/Model

1. **Create migration file:** `migrations/00X_create_table_name.sql`
   ```sql
   CREATE TABLE table_name (
       id SERIAL PRIMARY KEY,
       field VARCHAR(255) NOT NULL,
       created_at TIMESTAMP DEFAULT NOW()
   );
   ```

2. **Create model:** `internal/models/table_name.go`
   ```go
   package models

   type TableName struct {
       ID        int       `json:"id"`
       Field     string    `json:"field"`
       CreatedAt time.Time `json:"created_at"`
   }
   ```

3. **Create repository:** `internal/repository/table_name_repository.go`
   ```go
   package repository

   // See user_repository.go for example
   ```

4. **Add route handlers:** `internal/server/routes.go`
   ```go
   mux.HandleFunc("/table-name", s.handleTableName)
   ```

5. **Write tests:** `internal/repository/table_name_repository_test.go`

See [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md) for detailed examples.

---

## Important Reminders

### Security
- ✅ **Never commit `.env` file** - It contains passwords!
- ✅ **`.env` is already in `.gitignore`** - Don't remove it
- ⚠️ **Don't share passwords in Slack/Email** - Use a password manager or encrypted channel

### Database
- ⚠️ **We share the same database** - Be careful with destructive operations
- ✅ **Use migrations** for schema changes
- ✅ **Test on local database first** before testing on shared GCP database
- ⚠️ **Don't `DROP TABLE` in production** - Ask teammates first

### Git
- ✅ **Always pull before starting work**
- ✅ **Create feature branches** - Don't commit directly to main
- ✅ **Write descriptive commit messages**
- ✅ **Test before pushing**

---

### Resources

1. **Project Documentation:**
   - [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md) - Architecture overview

2. **Code Examples:**
   - `examples/user_operations.go` - Shows all CRUD operations
   - `internal/repository/user_repository.go` - Repository pattern example
   - `internal/repository/user_repository_test.go` - Testing examples

3. **External Resources:**
   - [Go Documentation](https://go.dev/doc/)
   - [pgx Documentation](https://github.com/jackc/pgx)
   - [PostgreSQL Docs](https://www.postgresql.org/docs/)


## Quick Reference

### Run Commands

```bash
# Run API server
./run.sh cmd/api/main.go

# Run example
./run.sh examples/user_operations.go

# Run tests
go test ./...

# Format code
go fmt ./...

# Build
go build cmd/api/main.go
```

### Database Commands

```bash
# Connect to database
psql -h YOUR_VM_IP -U postgres -d postgres

# Inside psql:
\l                    # List databases
\dt                   # List tables
\d table_name        # Describe table
SELECT * FROM users; # Query
\q                   # Quit
```

### Git Commands

```bash
# Pull latest
git pull origin main

# Create branch
git checkout -b feature/name

# Commit changes
git add .
git commit -m "message"

# Push
git push origin feature/name
```

---

## Checklist for Setup

Use this checklist to verify your setup is complete:

- [ ] Go 1.25.3+ installed (`go version`)
- [ ] Git installed (`git --version`)
- [ ] psql installed (`psql --version`)
- [ ] Repository cloned
- [ ] Dependencies downloaded (`go mod download`)
- [ ] `.env` file created with correct values
- [ ] Database connection tested with psql
- [ ] Example program runs successfully
- [ ] Tests pass (`go test ./...`)
- [ ] API server starts and health check works
- [ ] Read [PROJECT_STRUCTURE.md](PROJECT_STRUCTURE.md)



**Last Updated:** November 4, 2025
**Project:** StockNet - Social Network for Stocks
**Course:** CSCC43 - Introduction to Databases
