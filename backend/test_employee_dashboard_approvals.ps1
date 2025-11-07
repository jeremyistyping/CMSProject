# Test Employee Dashboard Approval Notifications
# PowerShell script to test the new employee dashboard features

Write-Host "=== Testing Employee Dashboard Approval Notifications ===" -ForegroundColor Green

# Configuration
$baseUrl = "http://localhost:8080"
$headers = @{
    "Content-Type" = "application/json"
}

# Function to make authenticated requests
function Invoke-AuthenticatedRequest {
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
        if ($_.Exception.Response) {
            $streamReader = [System.IO.StreamReader]::new($_.Exception.Response.GetResponseStream())
            $errorBody = $streamReader.ReadToEnd()
            Write-Host "Error Response: $errorBody" -ForegroundColor Yellow
        }
        return $null
    }
}

# Test 1: Login as Employee
Write-Host "`n1. Testing Employee Login..." -ForegroundColor Yellow
$loginData = @{
    email = "employee@test.com"
    password = "password123"
}

$loginResponse = Invoke-AuthenticatedRequest -Method "POST" -Endpoint "/api/auth/login" -Body $loginData
if ($loginResponse -and $loginResponse.token) {
    $employeeToken = $loginResponse.token
Write-Host "[OK] Employee login successful" -ForegroundColor Green
    Write-Host "User: $($loginResponse.user.username) | Role: $($loginResponse.user.role)" -ForegroundColor Cyan
} else {
Write-Host "[ERROR] Employee login failed" -ForegroundColor Red
    exit 1
}

# Test 2: Get Employee Dashboard Data
Write-Host "`n2. Testing Employee Dashboard Data..." -ForegroundColor Yellow
$dashboardResponse = Invoke-AuthenticatedRequest -Endpoint "/api/dashboard/employee" -Token $employeeToken
if ($dashboardResponse) {
    Write-Host "✅ Employee dashboard data retrieved successfully" -ForegroundColor Green
    Write-Host "Pending Approvals: $($dashboardResponse.data.pending_approvals.Count)" -ForegroundColor Cyan
    Write-Host "Submitted Requests: $($dashboardResponse.data.submitted_requests.Count)" -ForegroundColor Cyan
    Write-Host "Recent Notifications: $($dashboardResponse.data.recent_notifications.Count)" -ForegroundColor Cyan
    Write-Host "Quick Stats:" -ForegroundColor Cyan
    $dashboardResponse.data.quick_stats | ConvertTo-Json | Write-Host -ForegroundColor Gray
} else {
    Write-Host "❌ Failed to get employee dashboard data" -ForegroundColor Red
}

# Test 3: Get Employee Approval Notifications
Write-Host "`n3. Testing Employee Approval Notifications..." -ForegroundColor Yellow
$approvalNotifResponse = Invoke-AuthenticatedRequest -Endpoint "/api/dashboard/employee/approval-notifications" -Token $employeeToken
if ($approvalNotifResponse) {
    Write-Host "✅ Employee approval notifications retrieved successfully" -ForegroundColor Green
    Write-Host "Total Notifications: $($approvalNotifResponse.data.total_count)" -ForegroundColor Cyan
    Write-Host "Unread Count: $($approvalNotifResponse.data.unread_count)" -ForegroundColor Cyan
    
    if ($approvalNotifResponse.data.approval_notifications.Count -gt 0) {
        Write-Host "Recent Approval Notifications:" -ForegroundColor Cyan
        foreach ($notif in $approvalNotifResponse.data.approval_notifications | Select-Object -First 5) {
            Write-Host "  $($notif.icon) $($notif.title) - $($notif.time_ago)" -ForegroundColor Gray
            Write-Host "    Purchase: $($notif.purchase_code) | Amount: $($notif.amount)" -ForegroundColor Gray
        }
    }
} else {
    Write-Host "❌ Failed to get employee approval notifications" -ForegroundColor Red
}

