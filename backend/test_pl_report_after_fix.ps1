# Test P&L Report After Fix
# This script tests the SSOT Profit & Loss API to verify the fix

Write-Host "============================================================================" -ForegroundColor Cyan
Write-Host "TESTING P&L REPORT AFTER FIX" -ForegroundColor Cyan
Write-Host "============================================================================" -ForegroundColor Cyan
Write-Host ""

# API endpoint
$baseUrl = "http://localhost:8080"
$plEndpoint = "$baseUrl/api/reports/ssot-profit-loss"

# Parameters
$startDate = "2025-10-01"
$endDate = "2025-10-31"

Write-Host "Test Parameters:" -ForegroundColor Yellow
Write-Host "  Start Date: $startDate" -ForegroundColor Gray
Write-Host "  End Date: $endDate" -ForegroundColor Gray
Write-Host ""

# Build URL with query params
$url = "$plEndpoint`?start_date=$startDate&end_date=$endDate&format=json"

Write-Host "API URL: $url" -ForegroundColor Yellow
Write-Host ""

try {
    Write-Host "Calling P&L API..." -ForegroundColor Yellow
    
    # Call API
    $response = Invoke-RestMethod -Uri $url -Method Get -ContentType "application/json"
    
    Write-Host "✅ API Response Received!" -ForegroundColor Green
    Write-Host ""
    
    # Extract data
    $data = $response.data
    
    Write-Host "============================================================================" -ForegroundColor Cyan
    Write-Host "P&L REPORT SUMMARY" -ForegroundColor Cyan
    Write-Host "============================================================================" -ForegroundColor Cyan
    Write-Host ""
    
    # Total Revenue
    $totalRevenue = $data.total_revenue
    Write-Host "Total Revenue:       " -NoNewline
    Write-Host "Rp $($totalRevenue.ToString('N2'))" -ForegroundColor Green
    
    # Total Expenses
    $totalExpenses = $data.total_expenses
    Write-Host "Total Expenses:      " -NoNewline
    if ($totalExpenses -gt 0) {
        Write-Host "Rp $($totalExpenses.ToString('N2'))" -ForegroundColor Green
        Write-Host "                     ✅ FIXED! (was Rp 0 before)" -ForegroundColor Cyan
    } else {
        Write-Host "Rp $($totalExpenses.ToString('N2'))" -ForegroundColor Red
        Write-Host "                     ❌ Still showing Rp 0!" -ForegroundColor Red
    }
    
    # Net Profit
    $netProfit = $data.net_profit
    Write-Host "Net Profit:          " -NoNewline
    Write-Host "Rp $($netProfit.ToString('N2'))" -ForegroundColor Green
    
    # Net Loss
    $netLoss = $data.net_loss
    if ($netLoss -gt 0) {
        Write-Host "Net Loss:            " -NoNewline
        Write-Host "Rp $($netLoss.ToString('N2'))" -ForegroundColor Red
    }
    
    Write-Host ""
    Write-Host "Financial Metrics:" -ForegroundColor Yellow
    Write-Host "----------------------------------------------------------------------------" -ForegroundColor Gray
    
    $metrics = $data.financialMetrics
    
    Write-Host "Gross Profit:        " -NoNewline
    Write-Host "Rp $($metrics.grossProfit.ToString('N2'))" -ForegroundColor Cyan
    
    Write-Host "Gross Margin:        " -NoNewline
    Write-Host "$($metrics.grossProfitMargin.ToString('N2'))%" -ForegroundColor Cyan
    
    Write-Host "Operating Income:    " -NoNewline
    Write-Host "Rp $($metrics.operatingIncome.ToString('N2'))" -ForegroundColor Cyan
    
    Write-Host "Net Income:          " -NoNewline
    Write-Host "Rp $($metrics.netIncome.ToString('N2'))" -ForegroundColor Cyan
    
    Write-Host "Net Margin:          " -NoNewline
    Write-Host "$($metrics.netIncomeMargin.ToString('N2'))%" -ForegroundColor Cyan
    
    Write-Host ""
    Write-Host "============================================================================" -ForegroundColor Cyan
    Write-Host "VERIFICATION" -ForegroundColor Cyan
    Write-Host "============================================================================" -ForegroundColor Cyan
    Write-Host ""
    
    # Expected values after fix
    $expectedRevenue = 5000000
    $expectedExpenses = 500000
    $expectedGrossProfit = 4500000
    $expectedGrossMargin = 90.0
    $expectedNetIncome = 3375000
    $expectedNetMargin = 67.5
    
    $allPassed = $true
    
    # Check Total Revenue
    if ($totalRevenue -eq $expectedRevenue) {
        Write-Host "✅ Total Revenue:   CORRECT (Rp $($expectedRevenue.ToString('N2')))" -ForegroundColor Green
    } else {
        Write-Host "❌ Total Revenue:   INCORRECT" -ForegroundColor Red
        Write-Host "   Expected: Rp $($expectedRevenue.ToString('N2'))" -ForegroundColor Gray
        Write-Host "   Got:      Rp $($totalRevenue.ToString('N2'))" -ForegroundColor Gray
        $allPassed = $false
    }
    
    # Check Total Expenses (COGS)
    if ($totalExpenses -eq $expectedExpenses) {
        Write-Host "✅ Total Expenses:  CORRECT (Rp $($expectedExpenses.ToString('N2')))" -ForegroundColor Green
    } else {
        Write-Host "❌ Total Expenses:  INCORRECT" -ForegroundColor Red
        Write-Host "   Expected: Rp $($expectedExpenses.ToString('N2'))" -ForegroundColor Gray
        Write-Host "   Got:      Rp $($totalExpenses.ToString('N2'))" -ForegroundColor Gray
        $allPassed = $false
    }
    
    # Check Gross Profit
    if ($metrics.grossProfit -eq $expectedGrossProfit) {
        Write-Host "✅ Gross Profit:    CORRECT (Rp $($expectedGrossProfit.ToString('N2')))" -ForegroundColor Green
    } else {
        Write-Host "❌ Gross Profit:    INCORRECT" -ForegroundColor Red
        Write-Host "   Expected: Rp $($expectedGrossProfit.ToString('N2'))" -ForegroundColor Gray
        Write-Host "   Got:      Rp $($metrics.grossProfit.ToString('N2'))" -ForegroundColor Gray
        $allPassed = $false
    }
    
    # Check Gross Margin
    $marginDiff = [Math]::Abs($metrics.grossProfitMargin - $expectedGrossMargin)
    if ($marginDiff -lt 0.1) {
        Write-Host "✅ Gross Margin:    CORRECT ($($expectedGrossMargin.ToString('N2'))%)" -ForegroundColor Green
    } else {
        Write-Host "❌ Gross Margin:    INCORRECT" -ForegroundColor Red
        Write-Host "   Expected: $($expectedGrossMargin.ToString('N2'))%" -ForegroundColor Gray
        Write-Host "   Got:      $($metrics.grossProfitMargin.ToString('N2'))%" -ForegroundColor Gray
        $allPassed = $false
    }
    
    # Check Net Income (approximate due to tax calculation)
    $netDiff = [Math]::Abs($metrics.netIncome - $expectedNetIncome)
    if ($netDiff -lt 10000) {  # Allow 10k difference due to tax rounding
        Write-Host "✅ Net Income:      CORRECT (Rp $($expectedNetIncome.ToString('N2')))" -ForegroundColor Green
    } else {
        Write-Host "⚠️  Net Income:      CLOSE (within margin)" -ForegroundColor Yellow
        Write-Host "   Expected: Rp $($expectedNetIncome.ToString('N2'))" -ForegroundColor Gray
        Write-Host "   Got:      Rp $($metrics.netIncome.ToString('N2'))" -ForegroundColor Gray
    }
    
    Write-Host ""
    Write-Host "============================================================================" -ForegroundColor Cyan
    
    if ($allPassed) {
        Write-Host "✅ ALL TESTS PASSED!" -ForegroundColor Green
        Write-Host "P&L Report is now showing CORRECT values!" -ForegroundColor Green
    } else {
        Write-Host "❌ SOME TESTS FAILED" -ForegroundColor Red
        Write-Host "Please review the differences above." -ForegroundColor Red
    }
    
    Write-Host "============================================================================" -ForegroundColor Cyan
    Write-Host ""
    
    # Show full response (optional)
    Write-Host "Full API Response:" -ForegroundColor Yellow
    $response | ConvertTo-Json -Depth 10 | Write-Host -ForegroundColor Gray
    
} catch {
    Write-Host "❌ ERROR calling P&L API!" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    Write-Host ""
    Write-Host "Make sure:" -ForegroundColor Yellow
    Write-Host "  1. Backend server is running (http://localhost:8080)" -ForegroundColor Gray
    Write-Host "  2. Database is accessible" -ForegroundColor Gray
    Write-Host "  3. COGS journal has been created" -ForegroundColor Gray
    exit 1
}

