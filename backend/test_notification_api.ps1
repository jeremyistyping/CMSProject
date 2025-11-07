# Test Notification API Endpoints
# This script tests the notification API endpoints to verify they work correctly

param(
    [string]$BaseUrl = "http://localhost:8080",
    [string]$Token = ""
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  TESTING NOTIFICATION API ENDPOINTS" -ForegroundColor Cyan  
Write-Host "========================================" -ForegroundColor Cyan

if ($Token -eq "") {
    Write-Host "❌ Error: Token is required. Please provide a valid JWT token." -ForegroundColor Red
    Write-Host "Usage: ./test_notification_api.ps1 -Token 'your_jwt_token_here'" -ForegroundColor Yellow
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $Token"
    "Content-Type" = "application/json"
}

Write-Host "🔍 Testing notification endpoints..." -ForegroundColor Yellow
Write-Host ""

# Test 1: All notifications
Write-Host "1️⃣ Testing: GET /api/v1/notifications" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/v1/notifications" -Method GET -Headers $headers
    Write-Host "✅ SUCCESS: Found $($response.total) total notifications" -ForegroundColor Green
    
    if ($response.notifications.Count -gt 0) {
        Write-Host "📋 Latest notifications:" -ForegroundColor Cyan
        foreach ($notif in $response.notifications | Select-Object -First 3) {
            Write-Host "   - ID: $($notif.id) | Type: $($notif.type) | Read: $($notif.is_read)" -ForegroundColor White
            Write-Host "   - Title: $($notif.title)" -ForegroundColor White
            Write-Host "   - Created: $($notif.created_at)" -ForegroundColor White
            Write-Host ""
        }
    }
} catch {
    Write-Host "❌ FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 2: LOW_STOCK notifications
Write-Host "2️⃣ Testing: GET /api/v1/notifications/type/LOW_STOCK" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/v1/notifications/type/LOW_STOCK" -Method GET -Headers $headers
    Write-Host "✅ SUCCESS: Found $($response.total) LOW_STOCK notifications" -ForegroundColor Green
    
    if ($response.notifications.Count -gt 0) {
        Write-Host "🚨 LOW_STOCK notifications:" -ForegroundColor Red
        foreach ($notif in $response.notifications) {
            Write-Host "   - ID: $($notif.id) | Read: $($notif.is_read) | Priority: $($notif.priority)" -ForegroundColor White
            Write-Host "   - Message: $($notif.message)" -ForegroundColor White
            Write-Host "   - Created: $($notif.created_at)" -ForegroundColor White
            Write-Host ""
        }
    }
} catch {
    Write-Host "❌ FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 3: MIN_STOCK notifications (alternative endpoint)
Write-Host "3️⃣ Testing: GET /api/v1/notifications/type/MIN_STOCK" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/v1/notifications/type/MIN_STOCK" -Method GET -Headers $headers
    Write-Host "✅ SUCCESS: Found $($response.total) MIN_STOCK notifications" -ForegroundColor Green
    
    if ($response.notifications.Count -gt 0) {
        Write-Host "🔴 MIN_STOCK notifications:" -ForegroundColor Red
        foreach ($notif in $response.notifications) {
            Write-Host "   - ID: $($notif.id) | Read: $($notif.is_read)" -ForegroundColor White
            Write-Host "   - Message: $($notif.message)" -ForegroundColor White
            Write-Host ""
        }
    }
} catch {
    Write-Host "❌ FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 4: REORDER_ALERT notifications
Write-Host "4️⃣ Testing: GET /api/v1/notifications/type/REORDER_ALERT" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/v1/notifications/type/REORDER_ALERT" -Method GET -Headers $headers
    Write-Host "✅ SUCCESS: Found $($response.total) REORDER_ALERT notifications" -ForegroundColor Green
    
    if ($response.notifications.Count -gt 0) {
        Write-Host "📋 REORDER_ALERT notifications:" -ForegroundColor Yellow
        foreach ($notif in $response.notifications | Select-Object -First 2) {
            Write-Host "   - ID: $($notif.id) | Read: $($notif.is_read)" -ForegroundColor White
            Write-Host "   - Message: $($notif.message)" -ForegroundColor White
            Write-Host ""
        }
    }
} catch {
    Write-Host "❌ FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 5: Dashboard stock alerts
Write-Host "5️⃣ Testing: GET /api/v1/dashboard/stock-alerts" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/v1/dashboard/stock-alerts" -Method GET -Headers $headers
    Write-Host "✅ SUCCESS: Dashboard stock alerts endpoint working" -ForegroundColor Green
    
    $alertData = $response.data
    Write-Host "📊 Dashboard Alert Data:" -ForegroundColor Cyan
    Write-Host "   - Total alerts: $($alertData.total_count)" -ForegroundColor White
    Write-Host "   - Show banner: $($alertData.show_banner)" -ForegroundColor White
    
    if ($alertData.alerts.Count -gt 0) {
        Write-Host "🚨 Active alerts:" -ForegroundColor Red
        foreach ($alert in $alertData.alerts) {
            Write-Host "   - Product: $($alert.product_name) ($($alert.product_code))" -ForegroundColor White
            Write-Host "   - Current Stock: $($alert.current_stock) | Threshold: $($alert.threshold_stock)" -ForegroundColor White
            Write-Host "   - Alert Type: $($alert.alert_type) | Urgency: $($alert.urgency)" -ForegroundColor White
            Write-Host "   - Message: $($alert.message)" -ForegroundColor White
            Write-Host ""
        }
    }
} catch {
    Write-Host "❌ FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 6: Unread notification count
Write-Host "6️⃣ Testing: GET /api/v1/notifications/unread-count" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/v1/notifications/unread-count" -Method GET -Headers $headers
    Write-Host "✅ SUCCESS: Unread count: $($response.count)" -ForegroundColor Green
    
    if ($response.count -gt 0) {
        Write-Host "🔔 You have $($response.count) unread notifications!" -ForegroundColor Yellow
    } else {
        Write-Host "✅ No unread notifications" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ FAILED: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test 7: Approval notifications
Write-Host "7️⃣ Testing: GET /api/v1/notifications/approvals" -ForegroundColor Yellow
try {
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/v1/notifications/approvals" -Method GET -Headers $headers
    Write-Host "✅ SUCCESS: Found $($response.total) approval notifications" -ForegroundColor Green
} catch {
    Write-Host "❌ FAILED: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  API TESTING COMPLETED" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

Write-Host ""
Write-Host "🎯 SUMMARY:" -ForegroundColor Cyan
Write-Host "If all endpoints return SUCCESS with data, then:" -ForegroundColor White
Write-Host "✅ Backend notification system is working perfectly" -ForegroundColor Green
Write-Host "⚠️  Frontend should call these APIs to display notifications" -ForegroundColor Yellow
Write-Host ""
Write-Host "💡 FRONTEND INTEGRATION:" -ForegroundColor Cyan
Write-Host "1. Call /api/v1/notifications/type/LOW_STOCK for minimum stock alerts" -ForegroundColor White
Write-Host "2. Call /api/v1/dashboard/stock-alerts for dashboard banner" -ForegroundColor White
Write-Host "3. Call /api/v1/notifications/unread-count for notification badge" -ForegroundColor White
Write-Host "4. Poll these endpoints every 30 seconds for real-time updates" -ForegroundColor White