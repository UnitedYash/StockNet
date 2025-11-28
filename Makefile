

# Run the CLI application
run:
	@go run cmd/cli/main.go

# Build the CLI
build-cli:
	@echo "Building CLI..."
	@go build -o stocknet-cli cmd/cli/main.go


# Clean binaries
clean:
	@echo "Cleaning..."
	@rm -f main stocknet-cli

.PHONY: run build-cli clean watch
