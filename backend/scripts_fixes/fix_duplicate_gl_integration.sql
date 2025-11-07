-- 🔍 IDENTIFY DUPLICATE GL ACCOUNT INTEGRATIONS
-- This script helps identify cash bank accounts that might be linked to the same GL account

-- 1. Find cash bank accounts with their GL account details
SELECT 
    cb.id as cash_bank_id,
    cb.code as cash_bank_code, 
    cb.name as cash_bank_name,
    cb.type,
    cb.balance,
    cb.account_id as gl_account_id,
    a.code as gl_account_code,
    a.name as gl_account_name,
    cb.is_active
FROM cash_bank_accounts cb
LEFT JOIN accounts a ON cb.account_id = a.id
WHERE cb.is_active = 1
ORDER BY cb.type, cb.name;

-- 2. Find potential duplicates - GL accounts used by multiple cash bank accounts
SELECT 
    a.id as gl_account_id,
    a.code as gl_account_code,
    a.name as gl_account_name,
    COUNT(cb.id) as cash_bank_count,
    GROUP_CONCAT(cb.name SEPARATOR ' | ') as cash_bank_names,
    SUM(cb.balance) as total_balance
FROM accounts a
INNER JOIN cash_bank_accounts cb ON a.id = cb.account_id  
WHERE cb.is_active = 1
GROUP BY a.id, a.code, a.name
HAVING cash_bank_count > 1
ORDER BY total_balance DESC;

-- 3. Show summary by bank name (potential physical duplicates)
SELECT 
    cb.bank_name,
    cb.type,
    COUNT(*) as account_count,
    SUM(cb.balance) as total_balance,
    GROUP_CONCAT(CONCAT(cb.code, ' (', cb.balance, ')') SEPARATOR ' | ') as accounts_detail
FROM cash_bank_accounts cb 
WHERE cb.is_active = 1 
  AND cb.bank_name IS NOT NULL 
  AND cb.bank_name != ''
GROUP BY cb.bank_name, cb.type
HAVING account_count > 1
ORDER BY total_balance DESC;

-- 4. RECOMMENDED FIX: Ensure each cash bank account has unique GL account
-- Example fix for Bank BRI case:

-- If you have duplicate Bank BRI accounts, keep one and deactivate others:
-- UPDATE cash_bank_accounts 
-- SET is_active = 0 
-- WHERE bank_name = 'BANK BRI' 
--   AND id != (SELECT MIN(id) FROM (SELECT id FROM cash_bank_accounts WHERE bank_name = 'BANK BRI' AND is_active = 1) as temp);

-- Or consolidate balances into one account:
-- UPDATE cash_bank_accounts 
-- SET balance = (SELECT SUM(balance) FROM (SELECT balance FROM cash_bank_accounts WHERE bank_name = 'BANK BRI' AND is_active = 1) as temp)
-- WHERE bank_name = 'BANK BRI' 
--   AND id = (SELECT MIN(id) FROM (SELECT id FROM cash_bank_accounts WHERE bank_name = 'BANK BRI' AND is_active = 1) as temp);

-- UPDATE cash_bank_accounts 
-- SET is_active = 0 
-- WHERE bank_name = 'BANK BRI' 
--   AND id != (SELECT MIN(id) FROM (SELECT id FROM cash_bank_accounts WHERE bank_name = 'BANK BRI' AND is_active = 1) as temp);