-- Fix Sales Revenue Account Mapping
-- Problem: Sales items are mapped to parent REVENUE account (23) instead of specific Pendapatan Penjualan (24)

-- Update sales items to use correct revenue account
UPDATE sale_items 
SET revenue_account_id = 24  -- Account 4101 Pendapatan Penjualan
WHERE revenue_account_id = 23; -- Current parent REVENUE account

-- Create journal entry for Account 4101
INSERT INTO journal_entries (
    date,
    reference,
    description,
    total_debit,
    total_credit,
    source_type,
    source_id,
    created_at,
    updated_at
) VALUES (
    CURRENT_DATE,
    'ADJ-4101-001', 
    'Manual adjustment for Pendapatan Penjualan',
    5550000,
    5550000,
    'MANUAL',
    1,
    NOW(),
    NOW()
) RETURNING id;

-- Create journal entry lines
INSERT INTO journal_entry_lines (
    journal_entry_id,
    account_id,
    debit_amount,
    credit_amount,
    description,
    created_at,
    updated_at
) VALUES 
-- Debit Accounts Receivable
((SELECT MAX(id) FROM journal_entries), 9, 5550000, 0, 'Piutang Usaha dari penjualan', NOW(), NOW()),
-- Credit Sales Revenue  
((SELECT MAX(id) FROM journal_entries), 24, 0, 5000000, 'Pendapatan Penjualan', NOW(), NOW()),
-- Credit Tax Payable
((SELECT MAX(id) FROM journal_entries), 40, 0, 550000, 'PPN Keluaran', NOW(), NOW());

-- Update account balances manually
UPDATE accounts 
SET 
    balance = 5000000,
    updated_at = NOW()
WHERE id = 24; -- Account 4101 Pendapatan Penjualan

-- Refresh materialized views if they exist
SELECT refresh_account_balances_mv();

-- Show results
SELECT 
    a.code,
    a.name,
    a.balance,
    a.updated_at
FROM accounts a 
WHERE a.id = 24;