-- Fix constraint yang terlalu ketat untuk period closing
-- Error: constraint membatasi entry_date hingga CURRENT_DATE + 1 year atau bahkan lebih ketat
-- Solusi: Longgarkan hingga tahun 2099 untuk mendukung period closing masa depan

-- 1. Cek constraint yang ada saat ini
SELECT 
    conname as constraint_name,
    pg_get_constraintdef(oid) as constraint_definition
FROM pg_constraint 
WHERE conname = 'chk_journal_entries_date_valid'
  AND conrelid = 'journal_entries'::regclass;

-- 2. Drop dan recreate constraint dengan validasi yang lebih longgar
ALTER TABLE journal_entries 
DROP CONSTRAINT IF EXISTS chk_journal_entries_date_valid;

ALTER TABLE journal_entries 
ADD CONSTRAINT chk_journal_entries_date_valid 
CHECK (entry_date >= '2000-01-01'::date AND entry_date <= '2099-12-31'::date);

-- 3. Verifikasi constraint baru
SELECT 
    conname as constraint_name,
    pg_get_constraintdef(oid) as constraint_definition
FROM pg_constraint 
WHERE conname = 'chk_journal_entries_date_valid'
  AND conrelid = 'journal_entries'::regclass;

-- 4. Test insert untuk tahun 2026 (tidak akan benar-benar insert, hanya test)
-- DO $$
-- BEGIN
--     INSERT INTO journal_entries (
--         code, description, reference, reference_type,
--         entry_date, user_id, status, total_debit, total_credit,
--         is_balanced, is_auto_generated, created_at, updated_at
--     ) VALUES (
--         'TEST-2026', 'Test period closing 2026', 'TEST', 'CLOSING_BALANCE',
--         '2026-12-31'::date, 1, 'DRAFT', 0, 0,
--         true, true, NOW(), NOW()
--     );
--     
--     -- Rollback untuk tidak menyimpan test data
--     RAISE EXCEPTION 'Test successful - rolling back';
-- EXCEPTION
--     WHEN OTHERS THEN
--         RAISE NOTICE 'Test result: %', SQLERRM;
-- END $$;

-- Expected output: Constraint sekarang memperbolehkan tanggal dari 2000-01-01 hingga 2099-12-31
