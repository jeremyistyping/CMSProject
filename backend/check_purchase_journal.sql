-- Script untuk mengecek jurnal purchase transaction PO/2025/10/0003

-- 1. Cek transaksi purchase
SELECT 
    id,
    purchase_number,
    vendor_name,
    total_amount,
    ppn_amount,
    grand_total,
    status,
    approval_status,
    transaction_date,
    created_at
FROM purchases
WHERE purchase_number = 'PO/2025/10/0003'
ORDER BY created_at DESC;

-- 2. Cek journal entries untuk purchase ini
SELECT 
    j.id as journal_id,
    j.source_type,
    j.source_reference,
    j.entry_date,
    j.description,
    ji.account_id,
    a.code as account_code,
    a.name as account_name,
    a.type as account_type,
    ji.debit_amount,
    ji.credit_amount,
    ji.description as item_description
FROM unified_journal_ledger j
LEFT JOIN unified_journal_lines ji ON j.id = ji.journal_id
LEFT JOIN accounts a ON ji.account_id = a.id
WHERE j.source_reference = 'PO/2025/10/0003'
   OR j.source_reference LIKE '%0003%'
ORDER BY j.entry_date DESC, ji.debit_amount DESC;

-- 3. Cek semua journal entries di Oktober 2025
SELECT 
    j.id as journal_id,
    j.source_type,
    j.source_reference,
    j.entry_date,
    COUNT(*) as line_items,
    SUM(ji.debit_amount) as total_debit,
    SUM(ji.credit_amount) as total_credit
FROM unified_journal_ledger j
LEFT JOIN unified_journal_lines ji ON j.id = ji.journal_id
WHERE j.entry_date BETWEEN '2025-10-01' AND '2025-10-31'
  AND j.deleted_at IS NULL
GROUP BY j.id, j.source_type, j.source_reference, j.entry_date
ORDER BY j.entry_date DESC;

-- 4. Cek purchase items detail
SELECT 
    pi.id,
    pi.purchase_id,
    pi.product_id,
    p.name as product_name,
    pi.quantity,
    pi.unit_price,
    pi.total_price,
    pi.expense_account_id,
    ea.code as expense_account_code,
    ea.name as expense_account_name
FROM purchase_items pi
LEFT JOIN products p ON pi.product_id = p.id
LEFT JOIN accounts ea ON pi.expense_account_id = ea.id
WHERE pi.purchase_id IN (
    SELECT id FROM purchases WHERE purchase_number = 'PO/2025/10/0003'
);

-- 5. Cek balance account 6001 (Operating Expense)
SELECT 
    id,
    code,
    name,
    type,
    balance,
    is_active
FROM accounts
WHERE code LIKE '6%'
   AND deleted_at IS NULL
ORDER BY code;

-- 6. Cek apakah ada transaksi di account 6001
SELECT 
    ji.account_id,
    a.code,
    a.name,
    COUNT(*) as transaction_count,
    SUM(ji.debit_amount) as total_debit,
    SUM(ji.credit_amount) as total_credit,
    SUM(ji.debit_amount - ji.credit_amount) as net_balance
FROM unified_journal_lines ji
LEFT JOIN accounts a ON ji.account_id = a.id
WHERE a.code LIKE '6%'
  AND ji.deleted_at IS NULL
GROUP BY ji.account_id, a.code, a.name
ORDER BY a.code;

