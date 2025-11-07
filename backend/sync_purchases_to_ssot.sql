-- Sync Existing Purchases to SSOT Journal System
-- This will integrate the existing purchase data with SSOT journal

-- 1. First, let's check current integration status
SELECT 
    'Current Status' as check_type,
    COUNT(p.id) as total_purchases,
    COUNT(ujl.id) as purchases_with_journal,
    COUNT(p.id) - COUNT(ujl.id) as purchases_without_journal
FROM purchases p
LEFT JOIN unified_journal_ledger ujl ON (
    ujl.source_type = 'PURCHASE' AND ujl.source_id = p.id
)
WHERE p.deleted_at IS NULL;

-- 2. Get vendor information for Jerry Rolo Merentek
SELECT 
    'Vendor Check' as info,
    id,
    code,
    name,
    type
FROM contacts 
WHERE name ILIKE '%jerry%' OR name ILIKE '%merentek%'
ORDER BY name;

-- 3. Create SSOT journal entries for existing purchases
-- This will create the missing integration

-- For PO/2025/09/0036 (Rp 5,550,000)
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
)
SELECT 
    'PURCHASE',
    p.id,
    p.code,
    p.purchase_date,
    CONCAT('Purchase from ', c.name, ' - ', p.code),
    p.grand_total,
    p.grand_total,
    'POSTED',
    p.created_at,
    NOW()
FROM purchases p
LEFT JOIN contacts c ON p.vendor_id = c.id
LEFT JOIN unified_journal_ledger ujl_existing ON (
    ujl_existing.source_type = 'PURCHASE' AND ujl_existing.source_id = p.id
)
WHERE p.code = 'PO/2025/09/0036'
  AND p.deleted_at IS NULL
  AND ujl_existing.id IS NULL;

-- For PO/2025/09/0035 (Rp 3,885,000)
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
)
SELECT 
    'PURCHASE',
    p.id,
    p.code,
    p.purchase_date,
    CONCAT('Purchase from ', c.name, ' - ', p.code),
    p.grand_total,
    p.grand_total,
    'POSTED',
    p.created_at,
    NOW()
FROM purchases p
LEFT JOIN contacts c ON p.vendor_id = c.id
LEFT JOIN unified_journal_ledger ujl_existing ON (
    ujl_existing.source_type = 'PURCHASE' AND ujl_existing.source_id = p.id
)
WHERE p.code = 'PO/2025/09/0035'
  AND p.deleted_at IS NULL
  AND ujl_existing.id IS NULL;

-- 4. Generic insert for any other purchases without SSOT integration
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
)
SELECT 
    'PURCHASE',
    p.id,
    p.code,
    p.purchase_date,
    CONCAT('Purchase from ', COALESCE(c.name, 'Unknown Vendor'), ' - ', p.code),
    p.grand_total,
    p.grand_total,
    CASE WHEN p.status = 'APPROVED' THEN 'POSTED' ELSE 'DRAFT' END,
    p.created_at,
    NOW()
FROM purchases p
LEFT JOIN contacts c ON p.vendor_id = c.id
LEFT JOIN unified_journal_ledger ujl_existing ON (
    ujl_existing.source_type = 'PURCHASE' AND ujl_existing.source_id = p.id
)
WHERE p.deleted_at IS NULL
  AND ujl_existing.id IS NULL
  AND p.code NOT IN ('PO/2025/09/0036', 'PO/2025/09/0035'); -- Avoid duplicates

-- 5. Create corresponding journal lines for proper double-entry bookkeeping
-- This creates the detailed debit/credit entries

-- For each new journal entry, create the lines
DO $$
DECLARE 
    journal_rec RECORD;
    inventory_account_id INTEGER;
    payable_account_id INTEGER;
    tax_account_id INTEGER;
    purchase_amount DECIMAL(15,2);
    tax_amount DECIMAL(15,2);
    base_amount DECIMAL(15,2);
