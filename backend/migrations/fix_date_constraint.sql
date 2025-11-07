-- Fix constraint yang terlalu ketat untuk period closing
-- Masalah: constraint membatasi entry_date hingga CURRENT_DATE + 1 year
-- Solusi: Longgarkan hingga tahun 2099

ALTER TABLE journal_entries 
DROP CONSTRAINT IF EXISTS chk_journal_entries_date_valid,
ADD CONSTRAINT chk_journal_entries_date_valid 
CHECK (entry_date >= '2000-01-01' AND entry_date <= '2099-12-31');

-- Verifikasi constraint baru
SELECT 
    conname as constraint_name,
    pg_get_constraintdef(oid) as constraint_definition
FROM pg_constraint 
WHERE conname = 'chk_journal_entries_date_valid';
