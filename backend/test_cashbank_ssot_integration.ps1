# Test CashBank SSOT Integration Endpoints
# This script tests the integration between CashBank system and SSOT Journal system

Write-Host "=== Testing CashBank SSOT Integration ===" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Gray

$baseURL = "http://localhost:8080"
$token = "" # You'll need to set this with a valid JWT token

# Function to make authenticated API calls
function Invoke-AuthenticatedRequest {
    param(
        [string]$Uri,
        [string]$Method = "GET",
        [object]$Body = $null
    )
    
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }
    
    try {
        if ($Body) {
            $response = Invoke-RestMethod -Uri $Uri -Method $Method -Headers $headers -Body ($Body | ConvertTo-Json) -ErrorAction Stop
        } else {
            $response = Invoke-RestMethod -Uri $Uri -Method $Method -Headers $headers -ErrorAction Stop
        }
        return $response
    } catch {
        Write-Host "   ❌ Request failed: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            $errorResponse = $_.Exception.Response.GetResponseStream()
            $reader = New-Object System.IO.StreamReader($errorResponse)
            $responseBody = $reader.ReadToEnd()
            Write-Host "   Error response: $responseBody" -ForegroundColor Yellow
        }
        return $null
    }
}

# Check if token is set
if ([string]::IsNullOrEmpty($token)) {
    Write-Host "❌ Please set the JWT token in the script" -ForegroundColor Red
    Write-Host "   You can get a token by:" -ForegroundColor Yellow
    Write-Host "   1. Login via POST /api/v1/auth/login" -ForegroundColor Yellow
    Write-Host "   2. Copy the token from the response" -ForegroundColor Yellow
    Write-Host "   3. Set it in the `$token variable in this script" -ForegroundColor Yellow
    exit 1
}

# Test 1: Check backend health
Write-Host "1. Testing backend health..." -ForegroundColor Yellow
try {
    $healthResponse = Invoke-RestMethod -Uri "$baseURL/api/v1/health" -TimeoutSec 5 -ErrorAction Stop
    Write-Host "   ✅ Backend is running" -ForegroundColor Green
} catch {
    Write-Host "   ❌ Backend is not running!" -ForegroundColor Red
    exit 1
}

# Test 2: Get integrated summary
Write-Host "2. Testing integrated summary..." -ForegroundColor Yellow
$summaryResponse = Invoke-AuthenticatedRequest -Uri "$baseURL/api/v1/cashbank/integrated/summary"
if ($summaryResponse) {
    Write-Host "   ✅ Integrated summary retrieved successfully" -ForegroundColor Green
    Write-Host "   Total Accounts: $($summaryResponse.data.sync_status.total_accounts)" -ForegroundColor Cyan
    Write-Host "   Synced Accounts: $($summaryResponse.data.sync_status.synced_accounts)" -ForegroundColor Cyan
    Write-Host "   Variance Accounts: $($summaryResponse.data.sync_status.variance_accounts)" -ForegroundColor Cyan
    Write-Host "   Total Balance: $($summaryResponse.data.summary.total_balance)" -ForegroundColor Cyan
    Write-Host "   Total SSOT Balance: $($summaryResponse.data.summary.total_ssot_balance)" -ForegroundColor Cyan
    Write-Host "   Balance Variance: $($summaryResponse.data.summary.balance_variance)" -ForegroundColor Cyan
    
    # Store first account ID for subsequent tests
    if ($summaryResponse.data.accounts -and $summaryResponse.data.accounts.Count -gt 0) {
        $testAccountId = $summaryResponse.data.accounts[0].id
        Write-Host "   Using Account ID $testAccountId for detailed tests" -ForegroundColor Cyan
    } else {
        Write-Host "   ⚠️  No accounts found for detailed testing" -ForegroundColor Yellow
        $testAccountId = $null
    }
}