BEGIN
    -- Get account IDs
    SELECT id INTO inventory_account_id FROM accounts WHERE code = '1301' LIMIT 1; -- Inventory
    SELECT id INTO payable_account_id FROM accounts WHERE code = '2101' LIMIT 1;   -- Accounts Payable
    SELECT id INTO tax_account_id FROM accounts WHERE code = '2102' LIMIT 1;       -- VAT Input
    
    -- Process each new journal entry
    FOR journal_rec IN
        SELECT ujl.id as journal_id, p.total_amount, p.tax_amount, p.grand_total
        FROM unified_journal_ledger ujl
        JOIN purchases p ON ujl.source_id = p.id
        WHERE ujl.source_type = 'PURCHASE'
          AND ujl.created_at > NOW() - INTERVAL '1 minute' -- Only newly created entries
    LOOP
        purchase_amount := journal_rec.grand_total;
        tax_amount := COALESCE(journal_rec.tax_amount, 0);
        base_amount := journal_rec.total_amount;
        
        -- Skip if accounts not found
        IF inventory_account_id IS NULL OR payable_account_id IS NULL THEN
            RAISE NOTICE 'Skipping journal lines - accounts not found';
            CONTINUE;
        END IF;
        
        -- Insert Debit: Inventory (base amount)
        INSERT INTO unified_journal_lines (
            journal_id, account_id, debit_amount, credit_amount, description
        ) VALUES (
            journal_rec.journal_id, 
            inventory_account_id, 
            base_amount, 
            0, 
            'Inventory Purchase'
        ) ON CONFLICT DO NOTHING;
        
        -- Insert Debit: VAT Input (tax amount) if applicable
        IF tax_amount > 0 AND tax_account_id IS NOT NULL THEN
            INSERT INTO unified_journal_lines (
                journal_id, account_id, debit_amount, credit_amount, description
            ) VALUES (
                journal_rec.journal_id, 
                tax_account_id, 
                tax_amount, 
                0, 
                'VAT Input'
            ) ON CONFLICT DO NOTHING;
        END IF;
        
        -- Insert Credit: Accounts Payable (total amount)
        INSERT INTO unified_journal_lines (
            journal_id, account_id, debit_amount, credit_amount, description
        ) VALUES (
            journal_rec.journal_id, 
            payable_account_id, 
            0, 
            purchase_amount, 
            'Accounts Payable - Purchase'
        ) ON CONFLICT DO NOTHING;
        
        RAISE NOTICE 'Created journal lines for journal_id: %, amount: %', journal_rec.journal_id, purchase_amount;
    END LOOP;
END $$;

-- 6. Verification - check what we've created
SELECT 
    'SSOT Integration Verification' as result,
    COUNT(ujl.id) as journal_entries_created,
    SUM(ujl.total_debit) as total_debit,
    SUM(ujl.total_credit) as total_credit
FROM unified_journal_ledger ujl
WHERE ujl.source_type = 'PURCHASE'
  AND ujl.created_at > NOW() - INTERVAL '5 minutes';

-- 7. Check integration status after sync
SELECT 
    'Final Status' as check_type,
    COUNT(p.id) as total_purchases,
    COUNT(ujl.id) as purchases_with_journal,
    COUNT(p.id) - COUNT(ujl.id) as purchases_without_journal
FROM purchases p
LEFT JOIN unified_journal_ledger ujl ON (
    ujl.source_type = 'PURCHASE' AND ujl.source_id = p.id
)
WHERE p.deleted_at IS NULL;

-- 8. Test the Purchase Report query after integration
SELECT 
    'Purchase Report Test Query' as source,
    COUNT(*) as total_count,
    COUNT(CASE WHEN ujl.status = 'POSTED' THEN 1 END) as completed_count,
    COALESCE(SUM(ujl.total_debit), 0) as total_amount,
    COALESCE(SUM(CASE 
        WHEN ujl.description ILIKE '%paid%' OR ujl.description ILIKE '%cash%'
        THEN ujl.total_debit  
        ELSE 0           
    END), 0) as total_paid_detected
FROM unified_journal_ledger ujl
WHERE ujl.source_type = 'PURCHASE'
  AND ujl.entry_date BETWEEN '2025-09-01' AND '2025-09-30'  -- September range
  AND ujl.deleted_at IS NULL;