#!/bin/bash

# ================================================================
# Backup and Reset Database Script for Linux/Mac
# ================================================================
# This script will:
# 1. Backup the current database
# 2. Reset the database
# ================================================================

echo ""
echo "========================================"
echo "  Backup and Reset Database"
echo "========================================"
echo ""

# Check if .env file exists
if [ ! -f "../../.env" ]; then
    echo "ERROR: .env file not found in backend folder"
    echo "Please create .env file first"
    exit 1
fi

# Load database URL from .env
DATABASE_URL=$(grep "^DATABASE_URL=" ../../.env | cut -d '=' -f2)

if [ -z "$DATABASE_URL" ]; then
    echo "ERROR: DATABASE_URL not found in .env file"
    exit 1
fi

# Parse database name from URL
# Format: postgres://user:pass@host:port/dbname?options
DB_NAME=$(echo $DATABASE_URL | sed -n 's/.*\/\([^?]*\).*/\1/p')

if [ -z "$DB_NAME" ]; then
    echo "ERROR: Could not parse database name from DATABASE_URL"
    exit 1
fi

# Create backup directory if not exists
mkdir -p ../../backups

# Generate backup filename with timestamp
TIMESTAMP=$(date +"%Y-%m-%d_%H-%M-%S")
BACKUP_FILE="../../backups/${DB_NAME}_${TIMESTAMP}.sql"

echo "Database: $DB_NAME"
echo "Backup file: $BACKUP_FILE"
echo ""

# Ask for confirmation
read -p "Do you want to backup and reset? (yes/no): " CONFIRM
if [ "$CONFIRM" != "yes" ]; then
    echo "Operation cancelled"
    exit 0
fi

echo ""
echo "Step 1: Creating backup..."
echo ""

# Backup database using pg_dump
pg_dump -U postgres -d "$DB_NAME" -f "$BACKUP_FILE"

if [ $? -ne 0 ]; then
    echo "ERROR: Backup failed!"
    echo "Make sure PostgreSQL is installed and pg_dump is in PATH"
    exit 1
fi

echo "Backup created successfully: $BACKUP_FILE"
echo ""

echo "Step 2: Resetting database..."
echo ""

# Run reset script
go run main.go

echo ""
echo "========================================"
echo "  Backup and Reset Complete"
echo "========================================"
echo ""
echo "Backup saved to: $BACKUP_FILE"
echo ""
echo "To restore from backup:"
echo "  psql -U postgres -d $DB_NAME -f $BACKUP_FILE"
echo ""
