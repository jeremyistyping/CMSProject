# Makefile for Database Reset Tool

.PHONY: help reset build clean run

# Default target
help:
	@echo "Database Reset Tool - Available Commands:"
	@echo ""
	@echo "  make reset    - Reset database (drop all tables, sequences, types)"
	@echo "  make build    - Build executable"
	@echo "  make run      - Run without building"
	@echo "  make clean    - Remove built executable"
	@echo ""
	@echo "WARNING: 'make reset' will DELETE ALL DATA in the database!"

# Reset database
reset: run

# Build executable
build:
	@echo "Building reset_db executable..."
	@go build -o reset_db main.go
	@echo "Build complete: reset_db"

# Run without building
run:
	@echo "Running database reset..."
	@go run main.go

# Clean built files
clean:
	@echo "Cleaning built files..."
	@rm -f reset_db reset_db.exe
	@echo "Clean complete"

# Install dependencies
deps:
	@echo "Installing dependencies..."
	@cd ../.. && go mod download
	@echo "Dependencies installed"
