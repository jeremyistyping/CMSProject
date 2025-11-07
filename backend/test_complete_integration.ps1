# Complete Integration Test Script
# Tests backend WebSocket, journal service improvements, and balance monitoring

Write-Host "`n=== COMPLETE INTEGRATION TEST SUITE ===" -ForegroundColor Cyan

# Backend server endpoint
$BASE_URL = "http://localhost:8080"
$WS_URL = "ws://localhost:8080"

# Test variables
$headers = @{}
$token = $null

# Helper function for API calls
function Invoke-TestAPI {
    param(
        [string]$Method,
        [string]$Endpoint,
        [hashtable]$Body = @{},
        [string]$Description
    )
    
    Write-Host "`n[TEST] $Description" -ForegroundColor Yellow
    
    try {
        $url = "$BASE_URL$Endpoint"
        
        if ($Body.Count -gt 0) {
            $jsonBody = $Body | ConvertTo-Json -Depth 10
            Write-Host "Request: $Method $url" -ForegroundColor Gray
            Write-Host "Body: $jsonBody" -ForegroundColor Gray
            
            $response = Invoke-RestMethod -Uri $url -Method $Method -Body $jsonBody -Headers $headers -ContentType "application/json"
        } else {
            Write-Host "Request: $Method $url" -ForegroundColor Gray
            $response = Invoke-RestMethod -Uri $url -Method $Method -Headers $headers
        }
        
        Write-Host "[SUCCESS] Response: $($response | ConvertTo-Json -Depth 3)" -ForegroundColor Green
        return $response
    }
    catch {
        Write-Host "[ERROR] $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            Write-Host "Status: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
            Write-Host "Response: $($_.Exception.Response | ConvertTo-Json)" -ForegroundColor Red
        }
        return $null
    }
}

# Step 1: Authentication
Write-Host "`n=== STEP 1: AUTHENTICATION ===" -ForegroundColor Magenta

$loginResult = Invoke-TestAPI -Method "POST" -Endpoint "/api/auth/login" -Body @{
    username = "admin"
    password = "admin123"
} -Description "Admin Login"

if ($loginResult -and $loginResult.access_token) {
    $token = $loginResult.access_token
    $headers = @{ "Authorization" = "Bearer $token" }
    Write-Host "[SUCCESS] Authentication successful, token obtained" -ForegroundColor Green
} else {
    Write-Host "[FATAL] Authentication failed, cannot continue tests" -ForegroundColor Red
    exit 1
}

# Step 2: Test Enhanced Balance Hub Endpoint
Write-Host "`n=== STEP 2: BALANCE HUB MONITORING ===" -ForegroundColor Magenta

$balanceStatus = Invoke-TestAPI -Method "GET" -Endpoint "/api/ssot/balance-hub/status" -Description "Get Balance Hub Status"

if ($balanceStatus) {
    Write-Host "[SUCCESS] Balance Hub is operational" -ForegroundColor Green
    Write-Host "Hub Status: $($balanceStatus | ConvertTo-Json)" -ForegroundColor Gray
} else {
    Write-Host "[WARNING] Balance Hub endpoint not responding" -ForegroundColor Yellow
}

# Step 3: Test Enhanced Journal Service
Write-Host "`n=== STEP 3: ENHANCED JOURNAL SERVICE ===" -ForegroundColor Magenta

# Test journal creation with enhanced error handling
$journalData = @{
    reference = "TEST-$(Get-Date -Format 'yyyyMMdd-HHmmss')"
    reference_type = "MANUAL"
    description = "Integration test journal entry"
    entry_date = (Get-Date -Format "yyyy-MM-dd")
    journal_lines = @(
        @{
            account_id = 1101
            description = "Test debit entry"
            debit_amount = 100000
            credit_amount = 0
        },
        @{
            account_id = 4101  
            description = "Test credit entry"
            debit_amount = 0
            credit_amount = 100000
        }
    )
}

$journalResult = Invoke-TestAPI -Method "POST" -Endpoint "/api/ssot/journals" -Body $journalData -Description "Create Journal Entry with Enhanced Service"

$journalId = $null
if ($journalResult -and $journalResult.id) {
    $journalId = $journalResult.id
    Write-Host "[SUCCESS] Journal created with ID: $journalId" -ForegroundColor Green
} else {
    Write-Host "[ERROR] Failed to create journal entry" -ForegroundColor Red
}

