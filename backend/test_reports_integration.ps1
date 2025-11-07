# Test SSOT Reports Integration API
# Usage: .\test_reports_integration.ps1

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$Username = "admin@example.com",
    [string]$Password = "password123",
    [string]$StartDate = "2024-01-01",
    [string]$EndDate = "2024-12-31",
    [string]$AsOfDate = "2024-12-31"
)

Write-Host "🚀 Testing SSOT Reports API Integration" -ForegroundColor Cyan
Write-Host "Base URL: $BaseUrl" -ForegroundColor Gray
Write-Host ""

# Function to get JWT token
function Get-JWTToken {
    param($BaseUrl, $Username, $Password)
    
    Write-Host "🔐 Getting JWT Token..." -ForegroundColor Yellow
    
    $loginData = @{
        email = $Username
        password = $Password
    } | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl/api/v1/auth/login" -Method POST -Body $loginData -ContentType "application/json"
        if ($response.access_token) {
            Write-Host "✅ Successfully obtained JWT token" -ForegroundColor Green
            return $response.access_token
        } else {
            Write-Host "❌ Failed to get token: $($response.message)" -ForegroundColor Red
            return $null
        }
    }
    catch {
        Write-Host "❌ Login failed: $($_.Exception.Message)" -ForegroundColor Red
        Write-Host "💡 Make sure the server is running and credentials are correct" -ForegroundColor Yellow
        return $null
    }
}

# Function to test API endpoint
function Test-Endpoint {
    param($Url, $Headers, $Name)
    
    Write-Host "🧪 Testing: $Name" -ForegroundColor Cyan
    Write-Host "   URL: $Url" -ForegroundColor Gray
    
    try {
        $response = Invoke-RestMethod -Uri $Url -Method GET -Headers $Headers -ContentType "application/json"
        Write-Host "   ✅ Success" -ForegroundColor Green
        Write-Host "   📊 Status: $($response.status)" -ForegroundColor Gray
        
        if ($response.data) {
            Write-Host "   📈 Data available: Yes" -ForegroundColor Gray
            if ($response.data.generated_at) {
                Write-Host "   🕒 Generated at: $($response.data.generated_at)" -ForegroundColor Gray
            }
            if ($response.data.data_source_info) {
                Write-Host "   🔗 SSOT Version: $($response.data.data_source_info.ssot_version)" -ForegroundColor Gray
                Write-Host "   📝 Total Entries: $($response.data.data_source_info.total_journal_entries)" -ForegroundColor Gray
            }
        }
        return $true
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        Write-Host "   ❌ Failed (HTTP $statusCode)" -ForegroundColor Red
        
        try {
            $errorResponse = $_.ErrorDetails.Message | ConvertFrom-Json
            Write-Host "   💬 Error: $($errorResponse.message)" -ForegroundColor Red
        }
        catch {
            Write-Host "   💬 Error: $($_.Exception.Message)" -ForegroundColor Red
        }
        return $false
    }
}

# Main testing flow
Write-Host "=" * 60 -ForegroundColor Cyan

# Step 1: Get JWT Token
$token = Get-JWTToken -BaseUrl $BaseUrl -Username $Username -Password $Password

if (-not $token) {
    Write-Host ""
    Write-Host "❌ Cannot proceed without authentication token" -ForegroundColor Red
    Write-Host "💡 Please check:" -ForegroundColor Yellow
    Write-Host "   - Server is running on $BaseUrl" -ForegroundColor Yellow
    Write-Host "   - Credentials are correct (default: admin@example.com / password123)" -ForegroundColor Yellow
    Write-Host "   - Database is seeded with default admin user" -ForegroundColor Yellow
    exit 1
}

# Set headers for authenticated requests
$headers = @{
    "Authorization" = "Bearer $token"
}

Write-Host ""
Write-Host "=" * 60 -ForegroundColor Cyan

# Step 2: Test System Status
Write-Host "📊 TESTING SYSTEM STATUS" -ForegroundColor Magenta
$statusOk = Test-Endpoint -Url "$BaseUrl/api/v1/ssot-reports/status" -Headers $headers -Name "SSOT Reports Status"

Write-Host ""
Write-Host "=" * 60 -ForegroundColor Cyan

# Step 3: Test Main Integration Endpoint
Write-Host "🎯 TESTING MAIN INTEGRATION ENDPOINT" -ForegroundColor Magenta
$integratedOk = Test-Endpoint -Url "$BaseUrl/api/v1/ssot-reports/integrated?start_date=$StartDate&end_date=$EndDate" -Headers $headers -Name "Integrated Financial Reports"

