# Test Profit & Loss Summary Fields Fix
# This script verifies that the P&L API response includes the required summary fields

Write-Host "`n=== TESTING P&L SUMMARY FIELDS FIX ===" -ForegroundColor Cyan
Write-Host "This test verifies that the backend now returns:" -ForegroundColor White
Write-Host "  - total_revenue" -ForegroundColor Green
Write-Host "  - total_expenses" -ForegroundColor Green
Write-Host "  - net_profit" -ForegroundColor Green
Write-Host "  - net_loss" -ForegroundColor Green

# Load token from token.txt if it exists
$token = ""
$tokenFile = "token.txt"
if (Test-Path $tokenFile) {
    $token = Get-Content $tokenFile -Raw
    $token = $token.Trim()
    Write-Host "`n[INFO] Using token from $tokenFile" -ForegroundColor Yellow
} else {
    Write-Host "`n[WARNING] token.txt not found. You may need to login first." -ForegroundColor Red
    Write-Host "Run ./test_login_api.ps1 to get a token" -ForegroundColor Yellow
    exit 1
}

# API Configuration
$baseUrl = "http://localhost:8080"
$endpoint = "/api/v1/reports/ssot-profit-loss"

# Test parameters
$startDate = "2025-01-01"
$endDate = "2025-12-31"

$url = "${baseUrl}${endpoint}?start_date=${startDate}&end_date=${endDate}&format=json"

Write-Host "`n[TEST] Calling SSOT Profit & Loss API" -ForegroundColor Cyan
Write-Host "URL: $url" -ForegroundColor Gray

