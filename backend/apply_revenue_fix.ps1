# Apply Revenue Duplication Fix
# This script diagnoses and fixes the revenue duplication issue

Write-Host "`n=== REVENUE DUPLICATION FIX SCRIPT ===" -ForegroundColor Cyan
Write-Host "Problem: Account 4101 appears twice with different names" -ForegroundColor Yellow
Write-Host "Expected: Rp 10,000,000" -ForegroundColor White
Write-Host "Current: Rp 20,000,000`n" -ForegroundColor Red

# Database credentials - UPDATE THESE!
$dbHost = "localhost"
$dbPort = "3306"
$dbName = "accounting_db"
$dbUser = "root"
$dbPassword = "root"  # UPDATE THIS!

Write-Host "=== CONFIGURATION ===" -ForegroundColor Cyan
Write-Host "Database Host: $dbHost" -ForegroundColor White
Write-Host "Database Name: $dbName" -ForegroundColor White
Write-Host "Database User: $dbUser" -ForegroundColor White

# Check if mysql command is available
$mysqlCmd = Get-Command mysql -ErrorAction SilentlyContinue
if (-not $mysqlCmd) {
    Write-Host "`n[ERROR] MySQL command line tool not found!" -ForegroundColor Red
    Write-Host "Please run the SQL queries manually in HeidiSQL or phpMyAdmin" -ForegroundColor Yellow
    Write-Host "SQL file: fix_revenue_duplication.sql`n" -ForegroundColor Yellow
    
    Write-Host "=== MANUAL FIX STEPS ===" -ForegroundColor Cyan
    Write-Host @"
1. Open HeidiSQL or phpMyAdmin
2. Connect to database: $dbName
3. Run these queries:

-- Fix parent account
UPDATE accounts SET is_header = true WHERE code = '4000';

-- Standardize names in journal_entries
UPDATE journal_entries je
INNER JOIN accounts a ON a.code = je.account_code
SET je.account_name = a.name
WHERE je.account_code LIKE '4%' AND je.account_name != a.name;

-- Standardize names in unified_journal_lines
UPDATE unified_journal_lines ujl
INNER JOIN accounts a ON a.id = ujl.account_id
SET ujl.account_name = a.name
WHERE ujl.account_code LIKE '4%' AND ujl.account_name != a.name;

4. Restart backend
5. Test P&L report - should show Rp 10,000,000
"@ -ForegroundColor White
    
    exit 0
}

Write-Host "`n=== STEP 1: DIAGNOSTIC ===" -ForegroundColor Cyan

# Query 1: Check parent account 4000
$query1 = @"
SELECT 
    code,
    name,
    COALESCE(is_header, false) as is_header,
    CASE 
        WHEN COALESCE(is_header, false) = true THEN 'OK'
        ELSE 'PROBLEM - Should be TRUE!'
    END as status
FROM accounts WHERE code = '4000';
"@

Write-Host "Checking parent account 4000..." -ForegroundColor Yellow
$result1 = & mysql -h $dbHost -P $dbPort -u $dbUser -p$dbPassword $dbName -e $query1 2>&1

if ($LASTEXITCODE -eq 0) {
    Write-Host $result1 -ForegroundColor White
} else {
    Write-Host "[ERROR] Failed to connect to database!" -ForegroundColor Red
    Write-Host "Error: $result1" -ForegroundColor Red
    Write-Host "`nPlease check your database credentials at the top of this script." -ForegroundColor Yellow
    exit 1
}

# Query 2: Check for name variations
$query2 = @"
SELECT 
    je.account_code,
    je.account_name,
    COUNT(*) as count,
    SUM(je.credit - je.debit) as amount
FROM journal_entries je
INNER JOIN journals j ON j.id = je.journal_id
WHERE je.account_code = '4101' AND j.status = 'POSTED'
GROUP BY je.account_code, je.account_name;
"@

Write-Host "`nChecking journal_entries for name variations..." -ForegroundColor Yellow
$result2 = & mysql -h $dbHost -P $dbPort -u $dbUser -p$dbPassword $dbName -e $query2 2>&1
Write-Host $result2 -ForegroundColor White

$hasVariations = $result2 -match "4101.*\n.*4101"  # Check if 4101 appears multiple times

