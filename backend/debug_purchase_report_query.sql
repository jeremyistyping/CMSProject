-- Check if PO/2025/11/0001 exists in database
SELECT 
    p.id, 
    p.code, 
    p.date, 
    p.total, 
    p.status, 
    p.approval_status, 
    c.name as vendor_name,
    p.deleted_at as purchase_deleted
FROM purchases p 
LEFT JOIN contacts c ON c.id = p.vendor_id
WHERE p.code = 'PO/2025/11/0001'
ORDER BY p.id DESC;

-- Check if this purchase has SSOT journal entry
SELECT 
    ujl.id as journal_id,
    ujl.entry_date,
    ujl.description,
    ujl.status as journal_status,
    ujl.source_type,
    ujl.source_id,
    ujl.total_debit,
    ujl.total_credit,
    ujl.deleted_at as journal_deleted
FROM unified_journal_ledger ujl
WHERE ujl.source_type = 'PURCHASE'
  AND ujl.source_id IN (SELECT id FROM purchases WHERE code = 'PO/2025/11/0001')
  AND ujl.deleted_at IS NULL;

-- Check the exact query used in getPurchaseSummary
SELECT 
    COUNT(*) as total_count,
    COUNT(CASE WHEN ujl.status = 'POSTED' THEN 1 END) as completed_count,
    COALESCE(SUM(ujl.total_debit), 0) as total_amount,
    COALESCE(SUM(CASE 
        WHEN ujl.description ILIKE '%cash%' OR ujl.description ILIKE '%kas%' OR
             EXISTS(SELECT 1 FROM unified_journal_lines ujl2
                    JOIN accounts a ON ujl2.account_id = a.id 
                    WHERE ujl2.journal_id = ujl.id 
                      AND a.code IN ('1101', '1102', '1103', '1104', '1105')
                      AND ujl2.credit_amount > 0)
        THEN ujl.total_debit
        ELSE 0
    END), 0) as total_paid
FROM unified_journal_ledger ujl
INNER JOIN purchases p ON p.id = ujl.source_id
WHERE ujl.source_type = 'PURCHASE'
  AND ujl.entry_date BETWEEN '2025-01-01' AND '2025-12-31 23:59:59'
  AND ujl.deleted_at IS NULL
  AND p.deleted_at IS NULL
  AND (p.status = 'APPROVED' OR p.status = 'COMPLETED' OR p.approval_status = 'APPROVED');

-- Check vendor grouping query
SELECT 
    COALESCE(sje.source_id, 0) as vendor_id,
    CASE 
        WHEN sje.source_id IS NOT NULL AND sje.source_id > 0 
        THEN COALESCE((
            SELECT CASE
                WHEN description ~ 'Purchase from (.+) - ' 
                    THEN TRIM(SUBSTRING(description FROM 'Purchase from (.+) - '))
                WHEN description ~ 'Purchase Order [^-]+ - (.+)'
                    THEN TRIM(SUBSTRING(description FROM 'Purchase Order [^-]+ - (.+)'))
                WHEN description ~ '- (.+)$'
                    THEN TRIM(SUBSTRING(description FROM '- (.+)$'))
                ELSE NULL
            END
            FROM unified_journal_ledger
            WHERE source_id = sje.source_id 
              AND source_type = 'PURCHASE'
            LIMIT 1
        ), 'Vendor ID: ' || sje.source_id::text)
        ELSE 'Unknown Vendor'
    END as vendor_name,
    COUNT(*) as total_purchases,
    COALESCE(SUM(sje.total_debit), 0) as total_amount,
    string_agg(DISTINCT sje.description, ' | ') as descriptions
FROM unified_journal_ledger sje
INNER JOIN purchases p ON p.id = sje.source_id
WHERE sje.source_type = 'PURCHASE'
  AND sje.entry_date BETWEEN '2025-01-01' AND '2025-12-31 23:59:59'
  AND sje.deleted_at IS NULL
  AND p.deleted_at IS NULL
  AND (p.status = 'APPROVED' OR p.status = 'COMPLETED' OR p.approval_status = 'APPROVED')
GROUP BY sje.source_id
ORDER BY total_amount DESC;
