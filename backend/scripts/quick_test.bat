@echo off
REM ===============================================
REM QUICK SALES-PAYMENT FLOW TEST
REM ===============================================

echo.
echo 🚀 Starting Quick Sales-Payment Flow Test...
echo ===============================================

REM Check if Go is installed
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo ❌ Go is not installed or not in PATH
    echo Please install Go from https://golang.org/doc/install
    pause
    exit /b 1
)

REM Check if server is running
echo 🔍 Checking server status...
powershell -Command "try { $response = Invoke-WebRequest -Uri 'http://localhost:8080/api/v1/health' -Method GET -TimeoutSec 5; Write-Host '✅ Server is online' -ForegroundColor Green } catch { Write-Host '⚠️  Server appears to be offline - continuing anyway' -ForegroundColor Yellow }"

echo.
echo 📋 Running test script...
echo ===============================================

REM Run the test
go run scripts/test_sales_payment_flow.go

REM Check result
if %ERRORLEVEL% EQU 0 (
    echo.
    echo ===============================================
    echo 🎉 TEST COMPLETED SUCCESSFULLY!
    echo.
    echo ✅ All components working correctly:
    echo    - Sales creation and invoicing
    echo    - Payment recording and allocation
    echo    - Account balance updates
    echo    - Data integrity maintained
    echo.
    echo 🚀 System is ready for production!
    echo ===============================================
) else (
    echo.
    echo ===============================================
    echo ❌ TEST FAILED!
    echo.
    echo Please check the output above for errors.
    echo Common issues:
    echo - Server not running: go run cmd/main.go
    echo - Database connection problems
    echo - Missing test data
    echo ===============================================
    pause
    exit /b %ERRORLEVEL%
)

pause