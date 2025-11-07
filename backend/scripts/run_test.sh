#!/bin/bash

# ===============================================
# COMPREHENSIVE SALES-PAYMENT FLOW TEST RUNNER
# ===============================================
# 
# This script runs a comprehensive test of the sales-to-payment flow
# to ensure 100% data integrity in the accounting system.
#
# Test Coverage:
# - Sales creation and invoicing
# - Payment recording with allocation
# - Journal entries verification
# - Account balance updates
# - Data integrity checks
# ===============================================

set -e  # Exit on any error

# Default configuration
SERVER_URL="http://localhost:8080"
VERBOSE=false
WAIT_FOR_SERVER=false
MAX_RETRIES=3
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
TEST_SCRIPT="$SCRIPT_DIR/test_sales_payment_flow.go"

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Utility functions
print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

print_info() {
    echo -e "${CYAN}ℹ️  $1${NC}"
}

print_header() {
    local line="============================================================"
    echo -e "${BLUE}$line${NC}"
    echo -e "${BLUE}$1${NC}"
    echo -e "${BLUE}$line${NC}"
}

# Show usage help
show_help() {
    cat << EOF
USAGE:
    ./run_test.sh [OPTIONS]

OPTIONS:
    -u, --url URL           Server URL (default: http://localhost:8080)
    -v, --verbose          Show detailed output
    -w, --wait-server      Wait for server to be ready before testing
    -r, --max-retries N    Maximum number of retry attempts (default: 3)
    -h, --help             Show this help message

EXAMPLES:
    ./run_test.sh
    ./run_test.sh --verbose --wait-server
    ./run_test.sh --url http://localhost:8081
    ./run_test.sh --max-retries 5

DESCRIPTION:
    This script runs comprehensive tests to verify the sales-to-payment flow:
    
    1. Creates a sales order and converts to invoice
    2. Records a payment against the invoice
    3. Verifies journal entries are created properly
    4. Checks that account balances are updated correctly:
       - Receivable account decreases (piutang berkurang)
       - Cash/Bank account increases (kas/bank bertambah)
       - Revenue account increases (pendapatan bertambah)
    5. Validates 100% data integrity

    The test ensures that your accounting system maintains perfect
    data consistency across sales, payments, and general ledger.

EOF
}

# Check if server is running
test_server_connection() {
    local url="$1"
    
    if curl -s -f --max-time 5 "$url/api/v1/health" > /dev/null 2>&1; then
        return 0
    else
        return 1
    fi
}

# Wait for server to be ready
wait_for_server() {
    local url="$1"
    local max_wait="${2:-60}"
    
    print_info "Waiting for server to be ready at $url..."
    
    for ((i=0; i<max_wait; i++)); do
        if test_server_connection "$url"; then
            print_success "Server is ready!"
            return 0
        fi
        
        echo -n "."
        sleep 1
    done
    
    echo ""
    print_error "Server is not responding after $max_wait seconds"
    return 1
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -u|--url)
                SERVER_URL="$2"
                shift 2
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -w|--wait-server)
                WAIT_FOR_SERVER=true
                shift
                ;;
            -r|--max-retries)
                MAX_RETRIES="$2"
                shift 2
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                print_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# Main execution function
main() {
    print_header "🚀 COMPREHENSIVE SALES-PAYMENT FLOW TEST"
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed or not in PATH"
        print_info "Please install Go from https://golang.org/doc/install"
        exit 1
    fi
    
    local go_version
    go_version=$(go version)
    print_info "Go version: $go_version"
    
    # Check if test script exists
    if [[ ! -f "$TEST_SCRIPT" ]]; then
        print_error "Test script not found: $TEST_SCRIPT"
        exit 1
    fi
    
    # Wait for server if requested
    if [[ "$WAIT_FOR_SERVER" == true ]]; then
        if ! wait_for_server "$SERVER_URL"; then
            print_error "Cannot connect to server at $SERVER_URL"
            print_info "Please ensure the server is running with: go run cmd/main.go"
            exit 1
        fi
    else
        # Quick server check
        if ! test_server_connection "$SERVER_URL"; then
            print_warning "Server appears to be offline at $SERVER_URL"
            print_info "Starting test anyway... (use --wait-server to wait for server)"
        else
            print_success "Server is online at $SERVER_URL"
        fi
    fi
    
    print_header "📋 RUNNING TEST SUITE"
    
    # Run the test with retries
    local attempt=0
    local success=false
    
    while [[ $attempt -lt $MAX_RETRIES && "$success" == false ]]; do
        ((attempt++))
        
        if [[ $attempt -gt 1 ]]; then
            print_info "Retry attempt $attempt of $MAX_RETRIES"
            sleep 3
        fi
        
        print_info "Executing test script..."
        
        # Set environment variables for the test
        export TEST_BASE_URL="$SERVER_URL/api/v1"
        
        # Run the test
        local exit_code=0
        if [[ "$VERBOSE" == true ]]; then
            go run "$TEST_SCRIPT" || exit_code=$?
        else
            go run "$TEST_SCRIPT" 2>&1 || exit_code=$?
        fi
        
        if [[ $exit_code -eq 0 ]]; then
            print_success "Test completed successfully!"
            success=true
        else
            print_error "Test failed with exit code: $exit_code"
        fi
    done
    
    print_header "📊 TEST SUMMARY"
    
    if [[ "$success" == true ]]; then
        print_success "🎉 ALL TESTS PASSED!"
        echo ""
        print_success "✅ Sales creation and invoicing"
        print_success "✅ Payment recording and allocation"
        print_success "✅ Journal entries verification"
        print_success "✅ Account balance updates"
        print_success "✅ Data integrity verification"
        echo ""
        print_success "🚀 System is production-ready!"
    else
        print_error "❌ TESTS FAILED"
        echo ""
        print_warning "Please check the output above for error details."
        print_warning "Common issues:"
        print_warning "- Server not running (use --wait-server)"
        print_warning "- Database connection issues"
        print_warning "- Missing test data (customers, products)"
        print_warning "- API endpoint changes"
        exit 1
    fi
}

# Check dependencies
check_dependencies() {
    local missing_deps=()
    
    # Check for required commands
    if ! command -v curl &> /dev/null; then
        missing_deps+=("curl")
    fi
    
    if ! command -v go &> /dev/null; then
        missing_deps+=("go")
    fi
    
    if [[ ${#missing_deps[@]} -gt 0 ]]; then
        print_error "Missing required dependencies: ${missing_deps[*]}"
        print_info "Please install the missing dependencies and try again."
        
        # Provide installation hints
        for dep in "${missing_deps[@]}"; do
            case $dep in
                curl)
                    print_info "Install curl: sudo apt-get install curl (Ubuntu/Debian) or brew install curl (macOS)"
                    ;;
                go)
                    print_info "Install Go: https://golang.org/doc/install"
                    ;;
            esac
        done
        
        exit 1
    fi
}

# Handle script interruption
cleanup() {
    print_warning "Test interrupted by user"
    exit 130
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

# Main execution
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    # Parse command line arguments
    parse_args "$@"
    
    # Check dependencies
    check_dependencies
    
    # Run main function
    main
fi