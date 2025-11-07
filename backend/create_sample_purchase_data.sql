-- Create Sample Purchase Data for Testing Purchase Report
-- This will help verify that the Purchase Report works with real data

-- 1. Create sample vendor contact if not exists
INSERT INTO contacts (
    code, name, type, category, email, phone, address, 
    is_active, created_at, updated_at
) VALUES (
    'VENDOR-TEST-001',
    'PT Supplier Tester',
    'VENDOR',
    'WHOLESALE',
    'supplier@tester.com',
    '+62-21-1111111',
    'Jl. Test Vendor No. 123, Jakarta',
    true,
    NOW(),
    NOW()
) ON CONFLICT (code) DO NOTHING;

-- 2. Get vendor ID for reference
-- We'll use this in subsequent inserts

-- 3. Create sample purchase transaction in purchases table
INSERT INTO purchases (
    code, 
    vendor_id,
    purchase_date,
    due_date,
    total_amount,
    tax_amount,
    grand_total,
    payment_method,
    status,
    notes,
    created_at,
    updated_at
) VALUES (
    'PO-TEST-001',
    (SELECT id FROM contacts WHERE code = 'VENDOR-TEST-001' LIMIT 1),
    '2025-09-15'::date,
    '2025-10-15'::date,
    5000000.00,
    550000.00,  -- 11% VAT
    5550000.00,
    'CREDIT',
    'APPROVED',
    'Test purchase for Purchase Report testing',
    NOW(),
    NOW()
) ON CONFLICT (code) DO NOTHING;

-- 4. Create corresponding SSOT journal entry
INSERT INTO unified_journal_ledger (
    source_type,
    source_id,
    reference_number,
    entry_date,
    description,
    total_debit,
    total_credit,
    status,
    created_at,
    updated_at
) VALUES (
    'PURCHASE',
    (SELECT id FROM purchases WHERE code = 'PO-TEST-001' LIMIT 1),
    'PO-TEST-001',
    '2025-09-15'::date,
    'Purchase from PT Supplier Tester - Test Data',
    5550000.00,
    5550000.00,
    'POSTED',
    NOW(),
    NOW()
) ON CONFLICT (reference_number) DO NOTHING;

-- 5. Create journal lines for the purchase (Debit Inventory/Expense, Credit Payable)
-- Get the journal ID first
DO $$
DECLARE 
    journal_id_var INTEGER;
    inventory_account_id INTEGER;
    payable_account_id INTEGER;
    tax_account_id INTEGER;
BEGIN
    -- Get journal ID
    SELECT id INTO journal_id_var 
    FROM unified_journal_ledger 
    WHERE reference_number = 'PO-TEST-001' 
    LIMIT 1;
    
    -- Get account IDs
    SELECT id INTO inventory_account_id FROM accounts WHERE code = '1301' LIMIT 1; -- Inventory
    SELECT id INTO payable_account_id FROM accounts WHERE code = '2101' LIMIT 1;   -- Accounts Payable  
    SELECT id INTO tax_account_id FROM accounts WHERE code = '2102' LIMIT 1;       -- VAT Input
    
    -- Skip if journal not found or accounts not found
    IF journal_id_var IS NULL OR inventory_account_id IS NULL OR payable_account_id IS NULL THEN
        RAISE NOTICE 'Skipping journal lines creation - missing data';
        RETURN;
    END IF;
    
    -- Insert Debit: Inventory (5,000,000)
    INSERT INTO unified_journal_lines (
        journal_id, account_id, debit_amount, credit_amount, description
    ) VALUES (
        journal_id_var, inventory_account_id, 5000000.00, 0, 'Inventory Purchase - Test Data'
    ) ON CONFLICT DO NOTHING;
    
    -- Insert Debit: VAT Input (550,000) 
    IF tax_account_id IS NOT NULL THEN
        INSERT INTO unified_journal_lines (
            journal_id, account_id, debit_amount, credit_amount, description
        ) VALUES (
            journal_id_var, tax_account_id, 550000.00, 0, 'VAT Input - Test Data'
        ) ON CONFLICT DO NOTHING;
    END IF;
    
    -- Insert Credit: Accounts Payable (5,550,000)
    INSERT INTO unified_journal_lines (
        journal_id, account_id, debit_amount, credit_amount, description
    ) VALUES (
        journal_id_var, payable_account_id, 0, 5550000.00, 'Accounts Payable - PT Supplier Tester'
    ) ON CONFLICT DO NOTHING;
    
