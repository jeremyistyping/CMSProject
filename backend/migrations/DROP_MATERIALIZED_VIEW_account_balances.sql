-- ==========================================
-- Drop Materialized View: account_balances
-- ==========================================
-- 
-- TUJUAN: Menghapus materialized view yang sudah tidak dipakai
-- 
-- ALASAN: 
-- 1. Frontend sudah menggunakan SSOT endpoint (/ssot-reports/balance-sheet)
-- 2. SSOT implementation query langsung ke unified_journal_lines (real-time)
-- 3. Materialized view tidak digunakan untuk query apapun
-- 4. Maintenance overhead tidak diperlukan lagi
--
-- PENGGANTI:
-- - SSOT Balance Sheet Service (ssot_balance_sheet_service.go)
-- - Query langsung ke unified_journal_lines untuk real-time data
-- - Endpoint: /api/v1/ssot-reports/balance-sheet
--
-- CARA MENJALANKAN:
-- psql -U your_username -d your_database -f DROP_MATERIALIZED_VIEW_account_balances.sql
--
-- atau dari psql:
-- \i backend/migrations/DROP_MATERIALIZED_VIEW_account_balances.sql
--
-- ==========================================

-- Drop materialized view dan indexes terkait
DROP MATERIALIZED VIEW IF EXISTS account_balances CASCADE;

-- Verify deletion
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_matviews 
        WHERE schemaname = 'public' 
        AND matviewname = 'account_balances'
    ) THEN
        RAISE NOTICE '❌ Failed to drop materialized view account_balances';
    ELSE
        RAISE NOTICE '✅ Materialized view account_balances dropped successfully';
    END IF;
END $$;

-- ==========================================
-- CATATAN:
-- Setelah menjalankan script ini:
-- 1. Restart backend untuk memastikan tidak ada error
-- 2. Test frontend Balance Sheet masih berfungsi dengan SSOT endpoint
-- 3. Verify logs tidak ada error terkait account_balances
-- ==========================================
