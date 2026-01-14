#!/bin/bash

# ================================================================
# Backup Script
# ================================================================
# This script backs up the database and uploaded files
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
else
    print_error ".env.production not found"
    exit 1
fi

# Set backup directory
BACKUP_DIR="${BACKUP_DIR:-./backups}"
TIMESTAMP=$(date +"%Y%m%d_%H%M%S")
BACKUP_NAME="backup_${TIMESTAMP}"
BACKUP_PATH="${BACKUP_DIR}/${BACKUP_NAME}"

echo -e "${BLUE}================================================================${NC}"
echo -e "${BLUE}   Creating Backup${NC}"
echo -e "${BLUE}================================================================${NC}"
echo ""

# Create backup directory
mkdir -p "$BACKUP_DIR"
mkdir -p "$BACKUP_PATH"

# Step 1: Backup database
print_info "Step 1: Backing up database..."

if docker-compose exec -T postgres pg_dump -U "$POSTGRES_USER" "$POSTGRES_DB" > "${BACKUP_PATH}/database.sql"; then
    print_success "Database backed up successfully"
    
    # Get database size
    DB_SIZE=$(du -h "${BACKUP_PATH}/database.sql" | cut -f1)
    print_info "Database backup size: $DB_SIZE"
else
    print_error "Database backup failed"
    rm -rf "$BACKUP_PATH"
    exit 1
fi

echo ""

# Step 2: Backup uploaded files
print_info "Step 2: Backing up uploaded files..."

# Get the volume name
VOLUME_NAME=$(docker volume ls -q | grep backend_uploads || echo "")

if [ -n "$VOLUME_NAME" ]; then
    # Create temporary container to access volume
    docker run --rm \
        -v "$VOLUME_NAME":/source \
        -v "$BACKUP_PATH":/backup \
        alpine \
        tar czf /backup/uploads.tar.gz -C /source .
    
    if [ -f "${BACKUP_PATH}/uploads.tar.gz" ]; then
        print_success "Uploaded files backed up successfully"
        
        # Get uploads size
        UPLOADS_SIZE=$(du -h "${BACKUP_PATH}/uploads.tar.gz" | cut -f1)
        print_info "Uploads backup size: $UPLOADS_SIZE"
    else
        print_warning "No uploaded files found"
    fi
else
    print_warning "Uploads volume not found"
fi

echo ""

# Step 3: Save environment configuration
print_info "Step 3: Backing up configuration..."

if [ -f ".env.production" ]; then
    cp .env.production "${BACKUP_PATH}/env.production.backup"
    print_success "Configuration backed up"
fi

# Save Git commit info
git rev-parse HEAD > "${BACKUP_PATH}/git_commit.txt" 2>/dev/null || echo "unknown" > "${BACKUP_PATH}/git_commit.txt"
git rev-parse --abbrev-ref HEAD > "${BACKUP_PATH}/git_branch.txt" 2>/dev/null || echo "unknown" > "${BACKUP_PATH}/git_branch.txt"

echo ""

# Step 4: Create backup metadata
print_info "Step 4: Creating backup metadata..."

cat > "${BACKUP_PATH}/backup_info.txt" << EOF
Backup Information
==================
Timestamp: $(date)
Database: $POSTGRES_DB
Git Commit: $(cat "${BACKUP_PATH}/git_commit.txt")
Git Branch: $(cat "${BACKUP_PATH}/git_branch.txt")
Server Host: ${SERVER_HOST:-unknown}
EOF

print_success "Metadata created"
echo ""

# Step 5: Compress backup
print_info "Step 5: Compressing backup..."

cd "$BACKUP_DIR"
tar czf "${BACKUP_NAME}.tar.gz" "$BACKUP_NAME"

if [ -f "${BACKUP_NAME}.tar.gz" ]; then
    # Remove uncompressed backup
    rm -rf "$BACKUP_NAME"
    
    BACKUP_SIZE=$(du -h "${BACKUP_NAME}.tar.gz" | cut -f1)
    print_success "Backup compressed successfully"
    print_info "Total backup size: $BACKUP_SIZE"
else
    print_error "Compression failed"
    exit 1
fi

cd "$PROJECT_ROOT"
echo ""

# Step 6: Cleanup old backups
print_info "Step 6: Cleaning up old backups..."

RETENTION_DAYS="${BACKUP_RETENTION_DAYS:-30}"
print_info "Retention policy: $RETENTION_DAYS days"

# Find and delete old backups
DELETED_COUNT=0
while IFS= read -r old_backup; do
    rm -f "$old_backup"
    DELETED_COUNT=$((DELETED_COUNT + 1))
done < <(find "$BACKUP_DIR" -name "backup_*.tar.gz" -type f -mtime +$RETENTION_DAYS)

if [ $DELETED_COUNT -gt 0 ]; then
    print_success "Deleted $DELETED_COUNT old backup(s)"
else
    print_info "No old backups to delete"
fi

echo ""

# Display backup summary
echo -e "${GREEN}================================================================${NC}"
echo -e "${GREEN}   Backup Complete!${NC}"
echo -e "${GREEN}================================================================${NC}"
echo ""
echo -e "${BLUE}Backup Information:${NC}"
echo -e "  Location:  ${GREEN}${BACKUP_DIR}/${BACKUP_NAME}.tar.gz${NC}"
echo -e "  Size:      ${GREEN}${BACKUP_SIZE}${NC}"
echo -e "  Timestamp: ${GREEN}${TIMESTAMP}${NC}"
echo ""
echo -e "${BLUE}Backup Contents:${NC}"
echo -e "  ✓ Database dump"
echo -e "  ✓ Uploaded files"
echo -e "  ✓ Configuration"
echo -e "  ✓ Git information"
echo ""
echo -e "${YELLOW}To restore this backup:${NC}"
echo -e "  ${YELLOW}./scripts/rollback.sh ${BACKUP_NAME}${NC}"
echo ""

print_success "Backup successful!"
