@echo off
echo ================================================================
echo 🛡️  BALANCE PROTECTION SETUP
echo ================================================================
echo.
echo This script will setup automatic balance synchronization system
echo to prevent balance mismatch issues in the accounting system.
echo.
echo What this does:
echo   ✅ Install database triggers for auto-sync
echo   ✅ Install monitoring system  
echo   ✅ Install manual sync functions
echo   ✅ Fix any existing balance issues
echo.
echo ================================================================
echo.

REM Check if Go is installed
go version >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Go is not installed or not in PATH
    echo Please install Go first: https://golang.org/dl/
    pause
    exit /b 1
)

REM Check if .env file exists
if not exist ".env" (
    echo ❌ .env file not found
    echo Please make sure you're in the backend directory with .env file
    pause
    exit /b 1
)

echo 🚀 Running balance protection setup...
echo.

REM Step 1: Create balance sync system (reads DATABASE_URL from .env)
go run cmd/scripts/create_balance_sync_system.go

if %errorlevel% neq 0 (
    echo ⚠️  Creation failed. Attempting to grant DB permissions using CURRENT_USER from .env...
    go run cmd/scripts/grant_db_permissions.go
    echo 🔁 Retrying creation...
    go run cmd/scripts/create_balance_sync_system.go
)

REM Step 2: Verify installation status
go run cmd/scripts/verify_system_status.go

echo.
echo ================================================================
echo ✅ Setup process finished. See status above.
echo ================================================================
echo.
echo 💡 What's available:
echo   • Automatic balance sync triggers
echo   • Real-time monitoring system
echo   • Manual sync functions
echo   • Performance optimizations
echo.
echo 🔧 Manual SQL (replace with your tool or psql):
echo   • Health check:    SELECT * FROM account_balance_monitoring WHERE status='MISMATCH';
echo   • Manual sync:     SELECT * FROM sync_account_balances();
echo.
echo 📚 For more info, read: BALANCE_PREVENTION_GUIDE.md
echo.

pause
