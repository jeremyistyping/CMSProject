-- Script untuk memperbaiki sinkronisasi COA Balance dengan Cash & Bank Balance
-- Problem: Cash & Bank balance ter-update tapi COA balance tidak ter-update setelah payment

-- 1. Backup data penting sebelum melakukan perubahan
CREATE TABLE IF NOT EXISTS coa_backup_before_sync AS 
SELECT * FROM coa WHERE updated_at >= CURDATE() - INTERVAL 30 DAY;

-- 2. Identifikasi ketidakcocokan antara Cash & Bank balance dan COA balance
SELECT 
    cb.id as cash_bank_id,
    cb.name as cash_bank_name,
    cb.account_id,
    cb.balance as cash_bank_balance,
    coa.code as coa_code,
    coa.name as coa_name,
    coa.balance as coa_balance,
    (cb.balance - coa.balance) as difference,
    coa.type as account_type
FROM cash_banks cb
LEFT JOIN coa ON cb.account_id = coa.id
WHERE cb.is_active = 1 
  AND coa.is_active = 1
  AND ABS(cb.balance - coa.balance) > 0.01
ORDER BY ABS(cb.balance - coa.balance) DESC;

-- 3. Check journal entries yang tidak tersinkronisasi
SELECT 
    je.id as journal_id,
    je.reference,
    je.description,
    je.entry_date,
    je.status,
    COUNT(jl.id) as journal_lines_count,
    SUM(jl.debit_amount) as total_debit,
    SUM(jl.credit_amount) as total_credit
FROM ssot_journal_entries je
LEFT JOIN ssot_journal_lines jl ON je.id = jl.journal_entry_id
WHERE je.source_type IN ('PURCHASE', 'PAYMENT')
  AND je.status = 'POSTED'
  AND je.created_at >= CURDATE() - INTERVAL 7 DAY
GROUP BY je.id
HAVING total_debit != total_credit OR journal_lines_count = 0
ORDER BY je.created_at DESC;

-- 4. Update COA balance berdasarkan Cash & Bank balance untuk akun bank/kas
UPDATE coa 
JOIN cash_banks cb ON coa.id = cb.account_id
SET coa.balance = cb.balance,
    coa.updated_at = NOW()
WHERE cb.is_active = 1 
  AND coa.is_active = 1
  AND coa.type = 'ASSET'
  AND (coa.code LIKE '110%' OR coa.code LIKE '1101' OR coa.code LIKE '1102')
  AND ABS(cb.balance - coa.balance) > 0.01;

-- 5. Recalculate COA balance berdasarkan journal entries untuk semua akun
-- Update balance berdasarkan simple_ssot_journal_items yang POSTED
UPDATE coa SET balance = COALESCE((
    SELECT 
        CASE coa.type
            WHEN 'ASSET' THEN SUM(ssji.debit - ssji.credit)
            WHEN 'EXPENSE' THEN SUM(ssji.debit - ssji.credit)  
            WHEN 'LIABILITY' THEN SUM(ssji.credit - ssji.debit)
            WHEN 'EQUITY' THEN SUM(ssji.credit - ssji.debit)
            WHEN 'REVENUE' THEN SUM(ssji.credit - ssji.debit)
            ELSE 0
        END
    FROM simple_ssot_journal_items ssji
    JOIN simple_ssot_journals ssj ON ssji.journal_id = ssj.id
    WHERE ssji.account_id = coa.id 
      AND ssj.status = 'POSTED'
), 0),
updated_at = NOW()
WHERE coa.is_active = 1;

-- 6. Update balance berdasarkan SSOT journal entries (jika ada)
UPDATE coa SET balance = balance + COALESCE((
    SELECT 
        CASE coa.type
            WHEN 'ASSET' THEN SUM(jl.debit_amount - jl.credit_amount)
            WHEN 'EXPENSE' THEN SUM(jl.debit_amount - jl.credit_amount)  
            WHEN 'LIABILITY' THEN SUM(jl.credit_amount - jl.debit_amount)
            WHEN 'EQUITY' THEN SUM(jl.credit_amount - jl.debit_amount)
            WHEN 'REVENUE' THEN SUM(jl.credit_amount - jl.debit_amount)
            ELSE 0
        END
    FROM ssot_journal_lines jl
    JOIN ssot_journal_entries je ON jl.journal_entry_id = je.id
    WHERE jl.account_id = coa.id 
      AND je.status = 'POSTED'
      AND je.source_type IN ('PAYMENT', 'PURCHASE')
), 0),
updated_at = NOW()
WHERE coa.is_active = 1;

