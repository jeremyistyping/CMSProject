# Test Purchase Report - Fixed Version
Write-Host "PURCHASE REPORT TESTING" -ForegroundColor Cyan
Write-Host "=======================" -ForegroundColor Cyan

# Test 1: Check if Go server is running
Write-Host ""
Write-Host "Testing Go Backend Server..." -ForegroundColor Yellow
try {
    $loginBody = @{
        username = "admin"
        password = "admin123"
    } | ConvertTo-Json
    
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method Post -Body $loginBody -ContentType "application/json" -ErrorAction Stop
    
    if ($response.token) {
        $token = $response.token
        Write-Host "SUCCESS: Backend server is running and authentication works" -ForegroundColor Green
        Write-Host "Token obtained: $($token.Substring(0, 20))..." -ForegroundColor Gray
    } else {
        Write-Host "WARNING: Server running but no token received" -ForegroundColor Yellow
    }
} catch {
    Write-Host "ERROR: Backend server issue: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "SOLUTION: Make sure Go server is running: go run main.go" -ForegroundColor Yellow
    exit 1
}

# Test 2: Test Purchase Report API with different date ranges
Write-Host ""
Write-Host "Testing Purchase Report API..." -ForegroundColor Yellow

$dateRanges = @(
    @{ start = "2025-09-01"; end = "2025-09-30"; name = "September 2025 (Expected Data)" }
    @{ start = "2025-01-01"; end = "2025-12-31"; name = "Full Year 2025" }
    @{ start = "2024-01-01"; end = "2024-12-31"; name = "Year 2024" }
)

$headers = @{
    "Authorization" = "Bearer $token"
}

foreach ($range in $dateRanges) {
    Write-Host ""
    Write-Host "  Testing: $($range.name)" -ForegroundColor Cyan
    try {
        $url = "http://localhost:8080/api/v1/ssot-reports/purchase-report?start_date=$($range.start)&end_date=$($range.end)"
        Write-Host "     URL: $url" -ForegroundColor Gray
        
        $response = Invoke-RestMethod -Uri $url -Method Get -Headers $headers -ErrorAction Stop
        
        if ($response.success) {
            Write-Host "     SUCCESS!" -ForegroundColor Green
            Write-Host "     Total Purchases: $($response.data.total_purchases)" -ForegroundColor White
            Write-Host "     Total Amount: $($response.data.total_amount)" -ForegroundColor White
            Write-Host "     Active Vendors: $($response.data.active_vendors)" -ForegroundColor White
            Write-Host "     Outstanding: $($response.data.outstanding_payables)" -ForegroundColor White
            
            if ($response.data.purchases_by_vendor -and $response.data.purchases_by_vendor.Count -gt 0) {
                Write-Host "     Vendors:" -ForegroundColor White
                foreach ($vendor in $response.data.purchases_by_vendor) {
                    Write-Host "        - $($vendor.vendor_name): $($vendor.total_amount)" -ForegroundColor Gray
                }
            } else {
                Write-Host "     NO VENDORS FOUND" -ForegroundColor Red
            }
        } else {
            Write-Host "     FAILED: $($response.error)" -ForegroundColor Red
        }
    } catch {
        $errorMsg = $_.Exception.Message
        Write-Host "     REQUEST FAILED: $errorMsg" -ForegroundColor Red
        
        if ($_.Exception.Response) {
            $statusCode = $_.Exception.Response.StatusCode.value__
            Write-Host "     HTTP Status: $statusCode" -ForegroundColor Red
        }
    }
}

Write-Host ""
Write-Host "DIAGNOSIS SUMMARY" -ForegroundColor Cyan
Write-Host "=================" -ForegroundColor Cyan

Write-Host "Expected data based on Purchase Management:" -ForegroundColor White
Write-Host "- Purchase PO/2025/09/0036: Rp 5.550.000 (22/9/2025)" -ForegroundColor White  
Write-Host "- Purchase PO/2025/09/0035: Rp 3.885.000 (22/9/2025)" -ForegroundColor White
Write-Host "- Vendor: Jerry Rolo Merentek vendor" -ForegroundColor White
Write-Host "- Status: PAID/APPROVED" -ForegroundColor White

Write-Host ""
Write-Host "If September 2025 test shows 0 purchases:" -ForegroundColor Yellow
Write-Host "1. SSOT Integration Issue - Purchases not in unified_journal_ledger" -ForegroundColor White
Write-Host "2. Service Query Issue - Wrong date filtering or status mapping" -ForegroundColor White  
Write-Host "3. Database Connection Issue" -ForegroundColor White

Write-Host ""
Write-Host "NEXT STEPS" -ForegroundColor Cyan
Write-Host "==========" -ForegroundColor Cyan
Write-Host "1. Check the test results above" -ForegroundColor White
Write-Host "2. If September 2025 shows correct data - Problem solved!" -ForegroundColor Green
Write-Host "3. If September 2025 shows 0 data - Need SSOT sync" -ForegroundColor Red
Write-Host "4. Test Purchase Report in frontend with September date range" -ForegroundColor White

Write-Host ""
Write-Host "FRONTEND TEST RECOMMENDATION" -ForegroundColor Cyan
Write-Host "============================" -ForegroundColor Cyan
Write-Host "Open Purchase Report modal and test with:" -ForegroundColor White
Write-Host "- Start Date: 2025-09-01" -ForegroundColor Yellow
Write-Host "- End Date: 2025-09-30" -ForegroundColor Yellow
Write-Host "- Expected: 2 purchases, Rp 9.435.000 total" -ForegroundColor Green