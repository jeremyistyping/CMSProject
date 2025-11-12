# Complete Photo Fix Script
# This script:
# 1. Fixes old photo URLs in database
# 2. Restarts backend
# 3. Shows test instructions

Write-Host ""
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host "  Complete Daily Update Photo Fix" -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host ""

$backendDir = Join-Path $PSScriptRoot "backend"

# Step 1: Fix database URLs
Write-Host "📝 Step 1: Fixing old photo URLs in database..." -ForegroundColor Green
Write-Host ""

Set-Location $backendDir
go run ./cmd/scripts/fix_old_photo_urls.go

if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "❌ Database fix failed!" -ForegroundColor Red
    Write-Host "Press any key to exit..."
    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
    exit 1
}

Write-Host ""
Write-Host "✅ Database URLs fixed!" -ForegroundColor Green
Write-Host ""
Write-Host "Press any key to continue to backend restart..." -ForegroundColor Yellow
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")

# Step 2: Restart backend
Write-Host ""
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host "  Starting Backend Server" -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "📌 The backend will now start." -ForegroundColor Yellow
Write-Host "📌 After backend starts, refresh your browser page." -ForegroundColor Yellow
Write-Host ""
Write-Host "What to check:" -ForegroundColor Green
Write-Host "  1. Old photos should now display correctly" -ForegroundColor Gray
Write-Host "  2. New photos should continue to work" -ForegroundColor Gray
Write-Host "  3. Browser console should show:" -ForegroundColor Gray
Write-Host "     'Converting relative URL to absolute:'" -ForegroundColor DarkGray
Write-Host "     'uploads\\daily-updates\\... -> http://localhost:8080/uploads/daily-updates/...'" -ForegroundColor DarkGray
Write-Host ""
Write-Host "Press Ctrl+C to stop the server" -ForegroundColor Yellow
Write-Host ""

# Run backend
go run main.go

