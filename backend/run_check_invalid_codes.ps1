#!/usr/bin/env pwsh

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "     CHECK INVALID ACCOUNT CODES" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

# Run the check script
Write-Host "Running check script..." -ForegroundColor Yellow
go run scripts/check_invalid_account_codes.go

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "✅ Check completed successfully!" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "⚠️  Found issues - see output above" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "================================================" -ForegroundColor Cyan
