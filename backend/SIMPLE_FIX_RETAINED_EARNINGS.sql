-- SIMPLE FIX FOR RETAINED EARNINGS BALANCE
-- This recalculates Retained Earnings balance based ONLY on closing journals
-- Run this directly in pgAdmin or your SQL client

BEGIN;

-- Show current balance
SELECT 
    '=== BEFORE FIX ===' as info,
    code,
    name,
    balance
FROM accounts 
WHERE code = '3201';

-- Fix the balance by recalculating from closing journals only
UPDATE accounts
SET balance = (
    SELECT COALESCE(SUM(jl.credit_amount - jl.debit_amount), 0)
    FROM journal_lines jl
    JOIN journal_entries je ON je.id = jl.journal_entry_id
    WHERE je.reference_type = 'CLOSING'
        AND je.status = 'POSTED'
        AND je.deleted_at IS NULL
        AND jl.account_id = accounts.id
)
WHERE code = '3201' AND type = 'EQUITY';

-- Show new balance
SELECT 
    '=== AFTER FIX ===' as info,
    code,
    name,
    balance
FROM accounts 
WHERE code = '3201';

-- Verify balance sheet
SELECT 
    '=== BALANCE SHEET VERIFICATION ===' as info,
    (SELECT SUM(balance) FROM accounts WHERE type = 'ASSET' AND deleted_at IS NULL) as total_assets,
    (SELECT SUM(balance) FROM accounts WHERE type = 'LIABILITY' AND deleted_at IS NULL) as total_liabilities,
    (SELECT SUM(balance) FROM accounts WHERE type = 'EQUITY' AND deleted_at IS NULL) as total_equity,
    (SELECT SUM(balance) FROM accounts WHERE type = 'ASSET' AND deleted_at IS NULL) - 
    ((SELECT SUM(balance) FROM accounts WHERE type = 'LIABILITY' AND deleted_at IS NULL) + 
     (SELECT SUM(balance) FROM accounts WHERE type = 'EQUITY' AND deleted_at IS NULL)) as difference;

COMMIT;

-- Done!
SELECT '✅ Fix completed! Now restart your backend and refresh Balance Sheet.' as status;