# Test 3: Get integrated account details (if account available)
if ($testAccountId) {
    Write-Host "3. Testing integrated account details..." -ForegroundColor Yellow
    $accountResponse = Invoke-AuthenticatedRequest -Uri "$baseURL/api/v1/cashbank/integrated/accounts/$testAccountId"
    if ($accountResponse) {
        Write-Host "   ✅ Integrated account details retrieved successfully" -ForegroundColor Green
        Write-Host "   Account Name: $($accountResponse.data.account.name)" -ForegroundColor Cyan
        Write-Host "   CashBank Balance: $($accountResponse.data.account.balance)" -ForegroundColor Cyan
        Write-Host "   SSOT Balance: $($accountResponse.data.ssot_balance)" -ForegroundColor Cyan
        Write-Host "   Balance Difference: $($accountResponse.data.balance_difference)" -ForegroundColor Cyan
        Write-Host "   Reconciliation Status: $($accountResponse.data.reconciliation_status)" -ForegroundColor Cyan
        Write-Host "   Recent Transactions Count: $($accountResponse.data.recent_transactions.Count)" -ForegroundColor Cyan
        Write-Host "   Related Journal Entries Count: $($accountResponse.data.related_journal_entries.Count)" -ForegroundColor Cyan
    }
    
    # Test 4: Get account reconciliation
    Write-Host "4. Testing account reconciliation..." -ForegroundColor Yellow
    $reconciliationResponse = Invoke-AuthenticatedRequest -Uri "$baseURL/api/v1/cashbank/integrated/accounts/$testAccountId/reconciliation"
    if ($reconciliationResponse) {
        Write-Host "   ✅ Account reconciliation retrieved successfully" -ForegroundColor Green
        Write-Host "   Reconciliation Status: $($reconciliationResponse.data.reconciliation_status)" -ForegroundColor Cyan
        Write-Host "   Recommendations Count: $($reconciliationResponse.data.recommendations.Count)" -ForegroundColor Cyan
        
        if ($reconciliationResponse.data.recommendations) {
            Write-Host "   Recommendations:" -ForegroundColor Cyan
            foreach ($recommendation in $reconciliationResponse.data.recommendations) {
                Write-Host "   - $recommendation" -ForegroundColor White
            }
        }
    }
    
    # Test 5: Get journal entries
    Write-Host "5. Testing journal entries..." -ForegroundColor Yellow
    $journalResponse = Invoke-AuthenticatedRequest -Uri "$baseURL/api/v1/cashbank/integrated/accounts/$testAccountId/journal-entries?page=1&limit=5"
    if ($journalResponse) {
        Write-Host "   ✅ Journal entries retrieved successfully" -ForegroundColor Green
        Write-Host "   Journal Entries Count: $($journalResponse.data.journal_entries.Count)" -ForegroundColor Cyan
        Write-Host "   Current Page: $($journalResponse.data.pagination.page)" -ForegroundColor Cyan
        Write-Host "   Total Pages: $($journalResponse.data.pagination.total_pages)" -ForegroundColor Cyan
        
        # Show first journal entry details if available
        if ($journalResponse.data.journal_entries -and $journalResponse.data.journal_entries.Count -gt 0) {
            $firstEntry = $journalResponse.data.journal_entries[0]
            Write-Host "   First Entry:" -ForegroundColor Cyan
            Write-Host "   - Entry Number: $($firstEntry.entry_number)" -ForegroundColor White
            Write-Host "   - Description: $($firstEntry.description)" -ForegroundColor White
            Write-Host "   - Total Debit: $($firstEntry.total_debit)" -ForegroundColor White
            Write-Host "   - Status: $($firstEntry.status)" -ForegroundColor White
        }
    }
    
    # Test 6: Get transaction history
    Write-Host "6. Testing transaction history..." -ForegroundColor Yellow
    $transactionResponse = Invoke-AuthenticatedRequest -Uri "$baseURL/api/v1/cashbank/integrated/accounts/$testAccountId/transactions?page=1&limit=5"
    if ($transactionResponse) {
        Write-Host "   ✅ Transaction history retrieved successfully" -ForegroundColor Green
        Write-Host "   Transactions Count: $($transactionResponse.data.transactions.Count)" -ForegroundColor Cyan
        
        # Show first transaction details if available
        if ($transactionResponse.data.transactions -and $transactionResponse.data.transactions.Count -gt 0) {
            $firstTx = $transactionResponse.data.transactions[0]
            Write-Host "   First Transaction:" -ForegroundColor Cyan
            Write-Host "   - Amount: $($firstTx.amount)" -ForegroundColor White
            Write-Host "   - Type: $($firstTx.type)" -ForegroundColor White
            Write-Host "   - Date: $($firstTx.date)" -ForegroundColor White
            if ($firstTx.journal_entry_number) {
                Write-Host "   - Journal Entry: $($firstTx.journal_entry_number)" -ForegroundColor White
            }
        }
    }
} else {
    Write-Host "3-6. Skipping detailed tests - no accounts available" -ForegroundColor Yellow
}

# Test 7: Error handling - Invalid account ID
Write-Host "7. Testing error handling..." -ForegroundColor Yellow
$errorResponse = Invoke-AuthenticatedRequest -Uri "$baseURL/api/v1/cashbank/integrated/accounts/99999"
if ($errorResponse -eq $null) {
    Write-Host "   ✅ Error handling works correctly (invalid account ID rejected)" -ForegroundColor Green
} else {
    Write-Host "   ⚠️  Error handling may need improvement" -ForegroundColor Yellow
}

Write-Host "=========================================" -ForegroundColor Gray
Write-Host "Integration Testing Complete!" -ForegroundColor Green
Write-Host "" -ForegroundColor Gray
Write-Host "Next Steps for Frontend Integration:" -ForegroundColor Cyan
Write-Host "1. Update frontend to call these integrated endpoints" -ForegroundColor White
Write-Host "2. Create enhanced UI components for variance display" -ForegroundColor White
Write-Host "3. Implement reconciliation features" -ForegroundColor White
Write-Host "4. Add real-time balance monitoring" -ForegroundColor White
Write-Host "" -ForegroundColor Gray
Write-Host "Available Integrated Endpoints:" -ForegroundColor Yellow
Write-Host "- GET /api/v1/cashbank/integrated/summary" -ForegroundColor White
Write-Host "- GET /api/v1/cashbank/integrated/accounts/:id" -ForegroundColor White
Write-Host "- GET /api/v1/cashbank/integrated/accounts/:id/reconciliation" -ForegroundColor White
Write-Host "- GET /api/v1/cashbank/integrated/accounts/:id/journal-entries" -ForegroundColor White
Write-Host "- GET /api/v1/cashbank/integrated/accounts/:id/transactions" -ForegroundColor White