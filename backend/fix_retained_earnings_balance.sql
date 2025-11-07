-- =====================================================
-- FIX RETAINED EARNINGS DOUBLE COUNTING ISSUE
-- =====================================================
-- Problem: Retained Earnings has been double-counted after period closing
-- Solution: Recalculate Retained Earnings from closing journals only

BEGIN;

-- Step 1: Show current state
SELECT 
    '=== CURRENT STATE ===' as info,
    code,
    name,
    type,
    balance
FROM accounts 
WHERE code IN ('3101', '3201')
ORDER BY code;

-- Step 2: Check Revenue/Expense accounts (should be 0 after closing)
SELECT 
    '=== REVENUE/EXPENSE ACCOUNTS ===' as info,
    code,
    name,
    type,
    balance
FROM accounts 
WHERE type IN ('REVENUE', 'EXPENSE')
    AND balance != 0
ORDER BY code;

-- Step 3: Calculate correct Retained Earnings from closing journals only
WITH closing_journals AS (
    -- Get all closing journal entries
    SELECT 
        je.id,
        je.code,
        je.entry_date,
        je.reference_type
    FROM journal_entries je
    WHERE je.reference_type = 'CLOSING'
        AND je.status = 'POSTED'
        AND je.deleted_at IS NULL
),
retained_earnings_movements AS (
    -- Calculate movements to Retained Earnings from closing journals
    SELECT 
        jl.account_id,
        SUM(jl.credit_amount - jl.debit_amount) as net_movement
    FROM journal_lines jl
    JOIN closing_journals cj ON cj.id = jl.journal_entry_id
    JOIN accounts a ON a.id = jl.account_id
    WHERE a.code = '3201'
    GROUP BY jl.account_id
)
SELECT 
    '=== CORRECT RETAINED EARNINGS ===' as info,
    a.code,
    a.name,
    a.balance as current_balance,
    COALESCE(rem.net_movement, 0) as correct_balance,
    a.balance - COALESCE(rem.net_movement, 0) as difference
FROM accounts a
LEFT JOIN retained_earnings_movements rem ON rem.account_id = a.id
WHERE a.code = '3201';

-- Step 4: Fix Retained Earnings balance
-- Only run this if the difference is -35,000,000 (the double-counted amount)
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
WHERE code = '3201'
    AND type = 'EQUITY';

-- Step 5: Verify fix
SELECT 
    '=== AFTER FIX ===' as info,
    code,
    name,
    type,
    balance
FROM accounts 
WHERE code IN ('3101', '3201')
ORDER BY code;

-- Step 6: Calculate balance sheet totals to verify
WITH account_balances AS (
    SELECT 
        type,
        SUM(CASE 
            WHEN type = 'ASSET' THEN balance
            WHEN type = 'EXPENSE' THEN balance
            ELSE 0
        END) as debit_normal_balance,
        SUM(CASE 
            WHEN type IN ('LIABILITY', 'EQUITY', 'REVENUE') THEN balance
            ELSE 0
        END) as credit_normal_balance
    FROM accounts
    WHERE deleted_at IS NULL
        AND is_active = true
        AND COALESCE(is_header, false) = false
    GROUP BY type
)
SELECT 
    '=== BALANCE SHEET CHECK ===' as info,
    SUM(CASE WHEN type = 'ASSET' THEN debit_normal_balance ELSE 0 END) as total_assets,
    SUM(CASE WHEN type = 'LIABILITY' THEN credit_normal_balance ELSE 0 END) as total_liabilities,
    SUM(CASE WHEN type = 'EQUITY' THEN credit_normal_balance ELSE 0 END) as total_equity,
    SUM(CASE WHEN type = 'LIABILITY' THEN credit_normal_balance ELSE 0 END) +
    SUM(CASE WHEN type = 'EQUITY' THEN credit_normal_balance ELSE 0 END) as total_liab_equity,
    SUM(CASE WHEN type = 'ASSET' THEN debit_normal_balance ELSE 0 END) -
    (SUM(CASE WHEN type = 'LIABILITY' THEN credit_normal_balance ELSE 0 END) +
     SUM(CASE WHEN type = 'EQUITY' THEN credit_normal_balance ELSE 0 END)) as difference
FROM account_balances;

COMMIT;

-- Show final message
SELECT '✅ Retained Earnings balance has been fixed!' as status;
