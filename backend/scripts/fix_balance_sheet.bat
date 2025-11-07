@echo off
echo === BALANCE SHEET FIXER FOR SSOT JOURNAL ===
echo Script untuk membuat balance sheet seimbang
echo.

cd /d "D:\Project\app_sistem_akuntansi\backend"

echo Building balance sheet fixer...
go build -o scripts\balance_sheet_fixer.exe scripts\balance_sheet_fixer.go

if %ERRORLEVEL% neq 0 (
    echo ERROR: Build failed!
    pause
    exit /b 1
)

echo.
echo Running balance sheet fixer...
echo ============================================================
scripts\balance_sheet_fixer.exe

echo.
echo ============================================================
echo DONE! Next steps:
echo 1. Restart your backend service 
echo 2. Open frontend and generate SSOT Balance Sheet report
echo 3. Verify it shows 'Balanced' status
echo.
pause