@echo off
REM ================================================================
REM Backup and Reset Database Script for Windows
REM ================================================================
REM This script will:
REM 1. Backup the current database
REM 2. Reset the database
REM ================================================================

echo.
echo ========================================
echo   Backup and Reset Database
echo ========================================
echo.

REM Check if .env file exists
if not exist "..\..\..env" (
    echo ERROR: .env file not found in backend folder
    echo Please create .env file first
    pause
    exit /b 1
)

REM Load database URL from .env
for /f "tokens=1,2 delims==" %%a in ('findstr /r "^DATABASE_URL=" ..\..\..env') do set DATABASE_URL=%%b

if "%DATABASE_URL%"=="" (
    echo ERROR: DATABASE_URL not found in .env file
    pause
    exit /b 1
)

REM Parse database name from URL
REM Format: postgres://user:pass@host:port/dbname?options
for /f "tokens=5 delims=/:?" %%a in ("%DATABASE_URL%") do set DB_NAME=%%a

if "%DB_NAME%"=="" (
    echo ERROR: Could not parse database name from DATABASE_URL
    pause
    exit /b 1
)

REM Create backup directory if not exists
if not exist "..\..\backups" mkdir "..\..\backups"

REM Generate backup filename with timestamp
for /f "tokens=2 delims==" %%a in ('wmic OS Get localdatetime /value') do set "dt=%%a"
set "YYYY=%dt:~0,4%"
set "MM=%dt:~4,2%"
set "DD=%dt:~6,2%"
set "HH=%dt:~8,2%"
set "Min=%dt:~10,2%"
set "Sec=%dt:~12,2%"
set BACKUP_FILE=..\..\backups\%DB_NAME%_%YYYY%-%MM%-%DD%_%HH%-%Min%-%Sec%.sql

echo Database: %DB_NAME%
echo Backup file: %BACKUP_FILE%
echo.

REM Ask for confirmation
set /p CONFIRM="Do you want to backup and reset? (yes/no): "
if /i not "%CONFIRM%"=="yes" (
    echo Operation cancelled
    pause
    exit /b 0
)

echo.
echo Step 1: Creating backup...
echo.

REM Backup database using pg_dump
pg_dump -U postgres -d %DB_NAME% -f %BACKUP_FILE%

if %ERRORLEVEL% neq 0 (
    echo ERROR: Backup failed!
    echo Make sure PostgreSQL is installed and pg_dump is in PATH
    pause
    exit /b 1
)

echo Backup created successfully: %BACKUP_FILE%
echo.

echo Step 2: Resetting database...
echo.

REM Run reset script
go run main.go

echo.
echo ========================================
echo   Backup and Reset Complete
echo ========================================
echo.
echo Backup saved to: %BACKUP_FILE%
echo.
echo To restore from backup:
echo   psql -U postgres -d %DB_NAME% -f %BACKUP_FILE%
echo.

pause
