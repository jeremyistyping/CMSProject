# Simple Revenue Testing Script
# This script shows you the exact queries to run to investigate the duplication

Write-Host "`n=== REVENUE DUPLICATION INVESTIGATION ===" -ForegroundColor Cyan
Write-Host "Expected: Rp 10,000,000" -ForegroundColor White
Write-Host "Actual: Rp 20,000,000" -ForegroundColor Red
Write-Host "Issue: 100% duplication!`n" -ForegroundColor Yellow

Write-Host "=== KEY SQL QUERIES TO RUN ===" -ForegroundColor Cyan
Write-Host "`nRun these queries in your MySQL client (HeidiSQL, phpMyAdmin, etc.):`n" -ForegroundColor Yellow

Write-Host "-- QUERY 1: Check Account Balance --" -ForegroundColor Green
Write-Host @"
SELECT code, name, balance 
FROM accounts 
WHERE code LIKE '4%' 
ORDER BY code;
"@ -ForegroundColor White

Write-Host "`n-- QUERY 2: Check Journal Entries (the critical one!) --" -ForegroundColor Green
Write-Host @"
SELECT 
    je.account_code,
    je.account_name,
    SUM(je.credit - je.debit) as amount,
    COUNT(*) as entry_count
FROM journal_entries je
INNER JOIN journals j ON je.journal_id = j.id
WHERE je.account_code LIKE '4%'
  AND j.status = 'POSTED'
  AND j.date BETWEEN '2025-01-01' AND '2025-12-31'
GROUP BY je.account_code, je.account_name
ORDER BY je.account_code, je.account_name;
"@ -ForegroundColor White

Write-Host "`n-- QUERY 3: Check for Name Variations (likely culprit!) --" -ForegroundColor Green
Write-Host @"
SELECT 
    je.account_code,
    COUNT(DISTINCT je.account_name) as name_count,
    GROUP_CONCAT(DISTINCT je.account_name SEPARATOR ' | ') as all_names,
    SUM(je.credit - je.debit) as total_amount
FROM journal_entries je
WHERE je.account_code LIKE '4%'
GROUP BY je.account_code
HAVING COUNT(DISTINCT je.account_name) > 1;
"@ -ForegroundColor White

Write-Host "`n-- QUERY 4: Unified Journal Check (SSOT system) --" -ForegroundColor Green
Write-Host @"
SELECT 
    a.code,
    a.name,
    COALESCE(SUM(ujl.credit_amount - ujl.debit_amount), 0) as amount
FROM accounts a
LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id 
    AND uje.status = 'POSTED'
WHERE a.code LIKE '4%'
  AND uje.entry_date BETWEEN '2025-01-01' AND '2025-12-31'
GROUP BY a.id, a.code, a.name
HAVING amount != 0;
"@ -ForegroundColor White

Write-Host "`n-- QUERY 5: Legacy Journal Check --" -ForegroundColor Green
Write-Host @"
SELECT 
    a.code,
    a.name,
    COALESCE(SUM(jl.credit_amount - jl.debit_amount), 0) as amount
FROM accounts a
LEFT JOIN journal_lines jl ON jl.account_id = a.id
LEFT JOIN journal_entries je ON je.id = jl.journal_entry_id 
    AND je.status = 'POSTED'
WHERE a.code LIKE '4%'
  AND je.entry_date BETWEEN '2025-01-01' AND '2025-12-31'
GROUP BY a.id, a.code, a.name
HAVING amount != 0;
"@ -ForegroundColor White

Write-Host "`n=== WHAT TO LOOK FOR ===" -ForegroundColor Cyan
Write-Host @"
1. If QUERY 2 shows TWO rows for account 4101:
   - "PENDAPATAN PENJUALAN" = Rp 10,000,000
   - "Pendapatan Penjualan" = Rp 10,000,000
   => PROBLEM: Grouping by account_name causing duplicates!

2. If QUERY 3 returns ANY results:
   => PROBLEM: Same account code has different names in journal_entries

3. If QUERY 4 shows Rp 10M AND QUERY 5 shows Rp 10M:
   => PROBLEM: Both SSOT and Legacy systems are being counted!

4. If QUERY 4 shows Rp 20M:
   => PROBLEM: Duplicate entries in unified_journal system
"@ -ForegroundColor Yellow

Write-Host "`n=== MOST LIKELY FIX ===" -ForegroundColor Cyan
Write-Host @"
If QUERY 2 shows multiple rows with different account_name:

FIX 1: Standardize account names in journal_entries
UPDATE journal_entries je
INNER JOIN accounts a ON a.code = je.account_code
SET je.account_name = a.name
WHERE je.account_code = '4101';

FIX 2: Modify backend to GROUP BY account_code only (not account_name)
File: backend/services/ssot_profit_loss_service.go
Change: GROUP BY je.account_code ONLY (remove je.account_name)
"@ -ForegroundColor Green

Write-Host "`n=== INSTRUCTIONS ===" -ForegroundColor Cyan
Write-Host @"
1. Open your MySQL client (HeidiSQL, phpMyAdmin, MySQL Workbench, etc.)
2. Connect to database: accounting_db
3. Run QUERY 2 above (the most important one!)
4. Check the results:
   - If 1 row returned => Good, no duplication in journal_entries
   - If 2+ rows for same account_code => FOUND THE PROBLEM!
5. Take screenshot and share results
"@ -ForegroundColor Yellow

Write-Host "`n=== ALTERNATIVE: Direct MySQL Command ===" -ForegroundColor Cyan
Write-Host "If you have mysql command line, run this:" -ForegroundColor White
Write-Host @"

mysql -u root -p accounting_db -e "SELECT je.account_code, je.account_name, SUM(je.credit - je.debit) as amount, COUNT(*) as entries FROM journal_entries je INNER JOIN journals j ON je.journal_id = j.id WHERE je.account_code LIKE '4%' AND j.status = 'POSTED' AND j.date BETWEEN '2025-01-01' AND '2025-12-31' GROUP BY je.account_code, je.account_name ORDER BY je.account_code, je.account_name;"

"@ -ForegroundColor Gray

Write-Host "`nAll queries are also saved in: temp_pl_validation_queries.sql" -ForegroundColor Cyan
Write-Host "and: investigate_revenue_duplication.sql`n" -ForegroundColor Cyan

# Try to check if backend is running
Write-Host "=== CHECKING BACKEND STATUS ===" -ForegroundColor Cyan
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -Method Get -TimeoutSec 2 -ErrorAction SilentlyContinue
    Write-Host "[OK] Backend is running on port 8080" -ForegroundColor Green
} catch {
    Write-Host "[INFO] Backend might not be running on port 8080" -ForegroundColor Yellow
    Write-Host "If you want to test via API, start the backend first with: ..\bin\app.exe" -ForegroundColor Gray
}

Write-Host "`n[NEXT STEP] Please run QUERY 2 in your database and share the results!" -ForegroundColor Yellow

