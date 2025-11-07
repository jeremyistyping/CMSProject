# Check Payments in Database
Write-Host "Checking Payment Status in Database..." -ForegroundColor Yellow

# Login to get token
$loginBody = @{
    email = "admin@company.com"
    password = "password123"
} | ConvertTo-Json

$loginHeaders = @{
    "Content-Type" = "application/json"
}

try {
    Write-Host "1. Logging in..." -ForegroundColor Green
    $loginResponse = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/auth/login" -Method Post -Headers $loginHeaders -Body $loginBody
    $loginData = $loginResponse.Content | ConvertFrom-Json
    $token = $loginData.access_token
    Write-Host "✅ Login successful" -ForegroundColor Green

    # Get payments list
    $authHeaders = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }

    Write-Host "2. Getting current payments..." -ForegroundColor Green
    $paymentsResponse = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/payments?page=1&limit=20" -Method Get -Headers $authHeaders
    $paymentsData = $paymentsResponse.Content | ConvertFrom-Json

    Write-Host "✅ Found $($paymentsData.total) payments total" -ForegroundColor Green
    
    if ($paymentsData.data -and $paymentsData.data.Length -gt 0) {
        Write-Host "`nCurrent Payments in Database:" -ForegroundColor Cyan
        Write-Host "ID`tCode`t`t`tAmount`t`tStatus`t`tDate" -ForegroundColor White
        Write-Host "---`t----`t`t`t------`t`t------`t`t----" -ForegroundColor White
        
        foreach ($payment in $paymentsData.data) {
            $statusColor = if ($payment.status -eq "COMPLETED") { "Green" } elseif ($payment.status -eq "FAILED") { "Red" } else { "Yellow" }
            Write-Host "$($payment.id)`t$($payment.code)`t`tRp $($payment.amount)`t$($payment.status)`t`t$($payment.date)" -ForegroundColor $statusColor
        }
        
        # Check specifically for deleted payments (if any show up)
        $deletedCount = ($paymentsData.data | Where-Object { $_.status -eq "FAILED" }).Count
        if ($deletedCount -gt 0) {
            Write-Host "`n⚠️  Found $deletedCount payment(s) with FAILED status (may be deleted)" -ForegroundColor Red
        }
    } else {
        Write-Host "❌ No payments found in database" -ForegroundColor Red
    }
    
    # Also check dashboard stats
    Write-Host "`n3. Checking dashboard analytics..." -ForegroundColor Green
    try {
        $dashboardResponse = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/payments/analytics?start_date=2025-01-01&end_date=2025-12-31" -Method Get -Headers $authHeaders
        $dashboardData = $dashboardResponse.Content | ConvertFrom-Json
        
        Write-Host "✅ Dashboard Analytics:" -ForegroundColor Green
        Write-Host "   Total Received: Rp $($dashboardData.total_received)" -ForegroundColor Cyan
        Write-Host "   Total Paid: Rp $($dashboardData.total_paid)" -ForegroundColor Cyan
        Write-Host "   Net Flow: Rp $($dashboardData.net_flow)" -ForegroundColor Cyan
    }
    catch {
        Write-Host "❌ Could not get dashboard analytics" -ForegroundColor Yellow
    }

}
catch {
    Write-Host "❌ Error: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`nCheck completed." -ForegroundColor Yellow
