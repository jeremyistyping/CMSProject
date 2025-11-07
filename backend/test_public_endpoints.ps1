# PowerShell Test Script for Frontend AccountService Public Endpoints
# This script simulates the frontend calling the backend public endpoints

$API_BASE_URL = "http://localhost:8080"

Write-Host "🧪 Testing Frontend AccountService Public Endpoints" -ForegroundColor Cyan
Write-Host ""

try {
    # Test 1: Get all account catalog (simulates frontend getAccountCatalog())
    Write-Host "📝 Test 1: Get all account catalog" -ForegroundColor Yellow
    $response1 = Invoke-RestMethod -Uri "$API_BASE_URL/api/v1/accounts/catalog" -Method GET -ContentType "application/json"
    Write-Host "✅ SUCCESS: Retrieved $($response1.data.Count) accounts" -ForegroundColor Green
    if ($response1.data.Count -gt 0) {
        Write-Host "   Sample: $($response1.data[0].code) - $($response1.data[0].name)" -ForegroundColor Gray
    }
    Write-Host ""

    # Test 2: Get expense accounts (simulates frontend getExpenseAccounts())
    Write-Host "📝 Test 2: Get expense accounts for purchase form" -ForegroundColor Yellow
    $response2 = Invoke-RestMethod -Uri "$API_BASE_URL/api/v1/accounts/catalog?type=EXPENSE" -Method GET -ContentType "application/json"
    Write-Host "✅ SUCCESS: Retrieved $($response2.data.Count) expense accounts" -ForegroundColor Green
    if ($response2.data.Count -gt 0) {
        Write-Host "   Sample: $($response2.data[0].code) - $($response2.data[0].name)" -ForegroundColor Gray
    }
    Write-Host ""

    # Test 3: Get credit accounts (simulates frontend getCreditAccounts())
    Write-Host "📝 Test 3: Get credit accounts for payment method" -ForegroundColor Yellow
    $response3 = Invoke-RestMethod -Uri "$API_BASE_URL/api/v1/accounts/credit?type=LIABILITY" -Method GET -ContentType "application/json"
    Write-Host "✅ SUCCESS: Retrieved $($response3.data.Count) liability accounts" -ForegroundColor Green
    if ($response3.data.Count -gt 0) {
        Write-Host "   Sample: $($response3.data[0].code) - $($response3.data[0].name)" -ForegroundColor Gray
    } else {
        Write-Host "   (No LIABILITY accounts found, this might be expected)" -ForegroundColor Gray
    }
    Write-Host ""

    # Test 4: Get specific account types
    Write-Host "📝 Test 4: Get ASSET accounts" -ForegroundColor Yellow
    $response4 = Invoke-RestMethod -Uri "$API_BASE_URL/api/v1/accounts/catalog?type=ASSET" -Method GET -ContentType "application/json"
    Write-Host "✅ SUCCESS: Retrieved $($response4.data.Count) asset accounts" -ForegroundColor Green
    Write-Host ""

    Write-Host "🎉 ALL TESTS PASSED! Frontend service should work correctly." -ForegroundColor Green
    Write-Host ""

    Write-Host "📋 SUMMARY:" -ForegroundColor Cyan
    Write-Host "   - Total accounts: $($response1.data.Count)" -ForegroundColor White
    Write-Host "   - Expense accounts: $($response2.data.Count)" -ForegroundColor White
    Write-Host "   - Credit accounts: $($response3.data.Count)" -ForegroundColor White
    Write-Host "   - Asset accounts: $($response4.data.Count)" -ForegroundColor White
    Write-Host ""
    
    Write-Host "✅ The purchase form dropdowns should now work without 'Limited Access' errors!" -ForegroundColor Green
    Write-Host ""
    Write-Host "🔄 Next Steps:" -ForegroundColor Cyan
    Write-Host "   1. Hard refresh your browser (Ctrl+F5) to clear cache" -ForegroundColor White
    Write-Host "   2. Navigate to http://localhost:3000/purchases" -ForegroundColor White
    Write-Host "   3. Click 'Create New Purchase'" -ForegroundColor White
    Write-Host "   4. Verify dropdowns load without errors" -ForegroundColor White

} catch {
    Write-Host "❌ TEST FAILED: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
    Write-Host "🔍 Troubleshooting:" -ForegroundColor Yellow
    Write-Host "   1. Ensure backend server is running on port 8080" -ForegroundColor White
    Write-Host "   2. Verify public endpoints are correctly configured" -ForegroundColor White
    Write-Host "   3. Check network connectivity" -ForegroundColor White
    Write-Host "   4. Try: netstat -ano | findstr :8080" -ForegroundColor White
}