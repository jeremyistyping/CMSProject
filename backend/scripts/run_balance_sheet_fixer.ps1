# Balance Sheet Fixer Runner Script
# Script untuk menjalankan balance sheet fixer dengan mudah

Write-Host "=== BALANCE SHEET FIXER RUNNER ===" -ForegroundColor Green
Write-Host "Script untuk memperbaiki balance sheet SSOT jurnal" -ForegroundColor Yellow
Write-Host ""

# Check if Go is installed
try {
    $goVersion = go version 2>$null
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Go tidak ditemukan. Pastikan Go sudah terinstall." -ForegroundColor Red
        exit 1
    }
    Write-Host "✅ Go version: $goVersion" -ForegroundColor Green
} catch {
    Write-Host "❌ Error checking Go installation: $_" -ForegroundColor Red
    exit 1
}

# Change to backend directory
$backendPath = "D:\Project\app_sistem_akuntansi\backend"
if (!(Test-Path $backendPath)) {
    Write-Host "❌ Backend directory tidak ditemukan: $backendPath" -ForegroundColor Red
    exit 1
}

Set-Location $backendPath
Write-Host "📁 Changed to directory: $backendPath" -ForegroundColor Cyan

# Check if MySQL is running
Write-Host "🔍 Checking MySQL connection..." -ForegroundColor Yellow
try {
    $mysqlTest = mysql -u root -e "SELECT 1;" 2>$null
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ MySQL tidak dapat diakses. Pastikan MySQL running dan bisa diakses dengan user 'root' tanpa password." -ForegroundColor Red
        Write-Host "   Atau edit connection string di balance_sheet_fixer.go" -ForegroundColor Yellow
        exit 1
    }
    Write-Host "✅ MySQL connection OK" -ForegroundColor Green
} catch {
    Write-Host "⚠️  Cannot verify MySQL connection. Proceeding anyway..." -ForegroundColor Yellow
}

# Build and run the fixer
Write-Host ""
Write-Host "🔧 Building balance sheet fixer..." -ForegroundColor Yellow

try {
    go build -o scripts/balance_sheet_fixer.exe scripts/balance_sheet_fixer.go
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Build failed!" -ForegroundColor Red
        exit 1
    }
    Write-Host "✅ Build successful!" -ForegroundColor Green
} catch {
    Write-Host "❌ Error building: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "🚀 Running balance sheet fixer..." -ForegroundColor Yellow
Write-Host "=" * 60 -ForegroundColor Gray

try {
    & .\scripts\balance_sheet_fixer.exe
    if ($LASTEXITCODE -ne 0) {
        Write-Host ""
        Write-Host "⚠️  Fixer completed with warnings. Check output above." -ForegroundColor Yellow
    } else {
        Write-Host ""
        Write-Host "✅ Fixer completed successfully!" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ Error running fixer: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "=" * 60 -ForegroundColor Gray
Write-Host "🎯 Next Steps:" -ForegroundColor Cyan
Write-Host "1. Restart your backend service if it's running" -ForegroundColor White
Write-Host "2. Open the frontend application" -ForegroundColor White  
Write-Host "3. Generate SSOT Balance Sheet report" -ForegroundColor White
Write-Host "4. Verify that it shows 'Balanced' status" -ForegroundColor White

Write-Host ""
Write-Host "📝 Log file saved to: scripts/balance_sheet_fixer.log" -ForegroundColor Gray
Write-Host "Done! Press any key to exit..." -ForegroundColor Green
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")