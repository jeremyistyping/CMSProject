# Test Profit & Loss Database Validation
# This script checks actual database values vs P&L report to identify discrepancies

Write-Host "`n=== PROFIT & LOSS DATABASE VALIDATION ===" -ForegroundColor Cyan
Write-Host "Investigating Revenue Duplication Issue" -ForegroundColor Yellow

# Load token
$token = ""
$tokenFile = "token.txt"
if (Test-Path $tokenFile) {
    $token = Get-Content $tokenFile -Raw
    $token = $token.Trim()
} else {
    Write-Host "`n[ERROR] token.txt not found. Please login first." -ForegroundColor Red
    exit 1
}

$baseUrl = "http://localhost:8080"
$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

# Test parameters
$startDate = "2025-01-01"
$endDate = "2025-12-31"

Write-Host "`n=== STEP 1: CHECK CHART OF ACCOUNTS ===" -ForegroundColor Cyan
try {
    $accountsUrl = "${baseUrl}/api/v1/accounts"
    Write-Host "Fetching accounts from: $accountsUrl" -ForegroundColor Gray
    $accountsResponse = Invoke-RestMethod -Uri $accountsUrl -Method Get -Headers $headers
    
    # Filter revenue accounts (4xxx)
    $revenueAccounts = $accountsResponse.data | Where-Object { $_.code -like "4*" }
    
    Write-Host "`n[REVENUE ACCOUNTS]" -ForegroundColor Green
    $totalAccountBalance = 0
    foreach ($acc in $revenueAccounts) {
        $balance = [double]$acc.balance
        $totalAccountBalance += $balance
        Write-Host "  Code: $($acc.code) | Name: $($acc.name) | Balance: Rp $balance" -ForegroundColor White
    }
    Write-Host "`nTotal from Accounts Balance: Rp $totalAccountBalance" -ForegroundColor Yellow
    
} catch {
    Write-Host "[ERROR] Failed to fetch accounts: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n=== STEP 2: CHECK JOURNAL ENTRIES (SSOT) ===" -ForegroundColor Cyan
try {
    # Get journal entries for revenue accounts
    $journalUrl = "${baseUrl}/api/v1/journals?start_date=${startDate}&end_date=${endDate}"
    Write-Host "Fetching journal entries from: $journalUrl" -ForegroundColor Gray
    $journalResponse = Invoke-RestMethod -Uri $journalUrl -Method Get -Headers $headers
    
    if ($journalResponse.data -and $journalResponse.data.entries) {
        $journals = $journalResponse.data.entries
        Write-Host "`n[JOURNAL ENTRIES ANALYSIS]" -ForegroundColor Green
        Write-Host "Total Journal Entries: $($journals.Count)" -ForegroundColor White
        
        # Group by account for revenue accounts
        $revenueJournals = @{}
        $totalJournalRevenue = 0
        
        foreach ($journal in $journals) {
            if ($journal.entries) {
                foreach ($entry in $journal.entries) {
                    $accountCode = $entry.account_code
                    if ($accountCode -and $accountCode.StartsWith("4")) {
                        # Credit increases revenue, Debit decreases
                        $amount = [double]$entry.credit - [double]$entry.debit
                        
                        if (-not $revenueJournals.ContainsKey($accountCode)) {
                            $revenueJournals[$accountCode] = @{
                                AccountName = $entry.account_name
                                TotalAmount = 0
                                Count = 0
                            }
                        }
                        $revenueJournals[$accountCode].TotalAmount += $amount
                        $revenueJournals[$accountCode].Count++
                        $totalJournalRevenue += $amount
                    }
                }
            }
        }
        
        Write-Host "`n[REVENUE FROM JOURNAL ENTRIES]" -ForegroundColor Green
        foreach ($code in ($revenueJournals.Keys | Sort-Object)) {
            $data = $revenueJournals[$code]
            Write-Host "  Code: $code | Name: $($data.AccountName) | Amount: Rp $($data.TotalAmount) | Entries: $($data.Count)" -ForegroundColor White
        }
        Write-Host "`nTotal from Journal Entries: Rp $totalJournalRevenue" -ForegroundColor Yellow
    }
} catch {
    Write-Host "[ERROR] Failed to fetch journals: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n=== STEP 3: CHECK P&L REPORT API ===" -ForegroundColor Cyan
try {
    $plUrl = "${baseUrl}/api/v1/reports/ssot-profit-loss?start_date=${startDate}&end_date=${endDate}&format=json"
    Write-Host "Fetching P&L report from: $plUrl" -ForegroundColor Gray
    $plResponse = Invoke-RestMethod -Uri $plUrl -Method Get -Headers $headers
    
    $plData = $plResponse.data
    
    Write-Host "`n[P&L REPORT ANALYSIS]" -ForegroundColor Green
    Write-Host "Total Revenue (from API): Rp $($plData.total_revenue)" -ForegroundColor White
    
    # Check sections
    if ($plData.sections) {
        $revenueSection = $plData.sections | Where-Object { $_.name -eq "REVENUE" }
        if ($revenueSection) {
            Write-Host "`n[REVENUE SECTION BREAKDOWN]" -ForegroundColor Green
            Write-Host "Section Total: Rp $($revenueSection.total)" -ForegroundColor White
            
            if ($revenueSection.items) {
                Write-Host "`nRevenue Items:" -ForegroundColor White
                foreach ($item in $revenueSection.items) {
                    Write-Host "  Code: $($item.account_code) | Name: $($item.name) | Amount: Rp $($item.amount)" -ForegroundColor White
                }
            }
        }
    }
} catch {
    Write-Host "[ERROR] Failed to fetch P&L report: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n=== STEP 4: DIRECT DATABASE QUERY ===" -ForegroundColor Cyan
Write-Host "Executing raw SQL queries to verify data..." -ForegroundColor Gray

# SQL queries to check
$queries = @(
    @{
        Name = "Revenue Accounts Balance"
        SQL = "SELECT code, name, balance FROM accounts WHERE code LIKE '4%' ORDER BY code;"
    },
    @{
        Name = "Journal Entries for Revenue (Credit - Debit)"
        SQL = @"
SELECT 
    je.account_code,
    a.name as account_name,
    SUM(je.credit) as total_credit,
    SUM(je.debit) as total_debit,
    SUM(je.credit - je.debit) as net_amount,
    COUNT(*) as entry_count
FROM journal_entries je
LEFT JOIN accounts a ON je.account_code = a.code
WHERE je.account_code LIKE '4%'
AND je.status = 'POSTED'
GROUP BY je.account_code, a.name
ORDER BY je.account_code;
"@
    },
    @{
        Name = "All Journal Entries for Account 4101"
        SQL = @"
SELECT 
    j.id as journal_id,
    j.date,
    j.description,
    j.source_type,
    j.source_id,
    je.account_code,
    je.account_name,
    je.debit,
    je.credit,
    j.status
FROM journals j
INNER JOIN journal_entries je ON j.id = je.journal_id
WHERE je.account_code = '4101'
AND j.date BETWEEN '${startDate}' AND '${endDate}'
ORDER BY j.date, j.id;
"@
    },
    @{
        Name = "Check for Duplicate Account Codes"
        SQL = @"
SELECT 
    je.account_code,
    je.account_name,
    COUNT(DISTINCT je.account_name) as name_variations,
    GROUP_CONCAT(DISTINCT je.account_name) as all_names,
    SUM(je.credit - je.debit) as total_amount
FROM journal_entries je
WHERE je.account_code LIKE '4%'
GROUP BY je.account_code
HAVING COUNT(DISTINCT je.account_name) > 1
ORDER BY je.account_code;
"@
    },
    @{
        Name = "P&L Service Query Simulation"
        SQL = @"
SELECT 
    je.account_code,
    je.account_name,
    SUM(je.credit - je.debit) as amount
FROM journal_entries je
INNER JOIN journals j ON je.journal_id = j.id
WHERE je.account_code LIKE '4%'
AND j.status = 'POSTED'
AND j.date BETWEEN '${startDate}' AND '${endDate}'
GROUP BY je.account_code, je.account_name
ORDER BY je.account_code, je.account_name;
"@
    }
)

# Create SQL test file
$sqlFile = "temp_pl_validation_queries.sql"
$sqlContent = @"
-- Profit & Loss Database Validation Queries
-- Generated: $(Get-Date -Format "yyyy-MM-dd HH:mm:ss")

"@

foreach ($query in $queries) {
    $sqlContent += @"

-- ==========================================
-- $($query.Name)
-- ==========================================
$($query.SQL)

"@
}

$sqlContent | Out-File -FilePath $sqlFile -Encoding UTF8
Write-Host "`n[INFO] SQL queries saved to: $sqlFile" -ForegroundColor Cyan
Write-Host "You can execute these queries manually in your database client" -ForegroundColor Yellow

Write-Host "`n=== COMPARISON SUMMARY ===" -ForegroundColor Cyan
Write-Host @"

┌─────────────────────────────────────────────────────────┐
│ DATA SOURCE              │ TOTAL REVENUE                │
├─────────────────────────────────────────────────────────┤
│ Chart of Accounts        │ Rp $totalAccountBalance
│ Journal Entries (SSOT)   │ Rp $totalJournalRevenue
│ P&L Report API           │ Rp $($plData.total_revenue)
└─────────────────────────────────────────────────────────┘

"@ -ForegroundColor White

# Analysis
Write-Host "`n=== DISCREPANCY ANALYSIS ===" -ForegroundColor Cyan

if ($plData.total_revenue -eq $totalJournalRevenue) {
    Write-Host "[OK] P&L Report matches Journal Entries" -ForegroundColor Green
} else {
    Write-Host "[WARNING] P&L Report ($($plData.total_revenue)) != Journal Entries ($totalJournalRevenue)" -ForegroundColor Red
    Write-Host "Difference: Rp $($plData.total_revenue - $totalJournalRevenue)" -ForegroundColor Yellow
}

if ($totalJournalRevenue -eq $totalAccountBalance) {
    Write-Host "[OK] Journal Entries match Account Balances" -ForegroundColor Green
} else {
    Write-Host "[WARNING] Journal Entries ($totalJournalRevenue) != Account Balance ($totalAccountBalance)" -ForegroundColor Red
    Write-Host "Difference: Rp $($totalJournalRevenue - $totalAccountBalance)" -ForegroundColor Yellow
}

Write-Host "`n=== POSSIBLE CAUSES ===" -ForegroundColor Cyan
if ($plData.total_revenue -gt $totalAccountBalance) {
    Write-Host "1. Duplicate journal entries for same transaction" -ForegroundColor Yellow
    Write-Host "2. Account code/name case sensitivity causing duplicates" -ForegroundColor Yellow
    Write-Host "3. Multiple entries for same sale/transaction" -ForegroundColor Yellow
    Write-Host "4. Legacy journals + SSOT journals both counted" -ForegroundColor Yellow
}

Write-Host "`n=== NEXT STEPS ===" -ForegroundColor Cyan
Write-Host "1. Execute SQL queries in $sqlFile to see raw database data" -ForegroundColor White
Write-Host "2. Check if account_name has variations (e.g., 'PENDAPATAN PENJUALAN' vs 'Pendapatan Penjualan')" -ForegroundColor White
Write-Host "3. Verify journal_entries table for duplicate entries" -ForegroundColor White
Write-Host "4. Check if backend is grouping by both account_code AND account_name (causing duplicates)" -ForegroundColor White

Write-Host "`n[COMPLETED] Database validation complete. Check $sqlFile for detailed queries." -ForegroundColor Green