Write-Host ""
Write-Host "=" * 60 -ForegroundColor Cyan

# Step 4: Test Individual Report Endpoints
Write-Host "📈 TESTING INDIVIDUAL REPORT ENDPOINTS" -ForegroundColor Magenta

$endpoints = @(
    @{ Name = "Balance Sheet"; Url = "$BaseUrl/api/v1/reports/balance-sheet?as_of_date=$AsOfDate&format=json" }
    @{ Name = "Profit & Loss"; Url = "$BaseUrl/api/v1/reports/profit-loss?start_date=$StartDate&end_date=$EndDate&format=json" }
    @{ Name = "Trial Balance (SSOT)"; Url = "$BaseUrl/api/v1/ssot-reports/trial-balance?as_of_date=$AsOfDate&format=json" }
    @{ Name = "Vendor Analysis (SSOT)"; Url = "$BaseUrl/api/v1/ssot-reports/vendor-analysis?start_date=$StartDate&end_date=$EndDate&format=json" }
    @{ Name = "General Ledger (SSOT)"; Url = "$BaseUrl/api/v1/ssot-reports/general-ledger?account_id=all&start_date=$StartDate&end_date=$EndDate&format=json" }
    @{ Name = "Journal Analysis (SSOT)"; Url = "$BaseUrl/api/v1/ssot-reports/journal-analysis?start_date=$StartDate&end_date=$EndDate&format=json" }
)

$successCount = 0
foreach ($endpoint in $endpoints) {
    if (Test-Endpoint -Url $endpoint.Url -Headers $headers -Name $endpoint.Name) {
        $successCount++
    }
    Write-Host ""
}

Write-Host "=" * 60 -ForegroundColor Cyan

# Step 5: Test WebSocket Information
Write-Host "🔄 TESTING WEBSOCKET INFO" -ForegroundColor Magenta
$wsOk = Test-Endpoint -Url "$BaseUrl/api/v1/ssot-reports/realtime" -Headers $headers -Name "WebSocket Information"

Write-Host ""
Write-Host "=" * 60 -ForegroundColor Cyan

# Summary
Write-Host "📋 TEST SUMMARY" -ForegroundColor Magenta
Write-Host ""
Write-Host "🔐 Authentication: $(if($token) {"✅ Success"} else {"❌ Failed"})" -ForegroundColor $(if($token) {"Green"} else {"Red"})
Write-Host "📊 System Status: $(if($statusOk) {"✅ OK"} else {"❌ Failed"})" -ForegroundColor $(if($statusOk) {"Green"} else {"Red"})
Write-Host "🎯 Main Integration: $(if($integratedOk) {"✅ OK"} else {"❌ Failed"})" -ForegroundColor $(if($integratedOk) {"Green"} else {"Red"})
Write-Host "📈 Individual Reports: ✅ $successCount/$($endpoints.Count) endpoints working" -ForegroundColor Green
Write-Host "🔄 WebSocket Info: $(if($wsOk) {"✅ OK"} else {"❌ Failed"})" -ForegroundColor $(if($wsOk) {"Green"} else {"Red"})

Write-Host ""

if ($statusOk -and $integratedOk -and $successCount -gt 0) {
    Write-Host "🎉 INTEGRATION READY!" -ForegroundColor Green
    Write-Host ""
    Write-Host "✅ Backend API is ready for frontend integration at http://localhost:3000/reports" -ForegroundColor Green
    Write-Host ""
    Write-Host "📝 NEXT STEPS FOR FRONTEND:" -ForegroundColor Cyan
    Write-Host "1. Use the main endpoint: GET /api/v1/ssot-reports/integrated" -ForegroundColor Gray
    Write-Host "2. Include JWT token in Authorization header" -ForegroundColor Gray
    Write-Host "3. Handle start_date and end_date parameters" -ForegroundColor Gray
    Write-Host "4. Optionally connect to WebSocket at ws://localhost:8080/ws/balance" -ForegroundColor Gray
    Write-Host ""
    Write-Host "📖 Full documentation: REPORTS_API_INTEGRATION.md" -ForegroundColor Cyan
} else {
    Write-Host "⚠️  INTEGRATION INCOMPLETE" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Some endpoints are not working. Please check:" -ForegroundColor Yellow
    Write-Host "- Server is running and properly configured" -ForegroundColor Yellow
    Write-Host "- Database is migrated and has sample data" -ForegroundColor Yellow
    Write-Host "- SSOT Journal system is properly initialized" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "=" * 60 -ForegroundColor Cyan