-- ================================================
-- JURNAL PENYESUAIAN (ADJUSTMENT ENTRY)
-- ================================================
-- Purpose: Membuat jurnal untuk balance yang ada tanpa journal entries
-- Date: 2025-11-07

BEGIN;

-- ANALYSIS: Current Balances
-- Assets:      Rp 129.430.000 (1102: 108.78jt + 1104: 16.7jt + 1240: 4.95jt + 1301: -1jt)
-- Liabilities: Rp  45.730.000 (2101: 33.3jt + 2103: 12.43jt)
-- Equity:      Rp 152.000.000 (3101: 50jt + 3201: 102jt)
-- 
-- Equation Check:
-- Assets (129.43jt) = Liabilities (45.73jt) + Equity (152jt) 
-- 129.43jt ≠ 197.73jt --> DIFF: -68.3jt NOT BALANCED!
--
-- Problem: Persediaan negatif -1jt (salah!)
-- Fix: Total assets seharusnya = 130.43jt
--       130.43jt ≠ 197.73jt --> masih diff -67.3jt
--
-- ROOT CAUSE: Balance di accounts table tidak match dengan journal entries!
-- Only 2 journal entries exist (closing entries), but accounts have many balances

-- Step 1: Check current state
SELECT 
    'Current Balance Sheet Status' as info,
    SUM(CASE WHEN type = 'ASSET' THEN balance ELSE 0 END) AS total_assets,
    SUM(CASE WHEN type = 'LIABILITY' THEN balance ELSE 0 END) AS total_liabilities,
    SUM(CASE WHEN type = 'EQUITY' THEN balance ELSE 0 END) AS total_equity,
    SUM(CASE WHEN type = 'ASSET' THEN balance ELSE 0 END) - 
    (SUM(CASE WHEN type = 'LIABILITY' THEN balance ELSE 0 END) + 
     SUM(CASE WHEN type = 'EQUITY' THEN balance ELSE 0 END)) AS difference
FROM accounts
WHERE is_active = true AND COALESCE(is_header, false) = false;

-- Step 2: List accounts with balance but NO journal lines
SELECT 
    a.code,
    a.name,
    a.type,
    a.balance,
    COUNT(jl.id) as journal_line_count
FROM accounts a
LEFT JOIN journal_lines jl ON a.id = jl.account_id
LEFT JOIN journal_entries je ON jl.journal_entry_id = je.id AND je.status = 'POSTED'
WHERE a.is_active = true 
    AND COALESCE(a.is_header, false) = false
    AND a.balance != 0
GROUP BY a.id, a.code, a.name, a.type, a.balance
HAVING COUNT(jl.id) = 0
ORDER BY a.type, a.code;

-- Step 3: RECOMMENDED ACTION
-- Rather than creating fake journal entries, we should:
-- OPTION A: Reset all balances to 0 and re-import data properly
-- OPTION B: Create opening balance journal entry

SELECT 
    '⚠️  RECOMMENDATION' as action,
    'Do NOT create fake adjustment entries!' as warning,
    'Instead: Reset balances and re-import data with proper journal entries' as solution;

-- To reset balances (DANGEROUS - use with caution):
-- UPDATE accounts SET balance = 0 WHERE is_active = true AND is_header = false;

-- To create opening balance entry, we need:
-- 1. Modal Pemilik (3101) should be credited with net assets
-- 2. All asset accounts should be debited with their current balance
-- 3. All liability accounts should be credited with their current balance
-- 4. Persediaan negatif needs to be investigated and fixed first!

ROLLBACK;  -- Don't commit, this is just for analysis

-- INSTRUCTIONS:
-- 1. First fix Persediaan Barang Dagangan (1301) negative balance
-- 2. Then investigate where the balances came from
-- 3. Consider full database reset and proper data re-entry
-- 4. If opening balance entry is needed, create it manually with proper documentation
