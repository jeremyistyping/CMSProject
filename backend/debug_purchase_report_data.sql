-- Debug Purchase Report Data Issues
-- Check SSOT journal data for purchases

-- 1. Check if unified_journal_ledger table exists and has data
SELECT 'unified_journal_ledger' as table_name, COUNT(*) as total_records
FROM unified_journal_ledger
WHERE deleted_at IS NULL;

-- 2. Check purchase-related entries in SSOT journal
SELECT 
    source_type,
    COUNT(*) as count,
    MIN(entry_date) as earliest_date,
    MAX(entry_date) as latest_date,
    SUM(total_debit) as total_debit_amount,
    SUM(total_credit) as total_credit_amount
FROM unified_journal_ledger 
WHERE deleted_at IS NULL
GROUP BY source_type
ORDER BY count DESC;

-- 3. Check specific purchase transactions
SELECT 
    id,
    source_type,
    source_id,
    entry_date,
    description,
    total_debit,
    total_credit,
    reference_number
FROM unified_journal_ledger 
WHERE source_type = 'PURCHASE' 
  AND deleted_at IS NULL
  AND entry_date >= '2025-01-01'
  AND entry_date <= '2025-12-31'
LIMIT 10;

-- 4. Check if there are any purchase entries in the date range used in the modal
SELECT 
    COUNT(*) as purchase_count_in_range,
    SUM(total_debit) as total_debit,
    SUM(total_credit) as total_credit
FROM unified_journal_ledger 
WHERE source_type = 'PURCHASE'
  AND entry_date BETWEEN '2025-01-01' AND '2025-12-31'
  AND deleted_at IS NULL;

-- 5. Check alternative data - regular purchases table
SELECT 'purchases table' as source, COUNT(*) as total_records
FROM purchases
WHERE deleted_at IS NULL;

-- 6. Check if there are purchases in different date ranges
SELECT 
    DATE_TRUNC('month', purchase_date) as month,
    COUNT(*) as purchase_count,
    SUM(total_amount) as total_amount
FROM purchases 
WHERE deleted_at IS NULL
GROUP BY DATE_TRUNC('month', purchase_date)
ORDER BY month DESC
LIMIT 12;

-- 7. Check journal entries that might be purchases (by account codes)
SELECT 
    ujl.source_type,
    ujl.description,
    COUNT(*) as count,
    SUM(ujl.total_debit) as total_amount
FROM unified_journal_ledger ujl
LEFT JOIN unified_journal_lines ujll ON ujl.id = ujll.journal_id
LEFT JOIN accounts a ON ujll.account_id = a.id
WHERE ujl.deleted_at IS NULL
  AND (
    a.code LIKE '2%' OR  -- Payables
    a.code LIKE '5%' OR  -- Expenses
    a.code LIKE '6%'     -- Other expenses
  )
  AND ujl.entry_date >= '2025-01-01'
GROUP BY ujl.source_type, ujl.description
ORDER BY total_amount DESC
LIMIT 10;

-- 8. Check the Outstanding Payables source (Rp 5.550.000)
SELECT 
    account_code,
    account_name,
    balance,
    last_updated
FROM accounts 
WHERE balance <> 0
  AND (account_code LIKE '2%' OR account_code IN ('2101', '2102'))
ORDER BY ABS(balance) DESC;