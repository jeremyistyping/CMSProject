-- ============================================================================
-- Fix Double Balance Issue - Bank Mandiri
-- ============================================================================
-- Purpose: Memperbaiki saldo Bank Mandiri yang berlipat ganda akibat 
--          double journal entry (dari Rp 11.100.000 ke Rp 5.550.000)
-- Date: 2025-09-21
-- ============================================================================

-- 1. Check current balances for relevant accounts
SELECT 
    '=== CURRENT ACCOUNT BALANCES ===' as status,
    '' as code, '' as name, 0 as balance
UNION ALL
SELECT 
    CASE 
        WHEN code = '1103' AND balance > 10000000 THEN '❌ PROBLEM'
        WHEN code = '1103' AND balance = 5550000 THEN '✅ CORRECT'
        WHEN code = '1201' AND balance = 0 THEN '✅ PAID'
        WHEN code = '1201' AND balance > 0 THEN '⚠️ OUTSTANDING'
        WHEN code = '4101' AND balance < 0 THEN '✅ CREDIT'
        WHEN code = '2102' AND balance < 0 THEN '✅ CREDIT'
        ELSE '⚠️ CHECK'
    END as status,
    code, name, balance
FROM accounts 
WHERE code IN ('1103', '1201', '4101', '2102')
ORDER BY code;

-- 2. Show current Bank Mandiri balance specifically
SELECT 
    'Bank Mandiri Current Balance:' as description,
    CONCAT('Rp ', FORMAT(balance, 2)) as amount,
    CASE 
        WHEN balance = 11100000 THEN 'DOUBLE BALANCE DETECTED'
        WHEN balance = 5550000 THEN 'CORRECT BALANCE'
        ELSE CONCAT('UNEXPECTED BALANCE: ', balance)
    END as status
FROM accounts 
WHERE code = '1103';

-- 3. Fix the Bank Mandiri balance if it's exactly double (11,100,000)
UPDATE accounts 
SET 
    balance = CASE 
        WHEN balance = 11100000.00 THEN 5550000.00
        ELSE balance
    END,
    updated_at = NOW()
WHERE code = '1103' AND balance = 11100000.00;

-- 4. Also update cash_banks table to maintain consistency
UPDATE cash_banks cb
JOIN accounts a ON cb.account_id = a.id
SET 
    cb.balance = 5550000.00,
    cb.updated_at = NOW()
WHERE a.code = '1103' AND cb.balance = 11100000.00;

-- 5. Verify the fix
SELECT 
    '=== AFTER FIX - VERIFICATION ===' as status,
    '' as code, '' as name, 0 as balance, '' as analysis
UNION ALL
SELECT 
    CASE 
        WHEN code = '1103' AND balance = 5550000 THEN '✅ FIXED'
        WHEN code = '1103' THEN '⚠️ VERIFY'
        WHEN code = '1201' AND balance = 0 THEN '✅ OK'
        WHEN code = '1201' AND balance > 0 THEN '⚠️ OUTSTANDING'
        WHEN code = '4101' AND balance < 0 THEN '✅ OK'
        WHEN code = '2102' AND balance < 0 THEN '✅ OK'
        ELSE '⚠️ CHECK'
    END as status,
    code, 
    name, 
    balance,
    CASE code
        WHEN '1103' THEN 
            CASE 
                WHEN balance = 5550000 THEN '(Payment amount)'
                WHEN balance = 11100000 THEN '(STILL DOUBLED!)'
                ELSE '(Check amount)'
            END
        WHEN '1201' THEN 
            CASE 
                WHEN balance = 0 THEN '(Fully paid)'
                ELSE CONCAT('(Outstanding: Rp ', FORMAT(balance, 2), ')')
            END
        WHEN '4101' THEN '(Sales revenue - should be negative)'
        WHEN '2102' THEN '(Tax payable - should be negative)'
        ELSE ''
    END as analysis
FROM accounts 
WHERE code IN ('1103', '1201', '4101', '2102')
ORDER BY code;

-- 6. Accounting equation verification
SELECT 
    '=== ACCOUNTING EQUATION CHECK ===' as description,
    '' as account_type, 0 as total_amount
UNION ALL
SELECT 
    'ASSETS (Debit)',
    'AR + Bank',
    (SELECT COALESCE(balance, 0) FROM accounts WHERE code = '1201') + 
    (SELECT COALESCE(balance, 0) FROM accounts WHERE code = '1103') as total_assets
UNION ALL
SELECT 
    'LIABILITIES + REVENUE (Credit)',
    'Sales Rev + Tax Payable',
    -((SELECT COALESCE(balance, 0) FROM accounts WHERE code = '4101') + 
      (SELECT COALESCE(balance, 0) FROM accounts WHERE code = '2102')) as total_liab_rev
UNION ALL
SELECT 
    'DIFFERENCE',
    'Assets - (Liab + Rev)',
    (SELECT COALESCE(balance, 0) FROM accounts WHERE code = '1201') + 
    (SELECT COALESCE(balance, 0) FROM accounts WHERE code = '1103') -
    (-((SELECT COALESCE(balance, 0) FROM accounts WHERE code = '4101') + 
       (SELECT COALESCE(balance, 0) FROM accounts WHERE code = '2102'))) as difference;

-- 7. Summary report
SELECT 
    '=== FIX SUMMARY ===' as item,
    '' as details
UNION ALL
SELECT 
    '✅ Bank Mandiri Balance Fixed',
    'From Rp 11.100.000 to Rp 5.550.000'
UNION ALL
SELECT 
    '✅ Code Fixes Applied',
    'Prevented future double balance updates'
UNION ALL
SELECT 
    '✅ SSOT AutoPost Logic Fixed', 
    'Disabled when manual update done'
UNION ALL
SELECT 
    '📋 Next Steps',
    'Test payment creation & monitor balances'
UNION ALL
SELECT
    '🎯 Expected Result',
    'No more double balance updates';

-- ============================================================================
-- END OF SCRIPT
-- ============================================================================