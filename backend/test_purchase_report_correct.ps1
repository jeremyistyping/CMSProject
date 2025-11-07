# Test Purchase Report - Correct Credentials
Write-Host "PURCHASE REPORT API TESTING" -ForegroundColor Cyan
Write-Host "============================" -ForegroundColor Cyan

# Test 1: Check if Go server is running
Write-Host ""
Write-Host "Step 1: Testing Go Backend Server..." -ForegroundColor Yellow
try {
    $loginBody = @{
        email = "admin@company.com"
        password = "admin123"
    } | ConvertTo-Json
    
    Write-Host "Attempting login with email: admin@company.com" -ForegroundColor Gray
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method Post -Body $loginBody -ContentType "application/json" -ErrorAction Stop
    
    if ($response.token) {
        $token = $response.token
        Write-Host "SUCCESS: Backend server authenticated!" -ForegroundColor Green
        Write-Host "Token: $($token.Substring(0, 30))..." -ForegroundColor Gray
    } else {
        Write-Host "WARNING: No token in response" -ForegroundColor Yellow
        Write-Host "Response: $($response | ConvertTo-Json)" -ForegroundColor Gray
    }
} catch {
    Write-Host "ERROR: Authentication failed: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Trying without authentication..." -ForegroundColor Yellow
    $token = $null
}

# Test 2: Direct database check via SQL
Write-Host ""
Write-Host "Step 2: Creating database check queries..." -ForegroundColor Yellow

$dbQueries = @"
-- Check purchases data
SELECT 'Purchases Count' as check_type, COUNT(*) as count
FROM purchases WHERE deleted_at IS NULL;

-- Check September 2025 purchases specifically
SELECT 
    'September 2025 Purchases' as check_type,
    COUNT(*) as count,
    COALESCE(SUM(total_amount), 0) as total_amount
FROM purchases 
WHERE purchase_date BETWEEN '2025-09-01' AND '2025-09-30' 
AND deleted_at IS NULL;

-- Check SSOT journal entries for purchases
SELECT 'SSOT Purchase Entries' as check_type, COUNT(*) as count
FROM unified_journal_ledger 
WHERE source_type = 'PURCHASE' AND deleted_at IS NULL;

-- Check specific September purchases with details
SELECT 
    purchase_code,
    purchase_date,
    total_amount,
    status,
    vendor_name
FROM purchases p
LEFT JOIN vendors v ON p.vendor_id = v.id
WHERE purchase_date BETWEEN '2025-09-01' AND '2025-09-30'
AND p.deleted_at IS NULL
ORDER BY purchase_date DESC;
"@

$queryFile = "check_purchase_data.sql"
$dbQueries | Out-File -FilePath $queryFile -Encoding UTF8
Write-Host "Database queries saved to: $queryFile" -ForegroundColor Green
Write-Host "Run manually in PostgreSQL: psql -d sistem_akuntansi -f $queryFile" -ForegroundColor Yellow

# Test 3: Test Purchase Report API with various approaches
Write-Host ""
Write-Host "Step 3: Testing Purchase Report API..." -ForegroundColor Yellow

$dateRanges = @(
    @{ start = "2025-09-01"; end = "2025-09-30"; name = "September 2025 (TARGET)" }
    @{ start = "2025-01-01"; end = "2025-12-31"; name = "Full Year 2025" }
)

