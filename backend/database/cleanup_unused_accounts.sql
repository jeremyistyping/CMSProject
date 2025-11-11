-- =========================================================
-- Script untuk menghapus account yang tidak diperlukan
-- dari COA Cost Control Padel Bandung
-- =========================================================

-- BACKUP DULU sebelum jalankan script ini!
-- pg_dump -U postgres -d dbname > backup_before_cleanup.sql

BEGIN;

-- 1. Hapus Fixed Assets yang tidak diperlukan (1500-1509)
UPDATE accounts SET deleted_at = NOW() WHERE code IN ('1500', '1501', '1502', '1503', '1509') AND deleted_at IS NULL;

-- 2. Hapus Old Liability Structure yang duplikat
UPDATE accounts SET deleted_at = NOW() WHERE code IN ('2100', '2107', '2108') AND deleted_at IS NULL;

-- 3. Hapus Old Equity Structure
UPDATE accounts SET deleted_at = NOW() WHERE code = '3201' AND deleted_at IS NULL;

-- 4. Hapus Old Revenue yang tidak sesuai
UPDATE accounts SET deleted_at = NOW() WHERE code = '4900' AND deleted_at IS NULL;

-- 5. Hapus Old Expense Items yang duplikat/tidak sesuai
UPDATE accounts SET deleted_at = NOW() WHERE code IN ('5203', '5204', '5900') AND deleted_at IS NULL;

-- Cek hasil cleanup
SELECT 
    'DELETED' as status,
    code, 
    name, 
    type,
    deleted_at
FROM accounts 
WHERE code IN (
    '1500', '1501', '1502', '1503', '1509',  -- Fixed Assets
    '2100', '2107', '2108',                   -- Old Liabilities
    '3201',                                    -- Old Equity
    '4900',                                    -- Old Revenue
    '5203', '5204', '5900'                    -- Old Expenses
)
ORDER BY code;

-- Tampilkan COA yang tersisa (aktif)
SELECT 
    'ACTIVE' as status,
    code, 
    name, 
    type,
    balance
FROM accounts 
WHERE deleted_at IS NULL
ORDER BY code;

COMMIT;

-- Jika ada masalah, rollback dengan:
-- ROLLBACK;
