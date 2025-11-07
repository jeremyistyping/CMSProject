-- Direct PostgreSQL Fix for Bank Mandiri Balance Bug
-- Run this manually if auto-migration fails

-- Step 1: Check current balance
SELECT 
    cb.id,
    cb.name,
    cb.balance as cash_bank_balance,
    a.id as account_id,
    a.code as account_code,
    a.name as account_name,
    a.balance as account_balance
FROM cash_banks cb
JOIN accounts a ON cb.account_id = a.id 
WHERE cb.name ILIKE '%Mandiri%' OR a.code = '1103'
ORDER BY cb.id;

-- Step 2: Apply correction if needed
DO $$
DECLARE
    found_bank_id INTEGER;
    found_account_id INTEGER; 
    current_balance DECIMAL(15,2);
BEGIN
    -- Get Bank Mandiri info
    SELECT cb.id, cb.account_id, cb.balance 
    INTO found_bank_id, found_account_id, current_balance
    FROM cash_banks cb
    LEFT JOIN accounts a ON cb.account_id = a.id 
    WHERE (cb.name ILIKE '%Mandiri%' OR a.code = '1103')
      AND cb.balance = 11100000.00
    LIMIT 1;
    
    -- Apply correction if balance is wrong
    IF FOUND AND current_balance = 11100000.00 THEN
        RAISE NOTICE 'Found Bank Mandiri with wrong balance: %.2f. Applying correction...', current_balance;
        
        -- Update cash_banks balance
        UPDATE cash_banks 
        SET balance = 5550000.00, updated_at = NOW() 
        WHERE id = found_bank_id;
        
        -- Update accounts balance  
        UPDATE accounts 
        SET balance = 5550000.00, updated_at = NOW() 
        WHERE id = found_account_id;
        
        -- Insert corrective transaction
        INSERT INTO cash_bank_transactions (
            cash_bank_id, reference_type, reference_id, amount, balance_after,
            transaction_date, notes, created_at, updated_at
        ) VALUES (
            found_bank_id, 'ADJUSTMENT', 0, -5550000.00, 5550000.00,
            NOW(), 'Manual fix: Bank Mandiri balance correction from 11,100,000 to 5,550,000', NOW(), NOW()
        );
        
        RAISE NOTICE '✅ Bank Mandiri balance corrected from 11,100,000 to 5,550,000';
    ELSE
        RAISE NOTICE '✅ Bank Mandiri balance is already correct or not found';
    END IF;
END $$;

-- Step 3: Verify correction
SELECT 
    '=== AFTER CORRECTION ===' as status,
    cb.name,
    cb.balance as cash_bank_balance,
    a.code,
    a.name as account_name,
    a.balance as account_balance
FROM cash_banks cb
JOIN accounts a ON cb.account_id = a.id 
WHERE cb.name ILIKE '%Mandiri%' OR a.code = '1103'
ORDER BY cb.id;

-- Step 4: Check recent transactions
SELECT 
    'Recent Bank Mandiri Transactions:' as info,
    cbt.transaction_date,
    cbt.reference_type,
    cbt.amount,
    cbt.balance_after,
    cbt.notes
FROM cash_bank_transactions cbt
JOIN cash_banks cb ON cbt.cash_bank_id = cb.id
WHERE cb.name ILIKE '%Mandiri%'
ORDER BY cbt.transaction_date DESC
LIMIT 5;