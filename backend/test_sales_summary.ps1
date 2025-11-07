# Test Sales Summary API Endpoint
# Debug why sales summary shows 0

Write-Host "Testing Sales Summary API..." -ForegroundColor Yellow

# Configuration
$baseUrl = "http://localhost:8080/api/reports"
$token = ""  # You'll need to get this from login

# Test parameters - September 2025
$startDate = "2025-09-01"
$endDate = "2025-09-30"

# Build URL with parameters
$url = "$baseUrl/sales-summary?start_date=$startDate&end_date=$endDate&group_by=month&format=json"

Write-Host "Testing URL: $url" -ForegroundColor Green

try {
    # Test without auth first (if endpoint allows)
    $response = Invoke-RestMethod -Uri $url -Method GET -ContentType "application/json"
    
    Write-Host "Response Status: SUCCESS" -ForegroundColor Green
    Write-Host "Response Data:" -ForegroundColor Cyan
    $response | ConvertTo-Json -Depth 10 | Write-Host
    
} catch {
    Write-Host "Error occurred:" -ForegroundColor Red
    Write-Host $_.Exception.Message -ForegroundColor Red
    
    if ($_.Exception.Response) {
        $statusCode = $_.Exception.Response.StatusCode
        Write-Host "Status Code: $statusCode" -ForegroundColor Red
    }
}

# Test with different date ranges
Write-Host "`n--- Testing Different Date Ranges ---" -ForegroundColor Yellow

$testRanges = @(
    @{start="2024-01-01"; end="2024-12-31"; name="Full 2024"},
    @{start="2025-01-01"; end="2025-12-31"; name="Full 2025"},
    @{start="2024-09-01"; end="2024-09-30"; name="Sep 2024"},
    @{start="2025-08-01"; end="2025-08-31"; name="Aug 2025"}
)

foreach ($range in $testRanges) {
    $testUrl = "$baseUrl/sales-summary?start_date=$($range.start)&end_date=$($range.end)&group_by=month&format=json"
    Write-Host "`nTesting $($range.name): $testUrl" -ForegroundColor Cyan
    
    try {
        $testResponse = Invoke-RestMethod -Uri $testUrl -Method GET -ContentType "application/json"
        $totalRevenue = $testResponse.data.total_revenue
        $totalTransactions = $testResponse.data.total_transactions
        Write-Host "  Revenue: $totalRevenue, Transactions: $totalTransactions" -ForegroundColor Green
    } catch {
        Write-Host "  Error: $($_.Exception.Message)" -ForegroundColor Red
    }
}

Write-Host "`n--- Test Completed ---" -ForegroundColor Yellow