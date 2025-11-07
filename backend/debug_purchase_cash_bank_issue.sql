-- Debug script for investigating cash & bank balance issue in purchase
-- This script will check:
-- 1. Recent purchases with immediate payment methods
-- 2. Their approval status and workflow
-- 3. Related cash_bank_transactions records
-- 4. Associated journal entries and COA balances

-- 1. Recent purchases with immediate payment (CASH, BANK_TRANSFER)
SELECT 
    p.id,
    p.code,
    p.vendor_id,
    c.name as vendor_name,
    p.payment_method,
    p.bank_account_id,
    p.total_amount,
    p.status,
    p.approval_status,
    p.created_at,
    p.approved_at
FROM purchases p 
LEFT JOIN contacts c ON p.vendor_id = c.id
WHERE p.payment_method IN ('CASH', 'BANK_TRANSFER', 'CHECK')
ORDER BY p.created_at DESC 
LIMIT 10;

-- 2. Check approval requests for recent purchases
SELECT 
    ar.id as approval_request_id,
    ar.request_code,
    ar.entity_id as purchase_id,
    ar.amount,
    ar.status as approval_status,
    ar.created_at as approval_created,
    ar.completed_at,
    p.code as purchase_code,
    p.payment_method
FROM approval_requests ar
JOIN purchases p ON ar.entity_id = p.id
WHERE ar.entity_type = 'PURCHASE' 
    AND p.payment_method IN ('CASH', 'BANK_TRANSFER', 'CHECK')
ORDER BY ar.created_at DESC 
LIMIT 10;

-- 3. Check approval actions for recent purchase approvals
SELECT 
    aa.id,
    ar.request_code,
    aa.status,
    aa.approved_at,
    aa.is_active,
    astp.step_name,
    astp.approver_role,
    u.username as approver,
    p.code as purchase_code
FROM approval_actions aa
JOIN approval_requests ar ON aa.request_id = ar.id
JOIN approval_steps astp ON aa.step_id = astp.id
LEFT JOIN users u ON aa.approved_by = u.id
JOIN purchases p ON ar.entity_id = p.id
WHERE ar.entity_type = 'PURCHASE'
    AND p.payment_method IN ('CASH', 'BANK_TRANSFER', 'CHECK')
ORDER BY aa.created_at DESC
LIMIT 20;

-- 4. Check cash_bank_transactions for recent purchases
SELECT 
    cbt.id,
    cbt.cash_bank_id,
    cb.name as cash_bank_name,
    cbt.reference_type,
    cbt.reference_id,
    cbt.amount,
    cbt.balance_after,
    cbt.transaction_date,
    cbt.notes,
    p.code as purchase_code
FROM cash_bank_transactions cbt
JOIN cash_banks cb ON cbt.cash_bank_id = cb.id
LEFT JOIN purchases p ON cbt.reference_type = 'PURCHASE' AND cbt.reference_id = p.id
WHERE cbt.reference_type IN ('PURCHASE', 'PAYMENT')
ORDER BY cbt.transaction_date DESC
LIMIT 15;

-- 5. Check current cash_banks balance
SELECT 
    id,
    name,
    account_id,
    balance,
    updated_at
FROM cash_banks 
ORDER BY updated_at DESC;

-- 6. Check recent journal entries for purchases
SELECT 
    je.id,
    je.entry_date,
    je.description,
    je.reference_type,
    je.reference_id,
    je.reference,
    je.total_debit,
    je.total_credit,
    je.status,
    p.code as purchase_code
FROM journal_entries je
LEFT JOIN purchases p ON je.reference_type = 'PURCHASE' AND je.reference_id::text = p.id::text
WHERE je.reference_type IN ('PURCHASE', 'PAYMENT')
ORDER BY je.created_at DESC
LIMIT 10;

-- 7. Check account balances for cash/bank accounts
SELECT 
    a.id,
    a.code,
    a.name,
    a.account_type,
    a.balance,
    a.updated_at,
    cb.name as linked_cash_bank
FROM accounts a
LEFT JOIN cash_banks cb ON a.id = cb.account_id
WHERE a.account_type = 'ASSET' 
    AND (a.code LIKE '11%' OR a.name ILIKE '%kas%' OR a.name ILIKE '%bank%')
ORDER BY a.code;

-- 8. Check if there are any purchase workflows configured
SELECT 
    id,
    name,
    module,
    min_amount,
    max_amount,
    require_director,
    require_finance,
    is_active
FROM approval_workflows 
WHERE module = 'PURCHASE' AND is_active = true;