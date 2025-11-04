# StockNet Project Structure Guide

A comprehensive guide to understanding the Go project structure, database connectivity, and how everything works together.

---

## Table of Contents
1. [High-Level Architecture](#high-level-architecture)
2. [Folder Structure](#folder-structure)
3. [Database Driver: pgx vs pq](#database-driver-pgx-vs-pq)
4. [How the Application Works](#how-the-application-works)
5. [Request Flow](#request-flow)
6. [Database Connectivity](#database-connectivity)
7. [Key Concepts](#key-concepts)
8. [Adding New Features](#adding-new-features)

---

## High-Level Architecture

```
┌─────────────┐
│   Client    │
│  (Browser)  │
└──────┬──────┘
       │ HTTP Request
       ↓
┌─────────────────────────────────────────┐
│         cmd/api/main.go                 │
│  (Entry Point - Application Bootstrap)  │
└──────────────────┬──────────────────────┘
                   │
                   ↓
┌─────────────────────────────────────────┐
│      internal/server/server.go          │
│    (HTTP Server Configuration)          │
└──────────────────┬──────────────────────┘
                   │
                   ↓
┌─────────────────────────────────────────┐
│      internal/server/routes.go          │
│  (Route Handlers & Middleware)          │
└──────────────────┬──────────────────────┘
                   │
                   ↓
┌─────────────────────────────────────────┐
│   internal/repository/*.go              │
│  (Data Access Layer - CRUD Operations)  │
└──────────────────┬──────────────────────┘
                   │
                   ↓
┌─────────────────────────────────────────┐
│    internal/database/database.go        │
│   (Database Connection Management)      │
└──────────────────┬──────────────────────┘
                   │
                   ↓
┌─────────────────────────────────────────┐
│         PostgreSQL Database             │
│         (Docker Container)              │
└─────────────────────────────────────────┘
```

---

## Folder Structure

### Root Directory Files

```
StockNet/
├── go.mod              # Go module definition & dependencies
├── go.sum              # Dependency checksums for verification
├── .env                # Environment variables (DB credentials, ports)
├── .gitignore          # Files to ignore in git
├── .air.toml           # Live reload configuration for development
├── docker-compose.yml  # PostgreSQL container setup
├── Makefile            # Build automation commands
└── README.md           # Project documentation
```

**Purpose of Each File:**
- **go.mod**: Defines your module name (`StockNet`) and lists all dependencies with versions
- **go.sum**: Security feature - ensures dependencies haven't been tampered with
- **.env**: Stores configuration (database credentials, ports) - never commit to git
- **docker-compose.yml**: Defines PostgreSQL service configuration
- **Makefile**: Shortcuts for common commands (build, test, run)

---

### `cmd/` Directory - Application Entry Points

```
cmd/
└── api/
    └── main.go         # Main application entry point
```

**Purpose:**
- Contains **executable applications**
- Each subdirectory under `cmd/` is a separate program
- `cmd/api/main.go` is where your application starts

**What main.go Does:**
1. Initializes the database connection
2. Creates the HTTP server
3. Sets up graceful shutdown (clean exit on Ctrl+C)
4. Starts listening for HTTP requests

**Code Flow in main.go:**
```go
func main() {
    // 1. Initialize database
    dbService := database.New()
    defer dbService.Close()

    // 2. Create HTTP server with database
    server := server.NewServer(dbService)

    // 3. Set up graceful shutdown
    go func() {
        // Listen for interrupt signals
    }()

    // 4. Start serving HTTP requests
    server.ListenAndServe()
}
```

---

### `internal/` Directory - Private Application Code

The `internal/` directory is special in Go - code here **cannot be imported by external projects**. This enforces encapsulation.

```
internal/
├── database/           # Database connection management
├── server/             # HTTP server & routing
├── models/             # Data structures (Go structs)
└── repository/         # Database queries (Data Access Layer)
```

---

#### `internal/database/` - Database Connection Layer

```
internal/database/
├── database.go         # Connection management, singleton pattern
└── database_test.go    # Integration tests
```

**Purpose:**
- Manages PostgreSQL connection using **pgx driver**
- Implements singleton pattern (one connection for entire app)
- Provides health check functionality

**Key Components:**

```go
// Service interface - defines what database operations are available
type Service interface {
    Health() map[string]string    // Check database health
    Close() error                 // Close connection
    GetDB() *sql.DB              // Get raw database connection
}

// New() creates or returns existing database connection
func New() Service {
    // Singleton: reuse existing connection if available
    if dbInstance != nil {
        return dbInstance
    }

    // Create new connection
    connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s",
        username, password, host, port, database, schema)
    db, err := sql.Open("pgx", connStr)  // Using pgx driver!

    return &service{db: db}
}
```

**Connection String Breakdown:**
```
postgres://username:password@host:port/database?sslmode=disable&search_path=schema
         └────────┬────────┘ └─┬─┘└┬─┘└───┬───┘  └────────┬─────────┘└────┬────┘
               From .env   localhost 5432  blueprint    No encryption    public
```

---

#### `internal/models/` - Data Structures

```
internal/models/
└── user.go             # User struct definitions
```

**Purpose:**
- Defines **Go structs** that represent your data
- These map to database tables
- Include JSON tags for API responses

**Example:**
```go
// User represents a database row from the users table
type User struct {
    ID        int       `json:"id"`           // Maps to: users.id
    Username  string    `json:"username"`     // Maps to: users.username
    Email     string    `json:"email"`        // Maps to: users.email
    FullName  string    `json:"full_name"`    // Maps to: users.full_name
    CreatedAt time.Time `json:"created_at"`   // Maps to: users.created_at
    UpdatedAt time.Time `json:"updated_at"`   // Maps to: users.updated_at
}

// CreateUserRequest is used when creating new users
// Separate from User because it doesn't have ID/timestamps
type CreateUserRequest struct {
    Username string `json:"username"`
    Email    string `json:"email"`
    FullName string `json:"full_name,omitempty"`  // omitempty = optional field
}
```

**JSON Tags:**
- Define how the struct appears in JSON responses
- `json:"id"` means the field will be called "id" in JSON
- `json:"full_name,omitempty"` means omit if empty

---

#### `internal/repository/` - Data Access Layer

```
internal/repository/
├── user_repository.go      # CRUD operations for users
└── user_repository_test.go # Tests with real database
```

**Purpose:**
- Contains all **database queries**
- Implements business logic for data access
- Separates SQL from HTTP handlers (clean architecture)

**Repository Pattern:**
```go
// Interface defines what operations are available
type UserRepository interface {
    Create(ctx context.Context, req models.CreateUserRequest) (*models.User, error)
    GetByID(ctx context.Context, id int) (*models.User, error)
    GetByEmail(ctx context.Context, email string) (*models.User, error)
    GetAll(ctx context.Context) ([]models.User, error)
    Update(ctx context.Context, id int, req models.CreateUserRequest) (*models.User, error)
    Delete(ctx context.Context, id int) error
}

// Implementation holds the database connection
type userRepository struct {
    db *sql.DB  // PostgreSQL connection
}

// Example: Create a new user
func (r *userRepository) Create(ctx context.Context, req models.CreateUserRequest) (*models.User, error) {
    query := `
        INSERT INTO users (username, email, full_name, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, username, email, full_name, created_at, updated_at
    `

    user := &models.User{}
    err := r.db.QueryRowContext(ctx, query,
        req.Username, req.Email, req.FullName, now, now,
    ).Scan(&user.ID, &user.Username, &user.Email, &user.FullName, &user.CreatedAt, &user.UpdatedAt)

    return user, err
}
```

**Why Separate Repository?**
- **Testability**: Easy to test database logic in isolation
- **Reusability**: Same query logic can be used by multiple handlers
- **Maintainability**: All SQL in one place, not scattered across handlers
- **Swappability**: Easy to change database implementation

---

#### `internal/server/` - HTTP Server Layer

```
internal/server/
├── server.go           # Server configuration & initialization
├── routes.go           # Route handlers (HTTP endpoints)
└── routes_test.go      # HTTP handler tests
```

**server.go - Server Configuration:**
```go
type Server struct {
    port int
    db   database.Service  // Database injected via dependency injection
}

func NewServer(db database.Service) *http.Server {
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"  // Default
    }

    NewServer := &Server{
        port: port,
        db:   db,  // Store database reference
    }

    // Create HTTP server with timeouts
    server := &http.Server{
        Addr:         fmt.Sprintf(":%d", NewServer.port),
        Handler:      NewServer.RegisterRoutes(),  // Set up routes
        IdleTimeout:  time.Minute,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 30 * time.Second,
    }

    return server
}
```

**routes.go - HTTP Endpoints:**
```go
func (s *Server) RegisterRoutes() http.Handler {
    mux := http.NewServeMux()  // Create new router

    // Register endpoints
    mux.HandleFunc("/", s.HelloWorldHandler)      // GET /
    mux.HandleFunc("/health", s.healthHandler)    // GET /health

    // Wrap with middleware (CORS, logging, etc.)
    return s.corsMiddleware(mux)
}

// Example handler
func (s *Server) HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
    resp := make(map[string]string)
    resp["message"] = "Hello World"

    w.Header().Set("Content-Type", "application/json")
    jsonResp, _ := json.Marshal(resp)
    w.Write(jsonResp)
}
```

---

### `migrations/` - Database Schema

```
migrations/
└── 001_create_users_table.sql
```

**Purpose:**
- Contains SQL files to set up database schema
- Version controlled (001, 002, 003...)
- Run manually or via migration tools

---

### `examples/` - Example Code

```
examples/
├── user_operations.go  # Demo of all CRUD operations
└── README.md          # Usage guide
```

**Purpose:**
- Demonstrates how to use the codebase
- Useful for learning and documentation

---

## Database Driver: pgx vs pq

### What Driver Are We Using? **pgx** ✅

**Proof:**
```go
// In go.mod
require github.com/jackc/pgx/v5 v5.7.6

// In database.go
import _ "github.com/jackc/pgx/v5/stdlib"

// Connection string
db, err := sql.Open("pgx", connStr)  // "pgx" driver name
```

### pgx vs pq Comparison

| Feature | pgx (What we're using) | pq (Old driver) |
|---------|------------------------|-----------------|
| **Performance** | ⚡ Faster (native Go) | Slower |
| **Features** | More features (COPY, LISTEN/NOTIFY) | Basic features |
| **Maintenance** | ✅ Actively maintained | ⚠️ In maintenance mode |
| **Type Support** | Better PostgreSQL type support | Limited types |
| **Connection Pooling** | Built-in pgxpool | Must use external |
| **Prepared Statements** | Automatic | Manual |
| **Recommendation** | ✅ **Use this** | ❌ Deprecated |

**Why pgx is better:**
1. **Native Go**: Written in pure Go, no C dependencies
2. **Performance**: 20-40% faster than pq
3. **Active Development**: Regular updates and new features
4. **Better Errors**: More descriptive error messages
5. **PostgreSQL Specific**: Designed specifically for PostgreSQL

**Your project is already using pgx correctly!** 🎉

---

## How the Application Works

### Startup Sequence

```
1. Load environment variables from .env
   ↓
2. main.go starts execution
   ↓
3. database.New() creates PostgreSQL connection
   ↓
4. server.NewServer(db) creates HTTP server with DB
   ↓
5. RegisterRoutes() sets up URL endpoints
   ↓
6. ListenAndServe() starts accepting HTTP requests
   ↓
7. Application runs until interrupted (Ctrl+C)
   ↓
8. Graceful shutdown: close connections, cleanup
```

### Dependency Injection Flow

```
main.go
  └─> Creates database.Service
       └─> Passed to server.NewServer(db)
            └─> Server stores database reference
                 └─> Handlers can access database via s.db
                      └─> Create repository instances
                           └─> Repository executes SQL queries
```

**Why This Pattern?**
- **Testability**: Easy to inject mock database for testing
- **Flexibility**: Can swap implementations without changing code
- **Single Responsibility**: Each layer has one job

---

## Request Flow

Let's trace a request through the application:

### Example: Create a New User

```
1. Client sends HTTP POST to /users
   {
     "username": "johndoe",
     "email": "john@example.com",
     "full_name": "John Doe"
   }
   ↓

2. routes.go receives request
   mux.HandleFunc("/users", s.createUserHandler)
   ↓

3. Handler function in routes.go
   func (s *Server) createUserHandler(w http.ResponseWriter, r *http.Request) {
       // Parse JSON body into CreateUserRequest
       var req models.CreateUserRequest
       json.NewDecoder(r.Body).Decode(&req)

       // Create repository instance
       userRepo := repository.NewUserRepository(s.db.GetDB())

       // Call repository method
       user, err := userRepo.Create(r.Context(), req)

       // Return JSON response
       json.NewEncoder(w).Encode(user)
   }
   ↓

4. repository.Create() executes SQL
   INSERT INTO users (username, email, full_name, created_at, updated_at)
   VALUES ($1, $2, $3, $4, $5)
   RETURNING id, username, email, full_name, created_at, updated_at
   ↓

5. PostgreSQL processes query and returns data
   ↓

6. Repository scans result into User struct
   ↓

7. Handler sends JSON response to client
   {
     "id": 1,
     "username": "johndoe",
     "email": "john@example.com",
     "full_name": "John Doe",
     "created_at": "2025-11-04T12:00:00Z",
     "updated_at": "2025-11-04T12:00:00Z"
   }
```

---

## Database Connectivity

### Connection Lifecycle

```go
// 1. Application starts
dbService := database.New()

// 2. Connection created (first time only - singleton)
db, err := sql.Open("pgx", connStr)

// 3. Connection reused for all requests
userRepo := repository.NewUserRepository(db)

// 4. Repository executes queries
user, err := userRepo.Create(ctx, req)

// 5. Application shutdown
defer dbService.Close()
```

### Connection Pooling

Go's `database/sql` package automatically handles connection pooling:

```go
// Behind the scenes
db.SetMaxOpenConns(25)      // Max 25 connections
db.SetMaxIdleConns(25)      // Keep 25 idle connections
db.SetConnMaxLifetime(5min) // Recycle connections every 5 min
```

**What This Means:**
- Multiple requests can use the same connection pool
- Connections are reused (efficient)
- No need to manually manage connections
- Thread-safe (can use from multiple goroutines)

### Environment Variables (.env)

```bash
# Database configuration
BLUEPRINT_DB_HOST=localhost       # Where PostgreSQL is running
BLUEPRINT_DB_PORT=5432           # PostgreSQL port
BLUEPRINT_DB_DATABASE=blueprint  # Database name
BLUEPRINT_DB_USERNAME=melkey     # Database user
BLUEPRINT_DB_PASSWORD=password1234
BLUEPRINT_DB_SCHEMA=public       # PostgreSQL schema

# Application configuration
PORT=8080                        # HTTP server port
APP_ENV=local                    # Environment (local, dev, prod)
```

**How They're Loaded:**
```go
import _ "github.com/joho/godotenv/autoload"  // Auto-loads .env on import

// Access anywhere in code
host := os.Getenv("BLUEPRINT_DB_HOST")
```

---

## Key Concepts

### 1. Context (`context.Context`)

Used in every database operation:

```go
func (r *userRepository) Create(ctx context.Context, req models.CreateUserRequest) (*models.User, error)
```

**What is it?**
- Carries deadlines, cancellation signals, and request-scoped values
- Allows canceling long-running operations
- Essential for timeouts and graceful shutdown

**Example:**
```go
// Context with timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

user, err := userRepo.Create(ctx, req)  // Will timeout after 5 seconds
```

### 2. Error Handling

Go uses explicit error returns (no exceptions):

```go
user, err := userRepo.Create(ctx, req)
if err != nil {
    // Handle error
    log.Printf("Failed to create user: %v", err)
    http.Error(w, "Internal Server Error", http.StatusInternalServerError)
    return
}
// Use user safely
```

### 3. Pointers vs Values

```go
// Pointer (can be nil, shared memory)
func GetByID(id int) (*User, error)  // Returns *User (pointer)

// Value (copy, never nil)
func GetAll() ([]User, error)  // Returns []User (slice)
```

**When to use pointers:**
- Large structs (avoid copying)
- Need to return nil to indicate "not found"
- Need to modify the original value

### 4. Interfaces

```go
type UserRepository interface {
    Create(ctx context.Context, req models.CreateUserRequest) (*models.User, error)
}
```

**Why use interfaces?**
- **Abstraction**: Define behavior without implementation
- **Testing**: Easy to create mock implementations
- **Flexibility**: Swap implementations without changing client code

---

## Adding New Features

### Example: Add a "Products" Table

**1. Create model (`internal/models/product.go`):**
```go
package models

type Product struct {
    ID          int     `json:"id"`
    Name        string  `json:"name"`
    Description string  `json:"description"`
    Price       float64 `json:"price"`
}

type CreateProductRequest struct {
    Name        string  `json:"name"`
    Description string  `json:"description"`
    Price       float64 `json:"price"`
}
```

**2. Create migration (`migrations/002_create_products_table.sql`):**
```sql
CREATE TABLE products (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price DECIMAL(10,2) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);
```

**3. Create repository (`internal/repository/product_repository.go`):**
```go
package repository

import (
    "context"
    "database/sql"
    "StockNet/internal/models"
)

type ProductRepository interface {
    Create(ctx context.Context, req models.CreateProductRequest) (*models.Product, error)
    GetByID(ctx context.Context, id int) (*models.Product, error)
    GetAll(ctx context.Context) ([]models.Product, error)
}

type productRepository struct {
    db *sql.DB
}

func NewProductRepository(db *sql.DB) ProductRepository {
    return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, req models.CreateProductRequest) (*models.Product, error) {
    query := `
        INSERT INTO products (name, description, price)
        VALUES ($1, $2, $3)
        RETURNING id, name, description, price
    `

    product := &models.Product{}
    err := r.db.QueryRowContext(ctx, query, req.Name, req.Description, req.Price).
        Scan(&product.ID, &product.Name, &product.Description, &product.Price)

    return product, err
}

// Implement GetByID, GetAll, etc.
```

**4. Add route handlers (`internal/server/routes.go`):**
```go
func (s *Server) RegisterRoutes() http.Handler {
    mux := http.NewServeMux()

    // Existing routes
    mux.HandleFunc("/", s.HelloWorldHandler)
    mux.HandleFunc("/health", s.healthHandler)

    // New product routes
    mux.HandleFunc("/products", s.handleProducts)
    mux.HandleFunc("/products/", s.handleProduct)  // Note trailing slash for /products/{id}

    return s.corsMiddleware(mux)
}

func (s *Server) handleProducts(w http.ResponseWriter, r *http.Request) {
    productRepo := repository.NewProductRepository(s.db.GetDB())

    switch r.Method {
    case http.MethodGet:
        products, err := productRepo.GetAll(r.Context())
        // Return JSON
    case http.MethodPost:
        var req models.CreateProductRequest
        json.NewDecoder(r.Body).Decode(&req)
        product, err := productRepo.Create(r.Context(), req)
        // Return JSON
    }
}
```

**5. Run migration:**
```bash
docker exec -it stocknet-psql_bp-1 psql -U melkey -d blueprint -f /migrations/002_create_products_table.sql
```

**6. Test:**
```bash
# Create product
curl -X POST http://localhost:8080/products \
  -H "Content-Type: application/json" \
  -d '{"name":"Widget","description":"A great widget","price":19.99}'

# Get all products
curl http://localhost:8080/products
```

---

## Summary

### Project Architecture (Layered)

```
┌─────────────────────────────────────────┐
│  cmd/api (Entry Point)                  │  ← Application starts here
├─────────────────────────────────────────┤
│  internal/server (HTTP Layer)           │  ← Handles HTTP requests/responses
├─────────────────────────────────────────┤
│  internal/repository (Data Access)      │  ← SQL queries & business logic
├─────────────────────────────────────────┤
│  internal/database (Connection)         │  ← Manages PostgreSQL connection
├─────────────────────────────────────────┤
│  internal/models (Data Structures)      │  ← Structs representing data
└─────────────────────────────────────────┘
```

### Key Takeaways

1. **We're using pgx** (the best PostgreSQL driver for Go)
2. **Singleton pattern** for database connection (one connection pool)
3. **Repository pattern** separates SQL from HTTP handlers
4. **Dependency injection** makes code testable
5. **Clean architecture** with clear layer separation
6. **Standard project layout** following Go conventions

### Next Steps

1. Add more routes in `internal/server/routes.go`
2. Create more repositories in `internal/repository/`
3. Add authentication/authorization middleware
4. Implement business logic services
5. Add comprehensive tests

---

**Need help with anything specific? Check out:**
- `examples/README.md` for usage examples
- `internal/repository/user_repository_test.go` for testing patterns
- `go.mod` to see all dependencies