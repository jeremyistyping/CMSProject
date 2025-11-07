-- ============================================================================
-- QUICK DUPLICATE REMOVAL AND BALANCE SHEET FIX
-- Script untuk menghapus duplikat jurnal dan membuat balance sheet seimbang
-- ============================================================================

-- 1. Check current balance sheet status
WITH account_balances AS (
    SELECT 
        a.type as account_type,
        CASE 
            WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
                COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
            ELSE 
                COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
        END as net_balance
    FROM accounts a
    LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
    LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
    WHERE (uje.status = 'POSTED' AND uje.entry_date <= CURRENT_DATE) OR uje.status IS NULL
    GROUP BY a.id, a.type
    HAVING a.type IN ('ASSET', 'LIABILITY', 'EQUITY')
)
SELECT 
    'Before Fix' as status,
    COALESCE(SUM(CASE WHEN account_type = 'ASSET' THEN net_balance ELSE 0 END), 0) as total_assets,
    COALESCE(SUM(CASE WHEN account_type = 'LIABILITY' THEN net_balance ELSE 0 END), 0) as total_liabilities,
    COALESCE(SUM(CASE WHEN account_type = 'EQUITY' THEN net_balance ELSE 0 END), 0) as total_equity,
    COALESCE(SUM(CASE WHEN account_type = 'ASSET' THEN net_balance ELSE 0 END), 0) - 
    (COALESCE(SUM(CASE WHEN account_type = 'LIABILITY' THEN net_balance ELSE 0 END), 0) + 
     COALESCE(SUM(CASE WHEN account_type = 'EQUITY' THEN net_balance ELSE 0 END), 0)) as balance_difference
FROM account_balances;

-- 2. Find and show duplicate journal entries
WITH duplicate_lines AS (
    SELECT 
        a.code as account_code,
        a.name as account_name,
        DATE(uje.entry_date) as entry_date,
        ujl.debit_amount,
        ujl.credit_amount,
        ujl.description,
        COUNT(*) as count,
        ARRAY_AGG(ujl.id ORDER BY ujl.id) as line_ids,
        ARRAY_AGG(ujl.journal_id ORDER BY ujl.id) as journal_ids
    FROM unified_journal_lines ujl
    JOIN accounts a ON a.id = ujl.account_id
    JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
    WHERE uje.status = 'POSTED'
    GROUP BY a.code, a.name, DATE(uje.entry_date), ujl.debit_amount, ujl.credit_amount, ujl.description
    HAVING COUNT(*) > 1
)
SELECT 
    'Duplicates Found' as info,
    account_code,
    account_name,
    entry_date,
    debit_amount,
    credit_amount,
    count as duplicate_count,
    line_ids,
    journal_ids
FROM duplicate_lines
ORDER BY count DESC, entry_date DESC;

-- 3. Remove duplicate journal entries (keep the first, remove the rest)
-- WARNING: This will delete data! Make sure you have a backup!

-- First, let's create a temp table with IDs to delete
CREATE TEMP TABLE duplicate_lines_to_delete AS
WITH duplicate_analysis AS (
    SELECT 
        ujl.id as line_id,
        ujl.journal_id,
        a.code as account_code,
        ROW_NUMBER() OVER (
            PARTITION BY a.code, DATE(uje.entry_date), ujl.debit_amount, ujl.credit_amount, ujl.description 
            ORDER BY ujl.id
        ) as row_num
    FROM unified_journal_lines ujl
    JOIN accounts a ON a.id = ujl.account_id
    JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
    WHERE uje.status = 'POSTED'
)
SELECT line_id, journal_id, account_code
FROM duplicate_analysis 
WHERE row_num > 1;

-- Show what will be deleted
SELECT 'Lines to be deleted' as info, COUNT(*) as count FROM duplicate_lines_to_delete;
SELECT * FROM duplicate_lines_to_delete ORDER BY account_code, journal_id;

-- Delete the duplicate lines (uncomment to execute)
-- DELETE FROM unified_journal_lines WHERE id IN (SELECT line_id FROM duplicate_lines_to_delete);

-- Delete empty journal headers (uncomment to execute)
-- DELETE FROM unified_journal_ledger WHERE id NOT IN (SELECT DISTINCT journal_id FROM unified_journal_lines WHERE journal_id IS NOT NULL);

-- 4. Check balance sheet after duplicate removal (run after uncommenting delete statements)
/*
WITH account_balances AS (
    SELECT 
        a.type as account_type,
        CASE 
            WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
                COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
            ELSE 
                COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
        END as net_balance
    FROM accounts a
    LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
    LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
    WHERE (uje.status = 'POSTED' AND uje.entry_date <= CURRENT_DATE) OR uje.status IS NULL
    GROUP BY a.id, a.type
    HAVING a.type IN ('ASSET', 'LIABILITY', 'EQUITY')
)
SELECT 
    'After Duplicate Removal' as status,
    COALESCE(SUM(CASE WHEN account_type = 'ASSET' THEN net_balance ELSE 0 END), 0) as total_assets,
    COALESCE(SUM(CASE WHEN account_type = 'LIABILITY' THEN net_balance ELSE 0 END), 0) as total_liabilities,
    COALESCE(SUM(CASE WHEN account_type = 'EQUITY' THEN net_balance ELSE 0 END), 0) as total_equity,
    COALESCE(SUM(CASE WHEN account_type = 'ASSET' THEN net_balance ELSE 0 END), 0) - 
    (COALESCE(SUM(CASE WHEN account_type = 'LIABILITY' THEN net_balance ELSE 0 END), 0) + 
     COALESCE(SUM(CASE WHEN account_type = 'EQUITY' THEN net_balance ELSE 0 END), 0)) as balance_difference
FROM account_balances;
*/

-- 5. Create adjusting entry if needed (modify amount as needed)
-- First check table structure to see required columns
SELECT column_name, data_type, is_nullable, column_default 
FROM information_schema.columns 
WHERE table_name = 'unified_journal_ledger' 
ORDER BY ordinal_position;

-- ============================================================================
-- MANUAL EXECUTION INSTRUCTIONS:
-- 1. Run sections 1-2 to analyze current state
-- 2. Review the duplicate entries found
-- 3. Uncomment the DELETE statements in section 3 to remove duplicates
-- 4. Uncomment section 4 to check balance after duplicate removal
-- 5. If still not balanced, create manual adjusting entry based on table structure
-- ============================================================================