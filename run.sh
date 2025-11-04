#!/bin/bash
# Load environment variables and run the application

# Load .env file
export $(cat .env | grep -v '^#' | xargs)

# Run the application
go run "$@"