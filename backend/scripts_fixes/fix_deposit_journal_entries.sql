-- Script untuk mengoreksi journal entries deposit yang salah menggunakan Revenue
-- Mengganti dengan Modal Pemilik (Owner Equity) untuk balance sheet yang benar

-- 1. Identifikasi journal entries deposit yang menggunakan Revenue account
SELECT 
    je.id as journal_id,
    je.entry_number,
    je.description,
    jl.account_id,
    a.code as account_code,
    a.name as account_name,
    a.type as account_type,
    jl.credit_amount,
    je.created_at
FROM ssot_journal_entries je
JOIN ssot_journal_lines jl ON je.id = jl.journal_entry_id
JOIN accounts a ON jl.account_id = a.id
WHERE je.description LIKE '%Cash/Bank Deposit%'
  OR je.description LIKE '%Deposit to%'
  AND a.type = 'REVENUE'
ORDER BY je.created_at DESC;

-- 2. Cari Modal Pemilik account
SELECT id, code, name, type 
FROM accounts 
WHERE (code = '3101' OR name ILIKE '%Modal Pemilik%' OR name ILIKE '%Owner Equity%')
  AND type = 'EQUITY' 
  AND is_active = true;

-- 3. Update journal lines yang menggunakan Revenue account menjadi Modal Pemilik
-- HATI-HATI: Backup data terlebih dahulu sebelum menjalankan!

/*
-- Uncomment dan sesuaikan account_id setelah mengecek Modal Pemilik account ID

UPDATE ssot_journal_lines 
SET account_id = (
    SELECT id FROM accounts 
    WHERE (code = '3101' OR name ILIKE '%Modal Pemilik%') 
      AND type = 'EQUITY' 
      AND is_active = true 
    LIMIT 1
)
WHERE journal_entry_id IN (
    SELECT je.id 
    FROM ssot_journal_entries je
    JOIN ssot_journal_lines jl ON je.id = jl.journal_entry_id
    JOIN accounts a ON jl.account_id = a.id
    WHERE (je.description LIKE '%Cash/Bank Deposit%' OR je.description LIKE '%Deposit to%')
      AND a.type = 'REVENUE'
      AND jl.credit_amount > 0  -- Only credit entries (Revenue side)
);

-- Update description untuk journal lines yang dikoreksi
UPDATE ssot_journal_lines 
SET description = REPLACE(description, 'Revenue from', 'Capital deposit to')
WHERE journal_entry_id IN (
    SELECT DISTINCT je.id 
    FROM ssot_journal_entries je
    WHERE je.description LIKE '%Cash/Bank Deposit%' OR je.description LIKE '%Deposit to%'
)
AND description LIKE '%Revenue from%';

-- Update main journal entry description
UPDATE ssot_journal_entries 
SET description = REPLACE(description, 'Cash/Bank Deposit', 'Capital Deposit')
WHERE description LIKE '%Cash/Bank Deposit%';

*/

-- 4. Verifikasi perubahan
SELECT 
    je.id as journal_id,
    je.entry_number,
    je.description,
    jl.account_id,
    a.code as account_code,
    a.name as account_name,
    a.type as account_type,
    jl.debit_amount,
    jl.credit_amount,
    je.created_at
FROM ssot_journal_entries je
JOIN ssot_journal_lines jl ON je.id = jl.journal_entry_id
JOIN accounts a ON jl.account_id = a.id
WHERE je.description LIKE '%Capital Deposit%'
  OR je.description LIKE '%Deposit to%'
ORDER BY je.created_at DESC, jl.id;

-- 5. Refresh materialized view setelah perubahan
REFRESH MATERIALIZED VIEW account_balances;

-- 6. Check Balance Sheet balance setelah koreksi
SELECT 
    account_type,
    SUM(current_balance) as total_balance
FROM account_balances 
WHERE account_type IN ('ASSET', 'LIABILITY', 'EQUITY')
  AND is_active = true
GROUP BY account_type
ORDER BY 
    CASE account_type 
        WHEN 'ASSET' THEN 1 
        WHEN 'LIABILITY' THEN 2 
        WHEN 'EQUITY' THEN 3 
    END;