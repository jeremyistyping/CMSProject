-- =====================================================
-- INVESTIGATION: Revenue Duplication in P&L Report
-- Expected: Rp 10,000,000 (from Chart of Accounts)
-- Actual: Rp 20,000,000 (from P&L Report)
-- Difference: Rp 10,000,000 (100% duplication!)
-- =====================================================

-- ==========================================
-- STEP 1: Check Account Balance
-- ==========================================
SELECT 
    'Account Balance' as Source,
    code,
    name,
    type,
    balance,
    is_header
FROM accounts 
WHERE code LIKE '4%'
ORDER BY code;

-- ==========================================
-- STEP 2: Check Unified Journal Lines (SSOT)
-- ==========================================
SELECT 
    'Unified Journal Analysis' as Analysis,
    a.code as account_code,
    a.name as account_name_from_accounts,
    SUM(ujl.credit_amount) as total_credit,
    SUM(ujl.debit_amount) as total_debit,
    SUM(ujl.credit_amount - ujl.debit_amount) as net_amount,
    COUNT(*) as line_count,
    COUNT(DISTINCT ujl.journal_id) as journal_count
FROM accounts a
LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id 
    AND uje.status = 'POSTED' 
    AND uje.deleted_at IS NULL
WHERE a.code LIKE '4%'
    AND uje.entry_date >= '2025-01-01' 
    AND uje.entry_date <= '2025-12-31'
    AND COALESCE(a.is_header, false) = false
GROUP BY a.id, a.code, a.name
ORDER BY a.code;

-- ==========================================
-- STEP 3: Detail Journal Entries for 4101
-- ==========================================
SELECT 
    'Journal Detail for 4101' as Detail,
    uje.id as journal_id,
    uje.entry_date,
    uje.description,
    uje.source_type,
    uje.source_id,
    uje.reference_number,
    uje.status,
    ujl.account_id,
    a.code as account_code,
    a.name as account_name,
    ujl.debit_amount,
    ujl.credit_amount,
    ujl.created_at
FROM unified_journal_ledger uje
INNER JOIN unified_journal_lines ujl ON uje.id = ujl.journal_id
INNER JOIN accounts a ON ujl.account_id = a.id
WHERE a.code = '4101'
    AND uje.entry_date >= '2025-01-01' 
    AND uje.entry_date <= '2025-12-31'
    AND uje.status = 'POSTED'
    AND uje.deleted_at IS NULL
ORDER BY uje.entry_date, uje.id, ujl.id;

-- ==========================================
-- STEP 4: Check for Duplicate Journals
-- ==========================================
SELECT 
    'Potential Duplicates' as Check_Type,
    uje.source_type,
    uje.source_id,
    uje.reference_number,
    COUNT(*) as journal_count,
    GROUP_CONCAT(uje.id) as journal_ids,
    SUM(CASE WHEN ujl.account_code = '4101' THEN ujl.credit_amount ELSE 0 END) as total_revenue_4101
FROM unified_journal_ledger uje
INNER JOIN unified_journal_lines ujl ON uje.id = ujl.journal_id
WHERE uje.source_type IN ('SALE', 'PAYMENT', 'MANUAL')
    AND uje.entry_date >= '2025-01-01' 
    AND uje.entry_date <= '2025-12-31'
    AND uje.status = 'POSTED'
    AND uje.deleted_at IS NULL
GROUP BY uje.source_type, uje.source_id, uje.reference_number
HAVING COUNT(*) > 1
ORDER BY journal_count DESC;

-- ==========================================
-- STEP 5: Legacy Journal System Check
-- ==========================================
SELECT 
    'Legacy Journal Analysis' as Analysis,
    a.code as account_code,
    a.name as account_name,
    SUM(jl.credit_amount) as total_credit,
    SUM(jl.debit_amount) as total_debit,
    SUM(jl.credit_amount - jl.debit_amount) as net_amount,
    COUNT(*) as line_count
FROM accounts a
LEFT JOIN journal_lines jl ON jl.account_id = a.id
LEFT JOIN journal_entries je ON je.id = jl.journal_entry_id 
    AND je.status = 'POSTED' 
    AND je.deleted_at IS NULL
WHERE a.code LIKE '4%'
    AND je.entry_date >= '2025-01-01' 
    AND je.entry_date <= '2025-12-31'
GROUP BY a.id, a.code, a.name
ORDER BY a.code;

-- ==========================================
-- STEP 6: Check Both Systems Together
-- ==========================================
SELECT 
    'Combined Systems' as Source,
    'Unified Journal (SSOT)' as System,
    SUM(ujl.credit_amount - ujl.debit_amount) as total_revenue
