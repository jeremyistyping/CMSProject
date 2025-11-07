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

param(
    [string]$ServerURL = "http://localhost:8080",
    [switch]$Verbose,
    [switch]$WaitForServer,
    [int]$MaxRetries = 3
)

# Color functions
function Write-Success($message) {
    Write-Host "✅ $message" -ForegroundColor Green
}

function Write-Error($message) {
    Write-Host "❌ $message" -ForegroundColor Red
}

function Write-Warning($message) {
    Write-Host "⚠️  $message" -ForegroundColor Yellow
}

function Write-Info($message) {
    Write-Host "ℹ️  $message" -ForegroundColor Cyan
}

function Write-Header($message) {
    $line = "=" * 60
    Write-Host $line -ForegroundColor Blue
    Write-Host $message -ForegroundColor Blue
    Write-Host $line -ForegroundColor Blue
}

# Check if server is running
function Test-ServerConnection {
    param([string]$url)
    
    try {
        $response = Invoke-WebRequest -Uri "$url/api/v1/health" -Method GET -TimeoutSec 5 -ErrorAction Stop
        return $response.StatusCode -eq 200
    }
    catch {
        return $false
    }
}

# Wait for server to be ready
function Wait-ForServer {
    param([string]$url, [int]$maxWait = 60)
    
    Write-Info "Waiting for server to be ready at $url..."
    
    for ($i = 0; $i -lt $maxWait; $i++) {
        if (Test-ServerConnection $url) {
            Write-Success "Server is ready!"
            return $true
        }
        
        Write-Host "." -NoNewline
        Start-Sleep -Seconds 1
    }
    
    Write-Host ""
    Write-Error "Server is not responding after $maxWait seconds"
    return $false
}

# Main execution
function Main {
    Write-Header "🚀 COMPREHENSIVE SALES-PAYMENT FLOW TEST"
    
    # Check if Go is installed
    try {
        $goVersion = go version
        Write-Info "Go version: $goVersion"
    }
    catch {
        Write-Error "Go is not installed or not in PATH"
        exit 1
    }
    
    # Check if test script exists
    $testScript = "scripts\test_sales_payment_flow.go"
    if (-not (Test-Path $testScript)) {
        Write-Error "Test script not found: $testScript"
        exit 1
    }
    
    # Wait for server if requested
    if ($WaitForServer) {
        if (-not (Wait-ForServer $ServerURL)) {
            Write-Error "Cannot connect to server at $ServerURL"
            Write-Info "Please ensure the server is running with: go run cmd/main.go"
            exit 1
        }
    }
    else {
        # Quick server check
        if (-not (Test-ServerConnection $ServerURL)) {
            Write-Warning "Server appears to be offline at $ServerURL"
            Write-Info "Starting test anyway... (use -WaitForServer to wait for server)"
        }
        else {
            Write-Success "Server is online at $ServerURL"
        }
    }
    
    Write-Header "📋 RUNNING TEST SUITE"
    
    # Run the test with retries
    $attempt = 0
    $success = $false
    
    while ($attempt -lt $MaxRetries -and -not $success) {
        $attempt++
        
        if ($attempt -gt 1) {
            Write-Info "Retry attempt $attempt of $MaxRetries"
            Start-Sleep -Seconds 3
        }
        
        try {
            Write-Info "Executing test script..."
            
            # Set environment variables for the test
            $env:TEST_BASE_URL = $ServerURL + "/api/v1"
            
            # Run the test
            if ($Verbose) {
                $output = go run $testScript 2>&1
                Write-Host $output
                $success = $LASTEXITCODE -eq 0
            }
            else {
                $output = go run $testScript 2>&1
                Write-Host $output
                $success = $LASTEXITCODE -eq 0
            }
            
            if ($success) {
                Write-Success "Test completed successfully!"
            }
            else {
                Write-Error "Test failed with exit code: $LASTEXITCODE"
            }
        }
        catch {
            Write-Error "Error running test: $($_.Exception.Message)"
            $success = $false
        }
    }
    
    Write-Header "📊 TEST SUMMARY"
    
    if ($success) {
        Write-Success "🎉 ALL TESTS PASSED!"
        Write-Host ""
        Write-Host "✅ Sales creation and invoicing" -ForegroundColor Green
        Write-Host "✅ Payment recording and allocation" -ForegroundColor Green  
        Write-Host "✅ Journal entries verification" -ForegroundColor Green
        Write-Host "✅ Account balance updates" -ForegroundColor Green
        Write-Host "✅ Data integrity verification" -ForegroundColor Green
        Write-Host ""
        Write-Success "🚀 System is production-ready!"
    }
    else {
        Write-Error "❌ TESTS FAILED"
        Write-Host ""
        Write-Host "Please check the output above for error details." -ForegroundColor Yellow
        Write-Host "Common issues:" -ForegroundColor Yellow
        Write-Host "- Server not running (use -WaitForServer)" -ForegroundColor Yellow
        Write-Host "- Database connection issues" -ForegroundColor Yellow
        Write-Host "- Missing test data (customers, products)" -ForegroundColor Yellow
        Write-Host "- API endpoint changes" -ForegroundColor Yellow
        exit 1
    }
}

# Script usage help
function Show-Help {
    Write-Host @"
USAGE:
    .\run_test.ps1 [OPTIONS]

OPTIONS:
    -ServerURL <url>     Server URL (default: http://localhost:8080)
    -Verbose            Show detailed output
    -WaitForServer      Wait for server to be ready before testing
    -MaxRetries <n>     Maximum number of retry attempts (default: 3)
    -Help               Show this help message

EXAMPLES:
    .\run_test.ps1
    .\run_test.ps1 -Verbose -WaitForServer
    .\run_test.ps1 -ServerURL http://localhost:8081
    .\run_test.ps1 -MaxRetries 5

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
"@
}

# Handle help parameter
if ($args -contains "-Help" -or $args -contains "/?" -or $args -contains "-h") {
    Show-Help
    exit 0
}

# Run main function
try {
    Main
}
catch {
    Write-Error "Unexpected error: $($_.Exception.Message)"
    Write-Info "Stack trace:"
    Write-Host $_.ScriptStackTrace -ForegroundColor Gray
    exit 1
}