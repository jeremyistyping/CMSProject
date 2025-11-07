# Test Balance Summary API Endpoint
Write-Host "🔍 Testing Cash Bank Balance Summary API..." -ForegroundColor Green

try {
    # Test if backend is responsive
    $healthCheck = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method GET -TimeoutSec 10
    Write-Host "✅ Backend is running" -ForegroundColor Green
} catch {
    Write-Host "❌ Backend is not responding on localhost:8080" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Yellow
    exit 1
}

try {
    # Test balance summary endpoint (without auth for now)
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/cashbank/balance-summary" -Method GET -TimeoutSec 10
    
    Write-Host "📊 Balance Summary Response:" -ForegroundColor Cyan
    Write-Host "Total Cash: $($response.total_cash)" -ForegroundColor Yellow
    Write-Host "Total Bank: $($response.total_bank)" -ForegroundColor Yellow  
    Write-Host "Total Balance: $($response.total_balance)" -ForegroundColor Yellow
    
    if ($response.total_balance -gt 0) {
        Write-Host "✅ SUCCESS: Summary shows non-zero balance!" -ForegroundColor Green
        Write-Host "🎯 Fix is working - balance from COA is being used" -ForegroundColor Green
    } else {
        Write-Host "⚠️  WARNING: Total balance is still 0" -ForegroundColor Yellow
        Write-Host "This might indicate:" -ForegroundColor Yellow
        Write-Host "1. No active cash/bank accounts with COA integration" -ForegroundColor Yellow
        Write-Host "2. All COA account balances are actually 0" -ForegroundColor Yellow
        Write-Host "3. Query JOIN is not finding matching records" -ForegroundColor Yellow
    }
    
    # Display full response for debugging
    Write-Host "`n📋 Full API Response:" -ForegroundColor Cyan
    $response | ConvertTo-Json -Depth 3
    
} catch {
    Write-Host "❌ Failed to call balance summary API" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Yellow
    
    if ($_.Exception.Message -like "*401*" -or $_.Exception.Message -like "*Unauthorized*") {
        Write-Host "💡 This endpoint might require authentication" -ForegroundColor Blue
        Write-Host "Try accessing it through the frontend application" -ForegroundColor Blue
    }
}