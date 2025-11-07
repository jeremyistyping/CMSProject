# Purchase Report Final Test
Write-Host "PURCHASE REPORT FINAL TEST" -ForegroundColor Cyan
Write-Host "==========================" -ForegroundColor Cyan

# Step 1: Get authentication token
Write-Host ""
Write-Host "Step 1: Getting authentication token..." -ForegroundColor Yellow
try {
    $loginBody = @{
        email = "admin@company.com"
        password = "admin123"
    } | ConvertTo-Json
    
    $authResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method Post -Body $loginBody -ContentType "application/json" -ErrorAction Stop
    
    $token = $authResponse.access_token
    Write-Host "SUCCESS: Authentication successful!" -ForegroundColor Green
    Write-Host "User: $($authResponse.user.username) ($($authResponse.user.email))" -ForegroundColor Gray
    Write-Host "Token: $($token.Substring(0, 30))..." -ForegroundColor Gray
    
} catch {
    Write-Host "ERROR: Authentication failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Step 2: Test Purchase Report API
Write-Host ""
Write-Host "Step 2: Testing Purchase Report API..." -ForegroundColor Yellow

$headers = @{
    "Authorization" = "Bearer $token"
}

$testCases = @(
    @{ start = "2025-09-01"; end = "2025-09-30"; name = "September 2025 - TARGET DATA" }
    @{ start = "2025-01-01"; end = "2025-12-31"; name = "Full Year 2025" }
    @{ start = "2024-01-01"; end = "2024-12-31"; name = "Year 2024" }
)

$septemberResult = $null

foreach ($test in $testCases) {
    Write-Host ""
    Write-Host "  Testing: $($test.name)" -ForegroundColor Cyan
    
    $url = "http://localhost:8080/api/v1/ssot-reports/purchase-report?start_date=$($test.start)&end_date=$($test.end)"
    Write-Host "  URL: $url" -ForegroundColor Gray
    
    try {
        $response = Invoke-RestMethod -Uri $url -Method Get -Headers $headers -ErrorAction Stop
        
        if ($response.success) {
            Write-Host "  SUCCESS!" -ForegroundColor Green
            Write-Host "  Total Purchases: $($response.data.total_purchases)" -ForegroundColor White
            Write-Host "  Total Amount: Rp $($response.data.total_amount)" -ForegroundColor White
            Write-Host "  Active Vendors: $($response.data.active_vendors)" -ForegroundColor White
            Write-Host "  Outstanding: Rp $($response.data.outstanding_payables)" -ForegroundColor White
            
            if ($test.start -eq "2025-09-01") {
                $septemberResult = $response.data
            }
            
            if ($response.data.purchases_by_vendor -and $response.data.purchases_by_vendor.Count -gt 0) {
                Write-Host "  Vendor Details:" -ForegroundColor White
                foreach ($vendor in $response.data.purchases_by_vendor) {
                    Write-Host "    Vendor: $($vendor.vendor_name)" -ForegroundColor Gray
                    Write-Host "    Amount: Rp $($vendor.total_amount)" -ForegroundColor Gray
                    Write-Host "    Purchases: $($vendor.purchase_count)" -ForegroundColor Gray
                    Write-Host "    ---" -ForegroundColor Gray
                }
            } else {
                Write-Host "  NO VENDOR DATA FOUND!" -ForegroundColor Red
            }
            
            # Save response for detailed analysis
            $filename = "purchase_report_$($test.start)_to_$($test.end).json"
            $response | ConvertTo-Json -Depth 10 | Out-File -FilePath $filename -Encoding UTF8
            Write-Host "  Response saved: $filename" -ForegroundColor Gray
            
        } else {
            Write-Host "  API ERROR: $($response.error)" -ForegroundColor Red
        }
        
    } catch {
        Write-Host "  REQUEST FAILED: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            $statusCode = $_.Exception.Response.StatusCode.value__
            Write-Host "  HTTP Status: $statusCode" -ForegroundColor Red
        }
    }
}

# Step 3: Analysis
Write-Host ""
Write-Host "DETAILED ANALYSIS" -ForegroundColor Cyan
Write-Host "=================" -ForegroundColor Cyan

