# Simple Makefile for StockNet CLI

# Run the CLI application
run:
	@go run cmd/cli/main.go

# Build the CLI
build-cli:
	@echo "Building CLI..."
	@go build -o stocknet-cli cmd/cli/main.go

# Build the API
build-api:
	@echo "Building API..."
	@go build -o main cmd/api/main.go

# Clean binaries
clean:
	@echo "Cleaning..."
	@rm -f main stocknet-cli

# Live reload (requires air: go install github.com/air-verse/air@latest)
watch:
	@if command -v air > /dev/null; then \
		air; \
	else \
		echo "Error: 'air' is not installed. Install with: go install github.com/air-verse/air@latest"; \
		exit 1; \
	fi

.PHONY: run build-cli build-api clean watch
