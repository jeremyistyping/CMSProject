-- ============================================================================
-- BALANCE SHEET ADJUSTMENTS FOR SSOT JOURNAL SYSTEM
-- Script untuk membuat adjusting entries untuk balance sheet seimbang
-- ============================================================================

-- 1. Check current balance sheet status
SELECT 
    'Balance Sheet Analysis' as report_type,
    a.type as account_type,
    COUNT(*) as account_count,
    SUM(CASE 
        WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
            COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
        ELSE 
            COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
    END) as total_balance
FROM accounts a
LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
WHERE (uje.status = 'POSTED' AND uje.entry_date <= CURDATE()) OR uje.status IS NULL
GROUP BY a.type
HAVING a.type IN ('ASSET', 'LIABILITY', 'EQUITY')
ORDER BY a.type;

-- 2. Show detailed account balances that might need adjustment
SELECT 
    a.code as account_code,
    a.name as account_name,
    a.type as account_type,
    COALESCE(SUM(ujl.debit_amount), 0) as total_debit,
    COALESCE(SUM(ujl.credit_amount), 0) as total_credit,
    CASE 
        WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
            COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
        ELSE 
            COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
    END as net_balance
FROM accounts a
LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
WHERE (uje.status = 'POSTED' AND uje.entry_date <= CURDATE()) OR uje.status IS NULL
GROUP BY a.id, a.code, a.name, a.type
HAVING a.type IN ('ASSET', 'LIABILITY', 'EQUITY') 
    AND ABS(CASE 
        WHEN a.type IN ('ASSET', 'EXPENSE') THEN 
            COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0)
        ELSE 
            COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0)
    END) > 0
ORDER BY a.code;

-- 3. Fix account types if needed (uncomment to execute)
-- Update PPN Masukan to be ASSET type
-- UPDATE accounts SET type = 'ASSET' WHERE code = '2102' AND name LIKE '%PPN Masukan%';

-- Update any other misclassified accounts
-- UPDATE accounts SET type = 'ASSET' WHERE code LIKE '1%' AND type != 'ASSET';
-- UPDATE accounts SET type = 'LIABILITY' WHERE code LIKE '2%' AND code != '2102' AND type != 'LIABILITY';  
-- UPDATE accounts SET type = 'EQUITY' WHERE code LIKE '3%' AND type != 'EQUITY';

-- 4. Check for duplicate journal entries
SELECT 
    'Duplicate Check' as check_type,
    a.code as account_code,
    a.name as account_name,
    DATE(uje.entry_date) as entry_date,
    ujl.debit_amount,
    ujl.credit_amount,
    COUNT(*) as duplicate_count
FROM unified_journal_lines ujl
JOIN accounts a ON a.id = ujl.account_id
JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
WHERE uje.status = 'POSTED'
GROUP BY a.code, a.name, DATE(uje.entry_date), ujl.debit_amount, ujl.credit_amount
HAVING COUNT(*) > 1
ORDER BY duplicate_count DESC, entry_date DESC;

