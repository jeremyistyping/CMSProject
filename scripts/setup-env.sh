#!/bin/bash

# ================================================================
# Environment Setup Helper Script
# ================================================================
# This script helps create .env.production interactively
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

cd "$PROJECT_ROOT"

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

# Function to generate random string
generate_secret() {
    openssl rand -base64 32 2>/dev/null || cat /dev/urandom | tr -dc 'a-zA-Z0-9' | fold -w 32 | head -n 1
}

echo -e "${BLUE}================================================================${NC}"
echo -e "${BLUE}   Environment Configuration Setup${NC}"
echo -e "${BLUE}================================================================${NC}"
echo ""

print_info "This script will help you create .env.production"
echo ""

# Get VPS IP or domain
print_info "Step 1: Server Configuration"
echo ""
echo "Enter your server address (IP or domain):"
echo "  Examples:"
echo "    - IP only: 192.168.1.100"
echo "    - Domain: accounting.company.com"
echo ""
read -p "Server address: " SERVER_ADDRESS

if [ -z "$SERVER_ADDRESS" ]; then
    print_error "Server address is required"
    exit 1
fi

# Determine protocol
echo ""
read -p "Use HTTPS? (y/N): " -n 1 -r USE_HTTPS
echo ""

if [[ $REPLY =~ ^[Yy]$ ]]; then
    SERVER_HOST="https://$SERVER_ADDRESS"
    ENABLE_SSL="true"
else
    SERVER_HOST="http://$SERVER_ADDRESS"
    ENABLE_SSL="false"
fi

print_success "Server host: $SERVER_HOST"
echo ""

# Database configuration
print_info "Step 2: Database Configuration"
echo ""

read -p "Database name [sistem_akuntansi_prod]: " POSTGRES_DB
POSTGRES_DB=${POSTGRES_DB:-sistem_akuntansi_prod}

read -p "Database user [accounting_user]: " POSTGRES_USER
POSTGRES_USER=${POSTGRES_USER:-accounting_user}

echo ""
print_warning "Database password (leave empty to generate random):"
read -sp "Password: " POSTGRES_PASSWORD
echo ""

if [ -z "$POSTGRES_PASSWORD" ]; then
    POSTGRES_PASSWORD=$(generate_secret)
    print_info "Generated random password"
fi

print_success "Database configured"
echo ""

# JWT Secret
print_info "Step 3: Security Configuration"
echo ""

print_warning "JWT Secret (leave empty to generate random):"
read -sp "JWT Secret: " JWT_SECRET
echo ""

if [ -z "$JWT_SECRET" ]; then
    JWT_SECRET=$(generate_secret)
    print_info "Generated random JWT secret"
fi

print_success "Security configured"
echo ""

# Backup configuration
print_info "Step 4: Backup Configuration"
echo ""

read -p "Backup directory [/opt/backups/accounting_app]: " BACKUP_DIR
BACKUP_DIR=${BACKUP_DIR:-/opt/backups/accounting_app}

read -p "Backup retention days [30]: " BACKUP_RETENTION_DAYS
BACKUP_RETENTION_DAYS=${BACKUP_RETENTION_DAYS:-30}

print_success "Backup configured"
echo ""

# Create .env.production
print_info "Creating .env.production..."

cat > .env.production << EOF
# ================================================================
# PRODUCTION ENVIRONMENT CONFIGURATION
# ================================================================
# Generated on: $(date)
# ================================================================

# Project Configuration
COMPOSE_PROJECT_NAME=accounting_app

# Server Configuration
SERVER_HOST=$SERVER_HOST
HTTP_PORT=80
HTTPS_PORT=443

# Database Configuration
POSTGRES_DB=$POSTGRES_DB
POSTGRES_USER=$POSTGRES_USER
POSTGRES_PASSWORD=$POSTGRES_PASSWORD

# Backend Configuration
JWT_SECRET=$JWT_SECRET
ENVIRONMENT=production
SKIP_BALANCE_RESET=false
ENABLE_LEGACY_PAYMENT_JOURNALS=true
ENABLE_SSOT_PAYMENT_JOURNALS=true

# CORS Configuration
ALLOWED_ORIGINS=$SERVER_HOST

# Frontend Configuration
NEXT_PUBLIC_API_URL=$SERVER_HOST/api

# SSL Configuration
ENABLE_SSL=$ENABLE_SSL

# Backup Configuration
BACKUP_DIR=$BACKUP_DIR
BACKUP_RETENTION_DAYS=$BACKUP_RETENTION_DAYS
EOF

print_success ".env.production created"
echo ""

# Display summary
echo -e "${GREEN}================================================================${NC}"
echo -e "${GREEN}   Configuration Complete!${NC}"
echo -e "${GREEN}================================================================${NC}"
echo ""
echo -e "${BLUE}Configuration Summary:${NC}"
echo -e "  Server:           ${GREEN}$SERVER_HOST${NC}"
echo -e "  Database:         ${GREEN}$POSTGRES_DB${NC}"
echo -e "  Database User:    ${GREEN}$POSTGRES_USER${NC}"
echo -e "  SSL Enabled:      ${GREEN}$ENABLE_SSL${NC}"
echo -e "  Backup Dir:       ${GREEN}$BACKUP_DIR${NC}"
echo ""

print_warning "IMPORTANT: Keep these credentials secure!"
echo ""
echo -e "${BLUE}Database Password:${NC} ${YELLOW}$POSTGRES_PASSWORD${NC}"
echo -e "${BLUE}JWT Secret:${NC}        ${YELLOW}${JWT_SECRET:0:20}...${NC}"
echo ""

print_info "Configuration saved to .env.production"
print_warning "Never commit .env.production to Git!"
echo ""

# Offer to create backup directory
if [ ! -d "$BACKUP_DIR" ]; then
    read -p "Create backup directory now? (Y/n): " -n 1 -r
    echo ""
    if [[ ! $REPLY =~ ^[Nn]$ ]]; then
        sudo mkdir -p "$BACKUP_DIR"
        sudo chown $USER:$USER "$BACKUP_DIR"
        print_success "Backup directory created"
    fi
fi

print_success "Setup complete!"
