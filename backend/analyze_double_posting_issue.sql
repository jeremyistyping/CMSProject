-- ================================================
-- ANALISIS MASALAH DOUBLE POSTING LABA DITAHAN
-- ================================================
-- Masalah: Pada periode closing kedua, laba ditahan di-posting ulang
-- menyebabkan balance sheet tidak balance (diff Rp 35.000.000)

-- 1. CEK ACCOUNTING PERIODS (Periode Closing)
SELECT 
    id,
    start_date,
    end_date,
    description,
    is_closed,
    total_revenue,
    total_expense,
    net_income,
    closing_journal_id,
    closed_at,
    created_at
FROM accounting_periods
ORDER BY end_date DESC;

-- 2. CEK AKUN LABA DITAHAN (3201)
SELECT 
    id,
    code,
    name,
    type,
    balance,
    parent_id,
    is_active,
    is_header
FROM accounts
WHERE code = '3201' OR name ILIKE '%laba%ditahan%';

-- 3. CEK JOURNAL ENTRIES CLOSING
SELECT 
    je.id,
    je.code,
    je.description,
    je.reference,
    je.reference_type,
    je.entry_date,
    je.total_debit,
    je.total_credit,
    je.status,
    je.is_balanced,
    je.created_at
FROM journal_entries je
WHERE reference_type = 'CLOSING'
ORDER BY entry_date DESC;

-- 4. CEK JOURNAL LINES untuk CLOSING ENTRIES
SELECT 
    jl.journal_entry_id,
    je.code AS journal_code,
    je.entry_date,
    a.code AS account_code,
    a.name AS account_name,
    a.type AS account_type,
    jl.description,
    jl.debit_amount,
    jl.credit_amount,
    jl.line_number
FROM journal_lines jl
JOIN journal_entries je ON jl.journal_entry_id = je.id
JOIN accounts a ON jl.account_id = a.id
WHERE je.reference_type = 'CLOSING'
ORDER BY je.entry_date DESC, jl.line_number;

-- 5. CEK BALANCE AKUN REVENUE DAN EXPENSE SAAT INI
-- (Harusnya 0 setelah closing)
SELECT 
    code,
    name,
    type,
    balance,
    is_active
FROM accounts
WHERE type IN ('REVENUE', 'EXPENSE')
    AND is_active = true
    AND is_header = false
    AND balance != 0
ORDER BY type, code;

-- 6. HITUNG TOTAL POSTING KE LABA DITAHAN
SELECT 
    a.code,
    a.name,
    SUM(jl.credit_amount - jl.debit_amount) AS net_change_to_retained_earnings,
    COUNT(*) AS transaction_count
FROM journal_lines jl
JOIN accounts a ON jl.account_id = a.id
JOIN journal_entries je ON jl.journal_entry_id = je.id
WHERE a.code = '3201'
    AND je.reference_type = 'CLOSING'
    AND je.status = 'POSTED'
GROUP BY a.code, a.name;

-- 7. DETAIL SETIAP POSTING KE LABA DITAHAN DARI CLOSING
SELECT 
    je.id AS journal_id,
    je.code AS journal_code,
    je.entry_date,
    je.description,
    jl.description AS line_description,
    jl.debit_amount,
    jl.credit_amount,
    (jl.credit_amount - jl.debit_amount) AS net_effect
FROM journal_lines jl
JOIN journal_entries je ON jl.journal_entry_id = je.id
JOIN accounts a ON jl.account_id = a.id
WHERE a.code = '3201'
    AND je.reference_type = 'CLOSING'
    AND je.status = 'POSTED'
ORDER BY je.entry_date, jl.line_number;

-- 8. CEK APAKAH ADA DUPLICATE CLOSING UNTUK PERIODE YANG SAMA
SELECT 
    entry_date,
    COUNT(*) AS closing_count,
    STRING_AGG(code, ', ') AS journal_codes,
    SUM(total_debit) AS total_debit_sum
FROM journal_entries
WHERE reference_type = 'CLOSING'
GROUP BY entry_date
HAVING COUNT(*) > 1;

-- 9. BALANCE SHEET CHECK - Total Assets vs Total Liabilities + Equity
WITH balance_totals AS (
    SELECT 
        SUM(CASE WHEN type = 'ASSET' THEN balance ELSE 0 END) AS total_assets,
        SUM(CASE WHEN type = 'LIABILITY' THEN balance ELSE 0 END) AS total_liabilities,
        SUM(CASE WHEN type = 'EQUITY' THEN balance ELSE 0 END) AS total_equity,
        SUM(CASE WHEN type = 'REVENUE' THEN balance ELSE 0 END) AS total_revenue,
        SUM(CASE WHEN type = 'EXPENSE' THEN balance ELSE 0 END) AS total_expense
    FROM accounts
    WHERE is_active = true AND is_header = false
)
SELECT 
    total_assets,
    total_liabilities,
    total_equity,
    (total_liabilities + total_equity) AS total_liab_equity,
    (total_assets - (total_liabilities + total_equity)) AS difference,
    total_revenue,
    total_expense,
    (total_revenue - total_expense) AS net_income
FROM balance_totals;

-- 10. TEMUKAN MASALAH: Revenue/Expense yang masih punya balance setelah closing
-- (Ini indikasi bahwa closing tidak lengkap atau ada transaksi setelah closing)
SELECT 
    'Revenue with balance after closing' AS issue,
    code,
    name,
    balance,
    (SELECT MAX(end_date) FROM accounting_periods WHERE is_closed = true) AS last_closing_date
FROM accounts
WHERE type = 'REVENUE'
    AND is_active = true
    AND is_header = false
    AND balance != 0
UNION ALL
SELECT 
    'Expense with balance after closing' AS issue,
    code,
    name,
    balance,
    (SELECT MAX(end_date) FROM accounting_periods WHERE is_closed = true) AS last_closing_date
FROM accounts
WHERE type = 'EXPENSE'
    AND is_active = true
    AND is_header = false
    AND balance != 0;