Write-Host "`n=== DIAGNOSIS ===" -ForegroundColor Cyan
if ($hasVariations) {
    Write-Host "[FOUND] Multiple account names for 4101!" -ForegroundColor Red
    Write-Host "This causes backend to group separately, creating duplication." -ForegroundColor Yellow
} else {
    Write-Host "[INFO] No name variations detected in diagnostic query" -ForegroundColor Yellow
}

Write-Host "`n=== STEP 2: APPLY FIX ===" -ForegroundColor Cyan
Write-Host "Do you want to apply the fix now? (Y/N)" -ForegroundColor Yellow
$confirm = Read-Host

if ($confirm -eq "Y" -or $confirm -eq "y") {
    Write-Host "`nApplying fixes..." -ForegroundColor Cyan
    
    # Fix 1: Set parent account as header
    Write-Host "`n[1/3] Setting account 4000 as header..." -ForegroundColor Yellow
    $fix1 = "UPDATE accounts SET is_header = true WHERE code = '4000';"
    & mysql -h $dbHost -P $dbPort -u $dbUser -p$dbPassword $dbName -e $fix1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] Account 4000 marked as header" -ForegroundColor Green
    } else {
        Write-Host "[ERROR] Failed to update account 4000" -ForegroundColor Red
    }
    
    # Fix 2: Standardize journal_entries
    Write-Host "`n[2/3] Standardizing names in journal_entries..." -ForegroundColor Yellow
    $fix2 = @"
UPDATE journal_entries je
INNER JOIN accounts a ON a.code = je.account_code
SET je.account_name = a.name
WHERE je.account_code LIKE '4%' AND je.account_name != a.name;
"@
    & mysql -h $dbHost -P $dbPort -u $dbUser -p$dbPassword $dbName -e $fix2
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] journal_entries standardized" -ForegroundColor Green
    } else {
        Write-Host "[ERROR] Failed to update journal_entries" -ForegroundColor Red
    }
    
    # Fix 3: Standardize unified_journal_lines
    Write-Host "`n[3/3] Standardizing names in unified_journal_lines..." -ForegroundColor Yellow
    $fix3 = @"
UPDATE unified_journal_lines ujl
INNER JOIN accounts a ON a.id = ujl.account_id
SET ujl.account_name = a.name
WHERE ujl.account_code LIKE '4%' AND ujl.account_name != a.name;
"@
    & mysql -h $dbHost -P $dbPort -u $dbUser -p$dbPassword $dbName -e $fix3
    if ($LASTEXITCODE -eq 0) {
        Write-Host "[OK] unified_journal_lines standardized" -ForegroundColor Green
    } else {
        Write-Host "[ERROR] Failed to update unified_journal_lines" -ForegroundColor Red
    }
    
    Write-Host "`n=== STEP 3: VERIFICATION ===" -ForegroundColor Cyan
    
    # Verify no variations remain
    Write-Host "Checking if variations are fixed..." -ForegroundColor Yellow
    $verify = @"
SELECT 
    je.account_code,
    je.account_name,
    SUM(je.credit - je.debit) as amount
FROM journal_entries je
INNER JOIN journals j ON j.id = je.journal_id
WHERE je.account_code = '4101' AND j.status = 'POSTED'
GROUP BY je.account_code, je.account_name;
"@
    $verifyResult = & mysql -h $dbHost -P $dbPort -u $dbUser -p$dbPassword $dbName -e $verify
    Write-Host $verifyResult -ForegroundColor White
    
    $stillHasVariations = $verifyResult -match "4101.*\n.*4101"
    
    if ($stillHasVariations) {
        Write-Host "`n[WARNING] Variations still exist!" -ForegroundColor Red
    } else {
        Write-Host "`n[SUCCESS] Fix applied successfully!" -ForegroundColor Green
        Write-Host "Account 4101 now appears only once" -ForegroundColor Green
    }
    
    Write-Host "`n=== NEXT STEPS ===" -ForegroundColor Cyan
    Write-Host "1. Restart backend: ..\bin\app.exe" -ForegroundColor White
    Write-Host "2. Generate P&L Report in browser" -ForegroundColor White
    Write-Host "3. Verify Total Revenue = Rp 10,000,000 ✓" -ForegroundColor Green
    
} else {
    Write-Host "`nFix not applied. To apply manually, run: fix_revenue_duplication.sql" -ForegroundColor Yellow
}

Write-Host "`n[COMPLETED]" -ForegroundColor Cyan