-- 7. Verifikasi hasil setelah update
SELECT 
    'AFTER UPDATE' as status,
    cb.id as cash_bank_id,
    cb.name as cash_bank_name,
    cb.balance as cash_bank_balance,
    coa.code as coa_code,
    coa.name as coa_name,
    coa.balance as coa_balance,
    (cb.balance - coa.balance) as difference
FROM cash_banks cb
LEFT JOIN coa ON cb.account_id = coa.id
WHERE cb.is_active = 1 
  AND coa.is_active = 1
  AND ABS(cb.balance - coa.balance) > 0.01
ORDER BY ABS(cb.balance - coa.balance) DESC;

-- 8. Log perubahan yang dilakukan
INSERT INTO coa_balance_sync_log (
    sync_date,
    accounts_updated,
    total_difference_before,
    total_difference_after,
    notes
)
SELECT 
    NOW(),
    COUNT(*),
    SUM(ABS(cb.balance - coa_backup.balance)),
    SUM(ABS(cb.balance - coa.balance)),
    'COA balance synchronized with Cash & Bank balance'
FROM cash_banks cb
LEFT JOIN coa ON cb.account_id = coa.id
LEFT JOIN coa_backup_before_sync coa_backup ON coa.id = coa_backup.id
WHERE cb.is_active = 1 AND coa.is_active = 1;

-- 9. Check saldo Hutang Usaha untuk purchase credit
SELECT 
    coa.code,
    coa.name,
    coa.balance,
    (SELECT SUM(outstanding_amount) FROM purchases WHERE payment_method = 'CREDIT' AND status IN ('APPROVED', 'COMPLETED')) as total_outstanding_purchases,
    (coa.balance - COALESCE((SELECT SUM(outstanding_amount) FROM purchases WHERE payment_method = 'CREDIT' AND status IN ('APPROVED', 'COMPLETED')), 0)) as difference
FROM coa 
WHERE code = '2101' -- Hutang Usaha
  AND is_active = 1;

-- 10. Update Hutang Usaha balance jika perlu
UPDATE coa 
SET balance = COALESCE((
    SELECT SUM(outstanding_amount) 
    FROM purchases 
    WHERE payment_method = 'CREDIT' 
      AND status IN ('APPROVED', 'COMPLETED', 'PAID')
), 0),
updated_at = NOW()
WHERE code = '2101' 
  AND is_active = 1;

-- 11. Final validation report
SELECT 
    'SUMMARY' as report_type,
    COUNT(*) as total_coa_accounts,
    SUM(CASE WHEN balance > 0 THEN 1 ELSE 0 END) as positive_balance_accounts,
    SUM(CASE WHEN balance < 0 THEN 1 ELSE 0 END) as negative_balance_accounts,
    SUM(CASE WHEN balance = 0 THEN 1 ELSE 0 END) as zero_balance_accounts,
    SUM(balance) as total_balance_all_accounts
FROM coa 
WHERE is_active = 1

UNION ALL

SELECT 
    'CASH_BANK_SYNC' as report_type,
    COUNT(*) as total_accounts,
    SUM(CASE WHEN ABS(cb.balance - coa.balance) <= 0.01 THEN 1 ELSE 0 END) as synced_accounts,
    SUM(CASE WHEN ABS(cb.balance - coa.balance) > 0.01 THEN 1 ELSE 0 END) as unsynced_accounts,
    0 as zero_balance_accounts,
    SUM(ABS(cb.balance - coa.balance)) as total_difference
FROM cash_banks cb
LEFT JOIN coa ON cb.account_id = coa.id
WHERE cb.is_active = 1 AND coa.is_active = 1;

COMMIT;