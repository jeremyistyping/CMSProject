# Script untuk test API hierarchy endpoint dan debugging frontend issue

Write-Host "🔍 TESTING API HIERARCHY ENDPOINT" -ForegroundColor Yellow
Write-Host "=================================" -ForegroundColor Yellow

# Get JWT token (assuming you have one from login)
$token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6ImFkbWluQGV4YW1wbGUuY29tIiwiZXhwIjoxNzI4MTg4MDc3LCJpZCI6MSwicm9sZSI6ImFkbWluIiwidXNlcm5hbWUiOiJhZG1pbiJ9.dTVMY3uD9Q_MnuGTlROrhRHOLmqJCQTMTgGCLtPPeRs"

if (-not $token) {
    Write-Host "❌ No JWT token provided. Please login first to get a token." -ForegroundColor Red
    Write-Host ""
    Write-Host "💡 To get a token:" -ForegroundColor Cyan
    Write-Host "1. Open browser network tab" -ForegroundColor Gray
    Write-Host "2. Login to the app" -ForegroundColor Gray
    Write-Host "3. Look for Authorization header in any API call" -ForegroundColor Gray
    Write-Host "4. Copy the token after 'Bearer '" -ForegroundColor Gray
    exit 1
}

Write-Host ""
Write-Host "1️⃣ Testing /accounts/hierarchy endpoint..." -ForegroundColor Cyan

try {
    # Test the accounts hierarchy API
    $headers = @{
        'Authorization' = "Bearer $token"
        'Content-Type' = 'application/json'
        'Cache-Control' = 'no-cache'
    }
    
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/accounts/hierarchy" -Method Get -Headers $headers -UseBasicParsing
    
    Write-Host "✅ API Response received successfully" -ForegroundColor Green
    Write-Host ""
    
    # Look for Bank Mandiri in the response
    $bankMandiriFound = $false
    $bankMandiriBalance = 0
    
    function Search-BankMandiri($accounts) {
        foreach ($account in $accounts) {
            if ($account.code -eq "1103" -and $account.name -like "*Bank Mandiri*") {
                $script:bankMandiriFound = $true
                $script:bankMandiriBalance = $account.balance
                Write-Host "🎯 BANK MANDIRI FOUND:" -ForegroundColor Yellow
                Write-Host "   Code: $($account.code)" -ForegroundColor White
                Write-Host "   Name: $($account.name)" -ForegroundColor White
                Write-Host "   Balance: $($account.balance)" -ForegroundColor White
                Write-Host "   Is Header: $($account.is_header)" -ForegroundColor White
                return
            }
            
            if ($account.children -and $account.children.Count -gt 0) {
                Search-BankMandiri $account.children
            }
        }
    }
    
    if ($response.data) {
        Search-BankMandiri $response.data
    }
    
    if ($bankMandiriFound) {
        if ($bankMandiriBalance -eq 44450000) {
            Write-Host "✅ Bank Mandiri balance is CORRECT: Rp 44,450,000" -ForegroundColor Green
        } else {
            Write-Host "❌ Bank Mandiri balance is WRONG: Rp $($bankMandiriBalance.ToString('N0'))" -ForegroundColor Red
            Write-Host "   Expected: Rp 44,450,000" -ForegroundColor Gray
        }
    } else {
        Write-Host "❌ Bank Mandiri (1103) not found in API response" -ForegroundColor Red
    }
    
} catch {
    Write-Host "❌ Error calling API: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host ""
    
    if ($_.Exception.Response) {
        $statusCode = $_.Exception.Response.StatusCode.value__
        Write-Host "   Status Code: $statusCode" -ForegroundColor Red
        
        if ($statusCode -eq 401) {
            Write-Host "   Issue: Authentication failed - token might be expired" -ForegroundColor Yellow
        } elseif ($statusCode -eq 403) {
            Write-Host "   Issue: Permission denied - user might not have accounts view permission" -ForegroundColor Yellow
        } elseif ($statusCode -eq 404) {
            Write-Host "   Issue: Endpoint not found - check if backend is running" -ForegroundColor Yellow
        }
    }
}

Write-Host ""
Write-Host "2️⃣ Testing /coa/posted-balances endpoint..." -ForegroundColor Cyan

try {
    $response2 = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/coa/posted-balances" -Method Get -Headers $headers -UseBasicParsing
    
    Write-Host "✅ Posted balances API Response received" -ForegroundColor Green
    
    # Look for Bank Mandiri in posted balances
    $postedBalance = $response2.data | Where-Object { $_.account_code -eq "1103" }
    
    if ($postedBalance) {
        Write-Host "🎯 BANK MANDIRI POSTED BALANCE:" -ForegroundColor Yellow
        Write-Host "   Code: $($postedBalance.account_code)" -ForegroundColor White
        Write-Host "   Name: $($postedBalance.account_name)" -ForegroundColor White  
        Write-Host "   Raw Balance: $($postedBalance.raw_balance)" -ForegroundColor White
        Write-Host "   Display Balance: $($postedBalance.display_balance)" -ForegroundColor White
        
        if ($postedBalance.raw_balance -eq 44450000) {
            Write-Host "✅ Posted balance is CORRECT: Rp 44,450,000" -ForegroundColor Green
        } else {
            Write-Host "❌ Posted balance is WRONG: Rp $($postedBalance.raw_balance.ToString('N0'))" -ForegroundColor Red
        }
    } else {
        Write-Host "❌ Bank Mandiri not found in posted balances" -ForegroundColor Red
    }
    
} catch {
    Write-Host "❌ Error calling posted balances API: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "📋 DIAGNOSTIC SUMMARY:" -ForegroundColor Cyan
Write-Host "======================" -ForegroundColor Cyan

Write-Host ""
Write-Host "🔧 FRONTEND CACHE CLEARING STEPS:" -ForegroundColor Yellow
Write-Host "1. Open Chrome DevTools (F12)" -ForegroundColor White
Write-Host "2. Right-click refresh button → 'Empty Cache and Hard Reload'" -ForegroundColor White
Write-Host "3. Or go to Application tab → Storage → 'Clear site data'" -ForegroundColor White
Write-Host "4. Check Network tab for API calls with fresh data" -ForegroundColor White

Write-Host ""
Write-Host "🚀 IF PROBLEM PERSISTS:" -ForegroundColor Yellow
Write-Host "1. Restart frontend dev server (npm run dev)" -ForegroundColor White
Write-Host "2. Restart backend server" -ForegroundColor White
Write-Host "3. Check browser console for JavaScript errors" -ForegroundColor White
Write-Host "4. Verify the edited page.tsx was saved and compiled" -ForegroundColor White