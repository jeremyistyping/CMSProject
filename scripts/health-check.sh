#!/bin/bash

# ================================================================
# Health Check Script
# ================================================================
# This script checks the health of all services
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

echo -e "${BLUE}================================================================${NC}"
echo -e "${BLUE}   System Health Check${NC}"
echo -e "${BLUE}================================================================${NC}"
echo ""

OVERALL_HEALTH=0

# Check 1: Docker containers status
print_info "Checking container status..."
echo ""

CONTAINERS=("postgres" "backend" "frontend" "nginx")

for container in "${CONTAINERS[@]}"; do
    CONTAINER_NAME=$(docker-compose ps -q $container 2>/dev/null)
    
    if [ -z "$CONTAINER_NAME" ]; then
        print_error "$container: Not running"
        OVERALL_HEALTH=1
        continue
    fi
    
    STATUS=$(docker inspect --format='{{.State.Status}}' $CONTAINER_NAME 2>/dev/null)
    HEALTH=$(docker inspect --format='{{.State.Health.Status}}' $CONTAINER_NAME 2>/dev/null || echo "none")
    
    if [ "$STATUS" = "running" ]; then
        if [ "$HEALTH" = "healthy" ] || [ "$HEALTH" = "none" ]; then
            print_success "$container: Running"
        elif [ "$HEALTH" = "starting" ]; then
            print_warning "$container: Starting..."
            OVERALL_HEALTH=1
        else
            print_error "$container: Unhealthy"
            OVERALL_HEALTH=1
        fi
    else
        print_error "$container: $STATUS"
        OVERALL_HEALTH=1
    fi
done

echo ""

# Check 2: Database connectivity
print_info "Checking database connectivity..."

if docker-compose exec -T postgres pg_isready -U postgres > /dev/null 2>&1; then
    print_success "Database: Connected"
else
    print_error "Database: Connection failed"
    OVERALL_HEALTH=1
fi

echo ""

# Check 3: Backend API health endpoint
print_info "Checking backend API..."

# Get backend container IP or use localhost
BACKEND_URL="http://localhost:8080/health"

if curl -sf "$BACKEND_URL" > /dev/null 2>&1; then
    print_success "Backend API: Healthy"
else
    # Try through docker network
    if docker-compose exec -T backend wget -q -O- http://localhost:8080/health > /dev/null 2>&1; then
        print_success "Backend API: Healthy (internal)"
    else
        print_error "Backend API: Not responding"
        OVERALL_HEALTH=1
    fi
fi

echo ""

# Check 4: Frontend
print_info "Checking frontend..."

FRONTEND_URL="http://localhost:3000"

if curl -sf "$FRONTEND_URL" > /dev/null 2>&1; then
    print_success "Frontend: Accessible"
else
    # Try through docker network
    if docker-compose exec -T frontend wget -q -O- http://localhost:3000 > /dev/null 2>&1; then
        print_success "Frontend: Accessible (internal)"
    else
        print_error "Frontend: Not accessible"
        OVERALL_HEALTH=1
    fi
fi

echo ""

# Check 5: Nginx
print_info "Checking nginx..."

if curl -sf "http://localhost/health" > /dev/null 2>&1; then
    print_success "Nginx: Healthy"
else
    print_error "Nginx: Not responding"
    OVERALL_HEALTH=1
fi

echo ""

# Check 6: Disk space
print_info "Checking disk space..."

DISK_USAGE=$(df -h / | awk 'NR==2 {print $5}' | sed 's/%//')

if [ "$DISK_USAGE" -lt 80 ]; then
    print_success "Disk space: ${DISK_USAGE}% used"
elif [ "$DISK_USAGE" -lt 90 ]; then
    print_warning "Disk space: ${DISK_USAGE}% used (consider cleanup)"
else
    print_error "Disk space: ${DISK_USAGE}% used (critical!)"
    OVERALL_HEALTH=1
fi

echo ""

# Check 7: Docker volumes
print_info "Checking Docker volumes..."

VOLUMES=("postgres_data" "backend_uploads")
for volume in "${VOLUMES[@]}"; do
    if docker volume ls | grep -q "$volume"; then
        VOLUME_SIZE=$(docker system df -v | grep "$volume" | awk '{print $3}')
        print_success "$volume: ${VOLUME_SIZE:-Unknown size}"
    else
        print_warning "$volume: Not found"
    fi
done

echo ""

# Check 8: Memory usage
print_info "Checking memory usage..."

if command -v free &> /dev/null; then
    MEM_USAGE=$(free | grep Mem | awk '{printf "%.0f", $3/$2 * 100}')
    if [ "$MEM_USAGE" -lt 80 ]; then
        print_success "Memory: ${MEM_USAGE}% used"
    elif [ "$MEM_USAGE" -lt 90 ]; then
        print_warning "Memory: ${MEM_USAGE}% used"
    else
        print_error "Memory: ${MEM_USAGE}% used (high!)"
        OVERALL_HEALTH=1
    fi
else
    print_info "Memory: Unable to check"
fi

echo ""

# Summary
echo -e "${BLUE}================================================================${NC}"
if [ $OVERALL_HEALTH -eq 0 ]; then
    echo -e "${GREEN}   All Systems Healthy ✓${NC}"
    echo -e "${BLUE}================================================================${NC}"
    exit 0
else
    echo -e "${RED}   Some Issues Detected ✗${NC}"
    echo -e "${BLUE}================================================================${NC}"
    echo ""
    print_info "Check logs with: ./scripts/logs.sh"
    echo ""
    exit 1
fi