END $$;

-- 6. Update account balances to reflect the transaction
UPDATE accounts SET balance = balance + 5000000.00 WHERE code = '1301'; -- Inventory increase
UPDATE accounts SET balance = balance + 550000.00 WHERE code = '2102';  -- VAT Input increase  
UPDATE accounts SET balance = balance + 5550000.00 WHERE code = '2101'; -- Payable increase

-- 7. Create a partial payment to show payment analysis
INSERT INTO unified_journal_ledger (
    source_type,
    source_id,
    reference_number,
    entry_date,
    description,
    total_debit,
    total_credit,
    status,
    created_at,
    updated_at
) VALUES (
    'PAYMENT',
    (SELECT id FROM purchases WHERE code = 'PO-TEST-001' LIMIT 1),
    'PAY-TEST-001',
    '2025-09-20'::date,
    'Partial payment to PT Supplier Tester',
    2000000.00,
    2000000.00,
    'POSTED',
    NOW(),
    NOW()
) ON CONFLICT (reference_number) DO NOTHING;

-- 8. Create payment journal lines
DO $$
DECLARE 
    payment_journal_id INTEGER;
    cash_account_id INTEGER;
    payable_account_id INTEGER;
BEGIN
    -- Get payment journal ID
    SELECT id INTO payment_journal_id 
    FROM unified_journal_ledger 
    WHERE reference_number = 'PAY-TEST-001' 
    LIMIT 1;
    
    -- Get account IDs
    SELECT id INTO cash_account_id FROM accounts WHERE code = '1101' LIMIT 1;    -- Cash
    SELECT id INTO payable_account_id FROM accounts WHERE code = '2101' LIMIT 1; -- Accounts Payable
    
    -- Skip if not found
    IF payment_journal_id IS NULL OR cash_account_id IS NULL OR payable_account_id IS NULL THEN
        RAISE NOTICE 'Skipping payment journal lines creation - missing data';
        RETURN;
    END IF;
    
    -- Debit: Accounts Payable (reduce payable)
    INSERT INTO unified_journal_lines (
        journal_id, account_id, debit_amount, credit_amount, description
    ) VALUES (
        payment_journal_id, payable_account_id, 2000000.00, 0, 'Payment to PT Supplier Tester'
    ) ON CONFLICT DO NOTHING;
    
    -- Credit: Cash (reduce cash)
    INSERT INTO unified_journal_lines (
        journal_id, account_id, debit_amount, credit_amount, description  
    ) VALUES (
        payment_journal_id, cash_account_id, 0, 2000000.00, 'Cash payment for purchase'
    ) ON CONFLICT DO NOTHING;
    
END $$;

-- 9. Update account balances for payment
UPDATE accounts SET balance = balance - 2000000.00 WHERE code = '1101'; -- Cash decrease
UPDATE accounts SET balance = balance - 2000000.00 WHERE code = '2101'; -- Payable decrease

-- 10. Verification query - check what we created
SELECT 'Sample data created successfully' as status;

SELECT 
    'Purchase Transaction' as type,
    code,
    total_amount,
    status,
    created_at
FROM purchases 
WHERE code = 'PO-TEST-001';

SELECT 
    'SSOT Journal Entries' as type,
    source_type,
    reference_number,
    entry_date,
    total_debit,
    status
FROM unified_journal_ledger 
WHERE reference_number IN ('PO-TEST-001', 'PAY-TEST-001')
ORDER BY entry_date;

SELECT 
    'Account Balances' as type,
    code,
    name,
    balance
FROM accounts 
WHERE code IN ('1301', '2101', '2102', '1101')
ORDER BY code;