-- 5. Create adjusting entry if needed (EXAMPLE - modify amounts as needed)
-- This creates a general adjusting entry to balance the books
/*
-- First, calculate the difference
SET @balance_difference = (
    SELECT 
        SUM(CASE WHEN a.type = 'ASSET' THEN 
            COALESCE(SUM(ujl.debit_amount), 0) - COALESCE(SUM(ujl.credit_amount), 0) 
        ELSE 0 END) - 
        SUM(CASE WHEN a.type IN ('LIABILITY', 'EQUITY') THEN 
            COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0) 
        ELSE 0 END)
    FROM accounts a
    LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
    LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
    WHERE (uje.status = 'POSTED' AND uje.entry_date <= CURDATE()) OR uje.status IS NULL
    GROUP BY a.type
    HAVING a.type IN ('ASSET', 'LIABILITY', 'EQUITY')
);

-- Insert adjusting entry header
INSERT INTO unified_journal_ledger (
    entry_number,
    entry_date,
    description,
    status,
    created_at,
    updated_at
) VALUES (
    CONCAT('ADJ-', DATE_FORMAT(NOW(), '%Y%m%d'), '-001'),
    CURDATE(),
    'Balance Sheet Adjusting Entry - Automated',
    'POSTED',
    NOW(),
    NOW()
);

SET @journal_id = LAST_INSERT_ID();

-- Insert adjusting entry lines (example for retained earnings adjustment)
-- Adjust this based on your specific needs
INSERT INTO unified_journal_lines (
    journal_id,
    account_id,
    debit_amount,
    credit_amount,
    description,
    created_at,
    updated_at
) VALUES 
-- Debit retained earnings if assets > liabilities + equity
(
    @journal_id,
    (SELECT id FROM accounts WHERE code = '3201' LIMIT 1), -- Retained Earnings
    CASE WHEN @balance_difference > 0 THEN @balance_difference ELSE 0 END,
    CASE WHEN @balance_difference < 0 THEN ABS(@balance_difference) ELSE 0 END,
    'Adjusting entry for balance sheet balance',
    NOW(),
    NOW()
),
-- Credit/Debit adjustment account (use suspense account)
(
    @journal_id,
    (SELECT id FROM accounts WHERE code = '1901' OR name LIKE '%Suspense%' LIMIT 1), -- Suspense Account
    CASE WHEN @balance_difference < 0 THEN ABS(@balance_difference) ELSE 0 END,
    CASE WHEN @balance_difference > 0 THEN @balance_difference ELSE 0 END,
    'Adjusting entry for balance sheet balance',
    NOW(),
    NOW()
);
*/

-- 6. Final balance check query
SELECT 
    'Final Balance Check' as check_type,
    ROUND(SUM(CASE WHEN a.type = 'ASSET' THEN 
        COALESCE(bal.debit_total, 0) - COALESCE(bal.credit_total, 0) 
    ELSE 0 END), 2) as total_assets,
    
    ROUND(SUM(CASE WHEN a.type = 'LIABILITY' THEN 
        COALESCE(bal.credit_total, 0) - COALESCE(bal.debit_total, 0) 
    ELSE 0 END), 2) as total_liabilities,
    
    ROUND(SUM(CASE WHEN a.type = 'EQUITY' THEN 
        COALESCE(bal.credit_total, 0) - COALESCE(bal.debit_total, 0) 
    ELSE 0 END), 2) as total_equity,
    
    ROUND(SUM(CASE WHEN a.type = 'ASSET' THEN 
        COALESCE(bal.debit_total, 0) - COALESCE(bal.credit_total, 0) 
    ELSE 0 END) - 
    SUM(CASE WHEN a.type IN ('LIABILITY', 'EQUITY') THEN 
        COALESCE(bal.credit_total, 0) - COALESCE(bal.debit_total, 0) 
    ELSE 0 END), 2) as balance_difference,
    
    CASE WHEN ABS(
        SUM(CASE WHEN a.type = 'ASSET' THEN 
            COALESCE(bal.debit_total, 0) - COALESCE(bal.credit_total, 0) 
        ELSE 0 END) - 
        SUM(CASE WHEN a.type IN ('LIABILITY', 'EQUITY') THEN 
            COALESCE(bal.credit_total, 0) - COALESCE(bal.debit_total, 0) 
        ELSE 0 END)
    ) <= 0.01 THEN 'BALANCED' ELSE 'NOT BALANCED' END as status
FROM accounts a
LEFT JOIN (
    SELECT 
        ujl.account_id,
        SUM(ujl.debit_amount) as debit_total,
        SUM(ujl.credit_amount) as credit_total
    FROM unified_journal_lines ujl
    JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
    WHERE uje.status = 'POSTED' AND uje.entry_date <= CURDATE()
    GROUP BY ujl.account_id
) bal ON bal.account_id = a.id
WHERE a.type IN ('ASSET', 'LIABILITY', 'EQUITY');

-- ============================================================================
-- USAGE INSTRUCTIONS:
-- 1. Run sections 1-4 to analyze current balance sheet status
-- 2. Review the results and identify issues
-- 3. Uncomment and modify section 3 to fix account types if needed  
-- 4. Uncomment and modify section 5 to create adjusting entries if needed
-- 5. Run section 6 to verify the balance sheet is now balanced
-- ============================================================================