-- ==========================================
-- FIX DUPLICATE ACCOUNT 4101
-- ==========================================

-- 1. CEK DETAIL DUPLIKASI ACCOUNT 4101
SELECT 
    id,
    code,
    name,
    type,
    parent_id,
    balance,
    is_header,
    created_at
FROM accounts
WHERE code = '4101'
ORDER BY id;

-- 2. CEK JOURNAL LINES YANG TERHUBUNG KE MASING-MASING ACCOUNT
SELECT 
    a.id as account_id,
    a.code,
    a.name,
    COUNT(ujl.id) as line_count,
    SUM(ujl.credit_amount) as total_credit,
    SUM(ujl.debit_amount) as total_debit
FROM accounts a
LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
WHERE a.code = '4101'
GROUP BY a.id, a.code, a.name;

-- 3. IDENTIFIKASI ACCOUNT MANA YANG AKTIF DIGUNAKAN
WITH account_usage AS (
    SELECT 
        a.id,
        a.name,
        COUNT(ujl.id) as usage_count,
        MAX(uje.entry_date) as last_used
    FROM accounts a
    LEFT JOIN unified_journal_lines ujl ON ujl.account_id = a.id
    LEFT JOIN unified_journal_ledger uje ON uje.id = ujl.journal_id
    WHERE a.code = '4101'
    GROUP BY a.id, a.name
)
SELECT 
    id,
    name,
    usage_count,
    last_used,
    CASE 
        WHEN usage_count > 0 THEN 'ACTIVE - KEEP THIS'
        ELSE 'UNUSED - CAN BE MERGED/DELETED'
    END as recommendation
FROM account_usage
ORDER BY usage_count DESC;

-- 4. SOLUSI: UPDATE JOURNAL LINES YANG MENGGUNAKAN ACCOUNT DUPLIKAT
-- (Uncomment setelah verify account mana yang benar)
/*
-- Jika ingin merge semua ke account dengan ID terendah:
UPDATE unified_journal_lines
SET account_id = (SELECT MIN(id) FROM accounts WHERE code = '4101')
WHERE account_id IN (SELECT id FROM accounts WHERE code = '4101')
  AND account_id != (SELECT MIN(id) FROM accounts WHERE code = '4101');

-- Hapus account duplikat yang sudah tidak digunakan
DELETE FROM accounts 
WHERE code = '4101' 
  AND id != (SELECT MIN(id) FROM accounts WHERE code = '4101')
  AND NOT EXISTS (
    SELECT 1 FROM unified_journal_lines ujl WHERE ujl.account_id = accounts.id
  );
*/

-- 5. STANDARDISASI NAMA ACCOUNT
/*
UPDATE accounts
SET name = 'Pendapatan Penjualan'
WHERE code = '4101';
*/