Write-Host "Expected September 2025 Data:" -ForegroundColor White
Write-Host "- PO/2025/09/0036: Rp 5.550.000 (Jerry Rolo Merentek)" -ForegroundColor Yellow
Write-Host "- PO/2025/09/0035: Rp 3.885.000 (Jerry Rolo Merentek)" -ForegroundColor Yellow
Write-Host "- Total Expected: Rp 9.435.000" -ForegroundColor Yellow
Write-Host "- Expected Vendor Count: 1 (Jerry Rolo Merentek vendor)" -ForegroundColor Yellow

if ($septemberResult) {
    Write-Host ""
    Write-Host "SEPTEMBER 2025 API RESULTS:" -ForegroundColor Green
    Write-Host "- API Total Purchases: $($septemberResult.total_purchases)" -ForegroundColor White
    Write-Host "- API Total Amount: Rp $($septemberResult.total_amount)" -ForegroundColor White
    Write-Host "- API Active Vendors: $($septemberResult.active_vendors)" -ForegroundColor White
    
    if ($septemberResult.total_purchases -eq 0) {
        Write-Host ""
        Write-Host "PROBLEM: NO SEPTEMBER PURCHASES FOUND!" -ForegroundColor Red
        Write-Host "This means one of the following issues:" -ForegroundColor Red
        Write-Host "1. Purchases exist in 'purchases' table but not in 'unified_journal_ledger'" -ForegroundColor White
        Write-Host "2. SSOT integration is not working correctly" -ForegroundColor White
        Write-Host "3. Date filtering in the service is incorrect" -ForegroundColor White
        Write-Host "4. Purchase status filtering is excluding valid purchases" -ForegroundColor White
        
        Write-Host ""
        Write-Host "IMMEDIATE ACTIONS NEEDED:" -ForegroundColor Yellow
        Write-Host "1. Check database manually with: psql -d sistem_akuntansi -f check_purchase_data.sql" -ForegroundColor White
        Write-Host "2. Verify purchases exist in purchases table" -ForegroundColor White
        Write-Host "3. Check if unified_journal_ledger has PURCHASE entries" -ForegroundColor White
        Write-Host "4. Run SSOT sync if needed" -ForegroundColor White
        
    } elseif ($septemberResult.total_amount -eq 9435000) {
        Write-Host ""
        Write-Host "PERFECT! Data matches expected values!" -ForegroundColor Green
        Write-Host "✓ Total amount matches: Rp 9.435.000" -ForegroundColor Green
        Write-Host "✓ Purchases found: $($septemberResult.total_purchases)" -ForegroundColor Green
        
        Write-Host ""
        Write-Host "NEXT: Test Frontend Purchase Report Modal" -ForegroundColor Cyan
        Write-Host "1. Open your React app (port 3001)" -ForegroundColor White
        Write-Host "2. Navigate to Purchase Report" -ForegroundColor White
        Write-Host "3. Set date range: 2025-09-01 to 2025-09-30" -ForegroundColor White
        Write-Host "4. Verify the data matches the API response" -ForegroundColor White
        
    } else {
        Write-Host ""
        Write-Host "PARTIAL SUCCESS: Data found but amounts don't match!" -ForegroundColor Yellow
        Write-Host "Expected: Rp 9.435.000" -ForegroundColor Yellow
        Write-Host "Actual: Rp $($septemberResult.total_amount)" -ForegroundColor Yellow
        Write-Host "This could be due to:" -ForegroundColor Yellow
        Write-Host "- Different purchase statuses" -ForegroundColor White
        Write-Host "- Additional/missing purchases" -ForegroundColor White  
        Write-Host "- Data synchronization issues" -ForegroundColor White
    }
} else {
    Write-Host ""
    Write-Host "ERROR: Could not get September 2025 data!" -ForegroundColor Red
    Write-Host "Check the API responses above for details." -ForegroundColor Red
}

Write-Host ""
Write-Host "DATABASE CHECK AVAILABLE:" -ForegroundColor Cyan
Write-Host "File: check_purchase_data.sql" -ForegroundColor Gray
Write-Host "Run: psql -U postgres -d sistem_akuntansi -f check_purchase_data.sql" -ForegroundColor Gray