@echo off
echo Testing balance sheet fixer build...

cd /d "D:\Project\app_sistem_akuntansi\backend"

echo.
echo Building script...
go build -o scripts\balance_sheet_fixer_test.exe scripts\balance_sheet_fixer.go

if %ERRORLEVEL% neq 0 (
    echo ERROR: Build failed!
    pause
    exit /b 1
)

echo ✅ Build successful!
echo.
echo Script files created:
dir scripts\*.exe
echo.
echo Script is ready to run!
echo To execute: scripts\balance_sheet_fixer_test.exe
echo.
echo Note: Make sure MySQL is running and database exists before running.
pause