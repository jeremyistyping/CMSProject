-- Check SSOT Integration for Existing Purchase Transactions
-- This will help us understand why Purchase Report doesn't show the data

-- 1. Check existing purchase transactions
SELECT 
    'Existing Purchases' as source,
    code,
    vendor_id,
    purchase_date,
    total_amount,
    grand_total,
    status,
    created_at
FROM purchases 
WHERE deleted_at IS NULL
ORDER BY purchase_date DESC
LIMIT 10;

-- 2. Check if these purchases have corresponding SSOT journal entries
SELECT 
    'SSOT Journal Entries' as source,
    p.code as purchase_code,
    p.purchase_date,
    p.total_amount as purchase_amount,
    ujl.id as journal_id,
    ujl.source_type,
    ujl.reference_number,
    ujl.entry_date,
    ujl.total_debit,
    ujl.total_credit,
    ujl.status as journal_status
FROM purchases p
LEFT JOIN unified_journal_ledger ujl ON (
    ujl.source_type = 'PURCHASE' AND 
    (ujl.source_id = p.id OR ujl.reference_number = p.code)
)
WHERE p.deleted_at IS NULL
ORDER BY p.purchase_date DESC;

-- 3. Check SSOT journal entries for September 2025
SELECT 
    'September 2025 SSOT Entries' as source,
    source_type,
    reference_number,
    entry_date,
    description,
    total_debit,
    total_credit,
    status
FROM unified_journal_ledger
WHERE entry_date BETWEEN '2025-09-01' AND '2025-09-30'
  AND deleted_at IS NULL
ORDER BY entry_date DESC;

-- 4. Check if there are any PURCHASE type entries in SSOT
SELECT 
    'PURCHASE Type SSOT' as source,
    COUNT(*) as total_count,
    MIN(entry_date) as earliest_date,
    MAX(entry_date) as latest_date,
    SUM(total_debit) as total_debit_sum,
    SUM(total_credit) as total_credit_sum
FROM unified_journal_ledger
WHERE source_type = 'PURCHASE'
  AND deleted_at IS NULL;

-- 5. Check vendor information for the purchases
SELECT 
    'Vendor Information' as source,
    c.code as vendor_code,
    c.name as vendor_name,
    c.type,
    c.is_active,
    COUNT(p.id) as purchase_count,
    SUM(p.total_amount) as total_purchase_value
FROM purchases p
JOIN contacts c ON p.vendor_id = c.id
WHERE p.deleted_at IS NULL
  AND c.deleted_at IS NULL
GROUP BY c.code, c.name, c.type, c.is_active
ORDER BY total_purchase_value DESC;

-- 6. Check account balances that might relate to outstanding payables
SELECT 
    'Account Balances' as source,
    code,
    name,
    balance,
    account_type,
    last_updated
FROM accounts 
WHERE (code LIKE '2%' OR code IN ('2101', '2102', '1301'))
  AND balance <> 0
ORDER BY ABS(balance) DESC;

-- 7. Check if purchases need to be integrated to SSOT journal
SELECT 
    'Integration Status' as check_type,
    'Purchases without SSOT journal' as description,
    COUNT(*) as count
FROM purchases p
LEFT JOIN unified_journal_ledger ujl ON (
    ujl.source_type = 'PURCHASE' AND ujl.source_id = p.id
)
WHERE p.deleted_at IS NULL
  AND ujl.id IS NULL;

-- 8. Sample query that Purchase Report service might be using
-- This simulates what the Go service query looks like
SELECT 
    'Service Query Simulation' as source,
    COUNT(*) as total_count,
    COUNT(CASE WHEN ujl.status = 'POSTED' THEN 1 END) as completed_count,
    COALESCE(SUM(ujl.total_debit), 0) as total_amount,
    -- This is likely where the issue is - payment calculation
    COALESCE(SUM(CASE 
        WHEN ujl.description ILIKE '%cash%' OR ujl.description ILIKE '%kas%'
        THEN ujl.total_debit  
        ELSE 0           
    END), 0) as total_paid
FROM unified_journal_ledger ujl
WHERE ujl.source_type = 'PURCHASE'
  AND ujl.entry_date BETWEEN '2025-01-01' AND '2025-12-31'
  AND ujl.deleted_at IS NULL;