#!/bin/bash

# ================================================================
# Database Reset Script for Linux/Mac
# ================================================================
# This script will reset the database by dropping all tables,
# sequences, and custom types.
#
# WARNING: THIS WILL DELETE ALL DATA!
# ================================================================

echo ""
echo "========================================"
echo "  Database Reset Tool"
echo "========================================"
echo ""
echo "WARNING: This will DELETE ALL DATA!"
echo ""

# Check if .env file exists
if [ ! -f "../../.env" ]; then
    echo "ERROR: .env file not found in backend folder"
    echo "Please create .env file first"
    exit 1
fi

echo "Running database reset..."
echo ""

# Run the Go script
go run main.go

echo ""
echo "========================================"
echo "  Reset Complete"
echo "========================================"
echo ""
echo "Next steps:"
echo "1. Go to backend folder: cd ../.."
echo "2. Run application: go run main.go"
echo ""
