-- ==========================================
-- INVESTIGASI DUPLIKASI REVENUE DETAIL
-- ==========================================

-- 1. CEK TOTAL CREDIT DAN DEBIT UNTUK ACCOUNT 4101
SELECT 
    a.id as account_id,
    a.code as account_code,
    a.name as account_name,
    a.type as account_type,
    COUNT(ujl.id) as line_count,
    COUNT(DISTINCT ujl.journal_id) as unique_journal_count,
    COALESCE(SUM(ujl.debit_amount), 0) as total_debit,
    COALESCE(SUM(ujl.credit_amount), 0) as total_credit,
    COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0) as net_balance
FROM accounts a
LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id 
    AND uje.status = 'POSTED' 
    AND uje.deleted_at IS NULL
WHERE a.code = '4101'
  AND uje.entry_date >= '2025-01-01' 
  AND uje.entry_date <= '2025-12-31'
GROUP BY a.id, a.code, a.name, a.type;

-- 2. CEK DETAIL JOURNAL ENTRIES UNTUK ACCOUNT 4101
SELECT 
    uje.id as journal_id,
    uje.journal_number,
    uje.entry_date,
    uje.description,
    uje.source_type,
    uje.source_id,
    uje.status,
    ujl.id as line_id,
    ujl.account_id,
    a.code as account_code,
    a.name as account_name,
    ujl.debit_amount,
    ujl.credit_amount,
    ujl.description as line_description
FROM unified_journal_ledger uje
INNER JOIN unified_journal_lines ujl ON ujl.journal_id = uje.id
INNER JOIN accounts a ON a.id = ujl.account_id
WHERE a.code = '4101'
  AND uje.entry_date >= '2025-01-01' 
  AND uje.entry_date <= '2025-12-31'
  AND uje.status = 'POSTED'
  AND uje.deleted_at IS NULL
ORDER BY uje.entry_date, uje.id, ujl.id;

-- 3. CEK APA ADA JOURNAL ENTRY DUPLICATE (SAMA PERSIS)
SELECT 
    uje.source_type,
    uje.source_id,
    uje.entry_date,
    COUNT(*) as duplicate_count,
    STRING_AGG(uje.id::text, ', ') as journal_ids,
    STRING_AGG(uje.journal_number, ', ') as journal_numbers,
    SUM(ujl.credit_amount) as total_credit
FROM unified_journal_ledger uje
INNER JOIN unified_journal_lines ujl ON ujl.journal_id = uje.id
INNER JOIN accounts a ON a.id = ujl.account_id
WHERE a.code = '4101'
  AND uje.entry_date >= '2025-01-01' 
  AND uje.entry_date <= '2025-12-31'
  AND uje.status = 'POSTED'
  AND uje.deleted_at IS NULL
GROUP BY uje.source_type, uje.source_id, uje.entry_date
HAVING COUNT(*) > 1
ORDER BY duplicate_count DESC;

-- 4. CEK APA ADA MULTIPLE LINES UNTUK ACCOUNT 4101 DALAM SATU JOURNAL
SELECT 
    uje.id as journal_id,
    uje.journal_number,
    uje.entry_date,
    uje.source_type,
    uje.source_id,
    COUNT(ujl.id) as lines_count_for_4101,
    SUM(ujl.credit_amount) as total_credit,
    SUM(ujl.debit_amount) as total_debit
FROM unified_journal_ledger uje
INNER JOIN unified_journal_lines ujl ON ujl.journal_id = uje.id
INNER JOIN accounts a ON a.id = ujl.account_id
WHERE a.code = '4101'
  AND uje.entry_date >= '2025-01-01' 
  AND uje.entry_date <= '2025-12-31'
  AND uje.status = 'POSTED'
  AND uje.deleted_at IS NULL
GROUP BY uje.id, uje.journal_number, uje.entry_date, uje.source_type, uje.source_id
HAVING COUNT(ujl.id) > 1
ORDER BY lines_count_for_4101 DESC;

-- 5. CEK SEMUA REVENUE ACCOUNTS (4xxx)
SELECT 
    a.id as account_id,
    a.code as account_code,
    a.name as account_name,
    COUNT(ujl.id) as line_count,
    COUNT(DISTINCT ujl.journal_id) as unique_journal_count,
    COALESCE(SUM(ujl.credit_amount), 0) as total_credit,
    COALESCE(SUM(ujl.debit_amount), 0) as total_debit,
    COALESCE(SUM(ujl.credit_amount), 0) - COALESCE(SUM(ujl.debit_amount), 0) as net_revenue
FROM accounts a
LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id 
    AND uje.status = 'POSTED' 
    AND uje.deleted_at IS NULL
WHERE (a.code LIKE '4%' OR UPPER(a.type) = 'REVENUE')
  AND uje.entry_date >= '2025-01-01' 
  AND uje.entry_date <= '2025-12-31'
  AND COALESCE(a.is_header, false) = false
GROUP BY a.id, a.code, a.name
HAVING COALESCE(SUM(ujl.credit_amount), 0) > 0 OR COALESCE(SUM(ujl.debit_amount), 0) > 0
ORDER BY a.code;

-- 6. CEK APA ADA DUPLIKASI DARI SOURCE SALES
SELECT 
    s.id as sales_id,
    s.invoice_number,
    s.total_amount,
    COUNT(DISTINCT uje.id) as journal_count,
    STRING_AGG(DISTINCT uje.journal_number, ', ') as journal_numbers,
    SUM(ujl.credit_amount) as total_revenue_posted
FROM sales s
LEFT JOIN unified_journal_ledger uje ON uje.source_type = 'SALES' 
    AND uje.source_id = s.id 
    AND uje.status = 'POSTED'
    AND uje.deleted_at IS NULL
LEFT JOIN unified_journal_lines ujl ON ujl.journal_id = uje.id
LEFT JOIN accounts a ON a.id = ujl.account_id AND a.code = '4101'
WHERE s.created_at >= '2025-01-01' 
  AND s.created_at <= '2025-12-31'
GROUP BY s.id, s.invoice_number, s.total_amount
HAVING COUNT(DISTINCT uje.id) > 1
ORDER BY journal_count DESC;

