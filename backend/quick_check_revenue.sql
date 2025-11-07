-- Quick check for revenue account 4101
SELECT 
    'Account 4101 Summary' as check_type,
    a.code,
    a.name,
    COUNT(ujl.id) as total_lines,
    COUNT(DISTINCT ujl.journal_id) as unique_journals,
    SUM(ujl.credit_amount) as total_credit,
    SUM(ujl.debit_amount) as total_debit,
    SUM(ujl.credit_amount) - SUM(ujl.debit_amount) as net_revenue
FROM accounts a
LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id 
    AND uje.status = 'POSTED' 
    AND uje.deleted_at IS NULL
    AND uje.entry_date >= '2025-01-01' 
    AND uje.entry_date <= '2025-12-31'
WHERE a.code = '4101'
GROUP BY a.code, a.name;

-- Check if any sales has multiple journal entries
SELECT 
    'Sales with Multiple Journals' as check_type,
    s.id,
    s.invoice_number,
    s.total_amount,
    COUNT(DISTINCT uje.id) as journal_count,
    STRING_AGG(uje.journal_number, ', ' ORDER BY uje.id) as journal_numbers
FROM sales s
INNER JOIN unified_journal_ledger uje ON uje.source_type = 'SALES' 
    AND uje.source_id = s.id 
    AND uje.status = 'POSTED'
    AND uje.deleted_at IS NULL
WHERE s.created_at >= '2025-01-01' 
GROUP BY s.id, s.invoice_number, s.total_amount
HAVING COUNT(DISTINCT uje.id) > 1;

-- Check journal entries for account 4101 in detail
SELECT 
    'Journal Details for 4101' as check_type,
    uje.id,
    uje.journal_number,
    uje.entry_date,
    uje.source_type,
    uje.source_id,
    ujl.credit_amount,
    ujl.description
FROM unified_journal_ledger uje
INNER JOIN unified_journal_lines ujl ON ujl.journal_id = uje.id
INNER JOIN accounts a ON a.id = ujl.account_id
WHERE a.code = '4101'
  AND uje.entry_date >= '2025-01-01' 
  AND uje.entry_date <= '2025-12-31'
  AND uje.status = 'POSTED'
  AND uje.deleted_at IS NULL
ORDER BY uje.entry_date, uje.id;

