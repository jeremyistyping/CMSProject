-- Find all journal items that affected BANK account (1102)
SELECT 
    ji.id as item_id,
    j.id as journal_id,
    j.transaction_type,
    j.transaction_id,
    j.entry_number,
    j.description,
    j.created_at,
    ji.account_code,
    ji.account_name,
    ji.debit,
    ji.credit,
    (ji.debit - ji.credit) as net_change
FROM simple_ssot_journal_items ji
JOIN simple_ssot_journals j ON j.id = ji.journal_id
WHERE ji.account_code = '1102'
AND j.deleted_at IS NULL
ORDER BY j.created_at DESC;

-- Check if there are any cash_bank_transactions affecting Bank
SELECT 
    id,
    cash_bank_id,
    transaction_type,
    amount,
    description,
    created_at
FROM cash_bank_transactions
WHERE cash_bank_id = (SELECT id FROM cash_banks WHERE account_id = (SELECT id FROM accounts WHERE code = '1102' LIMIT 1))
ORDER BY created_at DESC;