try {
    # Make API request
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }
    
    Write-Host "`n[REQUEST] Sending GET request..." -ForegroundColor Yellow
    $response = Invoke-RestMethod -Uri $url -Method Get -Headers $headers -ErrorAction Stop
    
    Write-Host "`n[SUCCESS] API Response received!" -ForegroundColor Green
    
    # Extract data from response
    $data = $response.data
    
    if (-not $data) {
        Write-Host "`n[ERROR] No data field in response" -ForegroundColor Red
        Write-Host "Full Response:" -ForegroundColor Gray
        $response | ConvertTo-Json -Depth 10
        exit 1
    }
    
    # Check for required summary fields
    Write-Host "`n=== SUMMARY FIELDS CHECK ===" -ForegroundColor Cyan
    
    $allFieldsPresent = $true
    
    # Check total_revenue
    if ($null -ne $data.total_revenue) {
        Write-Host "[PASS] total_revenue: Rp $($data.total_revenue)" -ForegroundColor Green
    } else {
        Write-Host "[FAIL] total_revenue: MISSING" -ForegroundColor Red
        $allFieldsPresent = $false
    }
    
    # Check total_expenses
    if ($null -ne $data.total_expenses) {
        Write-Host "[PASS] total_expenses: Rp $($data.total_expenses)" -ForegroundColor Green
    } else {
        Write-Host "[FAIL] total_expenses: MISSING" -ForegroundColor Red
        $allFieldsPresent = $false
    }
    
    # Check net_profit
    if ($null -ne $data.net_profit) {
        Write-Host "[PASS] net_profit: Rp $($data.net_profit)" -ForegroundColor Green
    } else {
        Write-Host "[FAIL] net_profit: MISSING" -ForegroundColor Red
        $allFieldsPresent = $false
    }
    
    # Check net_loss
    if ($null -ne $data.net_loss) {
        Write-Host "[PASS] net_loss: Rp $($data.net_loss)" -ForegroundColor Green
    } else {
        Write-Host "[FAIL] net_loss: MISSING" -ForegroundColor Red
        $allFieldsPresent = $false
    }
    
    # Verify financial metrics are also present
    Write-Host "`n=== FINANCIAL METRICS CHECK ===" -ForegroundColor Cyan
    if ($data.financialMetrics) {
        Write-Host "[PASS] financialMetrics object present" -ForegroundColor Green
        Write-Host "  - grossProfit: Rp $($data.financialMetrics.grossProfit)" -ForegroundColor White
        Write-Host "  - operatingIncome: Rp $($data.financialMetrics.operatingIncome)" -ForegroundColor White
        Write-Host "  - netIncome: Rp $($data.financialMetrics.netIncome)" -ForegroundColor White
        Write-Host "  - netIncomeMargin: $($data.financialMetrics.netIncomeMargin)%" -ForegroundColor White
    } else {
        Write-Host "[FAIL] financialMetrics: MISSING" -ForegroundColor Red
        $allFieldsPresent = $false
    }
    
    # Check sections structure
    Write-Host "`n=== SECTIONS CHECK ===" -ForegroundColor Cyan
    if ($data.sections) {
        Write-Host "[PASS] sections array present with $($data.sections.Count) sections" -ForegroundColor Green
        foreach ($section in $data.sections) {
            Write-Host "  - $($section.name): Rp $($section.total)" -ForegroundColor White
        }
    } else {
        Write-Host "[FAIL] sections: MISSING" -ForegroundColor Red
        $allFieldsPresent = $false
    }
    
    # Accounting logic validation
    Write-Host "`n=== ACCOUNTING LOGIC VALIDATION ===" -ForegroundColor Cyan
    
    # Calculate expected values
    $totalRevenue = $data.total_revenue
    $totalExpenses = $data.total_expenses
    $netIncome = $data.financialMetrics.netIncome
    
    # Check if profit/loss calculation is correct
    if ($netIncome -gt 0) {
        if ($data.net_profit -eq $netIncome -and $data.net_loss -eq 0) {
            Write-Host "[PASS] Profit calculation correct (Net Income > 0, so Net Profit = Net Income, Net Loss = 0)" -ForegroundColor Green
        } else {
            Write-Host "[FAIL] Profit calculation incorrect" -ForegroundColor Red
            Write-Host "  Expected: net_profit = $netIncome, net_loss = 0" -ForegroundColor Gray
            Write-Host "  Actual: net_profit = $($data.net_profit), net_loss = $($data.net_loss)" -ForegroundColor Gray
        }
    } elseif ($netIncome -lt 0) {
        $expectedLoss = [Math]::Abs($netIncome)
        if ($data.net_loss -eq $expectedLoss -and $data.net_profit -eq 0) {
            Write-Host "[PASS] Loss calculation correct (Net Income < 0, so Net Loss = |Net Income|, Net Profit = 0)" -ForegroundColor Green
        } else {
            Write-Host "[FAIL] Loss calculation incorrect" -ForegroundColor Red
            Write-Host "  Expected: net_loss = $expectedLoss, net_profit = 0" -ForegroundColor Gray
            Write-Host "  Actual: net_loss = $($data.net_loss), net_profit = $($data.net_profit)" -ForegroundColor Gray
        }
    } else {
        Write-Host "[INFO] Break-even (Net Income = 0)" -ForegroundColor Yellow
    }
    
    # Check enhanced flag
    Write-Host "`n=== ENHANCED MODE CHECK ===" -ForegroundColor Cyan
    if ($data.enhanced -eq $true) {
        Write-Host "[PASS] Enhanced mode enabled" -ForegroundColor Green
    } else {
        Write-Host "[INFO] Enhanced mode not enabled" -ForegroundColor Yellow
    }
    
    # Data source info
    if ($data.data_source_label) {
        Write-Host "`n[INFO] Data Source: $($data.data_source_label)" -ForegroundColor Cyan
    }
    
    # Analysis message
    if ($data.message) {
        Write-Host "`n[INFO] Analysis Message:" -ForegroundColor Cyan
        Write-Host "  $($data.message)" -ForegroundColor White
    }
    
    # Final result
    Write-Host "`n=== TEST RESULT ===" -ForegroundColor Cyan
    if ($allFieldsPresent) {
        Write-Host "[SUCCESS] All required fields are present! ✓" -ForegroundColor Green
        Write-Host "`nThe frontend summary boxes should now display correctly." -ForegroundColor Green
        exit 0
    } else {
        Write-Host "[FAILURE] Some required fields are missing! ✗" -ForegroundColor Red
        Write-Host "`nPlease check the backend implementation." -ForegroundColor Red
        exit 1
    }
    
} catch {
    Write-Host "`n[ERROR] Request failed!" -ForegroundColor Red
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
    
    if ($_.Exception.Response) {
        $statusCode = $_.Exception.Response.StatusCode.value__
        Write-Host "HTTP Status Code: $statusCode" -ForegroundColor Yellow
        
        try {
            $errorStream = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($errorStream)
            $errorBody = $reader.ReadToEnd()
            Write-Host "Error Response:" -ForegroundColor Yellow
            Write-Host $errorBody -ForegroundColor Gray
        } catch {
            Write-Host "Could not read error response body" -ForegroundColor Gray
        }
    }
    
    exit 1
}

