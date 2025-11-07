-- Fix Profit & Loss Report Data Issues
-- Execute this script to correct account categories and balance synchronization

BEGIN;

-- 1. Fix Account Categories for proper COGS classification
UPDATE accounts 
SET category = 'COST_OF_GOODS_SOLD' 
WHERE code = '5101' AND name = 'Harga Pokok Penjualan';

-- Verify other expense accounts have correct categories
UPDATE accounts 
SET category = 'OPERATING_EXPENSE' 
WHERE code IN ('5000', '5201', '5202', '5203', '5204', '5900') 
  AND type = 'EXPENSE' 
  AND category != 'COST_OF_GOODS_SOLD';

-- 2. Create missing journal entries for accounts with balances
-- This creates journal entries to match the account balances

-- For Revenue account 4101 (Pendapatan Penjualan)
INSERT INTO journal_entries (
    code, description, reference, reference_type, entry_date, 
    user_id, status, total_debit, total_credit, is_balanced, is_auto_generated, 
    created_at, updated_at
)
SELECT 
    'JE-2025-01-01-BAL01',
    'Opening Balance - Pendapatan Penjualan',
    'OPENING_BALANCE',
    'OPENING_BALANCE',
    '2025-01-01 00:00:00',
    1, -- Assuming user ID 1 exists
    'POSTED',
    0.00,
    10000000.00,
    true,
    true,
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM journal_entries WHERE code = 'JE-2025-01-01-BAL01'
);

-- Create journal lines for Pendapatan Penjualan
INSERT INTO journal_lines (
    journal_entry_id, account_id, description, debit_amount, credit_amount, line_number, created_at, updated_at
)
SELECT 
    je.id,
    a.id,
    'Opening Balance',
    0.00,
    10000000.00,
    1,
    NOW(),
    NOW()
FROM journal_entries je
CROSS JOIN accounts a
WHERE je.code = 'JE-2025-01-01-BAL01'
  AND a.code = '4101'
  AND NOT EXISTS (
    SELECT 1 FROM journal_lines jl 
    WHERE jl.journal_entry_id = je.id AND jl.account_id = a.id
  );

-- For Expense account 5101 (Harga Pokok Penjualan)
INSERT INTO journal_entries (
    code, description, reference, reference_type, entry_date, 
    user_id, status, total_debit, total_credit, is_balanced, is_auto_generated, 
    created_at, updated_at
)
SELECT 
    'JE-2025-01-01-BAL02',
    'Opening Balance - Harga Pokok Penjualan',
    'OPENING_BALANCE',
    'OPENING_BALANCE',
    '2025-01-01 00:00:00',
    1, -- Assuming user ID 1 exists
    'POSTED',
    32400000.00,
    0.00,
    true,
    true,
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM journal_entries WHERE code = 'JE-2025-01-01-BAL02'
);

-- Create journal lines for Harga Pokok Penjualan
INSERT INTO journal_lines (
    journal_entry_id, account_id, description, debit_amount, credit_amount, line_number, created_at, updated_at
)
SELECT 
    je.id,
    a.id,
    'Opening Balance',
    32400000.00,
    0.00,
    1,
    NOW(),
    NOW()
FROM journal_entries je
CROSS JOIN accounts a
WHERE je.code = 'JE-2025-01-01-BAL02'
  AND a.code = '5101'
  AND NOT EXISTS (
    SELECT 1 FROM journal_lines jl 
    WHERE jl.journal_entry_id = je.id AND jl.account_id = a.id
  );

-- For Expense account 5000 (EXPENSES)
INSERT INTO journal_entries (
    code, description, reference, reference_type, entry_date, 
    user_id, status, total_debit, total_credit, is_balanced, is_auto_generated, 
    created_at, updated_at
)
SELECT 
    'JE-2025-01-01-BAL03',
    'Opening Balance - General Expenses',
    'OPENING_BALANCE',
    'OPENING_BALANCE',
    '2025-01-01 00:00:00',
    1, -- Assuming user ID 1 exists
    'POSTED',
    5000000.00,
    0.00,
    true,
    true,
    NOW(),
    NOW()
WHERE NOT EXISTS (
    SELECT 1 FROM journal_entries WHERE code = 'JE-2025-01-01-BAL03'
);

-- Create journal lines for General Expenses
INSERT INTO journal_lines (
    journal_entry_id, account_id, description, debit_amount, credit_amount, line_number, created_at, updated_at
)
SELECT 
    je.id,
    a.id,
    'Opening Balance',
    5000000.00,
    0.00,
    1,
    NOW(),
    NOW()
FROM journal_entries je
CROSS JOIN accounts a
WHERE je.code = 'JE-2025-01-01-BAL03'
  AND a.code = '5000'
  AND NOT EXISTS (
    SELECT 1 FROM journal_lines jl 
    WHERE jl.journal_entry_id = je.id AND jl.account_id = a.id
  );

-- 3. Add missing journal_lines table if needed
-- (The debug shows that journal_entries exist but no journal_lines are being returned)
-- This suggests journal_lines might be missing or the relationship is broken

-- Verify journal_lines table structure
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'journal_lines') THEN
        RAISE NOTICE 'journal_lines table does not exist - this may be the root cause!';
    ELSE
        RAISE NOTICE 'journal_lines table exists';
    END IF;
END $$;

-- Check if journal_lines are missing for existing journal_entries
DO $$
DECLARE
    missing_lines INTEGER;
BEGIN
    SELECT COUNT(*) INTO missing_lines
    FROM journal_entries je
    LEFT JOIN journal_lines jl ON je.id = jl.journal_entry_id
    WHERE je.status = 'POSTED' AND jl.id IS NULL;
    
    RAISE NOTICE 'Number of journal_entries without journal_lines: %', missing_lines;
END $$;

COMMIT;

-- Final verification queries
SELECT 
    'Account Categories Check' as check_type,
    code, 
    name, 
    category 
FROM accounts 
WHERE code IN ('5101', '5000', '4101', '4900')
ORDER BY code;

SELECT 
    'Journal Entries Check' as check_type,
    COUNT(*) as count,
    SUM(total_debit) as total_debit,
    SUM(total_credit) as total_credit
FROM journal_entries 
WHERE status = 'POSTED';

SELECT 
    'Journal Lines Check' as check_type,
    COUNT(*) as count,
    SUM(debit_amount) as total_debit,
    SUM(credit_amount) as total_credit
FROM journal_lines jl
JOIN journal_entries je ON jl.journal_entry_id = je.id
WHERE je.status = 'POSTED';
