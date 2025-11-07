-- Fix CREATE INDEX CONCURRENTLY issue
-- CONCURRENTLY cannot run in transaction, so we use regular CREATE INDEX

DO $$
BEGIN
    -- Drop and recreate without CONCURRENTLY
    DROP INDEX IF EXISTS accounts_code_active_unique;
    
    RAISE NOTICE '🔧 Creating unique index on accounts.code (non-concurrent)';
    
    -- Create regular unique index (not concurrent, but works in transaction)
    CREATE UNIQUE INDEX IF NOT EXISTS accounts_code_active_unique 
    ON accounts (code) 
    WHERE deleted_at IS NULL;
    
    RAISE NOTICE '✅ Created accounts_code_active_unique index';
    
    -- Mark migration as success if it was failed
    UPDATE migration_logs 
    SET status = 'SUCCESS', 
        message = 'Fixed by using non-concurrent index creation',
        updated_at = NOW()
    WHERE migration_name = 'add_accounts_code_unique_constraint.sql'
      AND status = 'FAILED';
    
    RAISE NOTICE '🎯 Unique constraint fix completed';
END $$;
