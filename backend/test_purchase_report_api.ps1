# Test Purchase Report API and analyze response
# This script will test the API endpoint and show us what data is being returned

$baseUrl = "http://localhost:8080/api/v1/ssot-reports"
$endpoint = "$baseUrl/purchase-report"

# Get authentication token first (if needed)
# You may need to replace this with actual token
$token = "your-auth-token-here"

# Test parameters
$startDate = "2025-01-01"
$endDate = "2025-12-31"

Write-Host "🔍 TESTING PURCHASE REPORT API" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan

# Test 1: Basic API health check
Write-Host "`n📡 Testing API Health..." -ForegroundColor Yellow
try {
    $healthResponse = Invoke-RestMethod -Uri "$baseUrl/health" -Method Get -ErrorAction SilentlyContinue
    Write-Host "✅ API Health: OK" -ForegroundColor Green
    $healthResponse | ConvertTo-Json -Depth 2
} catch {
    Write-Host "❌ API Health: Failed - $($_.Exception.Message)" -ForegroundColor Red
}

# Test 2: Purchase Report without auth (to see auth error)
Write-Host "`n📊 Testing Purchase Report (No Auth)..." -ForegroundColor Yellow
try {
    $url = "${endpoint}?start_date=$startDate&end_date=$endDate"
    Write-Host "URL: $url" -ForegroundColor Gray
    
    $response = Invoke-RestMethod -Uri $url -Method Get -ErrorAction SilentlyContinue
    Write-Host "✅ Purchase Report Response:" -ForegroundColor Green
    $response | ConvertTo-Json -Depth 5
} catch {
    Write-Host "❌ Purchase Report Failed: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $statusCode = $_.Exception.Response.StatusCode
        Write-Host "HTTP Status: $statusCode" -ForegroundColor Red
        
        # Try to read error response body
        try {
            $errorResponse = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($errorResponse)
            $errorBody = $reader.ReadToEnd()
            Write-Host "Error Body: $errorBody" -ForegroundColor Red
        } catch {
            Write-Host "Could not read error response body" -ForegroundColor Red
        }
    }
}

# Test 3: Different date range to see if there's data in other periods
Write-Host "`n📅 Testing Different Date Ranges..." -ForegroundColor Yellow

$dateRanges = @(
    @{ start = "2024-01-01"; end = "2024-12-31"; name = "2024 Full Year" }
    @{ start = "2025-09-01"; end = "2025-09-30"; name = "September 2025" }
    @{ start = "2025-01-01"; end = "2025-09-30"; name = "Jan-Sep 2025" }
)

foreach ($range in $dateRanges) {
    Write-Host "`n  Testing: $($range.name)" -ForegroundColor Cyan
    try {
        $url = "${endpoint}?start_date=$($range.start)&end_date=$($range.end)"
        $response = Invoke-RestMethod -Uri $url -Method Get -ErrorAction SilentlyContinue
        
        if ($response.success) {
            Write-Host "  ✅ Success for $($range.name)" -ForegroundColor Green
            Write-Host "  Total Purchases: $($response.data.total_purchases)" -ForegroundColor White
            Write-Host "  Total Amount: $($response.data.total_amount)" -ForegroundColor White
            Write-Host "  Outstanding: $($response.data.outstanding_payables)" -ForegroundColor White
            Write-Host "  Vendors Count: $($response.data.purchases_by_vendor.Count)" -ForegroundColor White
        } else {
            Write-Host "  ❌ Failed for $($range.name): $($response.error)" -ForegroundColor Red
        }
    } catch {
        Write-Host "  ❌ Request failed for $($range.name): $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 4: Check if server is responding to simpler endpoints
Write-Host "`n🏥 Testing Other SSOT Endpoints..." -ForegroundColor Yellow

$testEndpoints = @(
    "/health"
    "/purchase-summary?start_date=2025-09-01&end_date=2025-09-30"
    "/purchase-report/validate?start_date=2025-09-01&end_date=2025-09-30"
)

foreach ($testEndpoint in $testEndpoints) {
    try {
        Write-Host "`n  Testing: $baseUrl$testEndpoint" -ForegroundColor Cyan
        $response = Invoke-RestMethod -Uri "$baseUrl$testEndpoint" -Method Get -ErrorAction SilentlyContinue
        Write-Host "  ✅ Response received" -ForegroundColor Green
        
        # Show key response data
        if ($response.success) {
            Write-Host "  Success: $($response.message)" -ForegroundColor White
        } elseif ($response.data) {
            Write-Host "  Data keys: $($response.PSObject.Properties.Name -join ', ')" -ForegroundColor White
        } else {
            Write-Host "  Response: $(($response | ConvertTo-Json -Depth 1).Substring(0, [Math]::Min(100, ($response | ConvertTo-Json -Depth 1).Length)))" -ForegroundColor White
        }
    } catch {
        Write-Host "  ❌ Failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host "`n🎯 ANALYSIS SUMMARY" -ForegroundColor Cyan
Write-Host "==================" -ForegroundColor Cyan
Write-Host "Based on the screenshot data:" -ForegroundColor White
Write-Host "• Outstanding Payables: Rp 5,550,000 (likely from account balance)" -ForegroundColor White  
Write-Host "• Total Purchases: Rp 2 (suspiciously low - query issue?)" -ForegroundColor White
Write-Host "• All other metrics: 0 (no purchase data detected)" -ForegroundColor White
Write-Host ""
Write-Host "Recommendations:" -ForegroundColor Yellow
Write-Host "1. Check if unified_journal_ledger has PURCHASE entries" -ForegroundColor White
Write-Host "2. Verify date range has actual purchase data" -ForegroundColor White  
Write-Host "3. Check service logic for vendor detection" -ForegroundColor White
Write-Host "4. Consider adding sample test data" -ForegroundColor White