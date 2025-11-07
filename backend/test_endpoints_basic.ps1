# Basic Endpoint Testing Script
# Tests basic API endpoints without requiring authentication

param(
    [string]$BaseUrl = "http://localhost:8080"
)

$ErrorActionPreference = 'Continue'

Write-Host "🧪 Basic API Endpoint Testing" -ForegroundColor Green
Write-Host "============================" -ForegroundColor Green
Write-Host "Base URL: $BaseUrl" -ForegroundColor Cyan
Write-Host ""

# Test 1: Health Check
Write-Host "[1/4] Testing Health Check..." -ForegroundColor Cyan
try {
    $healthResp = Invoke-RestMethod -Uri "$BaseUrl/api/v1/health" -Method Get -TimeoutSec 5
    Write-Host "✅ Health Check: $($healthResp.status)" -ForegroundColor Green
} catch {
    Write-Host "❌ Health Check Failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 2: Check if WebSocket endpoint is available (should return 401 without auth)
Write-Host "[2/4] Testing WebSocket Endpoint Availability..." -ForegroundColor Cyan
try {
    $wsResp = Invoke-WebRequest -Uri "$BaseUrl/api/v1/journals/account-balances/ws" -Method Get -TimeoutSec 5
    if ($wsResp.StatusCode -eq 401) {
        Write-Host "✅ WebSocket Endpoint: Available (requires authentication)" -ForegroundColor Green
    } else {
        Write-Host "⚠️ WebSocket Endpoint: Unexpected response $($wsResp.StatusCode)" -ForegroundColor Yellow
    }
} catch {
    if ($_.Exception.Response.StatusCode -eq 401) {
        Write-Host "✅ WebSocket Endpoint: Available (requires authentication)" -ForegroundColor Green
    } else {
        Write-Host "❌ WebSocket Endpoint Failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 3: Check Journal API endpoint (should return 401 without auth)
Write-Host "[3/4] Testing Journal API Endpoint..." -ForegroundColor Cyan
try {
    $journalResp = Invoke-WebRequest -Uri "$BaseUrl/api/v1/journals" -Method Get -TimeoutSec 5
    if ($journalResp.StatusCode -eq 401) {
        Write-Host "✅ Journal API: Available (requires authentication)" -ForegroundColor Green
    } else {
        Write-Host "⚠️ Journal API: Unexpected response $($journalResp.StatusCode)" -ForegroundColor Yellow
    }
} catch {
    if ($_.Exception.Response.StatusCode -eq 401) {
        Write-Host "✅ Journal API: Available (requires authentication)" -ForegroundColor Green
    } else {
        Write-Host "❌ Journal API Failed: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test 4: Test if server supports WebSocket upgrade (connection will fail without auth, but server should respond)
Write-Host "[4/4] Testing WebSocket Connection..." -ForegroundColor Cyan
try {
    # Use .NET WebSocket client for testing
    $ws = New-Object System.Net.WebSockets.ClientWebSocket
    $uri = [System.Uri]::new("$($BaseUrl.Replace('http', 'ws'))/api/v1/journals/account-balances/ws")
    
    $cancellation = [System.Threading.CancellationToken]::None
    $task = $ws.ConnectAsync($uri, $cancellation)
    
    # Wait with timeout
    $timeout = [TimeSpan]::FromSeconds(3)
    if ($task.Wait($timeout)) {
        if ($ws.State -eq [System.Net.WebSockets.WebSocketState]::Open) {
            Write-Host "✅ WebSocket: Connection successful" -ForegroundColor Green
            $ws.CloseAsync([System.Net.WebSockets.WebSocketCloseStatus]::NormalClosure, "Test complete", $cancellation).Wait(1000)
        } else {
            Write-Host "⚠️ WebSocket: Connected but state is $($ws.State)" -ForegroundColor Yellow
        }
    } else {
        Write-Host "⚠️ WebSocket: Connection timeout (likely needs authentication)" -ForegroundColor Yellow
    }
} catch {
    Write-Host "⚠️ WebSocket: $($_.Exception.Message)" -ForegroundColor Yellow
    Write-Host "   (This is normal if authentication is required)" -ForegroundColor Gray
} finally {
    if ($ws) { $ws.Dispose() }
}

Write-Host ""
Write-Host "🎯 Summary:" -ForegroundColor Green
Write-Host "   - Health check should pass" -ForegroundColor Gray
Write-Host "   - Protected endpoints should return 401 (authentication required)" -ForegroundColor Gray
Write-Host "   - WebSocket upgrade should be supported by server" -ForegroundColor Gray
Write-Host ""

if (Test-Path "test_websocket_client.js") {
    Write-Host "💡 To test WebSocket with Node.js:" -ForegroundColor Yellow
    Write-Host "   node test_websocket_client.js $($BaseUrl.Replace('http', 'ws'))" -ForegroundColor Cyan
}

Write-Host ""
Write-Host "✨ Basic endpoint testing completed!" -ForegroundColor Green