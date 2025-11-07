# Test SSOT Reports Integration API - Simple Version
param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$Username = "admin@example.com",
    [string]$Password = "password123"
)

Write-Host "Testing SSOT Reports API Integration" -ForegroundColor Cyan
Write-Host "Base URL: $BaseUrl" -ForegroundColor Gray
Write-Host ""

# Function to get JWT token
function Get-JWTToken {
    param($BaseUrl, $Username, $Password)
    
    Write-Host "Getting JWT Token..." -ForegroundColor Yellow
    
    $loginData = @{
        email = $Username
        password = $Password
    } | ConvertTo-Json
    
    try {
        $response = Invoke-RestMethod -Uri "$BaseUrl/api/v1/auth/login" -Method POST -Body $loginData -ContentType "application/json"
        if ($response.access_token) {
            Write-Host "Successfully obtained JWT token" -ForegroundColor Green
            return $response.access_token
        } else {
            Write-Host "Failed to get token: $($response.message)" -ForegroundColor Red
            return $null
        }
    }
    catch {
        Write-Host "Login failed: $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
}

# Function to test API endpoint
function Test-Endpoint {
    param($Url, $Headers, $Name)
    
    Write-Host "Testing: $Name" -ForegroundColor Cyan
    Write-Host "URL: $Url" -ForegroundColor Gray
    
    try {
        $response = Invoke-RestMethod -Uri $Url -Method GET -Headers $Headers -ContentType "application/json"
        Write-Host "SUCCESS" -ForegroundColor Green
        Write-Host "Status: $($response.status)" -ForegroundColor Gray
        return $true
    }
    catch {
        $statusCode = $_.Exception.Response.StatusCode.value__
        Write-Host "FAILED (HTTP $statusCode)" -ForegroundColor Red
        
        try {
            $errorResponse = $_.ErrorDetails.Message | ConvertFrom-Json
            Write-Host "Error: $($errorResponse.message)" -ForegroundColor Red
        }
        catch {
            Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
        }
        return $false
    }
    Write-Host ""
}

# Main testing flow
Write-Host "============================================================" -ForegroundColor Cyan

# Step 1: Get JWT Token
$token = Get-JWTToken -BaseUrl $BaseUrl -Username $Username -Password $Password

if (-not $token) {
    Write-Host ""
    Write-Host "Cannot proceed without authentication token" -ForegroundColor Red
    exit 1
}

# Set headers for authenticated requests
$headers = @{
    "Authorization" = "Bearer $token"
}

Write-Host ""
Write-Host "============================================================" -ForegroundColor Cyan

# Step 2: Test System Status
Write-Host "TESTING SYSTEM STATUS" -ForegroundColor Magenta
$statusOk = Test-Endpoint -Url "$BaseUrl/api/v1/ssot-reports/status" -Headers $headers -Name "SSOT Reports Status"

Write-Host ""
Write-Host "============================================================" -ForegroundColor Cyan

# Step 3: Test Main Integration Endpoint
Write-Host "TESTING MAIN INTEGRATION ENDPOINT" -ForegroundColor Magenta
$integratedOk = Test-Endpoint -Url "$BaseUrl/api/v1/ssot-reports/integrated?start_date=2024-01-01&end_date=2024-12-31" -Headers $headers -Name "Integrated Financial Reports"

Write-Host ""
Write-Host "============================================================" -ForegroundColor Cyan

# Step 4: Test Individual Report Endpoints
Write-Host "TESTING INDIVIDUAL REPORT ENDPOINTS" -ForegroundColor Magenta

$endpoints = @(
    @{ Name = "Balance Sheet"; Url = "$BaseUrl/api/v1/reports/balance-sheet?as_of_date=2024-12-31&format=json" }
    @{ Name = "Trial Balance (SSOT)"; Url = "$BaseUrl/api/v1/ssot-reports/trial-balance?as_of_date=2024-12-31&format=json" }
)

$successCount = 0
foreach ($endpoint in $endpoints) {
    if (Test-Endpoint -Url $endpoint.Url -Headers $headers -Name $endpoint.Name) {
        $successCount++
    }
    Write-Host ""
}

Write-Host "============================================================" -ForegroundColor Cyan

# Summary
Write-Host "TEST SUMMARY" -ForegroundColor Magenta
Write-Host ""
Write-Host "Authentication: $(if($token) {"SUCCESS"} else {"FAILED"})" -ForegroundColor $(if($token) {"Green"} else {"Red"})
Write-Host "System Status: $(if($statusOk) {"OK"} else {"FAILED"})" -ForegroundColor $(if($statusOk) {"Green"} else {"Red"})
Write-Host "Main Integration: $(if($integratedOk) {"OK"} else {"FAILED"})" -ForegroundColor $(if($integratedOk) {"Green"} else {"Red"})
Write-Host "Individual Reports: $successCount/$($endpoints.Count) working" -ForegroundColor Green

Write-Host ""

if ($statusOk -and $integratedOk -and $successCount -gt 0) {
    Write-Host "INTEGRATION READY!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Backend API is ready for frontend integration at http://localhost:3000/reports" -ForegroundColor Green
    Write-Host ""
    Write-Host "NEXT STEPS FOR FRONTEND:" -ForegroundColor Cyan
    Write-Host "1. Use the main endpoint: GET /api/v1/ssot-reports/integrated" -ForegroundColor Gray
    Write-Host "2. Include JWT token in Authorization header" -ForegroundColor Gray
    Write-Host "3. Handle start_date and end_date parameters" -ForegroundColor Gray
    Write-Host "4. Optionally connect to WebSocket at ws://localhost:8080/ws/balance" -ForegroundColor Gray
    Write-Host ""
    Write-Host "Full documentation: REPORTS_API_INTEGRATION.md" -ForegroundColor Cyan
} else {
    Write-Host "INTEGRATION INCOMPLETE" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Some endpoints are not working. Please check:" -ForegroundColor Yellow
    Write-Host "- Server is running and properly configured" -ForegroundColor Yellow
    Write-Host "- Database is migrated and has sample data" -ForegroundColor Yellow
    Write-Host "- SSOT Journal system is properly initialized" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "============================================================" -ForegroundColor Cyan