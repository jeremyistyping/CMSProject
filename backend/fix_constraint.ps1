# Script untuk memperbaiki constraint journal_entries
# Menjalankan endpoint admin untuk force fix constraint

Write-Host "🔧 Fixing journal_entries date constraint..." -ForegroundColor Yellow
Write-Host ""

# Endpoint base URL (adjust if different)
$baseUrl = "http://localhost:8080/api/v1"

# Get JWT token (you need to login first or provide token)
Write-Host "⚠️  Note: You need to be logged in as admin" -ForegroundColor Cyan
Write-Host ""

# Check current constraint
Write-Host "1️⃣  Checking current constraint..." -ForegroundColor Green
$constraintUrl = "$baseUrl/admin/migrations/constraint-info"

try {
    $response = Invoke-RestMethod -Uri $constraintUrl -Method GET -Headers @{
        "Authorization" = "Bearer YOUR_JWT_TOKEN_HERE"
        "Content-Type" = "application/json"
    }
    
    Write-Host "Current constraint:" -ForegroundColor Cyan
    Write-Host $response.constraint_definition -ForegroundColor White
    Write-Host ""
} catch {
    Write-Host "❌ Failed to check constraint. Make sure you're logged in as admin." -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
    Write-Host "To get your JWT token:" -ForegroundColor Yellow
    Write-Host "1. Login via /api/v1/auth/login" -ForegroundColor White
    Write-Host "2. Copy the access_token from response" -ForegroundColor White
    Write-Host "3. Replace YOUR_JWT_TOKEN_HERE in this script" -ForegroundColor White
    exit 1
}

# Run the fix
Write-Host "2️⃣  Running constraint fix..." -ForegroundColor Green
$fixUrl = "$baseUrl/admin/migrations/fix-date-constraint"

try {
    $response = Invoke-RestMethod -Uri $fixUrl -Method POST -Headers @{
        "Authorization" = "Bearer YOUR_JWT_TOKEN_HERE"
        "Content-Type" = "application/json"
    }
    
    Write-Host "✅ Fix completed successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "New constraint:" -ForegroundColor Cyan
    Write-Host $response.constraint_definition -ForegroundColor White
    Write-Host ""
    Write-Host $response.info -ForegroundColor Yellow
    Write-Host ""
    Write-Host "✅ You can now close periods up to year 2099!" -ForegroundColor Green
    
} catch {
    Write-Host "❌ Failed to run constraint fix" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
}
