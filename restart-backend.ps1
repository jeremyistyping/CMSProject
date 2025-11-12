# Quick Restart Backend Script

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Restarting Backend Server" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "⚠️  IMPORTANT: Stop the current backend first (Ctrl+C)" -ForegroundColor Yellow
Write-Host ""
Write-Host "Press any key when backend is stopped, or Ctrl+C to cancel..."
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")

$backendDir = Join-Path $PSScriptRoot "backend"
Set-Location $backendDir

Write-Host ""
Write-Host "📝 Fix Applied:" -ForegroundColor Green
Write-Host "   - Fixed static file handler path joining" -ForegroundColor Gray
Write-Host "   - Added detailed logging for file serving" -ForegroundColor Gray
Write-Host "   - Removed leading slash from filepath param" -ForegroundColor Gray
Write-Host ""
Write-Host "Watch for these messages:" -ForegroundColor Yellow
Write-Host "   📁 Serving static files from: ..." -ForegroundColor Gray
Write-Host "   📝 Serving file: /uploads/... -> C:\...\uploads\..." -ForegroundColor Gray
Write-Host "   ✅ File served successfully: ..." -ForegroundColor Gray
Write-Host ""
Write-Host "Starting backend..." -ForegroundColor Green
Write-Host ""

go run main.go

