-- Fix Bank BCA Negative Balance Issue
-- Problem: Bank BCA (1102) shows -Rp 5,550,000 which is wrong for Asset account
-- Solution: Correct the balance based on transaction history or set proper opening balance

-- Step 1: Analyze current state
SELECT 'Current Bank BCA Status' as analysis;
SELECT 
    a.id, a.code, a.name, a.type, a.balance as account_balance,
    cb.id as cash_bank_id, cb.name as cash_bank_name, cb.balance as cash_bank_balance
FROM accounts a
LEFT JOIN cash_banks cb ON a.id = cb.account_id
WHERE a.code = '1102' OR a.name LIKE '%BCA%';

-- Step 2: Check transaction history for Bank BCA
SELECT 'Bank BCA Transaction History' as analysis;
SELECT 
    cbt.id, cbt.cash_bank_id, cbt.reference_type, cbt.reference_id,
    cbt.amount, cbt.balance_after, cbt.transaction_date, cbt.notes
FROM cash_bank_transactions cbt
JOIN cash_banks cb ON cbt.cash_bank_id = cb.id
JOIN accounts a ON cb.account_id = a.id
WHERE a.code = '1102' OR a.name LIKE '%BCA%'
ORDER BY cbt.transaction_date DESC;

-- Step 3: Check journal entries affecting Bank BCA
SELECT 'SSOT Journal Entries for Bank BCA' as analysis;
SELECT 
    sje.id, sje.entry_number, sje.source_type, sje.reference,
    sjl.debit_amount, sjl.credit_amount, sjl.description,
    sje.created_at
FROM ssot_journal_entries sje
JOIN ssot_journal_lines sjl ON sje.id = sjl.journal_entry_id
JOIN accounts a ON sjl.account_id = a.id
WHERE a.code = '1102' OR a.name LIKE '%BCA%'
ORDER BY sje.created_at DESC
LIMIT 10;

-- Step 4: SOLUTION - Fix the balance
-- Based on your data: BCA had 10M opening balance, then 5.55M purchase payment
-- So final balance should be: 10M - 5.55M = 4.45M
BEGIN;

-- Set correct balance for Bank BCA based on actual scenario
-- Opening balance 10M minus purchase 5.55M = 4.45M remaining
UPDATE accounts 
SET balance = 4450000.00, -- 4.45M remaining balance (10M - 5.55M)
    updated_at = CURRENT_TIMESTAMP
WHERE code = '1102';

-- Update corresponding CashBank balance
UPDATE cash_banks cb
JOIN accounts a ON cb.account_id = a.id
SET cb.balance = a.balance,
    cb.updated_at = CURRENT_TIMESTAMP
WHERE a.code = '1102';

-- Create opening balance journal entry for audit trail
INSERT INTO ssot_journal_entries (
    entry_number, source_type, source_id, reference, entry_date, description,
    status, total_amount, created_at, updated_at
) VALUES (
    CONCAT('OB-BCA-', DATE_FORMAT(NOW(), '%Y%m%d')),
    'OPENING_BALANCE',
    NULL,
    'OPENING_BALANCE_BCA_FIX',
    CURRENT_DATE,
    'Opening Balance Correction - Bank BCA (10M initial minus 5.55M purchase)',
    'POSTED',
    10000000.00,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Get the journal entry ID
SET @journal_id = LAST_INSERT_ID();

-- Create journal lines
INSERT INTO ssot_journal_lines (
    journal_entry_id, account_id, description, 
    debit_amount, credit_amount, line_number, created_at, updated_at
) VALUES 
-- Debit Bank BCA (Asset increase)
(
    @journal_id,
    (SELECT id FROM accounts WHERE code = '1102'),
    'Opening Balance - Bank BCA',
    10000000.00,
    0.00,
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
),
-- Credit Modal Pemilik (Equity increase) 
(
    @journal_id,
    (SELECT id FROM accounts WHERE code = '3101'),
    'Opening Balance - Capital contribution',
    0.00,
    10000000.00,
    2,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

-- Create corresponding cash bank transaction
INSERT INTO cash_bank_transactions (
    cash_bank_id, reference_type, reference_id, amount, balance_after,
    transaction_date, notes, created_at, updated_at
) 
SELECT 
    cb.id,
    'OPENING_BALANCE',
    @journal_id,
    10000000.00,
    4450000.00,
    CURRENT_DATE,
    'Opening Balance - Bank BCA correction',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM cash_banks cb
JOIN accounts a ON cb.account_id = a.id
WHERE a.code = '1102';

COMMIT;

-- Step 5: Verification
SELECT 'After Fix - Verification' as status;
SELECT 
    a.code, a.name, a.type, a.balance as account_balance,
    cb.name as cash_bank_name, cb.balance as cash_bank_balance,
    CASE 
        WHEN a.type = 'Asset' AND a.balance >= 0 THEN '✅ CORRECT'
        WHEN a.type = 'Asset' AND a.balance < 0 THEN '❌ STILL NEGATIVE'
        ELSE '? UNKNOWN'
    END as status
FROM accounts a
LEFT JOIN cash_banks cb ON a.id = cb.account_id
WHERE a.code = '1102';

-- Alternative Option B: If BCA should have zero balance
-- Uncomment below if you prefer to zero out the balance instead

/*
BEGIN;

UPDATE accounts 
SET balance = 0.00, 
    updated_at = CURRENT_TIMESTAMP
WHERE code = '1102';

UPDATE cash_banks cb
JOIN accounts a ON cb.account_id = a.id
SET cb.balance = 0.00,
    cb.updated_at = CURRENT_TIMESTAMP
WHERE a.code = '1102';

-- Create corrective journal entry
INSERT INTO ssot_journal_entries (
    entry_number, source_type, source_id, reference, entry_date, description,
    status, total_amount, created_at, updated_at
) VALUES (
    CONCAT('CORR-BCA-', DATE_FORMAT(NOW(), '%Y%m%d')),
    'ADJUSTMENT',
    NULL,
    'BCA_BALANCE_CORRECTION',
    CURRENT_DATE,
    'Correction - Bank BCA Negative Balance',
    'POSTED',
    5550000.00,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

SET @journal_id = LAST_INSERT_ID();

INSERT INTO ssot_journal_lines (
    journal_entry_id, account_id, description, 
    debit_amount, credit_amount, line_number, created_at, updated_at
) VALUES 
-- Debit Bank BCA to eliminate negative balance
(
    @journal_id,
    (SELECT id FROM accounts WHERE code = '1102'),
    'Correction - Eliminate negative balance',
    5550000.00,
    0.00,
    1,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
),
-- Credit Suspense account (or create correction account)
(
    @journal_id,
    (SELECT id FROM accounts WHERE code = '9999' OR name LIKE '%Suspense%' LIMIT 1),
    'Correction - BCA negative balance fix',
    0.00,
    5550000.00,
    2,
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
);

COMMIT;
*/