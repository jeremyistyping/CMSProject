-- Fix Bank Mandiri Balance Bug
-- Problem: Bank Mandiri balance shows Rp 11.100.000 instead of correct Rp 5.550.000
-- Root cause: Payment service used total payment amount instead of actual payment received

-- Step 1: Check current state
SELECT 
    cb.id,
    cb.name,
    cb.balance as current_balance,
    a.code as account_code,
    a.name as account_name,
    a.balance as account_balance
FROM cash_banks cb
JOIN accounts a ON cb.account_id = a.id 
WHERE cb.name LIKE '%Mandiri%' OR a.code = '1103';

-- Step 2: Check recent payment transactions
SELECT 
    cbt.id,
    cbt.cash_bank_id,
    cbt.reference_type,
    cbt.reference_id,
    cbt.amount,
    cbt.balance_after,
    cbt.transaction_date,
    cbt.notes
FROM cash_bank_transactions cbt
JOIN cash_banks cb ON cbt.cash_bank_id = cb.id
WHERE cb.name LIKE '%Mandiri%' 
ORDER BY cbt.transaction_date DESC
LIMIT 10;

-- Step 3: Check payment records
SELECT 
    p.id,
    p.code,
    p.contact_id,
    p.amount as payment_amount,
    p.date,
    p.status,
    c.name as customer_name
FROM payments p
JOIN contacts c ON p.contact_id = c.id
WHERE p.created_at >= CURRENT_DATE - INTERVAL 1 DAY
ORDER BY p.created_at DESC;

-- Step 4: Correction - Update Bank Mandiri balance
-- WARNING: Run this only after confirming the issue
BEGIN;

-- Update cash_banks table balance
UPDATE cash_banks 
SET 
    balance = 5550000.00,
    updated_at = CURRENT_TIMESTAMP
WHERE name LIKE '%Mandiri%' AND balance = 11100000.00;

-- Update corresponding account balance
UPDATE accounts 
SET 
    balance = 5550000.00,
    updated_at = CURRENT_TIMESTAMP  
WHERE code = '1103' AND balance = 11100000.00;

-- Create corrective cash bank transaction
INSERT INTO cash_bank_transactions (
    cash_bank_id,
    reference_type,
    reference_id,
    amount,
    balance_after,
    transaction_date,
    notes,
    created_at,
    updated_at
)
SELECT 
    cb.id,
    'ADJUSTMENT',
    0,
    -5550000.00,
    5550000.00,
    CURRENT_TIMESTAMP,
    'Manual correction - Bug fix for double payment recording',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM cash_banks cb
WHERE cb.name LIKE '%Mandiri%';

-- Create corrective journal entry
INSERT INTO journal_entries (
    code,
    entry_date,
    description,
    reference_type,
    reference,
    user_id,
    status,
    total_debit,
    total_credit,
    is_auto_generated,
    created_at,
    updated_at
) VALUES (
    CONCAT('ADJ/', YEAR(CURDATE()), '/', LPAD(MONTH(CURDATE()), 2, '0'), '/001'),
    CURRENT_DATE,
    'Correction - Bank Mandiri Balance Bug Fix',
    'ADJUSTMENT',
    'BANK_MANDIRI_FIX_001',
    1,
    'POSTED',
    5550000.00,
    5550000.00,
    true,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Get the journal entry ID for lines
SET @journal_id = LAST_INSERT_ID();

-- Add journal lines for the correction
INSERT INTO journal_lines (
    journal_entry_id,
    account_id,
    description,
    debit_amount,
    credit_amount,
    line_number,
    created_at,
    updated_at
) VALUES 
-- Debit: Suspense/Error account (create if needed)
(
    @journal_id,
    (SELECT id FROM accounts WHERE code = '9999' OR name LIKE '%Suspense%' LIMIT 1),
    'Correction - Excess Bank Mandiri balance',
    5550000.00,
    0.00,
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
),
-- Credit: Bank Mandiri account
(
    @journal_id,
    (SELECT id FROM accounts WHERE code = '1103'),
    'Correction - Reduce Bank Mandiri balance',
    0.00,
    5550000.00,
    2,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Step 5: Verify correction
SELECT 
    'After Correction:' as status,
    cb.name,
    cb.balance as cash_bank_balance,
    a.code,
    a.balance as account_balance
FROM cash_banks cb
JOIN accounts a ON cb.account_id = a.id 
WHERE cb.name LIKE '%Mandiri%' OR a.code = '1103';

-- Check if trial balance is still balanced
SELECT 
    'Trial Balance Check:' as status,
    SUM(CASE WHEN type IN ('ASSET', 'EXPENSE') THEN balance ELSE 0 END) as total_debits,
    SUM(CASE WHEN type IN ('LIABILITY', 'EQUITY', 'REVENUE') THEN balance ELSE 0 END) as total_credits,
    ABS(SUM(CASE WHEN type IN ('ASSET', 'EXPENSE') THEN balance ELSE -balance END)) as difference
FROM accounts 
WHERE deleted_at IS NULL;

COMMIT;