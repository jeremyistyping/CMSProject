#!/bin/bash

# ================================================================
# Rollback Script
# ================================================================
# This script restores the application from a backup
# ================================================================

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Function to print colored messages
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

cd "$PROJECT_ROOT"

# Load environment variables
if [ -f ".env.production" ]; then
    source .env.production
fi

BACKUP_DIR="${BACKUP_DIR:-./backups}"

echo -e "${BLUE}================================================================${NC}"
echo -e "${BLUE}   Rollback / Restore from Backup${NC}"
echo -e "${BLUE}================================================================${NC}"
echo ""

# Step 1: List available backups
print_info "Available backups:"
echo ""

BACKUPS=($(ls -t "$BACKUP_DIR"/backup_*.tar.gz 2>/dev/null || true))

if [ ${#BACKUPS[@]} -eq 0 ]; then
    print_error "No backups found in $BACKUP_DIR"
    exit 1
fi

# Display backups with numbers
for i in "${!BACKUPS[@]}"; do
    BACKUP_FILE=$(basename "${BACKUPS[$i]}")
    BACKUP_DATE=$(echo "$BACKUP_FILE" | sed 's/backup_\([0-9]\{8\}\)_\([0-9]\{6\}\).*/\1 \2/' | sed 's/\([0-9]\{4\}\)\([0-9]\{2\}\)\([0-9]\{2\}\) \([0-9]\{2\}\)\([0-9]\{2\}\)\([0-9]\{2\}\)/\1-\2-\3 \4:\5:\6/')
    BACKUP_SIZE=$(du -h "${BACKUPS[$i]}" | cut -f1)
    echo -e "  ${YELLOW}$((i+1))${NC}. $BACKUP_FILE"
    echo -e "     Date: $BACKUP_DATE | Size: $BACKUP_SIZE"
    echo ""
done

# Step 2: Select backup
if [ -n "$1" ]; then
    # Backup name provided as argument
    SELECTED_BACKUP="$BACKUP_DIR/$1.tar.gz"
    if [ ! -f "$SELECTED_BACKUP" ]; then
        print_error "Backup not found: $1"
        exit 1
    fi
else
    # Interactive selection
    read -p "Select backup number to restore (1-${#BACKUPS[@]}): " BACKUP_NUM
    
    if ! [[ "$BACKUP_NUM" =~ ^[0-9]+$ ]] || [ "$BACKUP_NUM" -lt 1 ] || [ "$BACKUP_NUM" -gt ${#BACKUPS[@]} ]; then
        print_error "Invalid selection"
        exit 1
    fi
    
    SELECTED_BACKUP="${BACKUPS[$((BACKUP_NUM-1))]}"
fi

BACKUP_NAME=$(basename "$SELECTED_BACKUP" .tar.gz)

echo ""
print_warning "You are about to restore from: $BACKUP_NAME"
print_warning "This will:"
echo "  - Stop the current application"
echo "  - Restore the database (current data will be lost)"
echo "  - Restore uploaded files"
echo "  - Restart the application"
echo ""

read -p "Are you sure you want to continue? (yes/NO): " CONFIRM

if [ "$CONFIRM" != "yes" ]; then
    print_info "Rollback cancelled"
    exit 0
fi

echo ""

# Step 3: Extract backup
print_info "Step 1: Extracting backup..."

TEMP_DIR=$(mktemp -d)
tar xzf "$SELECTED_BACKUP" -C "$TEMP_DIR"

BACKUP_CONTENT_DIR="$TEMP_DIR/$BACKUP_NAME"

if [ ! -d "$BACKUP_CONTENT_DIR" ]; then
    print_error "Invalid backup structure"
    rm -rf "$TEMP_DIR"
    exit 1
fi

print_success "Backup extracted"
echo ""

# Step 4: Display backup information
if [ -f "$BACKUP_CONTENT_DIR/backup_info.txt" ]; then
    print_info "Backup Information:"
    cat "$BACKUP_CONTENT_DIR/backup_info.txt"
    echo ""
fi

# Step 5: Stop containers
print_info "Step 2: Stopping containers..."

docker-compose stop

print_success "Containers stopped"
echo ""

# Step 6: Restore database
print_info "Step 3: Restoring database..."

if [ -f "$BACKUP_CONTENT_DIR/database.sql" ]; then
    # Start only postgres
    docker-compose up -d postgres
    
    # Wait for postgres to be ready
    print_info "Waiting for database to be ready..."
    sleep 10
    
    # Drop and recreate database
    docker-compose exec -T postgres psql -U "$POSTGRES_USER" -c "DROP DATABASE IF EXISTS $POSTGRES_DB;" || true
    docker-compose exec -T postgres psql -U "$POSTGRES_USER" -c "CREATE DATABASE $POSTGRES_DB;"
    
    # Restore database
    docker-compose exec -T postgres psql -U "$POSTGRES_USER" "$POSTGRES_DB" < "$BACKUP_CONTENT_DIR/database.sql"
    
    print_success "Database restored"
else
    print_error "Database backup not found"
    docker-compose start
    rm -rf "$TEMP_DIR"
    exit 1
fi

echo ""

# Step 7: Restore uploaded files
print_info "Step 4: Restoring uploaded files..."

if [ -f "$BACKUP_CONTENT_DIR/uploads.tar.gz" ]; then
    VOLUME_NAME=$(docker volume ls -q | grep backend_uploads || echo "")
    
    if [ -n "$VOLUME_NAME" ]; then
        # Remove old uploads
        docker run --rm \
            -v "$VOLUME_NAME":/uploads \
            alpine \
            sh -c "rm -rf /uploads/*"
        
        # Restore uploads
        docker run --rm \
            -v "$VOLUME_NAME":/uploads \
            -v "$BACKUP_CONTENT_DIR":/backup \
            alpine \
            tar xzf /backup/uploads.tar.gz -C /uploads
        
        print_success "Uploaded files restored"
    else
        print_warning "Uploads volume not found"
    fi
else
    print_warning "No uploaded files in backup"
fi

echo ""

# Step 8: Restore configuration (optional)
if [ -f "$BACKUP_CONTENT_DIR/env.production.backup" ]; then
    print_info "Step 5: Configuration backup found"
    read -p "Restore configuration? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        cp "$BACKUP_CONTENT_DIR/env.production.backup" .env.production
        print_success "Configuration restored"
    fi
    echo ""
fi

# Step 9: Restore Git commit (optional)
if [ -f "$BACKUP_CONTENT_DIR/git_commit.txt" ]; then
    BACKUP_COMMIT=$(cat "$BACKUP_CONTENT_DIR/git_commit.txt")
    if [ "$BACKUP_COMMIT" != "unknown" ]; then
        print_info "Step 6: Git commit in backup: $BACKUP_COMMIT"
        read -p "Checkout this commit? (y/N): " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            git checkout "$BACKUP_COMMIT"
            print_success "Git commit restored"
        fi
        echo ""
    fi
fi

# Step 10: Start containers
print_info "Step 7: Starting containers..."

docker-compose up -d

print_success "Containers started"
echo ""

# Step 11: Wait for services to be healthy
print_info "Step 8: Waiting for services to be healthy..."

sleep 15

# Step 12: Health check
print_info "Step 9: Performing health check..."

if bash "$SCRIPT_DIR/health-check.sh"; then
    print_success "Health check passed"
else
    print_warning "Health check failed"
    print_info "Check logs with: docker-compose logs"
fi

echo ""

# Cleanup
rm -rf "$TEMP_DIR"

# Display rollback summary
echo -e "${GREEN}================================================================${NC}"
echo -e "${GREEN}   Rollback Complete!${NC}"
echo -e "${GREEN}================================================================${NC}"
echo ""
echo -e "${BLUE}Restored from:${NC} ${GREEN}$BACKUP_NAME${NC}"
echo ""
echo -e "${BLUE}Useful commands:${NC}"
echo -e "  View logs:        ${YELLOW}docker-compose logs -f${NC}"
echo -e "  Check health:     ${YELLOW}./scripts/health-check.sh${NC}"
echo ""

print_success "Rollback successful!"
