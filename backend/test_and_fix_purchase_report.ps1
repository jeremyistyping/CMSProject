# Test and Fix Purchase Report
# This script will diagnose and fix the Purchase Report data issue

Write-Host "🔧 PURCHASE REPORT TESTING & FIXING" -ForegroundColor Cyan
Write-Host "===================================" -ForegroundColor Cyan

# Test 1: Check if Go server is running
Write-Host "`n📡 Testing Go Backend Server..." -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method Post -Body (@{
        username = "admin"
        password = "admin123"
    } | ConvertTo-Json) -ContentType "application/json" -ErrorAction SilentlyContinue
    
    if ($response.token) {
        $token = $response.token
        Write-Host "✅ Backend server is running and authentication works" -ForegroundColor Green
        Write-Host "Token obtained: $($token.Substring(0, 20))..." -ForegroundColor Gray
    } else {
        Write-Host "⚠️ Server running but no token received" -ForegroundColor Yellow
    }
} catch {
    Write-Host "❌ Backend server issue: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "💡 Make sure Go server is running: go run main.go" -ForegroundColor Yellow
}

# Test 2: Test Purchase Report API with different date ranges
Write-Host "`n📊 Testing Purchase Report API..." -ForegroundColor Yellow

$dateRanges = @(
    @{ start = "2025-09-01"; end = "2025-09-30"; name = "September 2025 (Expected Data)" }
    @{ start = "2025-01-01"; end = "2025-12-31"; name = "Full Year 2025" }
    @{ start = "2024-01-01"; end = "2024-12-31"; name = "Year 2024" }
)

$headers = @{}
if ($token) {
    $headers["Authorization"] = "Bearer $token"
}

foreach ($range in $dateRanges) {
    Write-Host "`n  🗓️  Testing: $($range.name)" -ForegroundColor Cyan
    try {
        $url = "http://localhost:8080/api/v1/ssot-reports/purchase-report?start_date=$($range.start)&end_date=$($range.end)"
        Write-Host "     URL: $url" -ForegroundColor Gray
        
        $response = Invoke-RestMethod -Uri $url -Method Get -Headers $headers -ErrorAction SilentlyContinue
        
        if ($response.success) {
            Write-Host "     ✅ SUCCESS!" -ForegroundColor Green
            Write-Host "     📈 Total Purchases: $($response.data.total_purchases)" -ForegroundColor White
            Write-Host "     💰 Total Amount: $($response.data.total_amount)" -ForegroundColor White
            Write-Host "     🏪 Active Vendors: $($response.data.active_vendors)" -ForegroundColor White
            Write-Host "     💳 Outstanding: $($response.data.outstanding_payables)" -ForegroundColor White
            Write-Host "     👥 Vendors Count: $($response.data.purchases_by_vendor.Count)" -ForegroundColor White
            
            if ($response.data.purchases_by_vendor -and $response.data.purchases_by_vendor.Count -gt 0) {
                Write-Host "     🏢 Vendor Names:" -ForegroundColor White
                foreach ($vendor in $response.data.purchases_by_vendor) {
                    Write-Host "        - $($vendor.vendor_name): $($vendor.total_amount)" -ForegroundColor Gray
                }
            }
        } else {
            Write-Host "     ❌ Failed: $($response.error)" -ForegroundColor Red
        }
    } catch {
        $errorMsg = $_.Exception.Message
        Write-Host "     ❌ Request failed: $errorMsg" -ForegroundColor Red
        
        if ($_.Exception.Response) {
            $statusCode = $_.Exception.Response.StatusCode.value__
            Write-Host "     HTTP Status: $statusCode" -ForegroundColor Red
        }
    }
}

Write-Host "`n🔍 DIAGNOSIS SUMMARY" -ForegroundColor Cyan
Write-Host "===================" -ForegroundColor Cyan

