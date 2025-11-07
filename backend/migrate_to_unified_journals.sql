-- ================================================================
-- MIGRATION TO UNIFIED JOURNALS SYSTEM (SSOT) - OPTION A
-- ================================================================
-- Purpose: Migrate from old journal_entries to unified_journal_ledger
-- Date: 2025-11-07
-- 
-- IMPORTANT: BACKUP DATABASE FIRST!
-- pg_dump -U postgres sistem_akuntansi > backup_before_migration.sql

BEGIN;

-- ================================================================
-- STEP 1: ANALYSIS & VALIDATION
-- ================================================================

SELECT 'STEP 1: PRE-MIGRATION ANALYSIS' as step;

-- Check current state
SELECT 
    'Old Journal Entries' as system,
    COUNT(*) as entry_count,
    SUM(CASE WHEN status = 'POSTED' THEN 1 ELSE 0 END) as posted_count
FROM journal_entries
UNION ALL
SELECT 
    'Unified Journal Ledger' as system,
    COUNT(*) as entry_count,
    SUM(CASE WHEN status = 'POSTED' THEN 1 ELSE 0 END) as posted_count
FROM unified_journal_ledger;

-- Check account balances before
CREATE TEMP TABLE pre_migration_balances AS
SELECT 
    code,
    name,
    type,
    balance as old_balance
FROM accounts
WHERE is_active = true AND COALESCE(is_header, false) = false;

-- ================================================================
-- STEP 2: DELETE OLD JOURNAL ENTRIES
-- ================================================================

SELECT 'STEP 2: DELETING OLD JOURNAL ENTRIES' as step;

-- Delete old journal lines first (foreign key)
DELETE FROM journal_lines;

-- Delete old journal entries
DELETE FROM journal_entries;

-- Delete accounting periods (old closing records)
DELETE FROM accounting_periods;

SELECT 'Old journal entries deleted' as status;

-- ================================================================
-- STEP 3: RESET ACCOUNTS BALANCE TO ZERO
-- ================================================================

SELECT 'STEP 3: RESETTING ACCOUNT BALANCES' as step;

-- Reset all account balances to 0
UPDATE accounts 
SET balance = 0, 
    updated_at = NOW()
WHERE is_active = true;

SELECT 
    'Account balances reset' as status,
    COUNT(*) as accounts_reset
FROM accounts 
WHERE is_active = true;

-- ================================================================
-- STEP 4: RECALCULATE BALANCES FROM UNIFIED JOURNALS
-- ================================================================

SELECT 'STEP 4: RECALCULATING BALANCES FROM UNIFIED JOURNALS' as step;

-- Update account balances from unified journals
WITH account_balances AS (
    SELECT 
        a.id,
        a.code,
        a.type,
        CASE 
            -- Asset & Expense: Debit increases, Credit decreases
            WHEN a.type IN ('ASSET', 'EXPENSE') 
            THEN COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
            -- Liability, Equity, Revenue: Credit increases, Debit decreases
            ELSE COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
        END as calculated_balance
    FROM accounts a
    LEFT JOIN unified_journal_lines ujl ON a.id = ujl.account_id
    LEFT JOIN unified_journal_ledger uj ON uj.id = ujl.journal_id
    WHERE uj.status = 'POSTED' OR uj.id IS NULL
    GROUP BY a.id, a.code, a.type
)
UPDATE accounts a
SET 
    balance = ab.calculated_balance,
    updated_at = NOW()
FROM account_balances ab
WHERE a.id = ab.id;

-- Show updated balances
SELECT 
    a.code,
    a.name,
    a.type,
    a.balance as new_balance
FROM accounts a
WHERE a.is_active = true 
    AND COALESCE(a.is_header, false) = false
    AND ABS(a.balance) > 0.01
ORDER BY a.code;

-- ================================================================
-- STEP 5: VALIDATE BALANCE SHEET EQUATION
-- ================================================================

SELECT 'STEP 5: VALIDATING BALANCE SHEET EQUATION' as step;