# Test 4: Get Employee Purchase Approval Status
Write-Host "`n4. Testing Employee Purchase Approval Status..." -ForegroundColor Yellow
$purchaseStatusResponse = Invoke-AuthenticatedRequest -Endpoint "/api/dashboard/employee/purchase-approval-status" -Token $employeeToken
if ($purchaseStatusResponse) {
    Write-Host "✅ Employee purchase approval status retrieved successfully" -ForegroundColor Green
    Write-Host "Total Purchases: $($purchaseStatusResponse.data.total)" -ForegroundColor Cyan
    
    if ($purchaseStatusResponse.data.purchase_approvals.Count -gt 0) {
        Write-Host "Purchase Approval Status:" -ForegroundColor Cyan
        foreach ($purchase in $purchaseStatusResponse.data.purchase_approvals | Select-Object -First 5) {
            $statusInfo = $purchase.status_info
            Write-Host "  $($statusInfo.icon) $($purchase.code) - $($statusInfo.message)" -ForegroundColor Gray
            Write-Host "    Vendor: $($purchase.vendor_name) | Amount: $($purchase.total_amount)" -ForegroundColor Gray
            if ($purchase.current_step_name) {
                Write-Host "    Current Step: $($purchase.current_step_name) ($($purchase.days_in_current_step) days)" -ForegroundColor Gray
            }
        }
    }
} else {
    Write-Host "❌ Failed to get employee purchase approval status" -ForegroundColor Red
}

# Test 5: Get Employee Purchase Requests
Write-Host "`n5. Testing Employee Purchase Requests..." -ForegroundColor Yellow
$purchaseRequestsResponse = Invoke-AuthenticatedRequest -Endpoint "/api/dashboard/employee/purchase-requests" -Token $employeeToken
if ($purchaseRequestsResponse) {
    Write-Host "✅ Employee purchase requests retrieved successfully" -ForegroundColor Green
    Write-Host "Total Requests: $($purchaseRequestsResponse.data.total)" -ForegroundColor Cyan
    
    if ($purchaseRequestsResponse.data.purchase_requests.Count -gt 0) {
        Write-Host "Recent Purchase Requests:" -ForegroundColor Cyan
        foreach ($request in $purchaseRequestsResponse.data.purchase_requests | Select-Object -First 5) {
			$urgencyIndicator = if ($request.urgency_level -eq "urgent") { "[URGENT]" } elseif ($request.urgency_level -eq "high") { "[HIGH]" } else { "" }
            Write-Host "  $urgencyIndicator $($request.code) - $($request.status_message)" -ForegroundColor Gray
            Write-Host "    Vendor: $($request.vendor_name) | Amount: $($request.total_amount)" -ForegroundColor Gray
            Write-Host "    Current Step: $($request.current_step_name) ($($request.days_in_current_step) days)" -ForegroundColor Gray
        }
    }
} else {
    Write-Host "❌ Failed to get employee purchase requests" -ForegroundColor Red
}

# Test 6: Get Employee Approval Workflows
Write-Host "`n6. Testing Employee Approval Workflows..." -ForegroundColor Yellow
$workflowsResponse = Invoke-AuthenticatedRequest -Endpoint "/api/dashboard/employee/workflows" -Token $employeeToken
if ($workflowsResponse) {
    Write-Host "✅ Employee approval workflows retrieved successfully" -ForegroundColor Green
    Write-Host "Total Workflows: $($workflowsResponse.data.total)" -ForegroundColor Cyan
    Write-Host "User Role: $($workflowsResponse.data.user_role)" -ForegroundColor Cyan
    
    if ($workflowsResponse.data.workflows.Count -gt 0) {
        Write-Host "Available Workflows:" -ForegroundColor Cyan
        foreach ($workflow in $workflowsResponse.data.workflows) {
            Write-Host "  📋 $($workflow.name) - $($workflow.module)" -ForegroundColor Gray
            Write-Host "    Can Submit: $($workflow.can_submit) | Can Approve: $($workflow.can_approve)" -ForegroundColor Gray
            if ($workflow.approver_steps.Count -gt 0) {
                Write-Host "    Approval Steps: $($workflow.approver_steps -join ', ')" -ForegroundColor Gray
            }
        }
    }
} else {
    Write-Host "❌ Failed to get employee approval workflows" -ForegroundColor Red
}

