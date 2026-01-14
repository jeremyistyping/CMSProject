#!/bin/bash

# ================================================================
# Log Viewing Script
# ================================================================
# This script helps view and search logs from containers
# ================================================================

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
print_info() {
    echo -e "${BLUE}ℹ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

# Function to display usage
usage() {
    echo "Usage: $0 [OPTIONS] [SERVICE]"
    echo ""
    echo "View logs from Docker containers"
    echo ""
    echo "Options:"
    echo "  -f, --follow       Follow log output (live tail)"
    echo "  -n, --lines NUM    Number of lines to show (default: 100)"
    echo "  -s, --search TEXT  Search for specific text in logs"
    echo "  -h, --help         Show this help message"
    echo ""
    echo "Services:"
    echo "  backend            Backend API logs"
    echo "  frontend           Frontend logs"
    echo "  postgres           Database logs"
    echo "  nginx              Nginx logs"
    echo "  all                All services (default)"
    echo ""
    echo "Examples:"
    echo "  $0                          # View all logs"
    echo "  $0 -f backend               # Follow backend logs"
    echo "  $0 -n 50 frontend           # Show last 50 lines of frontend logs"
    echo "  $0 -s \"error\" backend       # Search for 'error' in backend logs"
    echo "  $0 -f -s \"database\" all    # Follow all logs and filter for 'database'"
    exit 0
}

# Default values
FOLLOW=false
LINES=100
SEARCH=""
SERVICE="all"

# Parse arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -f|--follow)
            FOLLOW=true
            shift
            ;;
        -n|--lines)
            LINES="$2"
            shift 2
            ;;
        -s|--search)
            SEARCH="$2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        backend|frontend|postgres|nginx|all)
            SERVICE="$1"
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            usage
            ;;
    esac
done

# Check if docker-compose is available
if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
    print_error "Docker Compose is not installed"
    exit 1
fi

# Build docker-compose command
CMD="docker-compose logs"

# Add tail option if not following
if [ "$FOLLOW" = false ]; then
    CMD="$CMD --tail=$LINES"
fi

# Add follow option
if [ "$FOLLOW" = true ]; then
    CMD="$CMD -f"
fi

# Add service filter
if [ "$SERVICE" != "all" ]; then
    CMD="$CMD $SERVICE"
fi

# Display header
echo -e "${BLUE}================================================================${NC}"
echo -e "${BLUE}   Container Logs${NC}"
echo -e "${BLUE}================================================================${NC}"
echo ""

if [ "$SERVICE" != "all" ]; then
    print_info "Service: $SERVICE"
else
    print_info "Service: All services"
fi

if [ "$FOLLOW" = true ]; then
    print_info "Mode: Following (live tail)"
else
    print_info "Lines: $LINES"
fi

if [ -n "$SEARCH" ]; then
    print_info "Search: $SEARCH"
fi

echo ""
echo -e "${YELLOW}Press Ctrl+C to exit${NC}"
echo ""

# Execute command
if [ -n "$SEARCH" ]; then
    # With search filter
    $CMD 2>&1 | grep --color=always -i "$SEARCH"
else
    # Without search filter
    $CMD 2>&1
fi