FROM unified_journal_lines ujl
INNER JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
INNER JOIN accounts a ON ujl.account_id = a.id
WHERE a.code LIKE '4%'
    AND uje.entry_date >= '2025-01-01' 
    AND uje.entry_date <= '2025-12-31'
    AND uje.status = 'POSTED'
    AND uje.deleted_at IS NULL

UNION ALL

SELECT 
    'Combined Systems' as Source,
    'Legacy Journal' as System,
    SUM(jl.credit_amount - jl.debit_amount) as total_revenue
FROM journal_lines jl
INNER JOIN journal_entries je ON je.id = jl.journal_entry_id
INNER JOIN accounts a ON jl.account_id = a.id
WHERE a.code LIKE '4%'
    AND je.entry_date >= '2025-01-01' 
    AND je.entry_date <= '2025-12-31'
    AND je.status = 'POSTED'
    AND je.deleted_at IS NULL;

-- ==========================================
-- STEP 7: Find Account Name Variations
-- ==========================================
SELECT 
    'Account Name Check' as Check_Type,
    ujl.account_code,
    ujl.account_name as stored_in_journal_line,
    a.name as stored_in_accounts_table,
    CASE 
        WHEN ujl.account_name = a.name THEN 'MATCH'
        ELSE 'MISMATCH'
    END as comparison,
    COUNT(*) as occurrence_count,
    SUM(ujl.credit_amount - ujl.debit_amount) as total_amount
FROM unified_journal_lines ujl
INNER JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
INNER JOIN accounts a ON ujl.account_id = a.id
WHERE ujl.account_code LIKE '4%'
    AND uje.entry_date >= '2025-01-01' 
    AND uje.entry_date <= '2025-12-31'
    AND uje.status = 'POSTED'
GROUP BY ujl.account_code, ujl.account_name, a.name
ORDER BY ujl.account_code, occurrence_count DESC;

-- ==========================================
-- STEP 8: Sales Transactions Check
-- ==========================================
SELECT 
    'Sales Transaction Analysis' as Analysis,
    s.id as sale_id,
    s.invoice_number,
    s.transaction_date,
    s.total_amount,
    s.status,
    COUNT(DISTINCT uje.id) as journal_entry_count,
    SUM(CASE WHEN a.code LIKE '4%' THEN ujl.credit_amount ELSE 0 END) as revenue_recorded
FROM sales s
LEFT JOIN unified_journal_ledger uje ON uje.source_type = 'SALE' 
    AND uje.source_id = s.id
    AND uje.status = 'POSTED'
    AND uje.deleted_at IS NULL
LEFT JOIN unified_journal_lines ujl ON ujl.journal_id = uje.id
LEFT JOIN accounts a ON ujl.account_id = a.id
WHERE s.transaction_date >= '2025-01-01' 
    AND s.transaction_date <= '2025-12-31'
GROUP BY s.id, s.invoice_number, s.transaction_date, s.total_amount, s.status
HAVING journal_entry_count > 0
ORDER BY s.transaction_date, s.id;

-- ==========================================
-- STEP 9: Account ID Mapping Check
-- ==========================================
SELECT 
    'Account ID Mapping' as Check_Type,
    account_id,
    account_code,
    account_name,
    COUNT(*) as line_count,
    SUM(credit_amount - debit_amount) as total_amount
FROM unified_journal_lines
WHERE account_code LIKE '4%'
GROUP BY account_id, account_code, account_name
ORDER BY account_code, account_name;

-- ==========================================
-- STEP 10: Summary & Root Cause Analysis
-- ==========================================
SELECT 
    '=== SUMMARY ===' as Section,
    'Expected from Chart of Accounts' as Description,
    'Rp 10,000,000' as Amount
UNION ALL
SELECT 
    '=== SUMMARY ===' as Section,
    'Actual from P&L Report' as Description,
    'Rp 20,000,000' as Amount
UNION ALL
SELECT 
    '=== SUMMARY ===' as Section,
    'Discrepancy' as Description,
    'Rp 10,000,000 (100% duplication!)' as Amount;

-- ==========================================
-- EXPECTED FINDINGS:
-- ==========================================
-- 1. If unified_journal_lines shows 20M total → duplication at journal entry level
-- 2. If account 4101 appears multiple times with different names → name variation issue
-- 3. If journal_entry_count > 1 for same sale → multiple journal entries per sale
-- 4. If both SSOT and Legacy show amounts → both systems being counted
-- 5. If account_id mapping shows duplicates → account linking issue

-- ==========================================
-- REMEDIATION STEPS (based on findings):
-- ==========================================
-- Option 1: If duplicate journal entries exist
--   → DELETE or UPDATE status to 'CANCELLED' for duplicates
-- Option 2: If both SSOT and Legacy are counted
--   → Modify backend to use only one source
-- Option 3: If name variations cause duplicates
--   → Standardize account_name in journal_lines
-- Option 4: If query groups incorrectly
--   → Fix GROUP BY clause in backend service

