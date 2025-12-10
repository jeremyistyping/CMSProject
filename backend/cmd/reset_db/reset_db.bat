@echo off
REM ================================================================
REM Database Reset Script for Windows
REM ================================================================
REM This script will reset the database by dropping all tables,
REM sequences, and custom types.
REM
REM WARNING: THIS WILL DELETE ALL DATA!
REM ================================================================

echo.
echo ========================================
echo   Database Reset Tool
echo ========================================
echo.
echo WARNING: This will DELETE ALL DATA!
echo.

REM Check if .env file exists
if not exist "..\..\..env" (
    echo ERROR: .env file not found in backend folder
    echo Please create .env file first
    pause
    exit /b 1
)

echo Running database reset...
echo.

REM Run the Go script
go run main.go

echo.
echo ========================================
echo   Reset Complete
echo ========================================
echo.
echo Next steps:
echo 1. Go to backend folder: cd ..\..
echo 2. Run application: go run main.go
echo.

pause
