# Simple SSOT API Endpoint Testing
param(
    [string]$ServerUrl = "http://localhost:8080",
    [string]$TestUser = "admin@example.com", 
    [string]$TestPassword = "admin123"
)

Write-Host "SSOT API Endpoint Testing Suite" -ForegroundColor Green
Write-Host "==============================="
Write-Host "Server URL: $ServerUrl"
Write-Host ""

# Global test results
$script:TotalTests = 0
$script:PassedTests = 0
$script:FailedTests = 0
$script:AuthToken = ""

function Test-Endpoint {
    param(
        [string]$Method,
        [string]$Endpoint,
        [hashtable]$Headers = @{},
        [object]$Body = $null,
        [int]$ExpectedStatus = 200,
        [string]$TestName
    )
    
    $script:TotalTests++
    
    try {
        $uri = "$ServerUrl$Endpoint"
        $requestParams = @{
            Uri = $uri
            Method = $Method
            Headers = $Headers
            UseBasicParsing = $true
            TimeoutSec = 30
        }
        
        if ($Body -and $Method -ne "GET") {
            $requestParams.Body = ($Body | ConvertTo-Json -Depth 10)
            $requestParams.Headers["Content-Type"] = "application/json"
        }
        
        $startTime = Get-Date
        $response = Invoke-WebRequest @requestParams
        $duration = (Get-Date) - $startTime
        
        $actualStatus = $response.StatusCode
        $success = ($actualStatus -eq $ExpectedStatus)
        
        if ($success) {
            $script:PassedTests++
            $status = "PASS"
            $color = "Green"
        } else {
            $script:FailedTests++
            $status = "FAIL"
            $color = "Red"
        }
        
        Write-Host ("  {0,-6} {1} ({2}ms) - Status: {3} (Expected: {4})" -f 
            $status, $TestName, $duration.TotalMilliseconds, $actualStatus, $ExpectedStatus) -ForegroundColor $color
            
        return @{
            Success = $success
            StatusCode = $actualStatus
            Response = $response.Content
            Duration = $duration.TotalMilliseconds
        }
        
    } catch {
        $script:FailedTests++
        $statusCode = 0
        if ($_.Exception.Response) {
            $statusCode = $_.Exception.Response.StatusCode.value__
        }
        
        # Check if this was an expected error (like 401 for auth required)
        $success = ($statusCode -eq $ExpectedStatus)
        if ($success) {
            $script:PassedTests++
            $script:FailedTests--
            $status = "PASS"
            $color = "Green"
        } else {
            $status = "FAIL"
            $color = "Red"
        }
        
        Write-Host ("  {0,-6} {1} - Status: {2} (Expected: {3}) - {4}" -f 
            $status, $TestName, $statusCode, $ExpectedStatus, $_.Exception.Message) -ForegroundColor $color
            
        return @{
            Success = $success
            StatusCode = $statusCode
            Response = $null
            Error = $_.Exception.Message
        }
    }
}

# Test 1: Server Health Check
Write-Host "1. Testing Server Connectivity..." -ForegroundColor Cyan
$result = Test-Endpoint -Method "GET" -Endpoint "/api/v1/journals" -ExpectedStatus 401 -TestName "Server Health (should require auth)"

if (-not $result.Success -and $result.StatusCode -eq 0) {
    Write-Host "  ERROR: Server appears to be offline. Please start the SSOT server first." -ForegroundColor Red
    Write-Host ""
    Write-Host "To start the server, run: go run cmd/main.go" -ForegroundColor Yellow
    exit 1
}

# Test 2: Authentication
Write-Host ""
Write-Host "2. Testing Authentication..." -ForegroundColor Cyan

$loginData = @{
    email = $TestUser
    password = $TestPassword
}

$authResult = Test-Endpoint -Method "POST" -Endpoint "/api/v1/auth/login" -Body $loginData -ExpectedStatus 200 -TestName "User Login"

if ($authResult.Success) {
    try {
        $loginResponse = $authResult.Response | ConvertFrom-Json
        $script:AuthToken = $loginResponse.access_token
        Write-Host "  INFO: Authentication token acquired (length: $($script:AuthToken.Length))" -ForegroundColor Yellow
        Write-Host "  DEBUG: Token starts with: $($script:AuthToken.Substring(0, [Math]::Min(20, $script:AuthToken.Length)))..." -ForegroundColor DarkGray
    } catch {
        Write-Host "  WARNING: Could not extract token from login response" -ForegroundColor Yellow
        Write-Host "  DEBUG: Response was: $($authResult.Response)" -ForegroundColor DarkGray
    }
}

# Prepare auth headers for subsequent tests
$authHeaders = @{}
if ($script:AuthToken) {
    $authHeaders["Authorization"] = "Bearer $script:AuthToken"
}

# Test 3: SSOT Journal Endpoints
Write-Host ""
Write-Host "3. Testing SSOT Journal Endpoints..." -ForegroundColor Cyan

