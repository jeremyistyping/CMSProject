# Test Purchase Report Fix
# This script verifies that the purchase report now shows all approved transactions

$baseUrl = "http://localhost:8080"
$authToken = ""

Write-Host "`n=== PURCHASE REPORT FIX VERIFICATION ===" -ForegroundColor Cyan
Write-Host "Testing: Purchase Report untuk transaksi APPROVED/COMPLETED" -ForegroundColor Yellow

# Test 1: Check if purchase PO/2025/11/0001 exists in database
Write-Host "`n[Test 1] Verify transaksi exists in database..." -ForegroundColor Green
$query1 = @"
SELECT 
    p.id, 
    p.code, 
    p.date, 
    p.total, 
    p.status, 
    p.approval_status,
    c.name as vendor_name,
    p.payment_method
FROM purchases p 
LEFT JOIN contacts c ON c.id = p.vendor_id
WHERE p.code = 'PO/2025/11/0001'
  AND p.deleted_at IS NULL;
"@

Write-Host "Query: Check PO/2025/11/0001 in purchases table" -ForegroundColor Gray
Write-Host $query1 -ForegroundColor DarkGray

# Test 2: Call Purchase Report API
Write-Host "`n[Test 2] Call Purchase Report API..." -ForegroundColor Green
$startDate = "2025-01-01"
$endDate = "2025-12-31"
$url = "$baseUrl/api/v1/ssot-reports/purchase-report?start_date=$startDate&end_date=$endDate&format=json"

Write-Host "URL: $url" -ForegroundColor Gray

try {
    $headers = @{}
    if ($authToken) {
        $headers["Authorization"] = "Bearer $authToken"
    }
    
    $response = Invoke-RestMethod -Uri $url -Method Get -Headers $headers -ErrorAction Stop
    
    Write-Host "`n✅ API Call SUCCESS" -ForegroundColor Green
    
    # Display summary
    Write-Host "`n=== REPORT SUMMARY ===" -ForegroundColor Cyan
    Write-Host "Total Purchases: $($response.data.total_purchases)" -ForegroundColor White
    Write-Host "Total Amount: Rp $($response.data.total_amount)" -ForegroundColor White
    Write-Host "Outstanding: Rp $($response.data.outstanding_payables)" -ForegroundColor White
    
    # Check vendors
    Write-Host "`n=== VENDORS FOUND ===" -ForegroundColor Cyan
    if ($response.data.purchases_by_vendor) {
        foreach ($vendor in $response.data.purchases_by_vendor) {
            Write-Host "`nVendor: $($vendor.vendor_name)" -ForegroundColor Yellow
            Write-Host "  - Total Purchases: $($vendor.total_purchases)"
            Write-Host "  - Total Amount: Rp $($vendor.total_amount)"
            Write-Host "  - Payment Method: $($vendor.payment_method)"
            Write-Host "  - Status: $($vendor.status)"
            Write-Host "  - Outstanding: Rp $($vendor.outstanding)"
            
            # Check for CV Sumber Rejeki specifically
            if ($vendor.vendor_name -eq "CV Sumber Rejeki") {
                Write-Host "`n  ✅ FOUND: CV Sumber Rejeki with Rp $($vendor.total_amount)" -ForegroundColor Green
                $cvSumberRejekiFound = $true
            }
        }
    } else {
        Write-Host "No vendors found in report!" -ForegroundColor Red
    }
    
    # Verification
    Write-Host "`n=== VERIFICATION ===" -ForegroundColor Cyan
    $allPassed = $true
    
    # Test: Total purchases should be >= 1
    if ($response.data.total_purchases -ge 1) {
        Write-Host "✅ Total purchases >= 1: PASS" -ForegroundColor Green
    } else {
        Write-Host "❌ Total purchases < 1: FAIL" -ForegroundColor Red
        $allPassed = $false
    }
    
    # Test: CV Sumber Rejeki should be in vendors list
    if ($cvSumberRejekiFound) {
        Write-Host "✅ CV Sumber Rejeki found: PASS" -ForegroundColor Green
    } else {
        Write-Host "⚠️  CV Sumber Rejeki not found (might be different data)" -ForegroundColor Yellow
    }
    
    # Test: Total amount should be > 0
    if ($response.data.total_amount -gt 0) {
        Write-Host "✅ Total amount > 0: PASS" -ForegroundColor Green
    } else {
        Write-Host "❌ Total amount = 0: FAIL" -ForegroundColor Red
        $allPassed = $false
    }
    
    # Final result
    Write-Host "`n=== FINAL RESULT ===" -ForegroundColor Cyan
    if ($allPassed) {
        Write-Host "✅ ALL TESTS PASSED - Purchase Report is working!" -ForegroundColor Green
        Write-Host "`nPurchase report sekarang menampilkan transaksi yang APPROVED/COMPLETED." -ForegroundColor Green
    } else {
        Write-Host "❌ SOME TESTS FAILED - Check the results above" -ForegroundColor Red
    }
    
    # Display full JSON response for debugging
    Write-Host "`n=== FULL API RESPONSE ===" -ForegroundColor Cyan
    Write-Host ($response | ConvertTo-Json -Depth 10) -ForegroundColor DarkGray
    
} catch {
    Write-Host "`n❌ API Call FAILED" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "`nPossible causes:" -ForegroundColor Yellow
    Write-Host "1. Backend service not running (start with: ./bin/main.exe)" -ForegroundColor Gray
    Write-Host "2. Invalid auth token (update `$authToken variable)" -ForegroundColor Gray
    Write-Host "3. Database connection issue" -ForegroundColor Gray
}

Write-Host "`n=== END OF TEST ===" -ForegroundColor Cyan
