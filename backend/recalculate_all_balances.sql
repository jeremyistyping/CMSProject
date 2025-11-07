-- ==========================================
-- RECALCULATE ALL ACCOUNT BALANCES
-- Fix balance untuk semua accounts dari journal entries
-- ==========================================

BEGIN;

-- Backup current balances
CREATE TEMP TABLE backup_balances AS
SELECT id, code, name, balance, updated_at
FROM accounts
WHERE deleted_at IS NULL;

SELECT 'Backup created: ' || COUNT(*) || ' accounts' FROM backup_balances;

-- Update all account balances from journal entries
UPDATE accounts a
SET 
    balance = COALESCE((
        SELECT 
            CASE 
                -- ASSET & EXPENSE: Debit increases, Credit decreases
                WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
                    SUM(ujl.debit_amount - ujl.credit_amount)
                -- LIABILITY, EQUITY, REVENUE: Credit increases, Debit decreases
                ELSE 
                    SUM(ujl.credit_amount - ujl.debit_amount)
            END
        FROM unified_journal_lines ujl
        JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
        WHERE ujl.account_id = a.id 
          AND uje.status = 'POSTED'
          AND uje.deleted_at IS NULL
    ), 0),
    updated_at = NOW()
WHERE a.deleted_at IS NULL
  AND a.is_header = false;  -- Only update detail accounts (non-header)

-- Show summary of changes
SELECT 
    'Total Updated' as metric,
    COUNT(*) as count,
    SUM(ABS(a.balance - b.balance)) as total_difference
FROM accounts a
JOIN backup_balances b ON b.id = a.id
WHERE a.balance != b.balance;

-- Show top 10 biggest changes
SELECT 
    a.code,
    a.name,
    a.type,
    b.balance as old_balance,
    a.balance as new_balance,
    a.balance - b.balance as difference
FROM accounts a
JOIN backup_balances b ON b.id = a.id
WHERE a.balance != b.balance
ORDER BY ABS(a.balance - b.balance) DESC
LIMIT 10;

-- Verify key accounts
SELECT 
    a.code,
    a.name,
    a.type,
    b.balance as old_balance,
    a.balance as new_balance,
    (SELECT COUNT(*) 
     FROM unified_journal_lines ujl
     JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
     WHERE ujl.account_id = a.id 
       AND uje.status = 'POSTED'
       AND uje.deleted_at IS NULL) as journal_count
FROM accounts a
LEFT JOIN backup_balances b ON b.id = a.id
WHERE a.code IN ('4101', '5101', '1201', '1301', '2103', '1101', '1102')
ORDER BY a.code;

-- COMMIT or ROLLBACK based on verification
-- COMMIT;
-- ROLLBACK;

SELECT 'Review the results above, then run COMMIT; to apply changes or ROLLBACK; to cancel' as next_step;