foreach ($range in $dateRanges) {
    Write-Host ""
    Write-Host "  Testing: $($range.name)" -ForegroundColor Cyan
    
    $url = "http://localhost:8080/api/v1/ssot-reports/purchase-report?start_date=$($range.start)&end_date=$($range.end)"
    Write-Host "  URL: $url" -ForegroundColor Gray
    
    # Try with token first
    if ($token) {
        try {
            $headers = @{ "Authorization" = "Bearer $token" }
            $response = Invoke-RestMethod -Uri $url -Method Get -Headers $headers -ErrorAction Stop
            
            Write-Host "  SUCCESS WITH AUTH!" -ForegroundColor Green
            Write-Host "  Total Purchases: $($response.data.total_purchases)" -ForegroundColor White
            Write-Host "  Total Amount: $($response.data.total_amount)" -ForegroundColor White
            Write-Host "  Active Vendors: $($response.data.active_vendors)" -ForegroundColor White
            Write-Host "  Outstanding: $($response.data.outstanding_payables)" -ForegroundColor White
            
            if ($response.data.purchases_by_vendor -and $response.data.purchases_by_vendor.Count -gt 0) {
                Write-Host "  Vendors Found:" -ForegroundColor White
                foreach ($vendor in $response.data.purchases_by_vendor) {
                    Write-Host "    - $($vendor.vendor_name): Rp $($vendor.total_amount)" -ForegroundColor Gray
                    Write-Host "      Purchase Count: $($vendor.purchase_count)" -ForegroundColor Gray
                }
            } else {
                Write-Host "  NO VENDORS FOUND - This indicates SSOT integration issue!" -ForegroundColor Red
            }
            
            # Save successful response for analysis
            $response | ConvertTo-Json -Depth 10 | Out-File -FilePath "purchase_report_response_$($range.start).json"
            Write-Host "  Response saved to: purchase_report_response_$($range.start).json" -ForegroundColor Gray
            
            continue
        } catch {
            Write-Host "  AUTH REQUEST FAILED: $($_.Exception.Message)" -ForegroundColor Red
        }
    }
    
    # Try without authentication
    try {
        $response = Invoke-RestMethod -Uri $url -Method Get -ErrorAction Stop
        Write-Host "  SUCCESS WITHOUT AUTH!" -ForegroundColor Green
        Write-Host "  Response: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Gray
    } catch {
        Write-Host "  NO AUTH REQUEST FAILED: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            $statusCode = $_.Exception.Response.StatusCode.value__
            Write-Host "  HTTP Status: $statusCode" -ForegroundColor Red
        }
    }
}

Write-Host ""
Write-Host "ANALYSIS & NEXT STEPS" -ForegroundColor Cyan
Write-Host "=====================" -ForegroundColor Cyan

Write-Host "Expected September 2025 Data:" -ForegroundColor White
Write-Host "- PO/2025/09/0036: Rp 5.550.000" -ForegroundColor Yellow
Write-Host "- PO/2025/09/0035: Rp 3.885.000" -ForegroundColor Yellow
Write-Host "- Vendor: Jerry Rolo Merentek vendor" -ForegroundColor Yellow
Write-Host "- Total Expected: Rp 9.435.000" -ForegroundColor Yellow

Write-Host ""
Write-Host "IF SEPTEMBER TEST SHOWS 0 PURCHASES:" -ForegroundColor Red
Write-Host "1. Run database queries: psql -d sistem_akuntansi -f check_purchase_data.sql" -ForegroundColor White
Write-Host "2. Check if purchases exist in purchases table" -ForegroundColor White
Write-Host "3. Check if purchases are integrated in unified_journal_ledger" -ForegroundColor White
Write-Host "4. If missing from SSOT, run data sync script" -ForegroundColor White

Write-Host ""
Write-Host "IF SEPTEMBER TEST SHOWS CORRECT DATA:" -ForegroundColor Green
Write-Host "1. Test the frontend Purchase Report modal" -ForegroundColor White
Write-Host "2. Use date range: 2025-09-01 to 2025-09-30" -ForegroundColor White
Write-Host "3. Verify the data matches the backend response" -ForegroundColor White

Write-Host ""
Write-Host "Database Connection Info (from .env):" -ForegroundColor Cyan
Write-Host "Database: sistem_akuntansi" -ForegroundColor Gray  
Write-Host "User: postgres" -ForegroundColor Gray
Write-Host "Password: postgres" -ForegroundColor Gray
Write-Host "Host: localhost" -ForegroundColor Gray