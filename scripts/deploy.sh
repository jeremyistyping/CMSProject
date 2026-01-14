#!/bin/bash

# ================================================================
# Deployment/Update Script
# ================================================================
# This script updates the application with zero downtime
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

echo -e "${BLUE}================================================================${NC}"
echo -e "${BLUE}   Accounting System - Deployment Update${NC}"
echo -e "${BLUE}================================================================${NC}"
echo ""

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

# Step 1: Check if application is running
print_info "Step 1: Checking current deployment..."

if ! docker-compose ps | grep -q "Up"; then
    print_error "Application is not running"
    echo "Run ./scripts/setup.sh first"
    exit 1
fi

print_success "Application is running"
echo ""

# Step 2: Create backup before update
print_info "Step 2: Creating backup before update..."

if bash "$SCRIPT_DIR/backup.sh"; then
    print_success "Backup created successfully"
else
    print_error "Backup failed"
    read -p "Continue anyway? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

echo ""

# Step 3: Pull latest code from Git
print_info "Step 3: Pulling latest code from Git..."

# Get current branch
CURRENT_BRANCH=$(git rev-parse --abbrev-ref HEAD)
print_info "Current branch: $CURRENT_BRANCH"

# Get current commit
CURRENT_COMMIT=$(git rev-parse --short HEAD)
print_info "Current commit: $CURRENT_COMMIT"

# Stash any local changes
if ! git diff-index --quiet HEAD --; then
    print_warning "Local changes detected, stashing..."
    git stash
fi

# Pull latest changes
if git pull origin "$CURRENT_BRANCH"; then
    NEW_COMMIT=$(git rev-parse --short HEAD)
    if [ "$CURRENT_COMMIT" = "$NEW_COMMIT" ]; then
        print_info "Already up to date"
    else
        print_success "Updated from $CURRENT_COMMIT to $NEW_COMMIT"
    fi
else
    print_error "Failed to pull latest code"
    exit 1
fi

echo ""

# Step 4: Check if rebuild is needed
print_info "Step 4: Checking if rebuild is needed..."

REBUILD_NEEDED=false

# Check if Dockerfiles changed
if git diff --name-only "$CURRENT_COMMIT" "$NEW_COMMIT" | grep -q "Dockerfile"; then
    print_info "Dockerfile changes detected"
    REBUILD_NEEDED=true
fi

# Check if dependencies changed
if git diff --name-only "$CURRENT_COMMIT" "$NEW_COMMIT" | grep -qE "(go.mod|go.sum|package.json|package-lock.json)"; then
    print_info "Dependency changes detected"
    REBUILD_NEEDED=true
fi

if [ "$REBUILD_NEEDED" = true ]; then
    print_warning "Rebuild required"
    
    # Rebuild images
    print_info "Rebuilding Docker images..."
    if docker-compose build; then
        print_success "Images rebuilt successfully"
    else
        print_error "Failed to rebuild images"
        print_warning "Rolling back..."
        git reset --hard "$CURRENT_COMMIT"
        exit 1
    fi
else
    print_info "No rebuild needed"
fi

echo ""

# Step 5: Perform rolling update
print_info "Step 5: Performing rolling update..."

# Update containers one by one for zero downtime
print_info "Updating backend..."
docker-compose up -d --no-deps --build backend

print_info "Waiting for backend to be healthy..."
sleep 10

print_info "Updating frontend..."
docker-compose up -d --no-deps --build frontend

print_info "Waiting for frontend to be healthy..."
sleep 10

print_info "Updating nginx..."
docker-compose up -d --no-deps nginx

print_success "All services updated"
echo ""

# Step 6: Run database migrations
print_info "Step 6: Running database migrations..."

if docker-compose exec -T backend ./api migrate; then
    print_success "Migrations completed"
else
    print_warning "Migration command not available or failed"
    print_info "Migrations will run automatically on startup"
fi

echo ""

# Step 7: Health check
print_info "Step 7: Performing health check..."

sleep 5

if bash "$SCRIPT_DIR/health-check.sh"; then
    print_success "Health check passed"
else
    print_error "Health check failed"
    print_warning "Application may not be working correctly"
    
    read -p "Rollback to previous version? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        print_info "Rolling back..."
        bash "$SCRIPT_DIR/rollback.sh"
        exit 1
    fi
fi

echo ""

# Step 8: Cleanup old images
print_info "Step 8: Cleaning up old Docker images..."

docker image prune -f > /dev/null 2>&1 || true

print_success "Cleanup complete"
echo ""

# Display deployment summary
echo -e "${GREEN}================================================================${NC}"
echo -e "${GREEN}   Deployment Complete!${NC}"
echo -e "${GREEN}================================================================${NC}"
echo ""
echo -e "${BLUE}Deployment Summary:${NC}"
echo -e "  Previous commit: ${YELLOW}$CURRENT_COMMIT${NC}"
echo -e "  New commit:      ${GREEN}$NEW_COMMIT${NC}"
echo -e "  Branch:          ${BLUE}$CURRENT_BRANCH${NC}"
echo ""

if [ -f ".env.production" ]; then
    source .env.production
    echo -e "${BLUE}Application URL:${NC}"
    echo -e "  ${GREEN}${SERVER_HOST}${NC}"
    echo ""
fi

echo -e "${BLUE}Useful commands:${NC}"
echo -e "  View logs:        ${YELLOW}docker-compose logs -f${NC}"
echo -e "  Check health:     ${YELLOW}./scripts/health-check.sh${NC}"
echo -e "  Rollback:         ${YELLOW}./scripts/rollback.sh${NC}"
echo ""

print_success "Deployment successful!"
