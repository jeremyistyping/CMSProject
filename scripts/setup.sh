#!/bin/bash

# ================================================================
# Initial Setup Script for VPS Deployment
# ================================================================
# This script sets up the application for the first time on a VPS
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
echo -e "${BLUE}   Accounting System - Initial VPS Setup${NC}"
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

# Check if running as root
if [ "$EUID" -eq 0 ]; then
    print_warning "Please do not run this script as root"
    exit 1
fi

# Step 1: Check system requirements
print_info "Step 1: Checking system requirements..."

# Check Docker
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed"
    echo "Please install Docker first: https://docs.docker.com/engine/install/"
    exit 1
fi
print_success "Docker is installed"

# Check Docker Compose
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    print_error "Docker Compose is not installed"
    echo "Please install Docker Compose first: https://docs.docker.com/compose/install/"
    exit 1
fi
print_success "Docker Compose is installed"

# Check Git
if ! command -v git &> /dev/null; then
    print_error "Git is not installed"
    echo "Please install Git first: sudo apt install git"
    exit 1
fi
print_success "Git is installed"

# Check if user is in docker group
if ! groups | grep -q docker; then
    print_warning "Current user is not in docker group"
    echo "Run: sudo usermod -aG docker $USER"
    echo "Then log out and log back in"
    exit 1
fi
print_success "User is in docker group"

echo ""

# Step 2: Check if .env.production exists
print_info "Step 2: Checking environment configuration..."

cd "$PROJECT_ROOT"

if [ -f ".env.production" ]; then
    print_warning ".env.production already exists"
    read -p "Do you want to reconfigure? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_info "Using existing .env.production"
    else
        rm .env.production
        bash "$SCRIPT_DIR/setup-env.sh"
    fi
else
    print_info "Creating .env.production..."
    bash "$SCRIPT_DIR/setup-env.sh"
fi

print_success "Environment configuration ready"
echo ""

# Step 3: Create necessary directories
print_info "Step 3: Creating necessary directories..."

mkdir -p nginx/ssl
mkdir -p nginx/conf.d
mkdir -p backups
mkdir -p logs

print_success "Directories created"
echo ""

# Step 4: Build Docker images
print_info "Step 4: Building Docker images..."
print_warning "This may take several minutes..."

if docker-compose build; then
    print_success "Docker images built successfully"
else
    print_error "Failed to build Docker images"
    exit 1
fi

echo ""

# Step 5: Start containers
print_info "Step 5: Starting containers..."

if docker-compose up -d; then
    print_success "Containers started successfully"
else
    print_error "Failed to start containers"
    exit 1
fi

echo ""

# Step 6: Wait for services to be healthy
print_info "Step 6: Waiting for services to be healthy..."

MAX_WAIT=120
WAIT_TIME=0

while [ $WAIT_TIME -lt $MAX_WAIT ]; do
    if docker-compose ps | grep -q "unhealthy"; then
        echo -n "."
        sleep 5
        WAIT_TIME=$((WAIT_TIME + 5))
    else
        break
    fi
done

echo ""

if [ $WAIT_TIME -ge $MAX_WAIT ]; then
    print_warning "Services took longer than expected to start"
    print_info "Check logs with: docker-compose logs"
else
    print_success "All services are healthy"
fi

echo ""

# Step 7: Display access information
print_info "Step 7: Deployment complete!"
echo ""
echo -e "${GREEN}================================================================${NC}"
echo -e "${GREEN}   Application is now running!${NC}"
echo -e "${GREEN}================================================================${NC}"
echo ""

# Get SERVER_HOST from .env.production
if [ -f ".env.production" ]; then
    source .env.production
    echo -e "${BLUE}Access your application at:${NC}"
    echo -e "  ${GREEN}${SERVER_HOST}${NC}"
    echo ""
fi

echo -e "${BLUE}Useful commands:${NC}"
echo -e "  View logs:        ${YELLOW}docker-compose logs -f${NC}"
echo -e "  Stop app:         ${YELLOW}docker-compose stop${NC}"
echo -e "  Start app:        ${YELLOW}docker-compose start${NC}"
echo -e "  Restart app:      ${YELLOW}docker-compose restart${NC}"
echo -e "  Update app:       ${YELLOW}./scripts/deploy.sh${NC}"
echo -e "  Backup database:  ${YELLOW}./scripts/backup.sh${NC}"
echo -e "  View health:      ${YELLOW}./scripts/health-check.sh${NC}"
echo ""

echo -e "${YELLOW}Next steps:${NC}"
echo "  1. Test the application in your browser"
echo "  2. Set up SSL if you have a domain (see nginx/SSL_SETUP.md)"
echo "  3. Configure regular backups (cron job)"
echo "  4. Set up monitoring and alerts"
echo ""

print_success "Setup complete!"
