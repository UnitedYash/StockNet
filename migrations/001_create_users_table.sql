-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    username VARCHAR(100) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    full_name VARCHAR(255),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create index on email for faster lookups
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);

-- Insert sample data
INSERT INTO users (username, email, full_name) VALUES
    ('john_doe', 'john.doe@example.com', 'John Doe'),
    ('jane_smith', 'jane.smith@example.com', 'Jane Smith'),
    ('bob_wilson', 'bob.wilson@example.com', 'Bob Wilson'),
    ('alice_johnson', 'alice.johnson@example.com', 'Alice Johnson'),
    ('charlie_brown', 'charlie.brown@example.com', 'Charlie Brown')
ON CONFLICT (username) DO NOTHING;