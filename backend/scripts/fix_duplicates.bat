@echo off
echo === DUPLICATE JOURNAL FIXER ===
echo Script untuk menghapus duplikat jurnal dan balance sheet
echo.

cd /d "D:\Project\app_sistem_akuntansi\backend"

echo Building duplicate fixer script...
go build -o scripts\fix_duplicate_journals.exe scripts\fix_duplicate_journals.go

if %ERRORLEVEL% neq 0 (
    echo ERROR: Build failed!
    pause
    exit /b 1
)

echo ✅ Build successful!
echo.
echo Running duplicate journal fixer...
echo ============================================================
scripts\fix_duplicate_journals.exe

echo.
echo ============================================================
echo DONE! 
echo.
echo Next steps if balance sheet is now balanced:
echo 1. Restart your backend service 
echo 2. Open frontend and generate SSOT Balance Sheet report
echo 3. Verify it shows 'Balanced' status
echo.
pause