WITH balance_totals AS (
    SELECT 
        SUM(CASE WHEN type = 'ASSET' THEN balance ELSE 0 END) AS total_assets,
        SUM(CASE WHEN type = 'LIABILITY' THEN balance ELSE 0 END) AS total_liabilities,
        SUM(CASE WHEN type = 'EQUITY' THEN balance ELSE 0 END) AS total_equity,
        SUM(CASE WHEN type = 'REVENUE' THEN balance ELSE 0 END) AS total_revenue,
        SUM(CASE WHEN type = 'EXPENSE' THEN balance ELSE 0 END) AS total_expense
    FROM accounts
    WHERE is_active = true AND COALESCE(is_header, false) = false
)
SELECT 
    total_assets,
    total_liabilities,
    total_equity,
    total_revenue,
    total_expense,
    (total_liabilities + total_equity) as total_liab_equity,
    (total_assets - (total_liabilities + total_equity)) as difference_before_closing,
    (total_revenue - total_expense) as net_income,
    CASE 
        WHEN ABS(total_assets - (total_liabilities + total_equity + total_revenue - total_expense)) < 0.01 
        THEN '✓ BALANCED (including temp accounts)'
        WHEN ABS(total_assets - (total_liabilities + total_equity)) < 0.01
        THEN '✓ BALANCED'
        ELSE '✗ NOT BALANCED'
    END AS status
FROM balance_totals;

-- ================================================================
-- STEP 6: CHECK FOR PERIOD CLOSING NEEDED
-- ================================================================

SELECT 'STEP 6: CHECKING IF PERIOD CLOSING NEEDED' as step;

WITH temp_accounts AS (
    SELECT 
        SUM(CASE WHEN type = 'REVENUE' THEN balance ELSE 0 END) AS total_revenue,
        SUM(CASE WHEN type = 'EXPENSE' THEN balance ELSE 0 END) AS total_expense
    FROM accounts
    WHERE is_active = true AND COALESCE(is_header, false) = false
)
SELECT 
    total_revenue,
    total_expense,
    (total_revenue - total_expense) as net_income,
    CASE 
        WHEN ABS(total_revenue) > 0.01 OR ABS(total_expense) > 0.01
        THEN '⚠️  Revenue/Expense accounts have balances - PERIOD CLOSING NEEDED'
        ELSE '✓ No temporary accounts with balances'
    END as closing_status
FROM temp_accounts;

-- ================================================================
-- STEP 7: SUMMARY REPORT
-- ================================================================

SELECT 'STEP 7: MIGRATION SUMMARY' as step;

-- Compare before and after
SELECT 
    'BEFORE MIGRATION' as stage,
    SUM(ABS(old_balance)) as total_absolute_balance
FROM pre_migration_balances
UNION ALL
SELECT 
    'AFTER MIGRATION' as stage,
    SUM(ABS(balance)) as total_absolute_balance
FROM accounts
WHERE is_active = true AND COALESCE(is_header, false) = false;

-- Show account-by-account comparison
SELECT 
    COALESCE(pb.code, a.code) as account_code,
    COALESCE(pb.name, a.name) as account_name,
    pb.old_balance,
    a.balance as new_balance,
    (a.balance - COALESCE(pb.old_balance, 0)) as change
FROM accounts a
FULL OUTER JOIN pre_migration_balances pb ON pb.code = a.code
WHERE a.is_active = true AND COALESCE(a.is_header, false) = false
    AND (ABS(COALESCE(pb.old_balance, 0)) > 0.01 OR ABS(COALESCE(a.balance, 0)) > 0.01)
ORDER BY account_code;

-- ================================================================
-- FINAL STATUS
-- ================================================================

SELECT 
    '✓ MIGRATION COMPLETED' as status,
    'Next steps:' as action,
    '1. Review balance sheet - should show Revenue & Expense with balances' as step_1,
    '2. Run Period Closing through API to close Revenue/Expense to Retained Earnings' as step_2,
    '3. After closing, balance sheet should be balanced' as step_3,
    '4. All future transactions will use Unified Journals only' as step_4;

-- ================================================================
-- COMMIT OR ROLLBACK
-- ================================================================

-- ROLLBACK;  -- Uncomment to test without committing
COMMIT;  -- Uncomment to apply changes

-- ================================================================
-- POST-MIGRATION NOTES
-- ================================================================

-- After running this script:
-- 1. Old journal_entries table is now EMPTY
-- 2. accounts.balance reflects UNIFIED journals only
-- 3. Revenue & Expense accounts still have balances (not yet closed)
-- 4. You need to run PERIOD CLOSING via API/Application
-- 5. Balance sheet will be balanced after period closing

-- To run period closing via API:
-- POST /api/period-closing/preview
-- {
--   "start_date": "2025-01-01",
--   "end_date": "2025-12-31",
--   "description": "Annual Closing 2025"
-- }
--
-- Then if preview OK:
-- POST /api/period-closing/execute
-- (same payload)