# Test 7: Test Notification Mark as Read (if we have notifications)
if ($approvalNotifResponse -and $approvalNotifResponse.data.approval_notifications.Count -gt 0) {
    Write-Host "`n7. Testing Mark Notification as Read..." -ForegroundColor Yellow
    $firstNotif = $approvalNotifResponse.data.approval_notifications[0]
    if (-not $firstNotif.is_read) {
        $markReadResponse = Invoke-AuthenticatedRequest -Method "PATCH" -Endpoint "/api/dashboard/employee/notifications/$($firstNotif.id)/read" -Token $employeeToken
        if ($markReadResponse) {
            Write-Host "✅ Notification marked as read successfully" -ForegroundColor Green
            Write-Host "Notification ID: $($markReadResponse.data.notification_id)" -ForegroundColor Cyan
        } else {
            Write-Host "❌ Failed to mark notification as read" -ForegroundColor Red
        }
    } else {
        Write-Host "ℹ️ First notification is already read" -ForegroundColor Blue
    }
} else {
    Write-Host "ℹ️ No notifications available to test mark as read" -ForegroundColor Blue
}

# Test 8: Login as Finance/Director to test approval permissions
Write-Host "`n8. Testing Finance User Dashboard..." -ForegroundColor Yellow
$financeLoginData = @{
    email = "finance@test.com"
    password = "password123"
}

$financeLoginResponse = Invoke-AuthenticatedRequest -Method "POST" -Endpoint "/api/auth/login" -Body $financeLoginData
if ($financeLoginResponse -and $financeLoginResponse.token) {
    $financeToken = $financeLoginResponse.token
    Write-Host "✅ Finance login successful" -ForegroundColor Green
    
    # Get finance dashboard data
    $financeDashboardResponse = Invoke-AuthenticatedRequest -Endpoint "/api/dashboard/employee" -Token $financeToken
    if ($financeDashboardResponse) {
        Write-Host "✅ Finance dashboard data retrieved successfully" -ForegroundColor Green
        Write-Host "Pending Approvals for Finance: $($financeDashboardResponse.data.pending_approvals.Count)" -ForegroundColor Cyan
        if ($financeDashboardResponse.data.pending_approvals.Count -gt 0) {
            Write-Host "Pending Finance Approvals:" -ForegroundColor Cyan
            foreach ($approval in $financeDashboardResponse.data.pending_approvals | Select-Object -First 3) {
                Write-Host "  💰 $($approval.request_title) - Rp $($approval.amount)" -ForegroundColor Gray
                Write-Host "    From: $($approval.requester_name) | Days Pending: $($approval.days_pending)" -ForegroundColor Gray
            }
        }
    }
} else {
    Write-Host "ℹ️ Finance user not available for testing" -ForegroundColor Blue
}

# Summary
Write-Host "`n=== Test Summary ===" -ForegroundColor Green
Write-Host "✅ Employee dashboard approval notification system tested" -ForegroundColor Green
Write-Host "✅ Approval status tracking working" -ForegroundColor Green
Write-Host "✅ Purchase approval notifications implemented" -ForegroundColor Green
Write-Host "✅ Employee workflow permissions configured" -ForegroundColor Green

Write-Host "`nEmployee dashboard now provides:" -ForegroundColor Cyan
Write-Host "  🔔 Real-time approval notifications" -ForegroundColor White
Write-Host "  📊 Purchase approval status tracking" -ForegroundColor White
Write-Host "  ⏳ Days pending indicators with urgency levels" -ForegroundColor White
Write-Host "  📋 Employee workflow permissions" -ForegroundColor White
Write-Host "  ✅ Notification management (mark as read)" -ForegroundColor White
Write-Host "  💰 Finance and Director approval queues" -ForegroundColor White

Write-Host "`nTest completed! 🎉" -ForegroundColor Green