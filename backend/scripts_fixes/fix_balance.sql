-- Fix balance for testing payment integration
-- This adds sufficient balance to cash/bank account ID 6 for testing

UPDATE cash_banks 
SET balance = 10000000.00 
WHERE id = 6;

-- Also update other accounts for good measure
UPDATE cash_banks 
SET balance = CASE 
    WHEN type = 'CASH' THEN 5000000.00
    WHEN type = 'BANK' THEN 10000000.00
    ELSE balance 
END
WHERE is_active = true;

-- Check the results
SELECT id, code, name, type, balance, is_active 
FROM cash_banks 
ORDER BY id;
