-- ========================================================================
-- QUICK FIX: One-liner to remove problematic trigger
-- ========================================================================
-- Usage: 
--   psql -U postgres -d accounting_db -f backend/quick_fix_trigger.sql
-- Or copy-paste this single line into any PostgreSQL client:
-- ========================================================================

DROP TRIGGER IF EXISTS trg_refresh_account_balances ON unified_journal_lines CASCADE;

-- Verify removal
SELECT 
    CASE 
        WHEN EXISTS (SELECT 1 FROM pg_trigger WHERE tgname = 'trg_refresh_account_balances')
        THEN '❌ Trigger still exists!'
        ELSE '✅ Trigger removed successfully'
    END as status;
