-- Check purchases data
SELECT 'Purchases Count' as check_type, COUNT(*) as count
FROM purchases WHERE deleted_at IS NULL;

-- Check September 2025 purchases specifically
SELECT 
    'September 2025 Purchases' as check_type,
    COUNT(*) as count,
    COALESCE(SUM(total_amount), 0) as total_amount
FROM purchases 
WHERE purchase_date BETWEEN '2025-09-01' AND '2025-09-30' 
AND deleted_at IS NULL;

-- Check SSOT journal entries for purchases
SELECT 'SSOT Purchase Entries' as check_type, COUNT(*) as count
FROM unified_journal_ledger 
WHERE source_type = 'PURCHASE' AND deleted_at IS NULL;

-- Check specific September purchases with details
SELECT 
    purchase_code,
    purchase_date,
    total_amount,
    status,
    vendor_name
FROM purchases p
LEFT JOIN vendors v ON p.vendor_id = v.id
WHERE purchase_date BETWEEN '2025-09-01' AND '2025-09-30'
AND p.deleted_at IS NULL
ORDER BY purchase_date DESC;
