# ==========================================
# DUPLICATE ACCOUNTS MONITORING SCRIPT
# ==========================================

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "CHECKING FOR DUPLICATE ACCOUNTS" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Database connection details (using defaults)
$dbHost = "localhost"
$dbPort = "5432"
$dbName = "sistem_akuntansi"
$dbUser = "postgres"
$dbPassword = "postgres"

# Set PostgreSQL password
$env:PGPASSWORD = $dbPassword

Write-Host "Connecting to: $dbName @ $dbHost" -ForegroundColor Yellow
Write-Host ""

# Query 1: Check for active duplicates
Write-Host "1. ACTIVE DUPLICATE ACCOUNTS" -ForegroundColor Green
Write-Host "----------------------------------------" -ForegroundColor Green

$query1 = @"
SELECT 
    code,
    COUNT(*) as duplicate_count,
    STRING_AGG(id::text, ', ') as account_ids,
    STRING_AGG(name, ' | ') as account_names
FROM accounts
WHERE deleted_at IS NULL
  AND is_header = false
GROUP BY code
HAVING COUNT(*) > 1
ORDER BY code;
"@

$result1 = psql -h $dbHost -p $dbPort -U $dbUser -d $dbName -c $query1 2>&1

if ($result1 -match "\(0 rows\)") {
    Write-Host "✅ NO DUPLICATE ACCOUNTS FOUND!" -ForegroundColor Green
} else {
    Write-Host "⚠️  DUPLICATES DETECTED:" -ForegroundColor Red
    Write-Host $result1
}

Write-Host ""

# Query 2: Check monitoring view
Write-Host "2. MONITORING VIEW CHECK" -ForegroundColor Green
Write-Host "----------------------------------------" -ForegroundColor Green

$query2 = "SELECT COUNT(*) as duplicate_count FROM v_potential_duplicate_accounts;"
$result2 = psql -h $dbHost -p $dbPort -U $dbUser -d $dbName -t -c $query2 2>&1

$count = $result2.Trim()

if ($count -eq "0") {
    Write-Host "✅ Monitoring view shows 0 duplicates" -ForegroundColor Green
} else {
    Write-Host "⚠️  Monitoring view shows $count potential duplicate(s)" -ForegroundColor Yellow
    
    # Show details
    $detailQuery = "SELECT * FROM v_potential_duplicate_accounts LIMIT 5;"
    $details = psql -h $dbHost -p $dbPort -U $dbUser -d $dbName -c $detailQuery 2>&1
    Write-Host $details
}

Write-Host ""

# Query 3: Check if unique constraint exists
Write-Host "3. UNIQUE CONSTRAINT STATUS" -ForegroundColor Green
Write-Host "----------------------------------------" -ForegroundColor Green

$query3 = @"
SELECT 
    indexname,
    indexdef
FROM pg_indexes
WHERE tablename = 'accounts'
  AND indexname LIKE '%unique%'
ORDER BY indexname;
"@

$result3 = psql -h $dbHost -p $dbPort -U $dbUser -d $dbName -c $query3 2>&1

if ($result3 -match "idx_accounts_code_unique") {
    Write-Host "✅ Unique constraint is installed" -ForegroundColor Green
} else {
    Write-Host "⚠️  Unique constraint NOT found!" -ForegroundColor Red
    Write-Host "Run: .\backend\migrations\prevent_duplicate_accounts.sql" -ForegroundColor Yellow
}

Write-Host ""

# Query 4: Check if trigger exists
Write-Host "4. VALIDATION TRIGGER STATUS" -ForegroundColor Green
Write-Host "----------------------------------------" -ForegroundColor Green

$query4 = @"
SELECT 
    trigger_name,
    event_manipulation,
    event_object_table,
    action_statement
FROM information_schema.triggers
WHERE trigger_name = 'trg_prevent_duplicate_account_code';
"@

$result4 = psql -h $dbHost -p $dbPort -U $dbUser -d $dbName -c $query4 2>&1

if ($result4 -match "trg_prevent_duplicate_account_code") {
    Write-Host "✅ Validation trigger is installed" -ForegroundColor Green
} else {
    Write-Host "⚠️  Validation trigger NOT found!" -ForegroundColor Red
    Write-Host "Run: .\backend\migrations\prevent_duplicate_accounts.sql" -ForegroundColor Yellow
}

Write-Host ""

# Query 5: Account statistics
Write-Host "5. ACCOUNT STATISTICS" -ForegroundColor Green
Write-Host "----------------------------------------" -ForegroundColor Green

$query5 = @"
SELECT 
    COUNT(*) as total_accounts,
    COUNT(DISTINCT code) as unique_codes,
    COUNT(*) - COUNT(DISTINCT code) as duplicates,
    COUNT(CASE WHEN is_header = true THEN 1 END) as header_accounts,
    COUNT(CASE WHEN is_header = false THEN 1 END) as detail_accounts
FROM accounts
WHERE deleted_at IS NULL;
"@

$result5 = psql -h $dbHost -p $dbPort -U $dbUser -d $dbName -c $query5 2>&1
Write-Host $result5

Write-Host ""

# Summary
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "SUMMARY" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

if (($count -eq "0") -and ($result1 -match "\(0 rows\)")) {
    Write-Host "✅ SYSTEM STATUS: HEALTHY" -ForegroundColor Green
    Write-Host "   - No duplicate accounts detected" -ForegroundColor White
    Write-Host "   - Prevention systems active" -ForegroundColor White
} else {
    Write-Host "⚠️  SYSTEM STATUS: NEEDS ATTENTION" -ForegroundColor Yellow
    Write-Host "   - Duplicates detected or prevention not active" -ForegroundColor White
    Write-Host "   - Run migration to fix" -ForegroundColor White
}

Write-Host ""
Write-Host "Next actions:" -ForegroundColor Yellow
Write-Host "  - If duplicates found: Run fix migration" -ForegroundColor White
Write-Host "  - If constraints missing: Restart backend to apply" -ForegroundColor White
Write-Host "  - Schedule this check weekly/monthly" -ForegroundColor White
Write-Host ""

