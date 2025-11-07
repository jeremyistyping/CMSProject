-- Profit & Loss Database Validation Queries
-- Generated: 2025-10-17 06:54:14

-- ==========================================
-- Revenue Accounts Balance
-- ==========================================
SELECT code, name, balance FROM accounts WHERE code LIKE '4%' ORDER BY code;

-- ==========================================
-- Journal Entries for Revenue (Credit - Debit)
-- ==========================================
SELECT 
    je.account_code,
    a.name as account_name,
    SUM(je.credit) as total_credit,
    SUM(je.debit) as total_debit,
    SUM(je.credit - je.debit) as net_amount,
    COUNT(*) as entry_count
FROM journal_entries je
LEFT JOIN accounts a ON je.account_code = a.code
WHERE je.account_code LIKE '4%'
AND je.status = 'POSTED'
GROUP BY je.account_code, a.name
ORDER BY je.account_code;

-- ==========================================
-- All Journal Entries for Account 4101
-- ==========================================
SELECT 
    j.id as journal_id,
    j.date,
    j.description,
    j.source_type,
    j.source_id,
    je.account_code,
    je.account_name,
    je.debit,
    je.credit,
    j.status
FROM journals j
INNER JOIN journal_entries je ON j.id = je.journal_id
WHERE je.account_code = '4101'
AND j.date BETWEEN '2025-01-01' AND '2025-12-31'
ORDER BY j.date, j.id;

-- ==========================================
-- Check for Duplicate Account Codes
-- ==========================================
SELECT 
    je.account_code,
    je.account_name,
    COUNT(DISTINCT je.account_name) as name_variations,
    GROUP_CONCAT(DISTINCT je.account_name) as all_names,
    SUM(je.credit - je.debit) as total_amount
FROM journal_entries je
WHERE je.account_code LIKE '4%'
GROUP BY je.account_code
HAVING COUNT(DISTINCT je.account_name) > 1
ORDER BY je.account_code;

-- ==========================================
-- P&L Service Query Simulation
-- ==========================================
SELECT 
    je.account_code,
    je.account_name,
    SUM(je.credit - je.debit) as amount
FROM journal_entries je
INNER JOIN journals j ON je.journal_id = j.id
WHERE je.account_code LIKE '4%'
AND j.status = 'POSTED'
AND j.date BETWEEN '2025-01-01' AND '2025-12-31'
GROUP BY je.account_code, je.account_name
ORDER BY je.account_code, je.account_name;

