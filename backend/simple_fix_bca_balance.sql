-- SIMPLE FIX: Bank BCA Balance Issue
-- Problem: Bank BCA shows -5.550.000 instead of +4.450.000
-- Cause: Missing opening balance journal entry
-- Solution: Add the missing opening balance

-- Current situation analysis:
-- Initial balance: 10.000.000 (from seed/manual)
-- Purchase payment: 5.550.000 (correctly recorded as expense)  
-- Expected final: 4.450.000
-- Actual: -5.550.000 (wrong!)

-- Fix: Add the missing opening balance
BEGIN;

-- Step 1: Fix the account balance directly
UPDATE accounts 
SET balance = 4450000.00,
    updated_at = CURRENT_TIMESTAMP
WHERE code = '1102' AND name LIKE '%BCA%';

-- Step 2: Fix the corresponding cash_bank balance 
UPDATE cash_banks cb
JOIN accounts a ON cb.account_id = a.id
SET cb.balance = 4450000.00,
    cb.updated_at = CURRENT_TIMESTAMP
WHERE a.code = '1102';

-- Step 3: Create transaction record for the opening balance
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
    'OPENING_BALANCE',
    0,
    10000000.00,
    10000000.00,
    '2025-01-01',
    'Opening Balance - Bank BCA (missing initial balance)',
    CURRENT_TIMESTAMP,
    CURRENT_TIMESTAMP
FROM cash_banks cb
JOIN accounts a ON cb.account_id = a.id
WHERE a.code = '1102';

COMMIT;

-- Verification query:
SELECT 
    'After Fix' as status,
    a.code,
    a.name,
    a.balance as account_balance,
    cb.name as cash_bank_name,
    cb.balance as cash_bank_balance
FROM accounts a
LEFT JOIN cash_banks cb ON a.id = cb.account_id
WHERE a.code = '1102';

-- Expected result:
-- BCA Account Balance: 4.450.000 IDR  
-- BCA CashBank Balance: 4.450.000 IDR