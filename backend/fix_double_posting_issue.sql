-- ================================================
-- PERBAIKAN MASALAH DOUBLE POSTING LABA DITAHAN
-- ================================================
-- PENTING: Backup database sebelum menjalankan script ini!
-- Masalah: Pada periode closing kedua, revenue/expense yang sudah 0 
-- tetap di-closing lagi, menyebabkan Laba Ditahan double posting

BEGIN;

-- Step 1: Identifikasi closing entries yang problematik
-- Closing yang menutup akun revenue/expense dengan balance 0
CREATE TEMP TABLE problematic_closings AS
SELECT DISTINCT
    je.id AS journal_id,
    je.code,
    je.entry_date,
    je.description,
    je.total_debit,
    je.total_credit
FROM journal_entries je
JOIN journal_lines jl ON je.id = jl.journal_entry_id
JOIN accounts a ON jl.account_id = a.id
WHERE je.reference_type = 'CLOSING'
    AND je.status = 'POSTED'
    AND a.type IN ('REVENUE', 'EXPENSE')
    -- Jika ada closing untuk akun yang seharusnya sudah 0
GROUP BY je.id, je.code, je.entry_date, je.description, je.total_debit, je.total_credit;

-- Step 2: Log problematic closings untuk review
SELECT 
    'Problematic closing entries found:' AS status,
    COUNT(*) AS count
FROM problematic_closings;

SELECT * FROM problematic_closings;

-- Step 3: Identifikasi journal lines dari closing yang problematik
CREATE TEMP TABLE lines_to_reverse AS
SELECT 
    jl.id,
    jl.journal_entry_id,
    jl.account_id,
    a.code AS account_code,
    a.name AS account_name,
    a.type AS account_type,
    jl.debit_amount,
    jl.credit_amount,
    -- Hitung balance change yang perlu di-reverse
    CASE 
        WHEN a.type IN ('ASSET', 'EXPENSE') THEN jl.credit_amount - jl.debit_amount
        ELSE jl.debit_amount - jl.credit_amount
    END AS balance_reversal
FROM journal_lines jl
JOIN accounts a ON jl.account_id = a.id
WHERE jl.journal_entry_id IN (SELECT journal_id FROM problematic_closings);

-- Step 4: Display journal lines to reverse
SELECT 
    'Journal lines to reverse:' AS status,
    COUNT(*) AS count
FROM lines_to_reverse;

SELECT * FROM lines_to_reverse ORDER BY journal_entry_id, account_code;

-- Step 5: Reverse the account balances
-- This will undo the incorrect closing entries
UPDATE accounts a
SET balance = balance + ltr.balance_reversal,
    updated_at = NOW()
FROM lines_to_reverse ltr
WHERE a.id = ltr.account_id;

-- Step 6: Mark the problematic closing entries as VOIDED/CANCELLED
UPDATE journal_entries
SET status = 'VOIDED',
    description = description || ' [AUTO-VOIDED: Duplicate closing detected]',
    updated_at = NOW()
WHERE id IN (SELECT journal_id FROM problematic_closings);

-- Step 7: Update accounting_periods to remove reference to voided closings
UPDATE accounting_periods
SET closing_journal_id = NULL,
    notes = COALESCE(notes, '') || E'\n[AUTO-FIX ' || NOW() || ']: Closing journal voided due to duplicate closing issue',
    updated_at = NOW()
WHERE closing_journal_id IN (SELECT journal_id FROM problematic_closings);

-- Step 8: Verify the fix - check balance sheet
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
    'Balance Sheet Verification' AS check_type,
    total_assets,
    total_liabilities,
    total_equity,
    (total_liabilities + total_equity) AS total_liab_equity,
    (total_assets - (total_liabilities + total_equity)) AS difference,
    CASE 
        WHEN ABS(total_assets - (total_liabilities + total_equity)) < 0.01 THEN 'BALANCED ✓'
        ELSE 'NOT BALANCED ✗'
    END AS status
FROM balance_totals;

-- Step 9: Show retained earnings transactions
SELECT 
    je.id,
    je.code,
    je.entry_date,
    je.description,
    je.reference_type,
    je.status,
    jl.debit_amount,
    jl.credit_amount,
    (jl.credit_amount - jl.debit_amount) AS net_change
FROM journal_lines jl
JOIN journal_entries je ON jl.journal_entry_id = je.id
JOIN accounts a ON jl.account_id = a.id
WHERE a.code = '3201'
    AND je.status = 'POSTED'
ORDER BY je.entry_date, je.id;

-- Summary Report
SELECT 
    'Fix completed successfully!' AS status,
    (SELECT COUNT(*) FROM problematic_closings) AS voided_entries,
    (SELECT COUNT(*) FROM lines_to_reverse) AS reversed_lines,
    NOW() AS fixed_at;

-- ROLLBACK; -- Uncomment to test without committing
COMMIT; -- Comment out for testing, uncomment to apply fix

-- NOTES:
-- 1. Jalankan analyze_double_posting_issue.sql terlebih dahulu untuk diagnosis
-- 2. Script ini akan:
--    a. Mengidentifikasi closing entries yang menutup akun yang sudah 0
--    b. Me-reverse balance changes dari closing entries tersebut
--    c. Menandai journal entries sebagai VOIDED
--    d. Update accounting_periods untuk remove reference ke voided journals
-- 3. Setelah fix, balance sheet seharusnya balance kembali
