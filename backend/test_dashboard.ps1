# Test Dashboard Analytics
Write-Host "Testing Dashboard Analytics..." -ForegroundColor Yellow

# Login
$loginBody = @{
    email = "admin@company.com"
    password = "password123"
} | ConvertTo-Json

$loginHeaders = @{ "Content-Type" = "application/json" }

try {
    $loginResponse = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/auth/login" -Method Post -Headers $loginHeaders -Body $loginBody
    $loginData = $loginResponse.Content | ConvertFrom-Json
    $token = $loginData.access_token

    $authHeaders = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }

    # Test different dashboard endpoints
    Write-Host "1. Testing /api/v1/dashboard/analytics..." -ForegroundColor Green
    try {
        $analyticsResponse = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/dashboard/analytics" -Method Get -Headers $authHeaders
        $analyticsData = $analyticsResponse.Content | ConvertFrom-Json
        
        Write-Host "✅ Dashboard Analytics:" -ForegroundColor Green
        Write-Host "   Total Sales: $($analyticsData.totalSales)" -ForegroundColor Cyan
        Write-Host "   Total Purchases: $($analyticsData.totalPurchases)" -ForegroundColor Cyan
        Write-Host "   Accounts Receivable: $($analyticsData.accountsReceivable)" -ForegroundColor Cyan
        Write-Host "   Accounts Payable: $($analyticsData.accountsPayable)" -ForegroundColor Cyan
    }
    catch {
        Write-Host "❌ Analytics failed: $($_.Exception.Message)" -ForegroundColor Red
    }

    Write-Host "`n2. Testing /api/v1/dashboard/summary..." -ForegroundColor Green
    try {
        $summaryResponse = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/dashboard/summary" -Method Get -Headers $authHeaders
        $summaryData = $summaryResponse.Content | ConvertFrom-Json
        
        Write-Host "✅ Dashboard Summary retrieved" -ForegroundColor Green
        Write-Host "   Data keys: $($summaryData.data.PSObject.Properties.Name -join ', ')" -ForegroundColor Cyan
    }
    catch {
        Write-Host "❌ Summary failed: $($_.Exception.Message)" -ForegroundColor Red
    }

    Write-Host "`n3. Testing /api/v1/dashboard/quick-stats..." -ForegroundColor Green
    try {
        $quickStatsResponse = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/dashboard/quick-stats" -Method Get -Headers $authHeaders
        $quickStatsData = $quickStatsResponse.Content | ConvertFrom-Json
        
        Write-Host "✅ Quick Stats:" -ForegroundColor Green
        Write-Host "   Response: $($quickStatsResponse.Content.Substring(0, [Math]::Min(200, $quickStatsResponse.Content.Length)))..." -ForegroundColor Cyan
    }
    catch {
        Write-Host "❌ Quick stats failed: $($_.Exception.Message)" -ForegroundColor Red
    }

    # Check sales table directly
    Write-Host "`n4. Testing sales data directly..." -ForegroundColor Green
    try {
        $salesResponse = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/sales?page=1&limit=5" -Method Get -Headers $authHeaders
        $salesData = $salesResponse.Content | ConvertFrom-Json
        
        Write-Host "✅ Sales Data:" -ForegroundColor Green
        Write-Host "   Total sales records: $($salesData.total)" -ForegroundColor Cyan
        if ($salesData.data -and $salesData.data.Length -gt 0) {
            foreach ($sale in $salesData.data) {
                Write-Host "   - Sale ID $($sale.id): Rp $($sale.total_amount) (Outstanding: Rp $($sale.outstanding_amount))" -ForegroundColor White
            }
        }
    }
    catch {
        Write-Host "❌ Sales data failed: $($_.Exception.Message)" -ForegroundColor Red
    }

}
catch {
    Write-Host "❌ Login failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`nTest completed." -ForegroundColor Yellow