# Test getting journals (with auth)
Write-Host "  DEBUG: Checking auth token: '$($script:AuthToken)' (length: $($script:AuthToken.Length))" -ForegroundColor DarkGray
if ($script:AuthToken -and $script:AuthToken.Length -gt 0) {
    $authHeaders = @{"Authorization" = "Bearer $($script:AuthToken)"}
    Test-Endpoint -Method "GET" -Endpoint "/api/v1/journals" -Headers $authHeaders -ExpectedStatus 200 -TestName "Get All Journals"
    Test-Endpoint -Method "GET" -Endpoint "/api/v1/journals/summary" -Headers $authHeaders -ExpectedStatus 200 -TestName "Get Journal Summary"
    Test-Endpoint -Method "GET" -Endpoint "/api/v1/journals/account-balances" -Headers $authHeaders -ExpectedStatus 200 -TestName "Get Account Balances"
} else {
    Write-Host "  SKIP: Journal endpoints (no auth token)" -ForegroundColor Yellow
    $script:TotalTests += 3
}

# Test 4: Create Journal Entry
Write-Host ""
Write-Host "4. Testing Journal Creation..." -ForegroundColor Cyan

if ($script:AuthToken -and $script:AuthToken.Length -gt 0) {
    $authHeaders = @{"Authorization" = "Bearer $($script:AuthToken)"}
    $journalData = @{
        entry_date = (Get-Date -Format "yyyy-MM-dd")
        description = "PowerShell API Test Entry"
        reference = "PS-TEST-$(Get-Date -Format 'yyyyMMdd-HHmmss')"
        notes = "Created via PowerShell test suite"
        lines = @(
            @{
                account_id = 1
                description = "Test Debit Entry"
                debit_amount = 50000
                credit_amount = 0
            },
            @{
                account_id = 2  
                description = "Test Credit Entry"
                debit_amount = 0
                credit_amount = 50000
            }
        )
    }
    
    $createResult = Test-Endpoint -Method "POST" -Endpoint "/api/v1/journals" -Headers $authHeaders -Body $journalData -ExpectedStatus 201 -TestName "Create Journal Entry"
    
    # Extract journal ID if creation was successful
    $journalId = $null
    if ($createResult.Success -and $createResult.Response) {
        try {
            $createResponse = $createResult.Response | ConvertFrom-Json
            $journalId = $createResponse.data.id
            Write-Host "  INFO: Created journal with ID: $journalId" -ForegroundColor Yellow
        } catch {
            Write-Host "  WARNING: Could not extract journal ID from response" -ForegroundColor Yellow
        }
    }
    
    # Test getting specific journal
    if ($journalId) {
        Test-Endpoint -Method "GET" -Endpoint "/api/v1/journals/$journalId" -Headers $authHeaders -ExpectedStatus 200 -TestName "Get Specific Journal"
    }
    
} else {
    Write-Host "  SKIP: Journal creation (no auth token)" -ForegroundColor Yellow
    $script:TotalTests += 1
}

# Test 5: Performance Check
Write-Host ""
Write-Host "5. Testing Performance..." -ForegroundColor Cyan

if ($script:AuthToken -and $script:AuthToken.Length -gt 0) {
    $authHeaders = @{"Authorization" = "Bearer $($script:AuthToken)"}
    $perfStart = Get-Date
    $perfResult = Test-Endpoint -Method "GET" -Endpoint "/api/v1/journals?limit=100" -Headers $authHeaders -ExpectedStatus 200 -TestName "Performance Test (100 records)"
    $perfDuration = ((Get-Date) - $perfStart).TotalMilliseconds
    
    if ($perfDuration -gt 5000) {
        Write-Host "  WARNING: Performance issue detected - took ${perfDuration}ms (>5000ms)" -ForegroundColor Yellow
    } else {
        Write-Host "  INFO: Performance acceptable - took ${perfDuration}ms" -ForegroundColor Green
    }
} else {
    Write-Host "  SKIP: Performance test (no auth token)" -ForegroundColor Yellow
    $script:TotalTests++
}

# Final Results
Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Test Results Summary" -ForegroundColor Cyan  
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ("Total Tests: {0}" -f $script:TotalTests)
Write-Host ("Passed: {0}" -f $script:PassedTests) -ForegroundColor Green
Write-Host ("Failed: {0}" -f $script:FailedTests) -ForegroundColor Red

$successRate = 0
if ($script:TotalTests -gt 0) {
    $successRate = [math]::Round(($script:PassedTests / $script:TotalTests) * 100, 1)
}
Write-Host ("Success Rate: {0}%" -f $successRate)

Write-Host ""
if ($script:FailedTests -eq 0) {
    Write-Host "🎉 All tests passed! SSOT API is working correctly." -ForegroundColor Green
    $exitCode = 0
} else {
    Write-Host "⚠️  Some tests failed. Please check the issues above." -ForegroundColor Yellow
    $exitCode = 1
}

Write-Host ""
Write-Host "Next Steps:" -ForegroundColor Cyan
Write-Host "• Run comprehensive tests: go run cmd/scripts/comprehensive_ssot_api_tester.go"
Write-Host "• Test frontend integration"
Write-Host "• Monitor system performance"
Write-Host "• Update documentation"

exit $exitCode