# ==========================================
# SCRIPT UNTUK CEK DUPLIKASI REVENUE
# ==========================================

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "CHECKING REVENUE DUPLICATION IN DATABASE" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Load environment
$envFile = ".env"
if (Test-Path $envFile) {
    Get-Content $envFile | ForEach-Object {
        if ($_ -match '^\s*([^#][^=]+?)\s*=\s*(.*?)\s*$') {
            $name = $matches[1]
            $value = $matches[2]
            Set-Item -Path "env:$name" -Value $value
        }
    }
}

# Database connection details
$dbHost = if ($env:DB_HOST) { $env:DB_HOST } else { "localhost" }
$dbPort = if ($env:DB_PORT) { $env:DB_PORT } else { "5432" }
$dbName = if ($env:DB_NAME) { $env:DB_NAME } else { "accounting_db" }
$dbUser = if ($env:DB_USER) { $env:DB_USER } else { "postgres" }
$dbPassword = if ($env:DB_PASSWORD) { $env:DB_PASSWORD } else { "postgres" }

# Set PostgreSQL password environment variable
$env:PGPASSWORD = $dbPassword

Write-Host "Connecting to database: $dbName @ $dbHost : $dbPort" -ForegroundColor Yellow
Write-Host ""

# Function to run query and display results
function Run-Query {
    param(
        [string]$QueryTitle,
        [string]$Query
    )
    
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "  $QueryTitle" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host ""
    
    $result = psql -h $dbHost -p $dbPort -U $dbUser -d $dbName -c $Query
    Write-Host $result
    Write-Host ""
}

# Query 1: Check Account 4101 Summary
$query1 = @"
SELECT 
    a.id as account_id,
    a.code as account_code,
    a.name as account_name,
    COUNT(ujl.id) as line_count,
    COUNT(DISTINCT ujl.journal_id) as unique_journal_count,
    COALESCE(SUM(ujl.debit_amount), 0) as total_debit,
    COALESCE(SUM(ujl.credit_amount), 0) as total_credit,
    COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0) as net_balance
FROM accounts a
LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id 
    AND uje.status = 'POSTED' 
    AND uje.deleted_at IS NULL
WHERE a.code = '4101'
  AND uje.entry_date >= '2025-01-01' 
  AND uje.entry_date <= '2025-12-31'
GROUP BY a.id, a.code, a.name;
"@

Run-Query "1. ACCOUNT 4101 SUMMARY" $query1

# Query 2: Check for duplicate journals from same source
$query2 = @"
SELECT 
    uje.source_type,
    uje.source_id,
    uje.entry_date,
    COUNT(DISTINCT uje.id) as duplicate_count,
    STRING_AGG(uje.id::text, ', ') as journal_ids,
    STRING_AGG(uje.journal_number, ', ') as journal_numbers
FROM unified_journal_ledger uje
INNER JOIN unified_journal_lines ujl ON ujl.journal_id = uje.id
INNER JOIN accounts a ON a.id = ujl.account_id
WHERE a.code = '4101'
  AND uje.entry_date >= '2025-01-01' 
  AND uje.entry_date <= '2025-12-31'
  AND uje.status = 'POSTED'
  AND uje.deleted_at IS NULL
GROUP BY uje.source_type, uje.source_id, uje.entry_date
HAVING COUNT(DISTINCT uje.id) > 1
ORDER BY duplicate_count DESC;
"@

Run-Query "2. DUPLICATE JOURNALS FROM SAME SOURCE" $query2

# Query 3: Check for multiple lines of account 4101 in same journal
$query3 = @"
SELECT 
    uje.id as journal_id,
    uje.journal_number,
    uje.entry_date,
    uje.source_type,
    uje.source_id,
    COUNT(ujl.id) as lines_for_4101,
    SUM(ujl.credit_amount) as total_credit
FROM unified_journal_ledger uje
INNER JOIN unified_journal_lines ujl ON ujl.journal_id = uje.id
INNER JOIN accounts a ON a.id = ujl.account_id
WHERE a.code = '4101'
  AND uje.entry_date >= '2025-01-01' 
  AND uje.entry_date <= '2025-12-31'
  AND uje.status = 'POSTED'
  AND uje.deleted_at IS NULL
GROUP BY uje.id, uje.journal_number, uje.entry_date, uje.source_type, uje.source_id
HAVING COUNT(ujl.id) > 1;
"@

Run-Query "3. MULTIPLE LINES FOR ACCOUNT 4101 IN SAME JOURNAL" $query3

# Query 4: All revenue accounts summary
$query4 = @"
SELECT 
    a.code as account_code,
    a.name as account_name,
    COUNT(DISTINCT ujl.journal_id) as journal_count,
    COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0) as net_revenue
FROM accounts a
LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id 
    AND uje.status = 'POSTED' 
    AND uje.deleted_at IS NULL
WHERE (a.code LIKE '4%' OR UPPER(a.type) = 'REVENUE')
  AND uje.entry_date >= '2025-01-01' 
  AND uje.entry_date <= '2025-12-31'
  AND COALESCE(a.is_header, false) = false
GROUP BY a.code, a.name
HAVING COALESCE(SUM(ujl.credit_amount), 0) > 0 OR COALESCE(SUM(ujl.debit_amount), 0) > 0
ORDER BY a.code;
"@

Run-Query "4. ALL REVENUE ACCOUNTS (4xxx)" $query4

# Query 5: Check sales to journal mapping
$query5 = @"
SELECT 
    s.id as sales_id,
    s.invoice_number,
    s.total_amount,
    COUNT(DISTINCT uje.id) as journal_count,
    STRING_AGG(DISTINCT uje.journal_number, ', ') as journal_numbers
FROM sales s
LEFT JOIN unified_journal_ledger uje ON uje.source_type = 'SALES' 
    AND uje.source_id = s.id 
    AND uje.status = 'POSTED'
    AND uje.deleted_at IS NULL
WHERE s.created_at >= '2025-01-01' 
  AND s.created_at <= '2025-12-31'
GROUP BY s.id, s.invoice_number, s.total_amount
HAVING COUNT(DISTINCT uje.id) > 1
ORDER BY journal_count DESC;
"@

Run-Query "5. SALES WITH MULTIPLE JOURNALS (POTENTIAL DUPLICATES)" $query5

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "ANALYSIS COMPLETE" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "1. Review the output above" -ForegroundColor White
Write-Host "2. If duplicate journals found, use fix script to clean them" -ForegroundColor White
Write-Host "3. If multiple lines in same journal, check journal creation logic" -ForegroundColor White
Write-Host ""

