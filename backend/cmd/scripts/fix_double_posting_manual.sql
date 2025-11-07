-- Fix Double Posting Balance for BNK-2025-0004
-- This script corrects the doubled balance issue

-- Step 1: Check current balance
SELECT 
    cb.account_code,
    cb.name,
    cb.balance as current_balance,
    a.balance as gl_balance
FROM cash_banks cb
JOIN accounts a ON cb.account_id = a.id
WHERE cb.account_code = 'BNK-2025-0004';

-- Step 2: Analyze journal entries for this cash bank account
SELECT 
    je.id,
    je.code,
    je.entry_date,
    je.description,
    je.total_debit,
    je.total_credit,
    je.status,
    jl.account_id,
    jl.debit_amount,
    jl.credit_amount,
    a.code as account_code,
    a.name as account_name
FROM journal_entries je
JOIN journal_lines jl ON je.id = jl.journal_entry_id
JOIN accounts a ON jl.account_id = a.id
JOIN cash_banks cb ON a.id = cb.account_id
WHERE cb.account_code = 'BNK-2025-0004'
ORDER BY je.entry_date DESC, je.id DESC;

-- Step 3: Check for duplicate transactions
SELECT 
    reference_id,
    reference_type,
    COUNT(*) as count,
    SUM(debit_amount) as total_debit,
    SUM(credit_amount) as total_credit
FROM journal_lines jl
JOIN accounts a ON jl.account_id = a.id
JOIN cash_banks cb ON a.id = cb.account_id
WHERE cb.account_code = 'BNK-2025-0004'
GROUP BY reference_id, reference_type
HAVING COUNT(*) > 1;

-- Step 4: Fix the doubled balance - reduce by half if it's doubled
-- CAUTION: Only run this if you've confirmed the balance is exactly doubled
UPDATE cash_banks 
SET 
    balance = balance / 2,
    updated_at = NOW()
WHERE 
    account_code = 'BNK-2025-0004'
    AND balance = 20000000.00  -- Only if current balance is exactly IDR 20,000,000
    AND EXISTS (
        -- Verify this is a double posting case
        SELECT 1 FROM cash_bank_transactions 
        WHERE cash_bank_id = cash_banks.id 
        AND amount = 10000000.00
    );

-- Step 5: Also update the corresponding GL account balance
UPDATE accounts 
SET 
    balance = balance / 2,
    updated_at = NOW()
WHERE 
    id IN (
        SELECT account_id 
        FROM cash_banks 
        WHERE account_code = 'BNK-2025-0004'
    )
    AND balance = 20000000.00;  -- Only if current balance is exactly IDR 20,000,000

-- Step 6: Verify the fix
SELECT 
    'After Fix' as status,
    cb.account_code,
    cb.name,
    cb.balance as cash_bank_balance,
    a.balance as gl_account_balance,
    CASE 
        WHEN cb.balance = a.balance THEN 'CONSISTENT'
        ELSE 'INCONSISTENT'
    END as consistency_status
FROM cash_banks cb
JOIN accounts a ON cb.account_id = a.id
WHERE cb.account_code = 'BNK-2025-0004';

-- Step 7: Add a record to track this fix
INSERT INTO migration_records (migration_id, description, version, applied_at, created_at, updated_at)
VALUES (
    'fix_double_posting_BNK_2025_0004',
    'Manual fix for double posting balance in BNK-2025-0004 account',
    '1.0',
    NOW(),
    NOW(),
    NOW()
)
ON CONFLICT (migration_id) DO NOTHING;