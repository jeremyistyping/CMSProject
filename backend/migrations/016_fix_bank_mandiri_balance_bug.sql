-- Migration: Fix Bank Mandiri Balance Bug (Double Payment Recording)
-- Date: 2025-09-18
-- Issue: Bank Mandiri balance was incorrectly updated with full payment amount instead of allocated amount

-- PostgreSQL version - Check and fix Bank Mandiri balance

-- Apply correction if needed (Simple UPDATE approach)
-- Update Bank Mandiri balance from 11,100,000 to 5,550,000 if exists
UPDATE cash_banks 
SET balance = 5550000.00, updated_at = NOW() 
WHERE name ILIKE '%Mandiri%' AND balance = 11100000.00;

-- Update corresponding account balance
UPDATE accounts 
SET balance = 5550000.00, updated_at = NOW() 
WHERE code = '1103' AND balance = 11100000.00;

-- Insert corrective transaction for Bank Mandiri if correction was applied
INSERT INTO cash_bank_transactions (
    cash_bank_id, reference_type, reference_id, amount, balance_after,
    transaction_date, notes, created_at, updated_at
)
SELECT 
    cb.id, 'ADJUSTMENT', 0, -5550000.00, 5550000.00,
    NOW(), 'Auto-fix: Bank Mandiri balance correction', NOW(), NOW()
FROM cash_banks cb
WHERE cb.name ILIKE '%Mandiri%' AND cb.balance = 5550000.00
LIMIT 1;



-- Migration completed
