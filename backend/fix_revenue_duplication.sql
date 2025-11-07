-- =====================================================
-- FIX: Revenue Duplication Issue
-- Problem: Account 4101 appears twice with different names
-- =====================================================

-- ==========================================
-- STEP 1: DIAGNOSTIC - Check Current State
-- ==========================================

-- 1.1 Check parent account (4000) is_header flag
SELECT 
    'Parent Account Check' as Check_Type,
    code,
    name,
    type,
    balance,
    COALESCE(is_header, false) as is_header,
    CASE 
        WHEN COALESCE(is_header, false) = true THEN 'OK - Will be skipped'
        WHEN COALESCE(is_header, false) = false THEN 'PROBLEM - Will be included!'
    END as status
FROM accounts 
WHERE code = '4000';

-- 1.2 Check child account (4101)
SELECT 
    'Child Account Check' as Check_Type,
    code,
    name,
    type,
    balance,
    COALESCE(is_header, false) as is_header
FROM accounts 
WHERE code = '4101';

-- 1.3 Check for account name variations in journal_entries
SELECT 
    'Journal Entries Name Variations' as Check_Type,
    je.account_code,
    je.account_name,
    COUNT(*) as entry_count,
    SUM(je.credit - je.debit) as total_amount
FROM journal_entries je
INNER JOIN journals j ON j.id = je.journal_id
WHERE je.account_code = '4101'
  AND j.status = 'POSTED'
GROUP BY je.account_code, je.account_name
ORDER BY je.account_name;

-- 1.4 Check for account name variations in unified_journal_lines
SELECT 
    'Unified Journal Name Variations' as Check_Type,
    ujl.account_code,
    ujl.account_name,
    COUNT(*) as line_count,
    SUM(ujl.credit_amount - ujl.debit_amount) as total_amount
FROM unified_journal_lines ujl
INNER JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
WHERE ujl.account_code = '4101'
  AND uje.status = 'POSTED'
GROUP BY ujl.account_code, ujl.account_name
ORDER BY ujl.account_name;

-- ==========================================
-- STEP 2: FIXES
-- ==========================================

-- FIX 2.1: Ensure parent account 4000 is marked as header
UPDATE accounts 
SET is_header = true
WHERE code = '4000' 
  AND COALESCE(is_header, false) = false;

-- Verify
SELECT 'After Fix 2.1' as Status, code, name, is_header 
FROM accounts 
WHERE code = '4000';

-- FIX 2.2: Standardize account_name in journal_entries to match accounts table
UPDATE journal_entries je
INNER JOIN accounts a ON a.code = je.account_code
SET je.account_name = a.name
WHERE je.account_code = '4101'
  AND je.account_name != a.name;

-- Verify
SELECT 'After Fix 2.2' as Status, 
       je.account_code, 
       je.account_name, 
       COUNT(*) as count
FROM journal_entries je
WHERE je.account_code = '4101'
GROUP BY je.account_code, je.account_name;

-- FIX 2.3: Standardize account_name in unified_journal_lines
UPDATE unified_journal_lines ujl
INNER JOIN accounts a ON a.id = ujl.account_id
SET ujl.account_name = a.name
WHERE ujl.account_code = '4101'
  AND ujl.account_name != a.name;

-- Verify
SELECT 'After Fix 2.3' as Status,
       ujl.account_code,
       ujl.account_name,
       COUNT(*) as count
FROM unified_journal_lines ujl
WHERE ujl.account_code = '4101'
GROUP BY ujl.account_code, ujl.account_name;

-- FIX 2.4: Apply fix to ALL revenue accounts (4xxx) not just 4101
UPDATE journal_entries je
INNER JOIN accounts a ON a.code = je.account_code
SET je.account_name = a.name
WHERE je.account_code LIKE '4%'
  AND je.account_name != a.name;

UPDATE unified_journal_lines ujl
INNER JOIN accounts a ON a.id = ujl.account_id
SET ujl.account_name = a.name
WHERE ujl.account_code LIKE '4%'
  AND ujl.account_name != a.name;

-- ==========================================
-- STEP 3: VERIFICATION
-- ==========================================

-- 3.1 Verify no name variations remain
SELECT 
    'Final Verification' as Check_Type,
    je.account_code,
    COUNT(DISTINCT je.account_name) as distinct_names,
    GROUP_CONCAT(DISTINCT je.account_name) as all_names,
    SUM(je.credit - je.debit) as total_amount
FROM journal_entries je
INNER JOIN journals j ON j.id = je.journal_id
WHERE je.account_code LIKE '4%'
  AND j.status = 'POSTED'
GROUP BY je.account_code
HAVING COUNT(DISTINCT je.account_name) > 1;

-- Should return 0 rows if fixed!

-- 3.2 Final revenue calculation
SELECT 
    'Final Revenue Check' as Check_Type,
    je.account_code,
    je.account_name,
    SUM(je.credit - je.debit) as amount
FROM journal_entries je
INNER JOIN journals j ON j.id = je.journal_id
WHERE je.account_code LIKE '4%'
  AND j.status = 'POSTED'
  AND j.date BETWEEN '2025-01-01' AND '2025-12-31'
GROUP BY je.account_code, je.account_name
ORDER BY je.account_code;

-- Should show only ONE row per account code!

-- ==========================================
-- STEP 4: SUMMARY
-- ==========================================

SELECT '=== FIX SUMMARY ===' as Summary;
SELECT 'Expected Result: Account 4101 appears only ONCE with Rp 10,000,000' as Expected;
SELECT 'Run backend API test to verify P&L now shows Rp 10,000,000 total' as Next_Step;

-- ==========================================
-- ROLLBACK (if needed)
-- ==========================================
-- If something goes wrong, restore from backup
-- Or manually update back to original names

-- ==========================================
-- PREVENTION: Add CHECK constraint
-- ==========================================

-- Ensure all parent accounts are marked as headers
UPDATE accounts 
SET is_header = true
WHERE code IN ('1000', '1100', '1500', '2000', '2100', '3000', '4000', '5000')
  AND COALESCE(is_header, false) = false;

-- Verify all parent accounts
SELECT 
    'Parent Accounts Final Check' as Check_Type,
    code,
    name,
    is_header
FROM accounts
WHERE code IN ('1000', '1100', '1500', '2000', '2100', '3000', '4000', '5000')
ORDER BY code;

-- All should have is_header = true

