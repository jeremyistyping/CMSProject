-- Fix migration_logs status for migrations that already completed but were marked as FAILED
-- due to transaction abort issues

-- Only mark as SUCCESS if the tables/objects actually exist
DO $$
BEGIN
    -- Fix 020_create_unified_journal_ssot.sql if tables exist
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'unified_journal_ledger')
       AND EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'unified_journal_lines') THEN
        UPDATE migration_logs 
        SET status = 'SUCCESS', 
            message = 'Migration completed (auto-fixed after transaction abort issue)',
            updated_at = NOW()
        WHERE migration_name = '020_create_unified_journal_ssot.sql'
          AND status = 'FAILED';
        RAISE NOTICE '✅ Fixed migration log for 020_create_unified_journal_ssot.sql';
    END IF;

    -- Add similar fixes for other migrations if needed
    RAISE NOTICE '✅ Migration log cleanup completed';
END $$;