Write-Host "Based on Purchase Management screenshot:" -ForegroundColor White
Write-Host "• Purchase PO/2025/09/0036: Rp 5.550.000 (22/9/2025)" -ForegroundColor White
Write-Host "• Purchase PO/2025/09/0035: Rp 3.885.000 (22/9/2025)" -ForegroundColor White
Write-Host "• Vendor: Jerry Rolo Merentek vendor" -ForegroundColor White
Write-Host "• Status: PAID/APPROVED" -ForegroundColor White

Write-Host "`nIf September 2025 test shows 0 purchases:" -ForegroundColor Yellow
Write-Host "1. ❌ SSOT Integration Issue - Purchases not in unified_journal_ledger" -ForegroundColor White
Write-Host "2. ❌ Service Query Issue - Wrong date filtering or status mapping" -ForegroundColor White
Write-Host "3. ❌ Database Connection Issue" -ForegroundColor White

Write-Host "`n🔧 FIXING OPTIONS" -ForegroundColor Cyan
Write-Host "=================" -ForegroundColor Cyan

Write-Host "Option 1: Run database sync to integrate existing purchases" -ForegroundColor Yellow
Write-Host "Option 2: Debug Go service query logic" -ForegroundColor Yellow
Write-Host "Option 3: Check SSOT journal tables manually" -ForegroundColor Yellow

$continue = Read-Host "`nDo you want to proceed with database diagnosis? (y/n)"

if ($continue -eq 'y' -or $continue -eq 'Y') {
    Write-Host "`n💾 RUNNING DATABASE CHECKS..." -ForegroundColor Green
    Write-Host "=============================" -ForegroundColor Green
    
    # Check if we have psql or database access
    try {
        # Try to connect to PostgreSQL (assuming default setup)
        $dbCheck = @"
-- Quick database check for Purchase Report data
SELECT 'Purchase Transactions' as check_type, COUNT(*) as count, MAX(purchase_date) as latest_date
FROM purchases WHERE deleted_at IS NULL;

SELECT 'SSOT Purchase Entries' as check_type, COUNT(*) as count, MAX(entry_date) as latest_date  
FROM unified_journal_ledger WHERE source_type = 'PURCHASE' AND deleted_at IS NULL;

SELECT 'September Purchases' as check_type, COUNT(*) as count
FROM purchases WHERE purchase_date BETWEEN '2025-09-01' AND '2025-09-30' AND deleted_at IS NULL;
"@
        
        # Save query to temp file
        $queryFile = "temp_db_check.sql"
        $dbCheck | Out-File -FilePath $queryFile -Encoding UTF8
        
        Write-Host "✅ Database check query created: $queryFile" -ForegroundColor Green
        Write-Host "💡 Run this manually in your database to check SSOT integration" -ForegroundColor Yellow
        
    } catch {
        Write-Host "❌ Database check setup failed: $($_.Exception.Message)" -ForegroundColor Red
    }
} else {
    Write-Host "⏭️ Skipping database diagnosis" -ForegroundColor Yellow
}

Write-Host "`n🎯 NEXT STEPS" -ForegroundColor Cyan
Write-Host "=============" -ForegroundColor Cyan
Write-Host "1. Check the test results above" -ForegroundColor White
Write-Host "2. If September 2025 shows correct data ✅ - Problem solved!" -ForegroundColor Green
Write-Host "3. If September 2025 shows 0 data ❌ - Need SSOT sync" -ForegroundColor Red
Write-Host "4. Run database sync script if needed" -ForegroundColor White
Write-Host "5. Test Purchase Report in frontend with September date range" -ForegroundColor White

Write-Host "`n📱 FRONTEND TEST RECOMMENDATION" -ForegroundColor Cyan
Write-Host "===============================" -ForegroundColor Cyan
Write-Host "Open Purchase Report modal and test with:" -ForegroundColor White
Write-Host "• Start Date: 2025-09-01" -ForegroundColor Yellow
Write-Host "• End Date: 2025-09-30" -ForegroundColor Yellow
Write-Host "• Expected: 2 purchases, Rp 9.435.000 total" -ForegroundColor Green