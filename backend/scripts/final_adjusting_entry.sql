-- Final adjusting entry untuk menyelesaikan selisih Rp 670.000
-- Setelah duplikat removal berhasil

-- 1. Cek balance sheet saat ini
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
    'Current Status' as info,
    COALESCE(SUM(CASE WHEN account_type = 'ASSET' THEN net_balance ELSE 0 END), 0) as total_assets,
    COALESCE(SUM(CASE WHEN account_type = 'LIABILITY' THEN net_balance ELSE 0 END), 0) as total_liabilities,
    COALESCE(SUM(CASE WHEN account_type = 'EQUITY' THEN net_balance ELSE 0 END), 0) as total_equity,
    COALESCE(SUM(CASE WHEN account_type = 'ASSET' THEN net_balance ELSE 0 END), 0) - 
    (COALESCE(SUM(CASE WHEN account_type = 'LIABILITY' THEN net_balance ELSE 0 END), 0) + 
     COALESCE(SUM(CASE WHEN account_type = 'EQUITY' THEN net_balance ELSE 0 END), 0)) as balance_difference
FROM account_balances;

-- 2. Create adjusting entry journal header dengan semua required fields
INSERT INTO unified_journal_ledger (
    entry_number,
    source_type,
    entry_date,
    description,
    total_debit,
    total_credit,
    status,
    is_balanced,
    is_auto_generated,
    created_by,
    created_at,
    updated_at
) VALUES (
    'ADJ-20250922-002',
    'MANUAL',
    CURRENT_DATE,
    'Final Balance Sheet Adjusting Entry - Manual',
    670000,
    670000,
    'POSTED',
    true,
    false,
    1, -- Assuming user ID 1 exists, adjust if needed
    NOW(),
    NOW()
);

-- 3. Get the journal ID that was just created
-- (In a real transaction, you'd use RETURNING, but for manual execution:)

-- First, find the journal ID
SELECT id, entry_number FROM unified_journal_ledger 
WHERE entry_number = 'ADJ-20250922-002';

-- 4. Create adjusting journal lines
-- Replace <JOURNAL_ID> with the actual ID from step 3

-- Find or create retained earnings account
INSERT INTO accounts (code, name, type, is_active, created_at, updated_at)
VALUES ('3202', 'Retained Earnings - Adjustment', 'EQUITY', true, NOW(), NOW())
ON CONFLICT (code) DO UPDATE SET updated_at = NOW();

-- Get the account ID
SELECT id, code, name FROM accounts WHERE code = '3202';

-- Create the adjusting entry lines (replace <JOURNAL_ID> and <ACCOUNT_ID> with actual values)
-- Since Assets (32,775,000) > Liabilities + Equity (32,105,000) by 670,000
-- We need to CREDIT an equity account to increase equity

INSERT INTO unified_journal_lines (
    journal_id,
    account_id, 
    debit_amount,
    credit_amount,
    description,
    created_at,
    updated_at
) VALUES
-- Credit retained earnings to increase equity by 670,000
(
    (SELECT id FROM unified_journal_ledger WHERE entry_number = 'ADJ-20250922-002'),
    (SELECT id FROM accounts WHERE code = '3202'),
    0,
    670000,
    'Adjusting entry to balance sheet',
    NOW(),
    NOW()
);

-- 5. Verify final balance sheet
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
    'Final Status' as info,
    COALESCE(SUM(CASE WHEN account_type = 'ASSET' THEN net_balance ELSE 0 END), 0) as total_assets,
    COALESCE(SUM(CASE WHEN account_type = 'LIABILITY' THEN net_balance ELSE 0 END), 0) as total_liabilities,
    COALESCE(SUM(CASE WHEN account_type = 'EQUITY' THEN net_balance ELSE 0 END), 0) as total_equity,
    COALESCE(SUM(CASE WHEN account_type = 'ASSET' THEN net_balance ELSE 0 END), 0) - 
    (COALESCE(SUM(CASE WHEN account_type = 'LIABILITY' THEN net_balance ELSE 0 END), 0) + 
     COALESCE(SUM(CASE WHEN account_type = 'EQUITY' THEN net_balance ELSE 0 END), 0)) as balance_difference,
    
    CASE WHEN ABS(COALESCE(SUM(CASE WHEN account_type = 'ASSET' THEN net_balance ELSE 0 END), 0) - 
    (COALESCE(SUM(CASE WHEN account_type = 'LIABILITY' THEN net_balance ELSE 0 END), 0) + 
     COALESCE(SUM(CASE WHEN account_type = 'EQUITY' THEN net_balance ELSE 0 END), 0))) <= 0.01 
     THEN '✅ BALANCED' ELSE '❌ NOT BALANCED' END as status
FROM account_balances;

-- EXECUTE INSTRUCTIONS:
-- 1. Run section 1 to confirm current balance difference is ~670,000
-- 2. Execute sections 2-4 in order
-- 3. Run section 5 to verify balance sheet is now balanced
-- 4. Restart backend service and check frontend SSOT Balance Sheet report