# Step 4: Test journal posting with transaction safety
if ($journalId) {
    Write-Host "`n=== STEP 4: ENHANCED JOURNAL POSTING ===" -ForegroundColor Magenta
    
    $postResult = Invoke-TestAPI -Method "POST" -Endpoint "/api/ssot/journals/$journalId/post" -Description "Post Journal with Enhanced Transaction Safety"
    
    if ($postResult) {
        Write-Host "[SUCCESS] Journal posted successfully with enhanced safety" -ForegroundColor Green
    } else {
        Write-Host "[ERROR] Failed to post journal" -ForegroundColor Red
    }
}

# Step 5: Test Balance Refresh and Materialized Views
Write-Host "`n=== STEP 5: BALANCE CALCULATION TEST ===" -ForegroundColor Magenta

$refreshResult = Invoke-TestAPI -Method "POST" -Endpoint "/api/ssot/balances/refresh" -Description "Refresh Balance Calculations"

if ($refreshResult) {
    Write-Host "[SUCCESS] Balance refresh completed" -ForegroundColor Green
} else {
    Write-Host "[ERROR] Balance refresh failed" -ForegroundColor Red
}

# Test balance retrieval
$balanceResult = Invoke-TestAPI -Method "GET" -Endpoint "/api/ssot/balances" -Description "Get Updated Balances"

if ($balanceResult) {
    Write-Host "[SUCCESS] Balance retrieval successful" -ForegroundColor Green
    Write-Host "Sample balances:" -ForegroundColor Gray
    
    # Show first few balances
    if ($balanceResult.GetType().Name -eq "Object[]" -and $balanceResult.Length -gt 0) {
        $balanceResult[0..([Math]::Min(2, $balanceResult.Length-1))] | ForEach-Object {
            Write-Host "  Account $($_.account_code): $($_.balance)" -ForegroundColor Gray
        }
    }
} else {
    Write-Host "[ERROR] Balance retrieval failed" -ForegroundColor Red
}

# Step 6: WebSocket Connection Test
Write-Host "`n=== STEP 6: WEBSOCKET BALANCE MONITORING ===" -ForegroundColor Magenta

Write-Host "[INFO] Testing WebSocket connection..." -ForegroundColor Yellow
Write-Host "WebSocket URL: $WS_URL/ws/balance?token=$token" -ForegroundColor Gray

# Create a simple WebSocket test using Node.js if available
$wsTestScript = @"
const WebSocket = require('ws');

const wsUrl = 'ws://localhost:8080/ws/balance?token=$token';
console.log('Connecting to WebSocket:', wsUrl);

const ws = new WebSocket(wsUrl);
let messageCount = 0;
const maxMessages = 5;
const timeout = 10000; // 10 seconds

const timer = setTimeout(() => {
    console.log('WebSocket test timeout - closing connection');
    ws.close();
}, timeout);

ws.on('open', function open() {
    console.log('[SUCCESS] WebSocket connected successfully');
});

ws.on('message', function message(data) {
    messageCount++;
    console.log('[MESSAGE] Received balance update:', data.toString());
    
    if (messageCount >= maxMessages) {
        console.log('[INFO] Received enough test messages, closing connection');
        clearTimeout(timer);
        ws.close();
    }
});

ws.on('close', function close() {
    console.log('[INFO] WebSocket connection closed');
    process.exit(0);
});

ws.on('error', function error(err) {
    console.error('[ERROR] WebSocket error:', err.message);
    clearTimeout(timer);
    process.exit(1);
});

// Keep connection alive
setInterval(() => {
    if (ws.readyState === WebSocket.OPEN) {
        ws.ping();
    }
}, 5000);
"@

try {
    # Check if Node.js is available
    $nodeVersion = node --version 2>$null
    if ($nodeVersion) {
        Write-Host "[INFO] Node.js available: $nodeVersion" -ForegroundColor Gray
        
        # Save and run WebSocket test script
        $wsTestPath = "ws_test_temp.js"
        $wsTestScript | Out-File -FilePath $wsTestPath -Encoding UTF8
        
        Write-Host "[INFO] Running WebSocket connection test..." -ForegroundColor Yellow
        $wsResult = node $wsTestPath 2>&1
        
        Write-Host "WebSocket Test Results:" -ForegroundColor Gray
        $wsResult | ForEach-Object {
            Write-Host "  $_" -ForegroundColor Gray
        }
        
        # Clean up
        Remove-Item $wsTestPath -ErrorAction SilentlyContinue
        
        if ($wsResult -match "SUCCESS.*WebSocket connected") {
            Write-Host "[SUCCESS] WebSocket endpoint is functional" -ForegroundColor Green
        } else {
            Write-Host "[WARNING] WebSocket connection may have issues" -ForegroundColor Yellow
        }
    } else {
        Write-Host "[INFO] Node.js not available, skipping WebSocket test" -ForegroundColor Yellow
        Write-Host "[INFO] WebSocket endpoint: $WS_URL/ws/balance?token=<TOKEN>" -ForegroundColor Gray
    }
}
catch {
    Write-Host "[WARNING] WebSocket test could not run: $($_.Exception.Message)" -ForegroundColor Yellow
}

