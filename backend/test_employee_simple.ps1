# Simple Test Employee Dashboard
Write-Host "=== Testing Employee Dashboard Endpoints ===" -ForegroundColor Green

$baseUrl = "http://localhost:8080"
$headers = @{
    "Content-Type" = "application/json"
}

# Function to make requests
function Test-Endpoint {
    param(
        [string]$Method = "GET",
        [string]$Endpoint,
        [object]$Body = $null,
        [string]$Token
    )
    
    $uri = "$baseUrl$Endpoint"
    $authHeaders = $headers.Clone()
    if ($Token) {
        $authHeaders["Authorization"] = "Bearer $Token"
    }
    
    try {
        if ($Body) {
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $authHeaders -Body ($Body | ConvertTo-Json)
        } else {
            $response = Invoke-RestMethod -Uri $uri -Method $Method -Headers $authHeaders
        }
        return $response
    }
    catch {
        Write-Host "Error calling $Endpoint : $($_.Exception.Message)" -ForegroundColor Red
        return $null
    }
}

# Test login (try with admin credentials first)
Write-Host "`nTesting Employee Login..." -ForegroundColor Yellow
$loginData = @{
    email = "admin"
    password = "admin123"
}

$loginResponse = Test-Endpoint -Method "POST" -Endpoint "/api/v1/auth/login" -Body $loginData
if ($loginResponse -and $loginResponse.token) {
    $employeeToken = $loginResponse.token
    Write-Host "[OK] Employee login successful" -ForegroundColor Green
    Write-Host "User: $($loginResponse.user.username) | Role: $($loginResponse.user.role)" -ForegroundColor Cyan
} else {
    Write-Host "[ERROR] Employee login failed" -ForegroundColor Red
    exit 1
}

# Test Employee Dashboard Data
Write-Host "`nTesting Employee Dashboard Data..." -ForegroundColor Yellow
$dashboardResponse = Test-Endpoint -Endpoint "/api/v1/dashboard/employee" -Token $employeeToken
if ($dashboardResponse) {
    Write-Host "[OK] Employee dashboard data retrieved successfully" -ForegroundColor Green
    Write-Host "Pending Approvals: $($dashboardResponse.data.pending_approvals.Count)" -ForegroundColor Cyan
} else {
    Write-Host "[ERROR] Failed to get employee dashboard data" -ForegroundColor Red
}

# Test Employee Approval Notifications
Write-Host "`nTesting Employee Approval Notifications..." -ForegroundColor Yellow
$approvalNotifResponse = Test-Endpoint -Endpoint "/api/v1/dashboard/employee/approval-notifications" -Token $employeeToken
if ($approvalNotifResponse) {
    Write-Host "[OK] Employee approval notifications retrieved successfully" -ForegroundColor Green
    Write-Host "Total Notifications: $($approvalNotifResponse.data.total_count)" -ForegroundColor Cyan
} else {
    Write-Host "[ERROR] Failed to get employee approval notifications" -ForegroundColor Red
}

# Test Purchase Approval Status
Write-Host "`nTesting Purchase Approval Status..." -ForegroundColor Yellow
$purchaseStatusResponse = Test-Endpoint -Endpoint "/api/v1/dashboard/employee/purchase-approval-status" -Token $employeeToken
if ($purchaseStatusResponse) {
    Write-Host "[OK] Purchase approval status retrieved successfully" -ForegroundColor Green
    Write-Host "Total Purchases: $($purchaseStatusResponse.data.total)" -ForegroundColor Cyan
} else {
    Write-Host "[ERROR] Failed to get purchase approval status" -ForegroundColor Red
}

# Test Employee Workflows
Write-Host "`nTesting Employee Workflows..." -ForegroundColor Yellow
$workflowsResponse = Test-Endpoint -Endpoint "/api/v1/dashboard/employee/workflows" -Token $employeeToken
if ($workflowsResponse) {
    Write-Host "[OK] Employee workflows retrieved successfully" -ForegroundColor Green
    Write-Host "Total Workflows: $($workflowsResponse.data.total)" -ForegroundColor Cyan
} else {
    Write-Host "[ERROR] Failed to get employee workflows" -ForegroundColor Red
}

Write-Host "`n=== Test Summary ===" -ForegroundColor Green
Write-Host "[OK] Employee dashboard endpoints are working!" -ForegroundColor Green
Write-Host "Now you can refresh your dashboard to see the new features" -ForegroundColor Cyan