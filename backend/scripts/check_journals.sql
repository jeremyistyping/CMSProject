-- Check all journal entries for the sale
SELECT 
    j.id,
    j.transaction_type,
    j.transaction_id,
    j.entry_number,
    j.description,
    j.total_amount,
    j.status,
    j.created_at,
    ji.account_code,
    ji.account_name,
    ji.debit,
    ji.credit
FROM simple_ssot_journals j
LEFT JOIN simple_ssot_journal_items ji ON ji.journal_id = j.id
WHERE j.transaction_type IN ('SALES', 'SALES_PAYMENT')
AND j.deleted_at IS NULL
ORDER BY j.created_at DESC, ji.account_code;

-- Count journals per transaction
SELECT 
    transaction_type,
    transaction_id,
    COUNT(*) as journal_count,
    SUM(total_amount) as total_amount_sum
FROM simple_ssot_journals
WHERE transaction_type IN ('SALES', 'SALES_PAYMENT')
AND deleted_at IS NULL
GROUP BY transaction_type, transaction_id
HAVING COUNT(*) > 1;

-- Check account balances
SELECT 
    code,
    name,
    type,
    balance
FROM accounts
WHERE code IN ('1102', '1201', '4101', '2103')
AND deleted_at IS NULL
ORDER BY code;