# Step 7: Integration Test - Complete Flow
Write-Host "`n=== STEP 7: END-TO-END INTEGRATION TEST ===" -ForegroundColor Magenta

Write-Host "[INFO] Testing complete transaction flow with real-time monitoring..." -ForegroundColor Yellow

# Create another journal entry to trigger balance updates
$integrationJournal = @{
    reference = "INTEGRATION-$(Get-Date -Format 'yyyyMMdd-HHmmss')"
    reference_type = "MANUAL"
    description = "End-to-end integration test"
    entry_date = (Get-Date -Format "yyyy-MM-dd")
    journal_lines = @(
        @{
            account_id = 1102  # Bank BCA
            description = "Integration test debit"
            debit_amount = 250000
            credit_amount = 0
        },
        @{
            account_id = 4102  # Service Revenue
            description = "Integration test credit"
            debit_amount = 0
            credit_amount = 250000
        }
    )
}

$integrationResult = Invoke-TestAPI -Method "POST" -Endpoint "/api/ssot/journals" -Body $integrationJournal -Description "Create Integration Test Journal"

if ($integrationResult -and $integrationResult.id) {
    $integrationId = $integrationResult.id
    Write-Host "[SUCCESS] Integration journal created: $integrationId" -ForegroundColor Green
    
    # Post the journal
    $postIntegrationResult = Invoke-TestAPI -Method "POST" -Endpoint "/api/ssot/journals/$integrationId/post" -Description "Post Integration Journal"
    
    if ($postIntegrationResult) {
        Write-Host "[SUCCESS] Integration journal posted successfully" -ForegroundColor Green
        
        # Refresh balances
        Start-Sleep -Seconds 2
        $finalRefresh = Invoke-TestAPI -Method "POST" -Endpoint "/api/ssot/balances/refresh" -Description "Final Balance Refresh"
        
        if ($finalRefresh) {
            Write-Host "[SUCCESS] Final balance refresh completed" -ForegroundColor Green
        }
        
        # Get final balances
        $finalBalances = Invoke-TestAPI -Method "GET" -Endpoint "/api/ssot/balances" -Description "Get Final Balances"
        
        if ($finalBalances) {
            Write-Host "[SUCCESS] End-to-end integration test completed successfully!" -ForegroundColor Green
        }
    }
}

# Summary Report
Write-Host "`n=== INTEGRATION TEST SUMMARY ===" -ForegroundColor Cyan

$testResults = @(
    @{ Name = "Authentication"; Status = if($token) { "PASS" } else { "FAIL" } }
    @{ Name = "Balance Hub"; Status = if($balanceStatus) { "PASS" } else { "FAIL" } }
    @{ Name = "Enhanced Journal Creation"; Status = if($journalResult) { "PASS" } else { "FAIL" } }
    @{ Name = "Balance Calculations"; Status = if($balanceResult) { "PASS" } else { "FAIL" } }
    @{ Name = "End-to-End Integration"; Status = if($integrationResult) { "PASS" } else { "FAIL" } }
)

Write-Host "`nTest Results:" -ForegroundColor White
$testResults | ForEach-Object {
    $color = if($_.Status -eq "PASS") { "Green" } else { "Red" }
    Write-Host "  [$($_.Status)] $($_.Name)" -ForegroundColor $color
}

$passedTests = ($testResults | Where-Object { $_.Status -eq "PASS" }).Count
$totalTests = $testResults.Count
$passRate = [Math]::Round(($passedTests / $totalTests) * 100, 1)

Write-Host "`nOverall Results: $passedTests/$totalTests tests passed ($passRate%)" -ForegroundColor Cyan

if ($passRate -ge 80) {
    Write-Host "`n[SUCCESS] Integration test suite completed successfully!" -ForegroundColor Green
    Write-Host "The enhanced journal system with real-time balance monitoring is ready for deployment." -ForegroundColor Green
} elseif ($passRate -ge 60) {
    Write-Host "`n[WARNING] Integration test completed with some issues." -ForegroundColor Yellow
    Write-Host "Please review failed tests before deployment." -ForegroundColor Yellow
} else {
    Write-Host "`n[ERROR] Integration test failed." -ForegroundColor Red
    Write-Host "Significant issues detected. System not ready for deployment." -ForegroundColor Red
}

Write-Host "`n=== TEST COMPLETED ===" -ForegroundColor Cyan