-- ========================================================================
-- EMERGENCY FIX: Remove Problematic Trigger Manually
-- ========================================================================
-- Use this script if backend auto-fix fails or if you need immediate fix
-- Run this directly in psql, pgAdmin, or any PostgreSQL client
-- ========================================================================

\echo '🔧 Starting emergency trigger fix...'

-- 1. Show current triggers (for verification)
\echo ''
\echo '📋 Current triggers on unified_journal_lines:'
SELECT 
    tgname as trigger_name,
    pg_get_triggerdef(oid) as trigger_definition
FROM pg_trigger 
WHERE tgrelid = 'unified_journal_lines'::regclass 
AND tgname NOT LIKE 'RI_%'  -- Exclude system triggers
ORDER BY tgname;

-- 2. Drop the problematic trigger
\echo ''
\echo '🗑️  Dropping trg_refresh_account_balances...'
DROP TRIGGER IF EXISTS trg_refresh_account_balances ON unified_journal_lines CASCADE;

-- 3. Verify trigger is removed
\echo ''
\echo '✅ Verifying trigger removal...'
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_trigger 
        WHERE tgname = 'trg_refresh_account_balances'
    ) THEN
        RAISE EXCEPTION '❌ Failed to remove trigger!';
    ELSE
        RAISE NOTICE '✅ Trigger successfully removed';
    END IF;
END $$;

-- 4. Create manual refresh function (if not exists)
\echo ''
\echo '🔧 Creating manual refresh helper functions...'

CREATE OR REPLACE FUNCTION manual_refresh_account_balances()
RETURNS TABLE(
    success BOOLEAN,
    message TEXT,
    refreshed_at TIMESTAMPTZ
) AS $$
DECLARE
    start_time TIMESTAMPTZ;
    end_time TIMESTAMPTZ;
    duration INTERVAL;
BEGIN
    start_time := clock_timestamp();
    
    -- Refresh the materialized view
    REFRESH MATERIALIZED VIEW account_balances;
    
    end_time := clock_timestamp();
    duration := end_time - start_time;
    
    RETURN QUERY SELECT 
        TRUE as success,
        format('Account balances refreshed in %s', duration) as message,
        end_time as refreshed_at;
        
EXCEPTION WHEN OTHERS THEN
    RETURN QUERY SELECT 
        FALSE as success,
        format('Refresh failed: %s', SQLERRM) as message,
        clock_timestamp() as refreshed_at;
END;
$$ LANGUAGE plpgsql;

-- 5. Create freshness check function (if not exists)
CREATE OR REPLACE FUNCTION check_account_balances_freshness()
RETURNS TABLE(
    last_updated TIMESTAMPTZ,
    age_minutes INTEGER,
    needs_refresh BOOLEAN
) AS $$
DECLARE
    last_update TIMESTAMPTZ;
    age_mins INTEGER;
BEGIN
    -- Get the last_updated timestamp from the materialized view
    SELECT MAX(ab.last_updated) INTO last_update
    FROM account_balances ab;
    
    -- Calculate age in minutes
    age_mins := EXTRACT(EPOCH FROM (NOW() - last_update)) / 60;
    
    RETURN QUERY SELECT 
        last_update,
        age_mins,
        age_mins > 60 as needs_refresh;
END;
$$ LANGUAGE plpgsql;

-- 6. Test the fix
\echo ''
\echo '🧪 Testing manual refresh function...'
SELECT * FROM manual_refresh_account_balances();

\echo ''
\echo '🧪 Testing freshness check function...'
SELECT * FROM check_account_balances_freshness();

-- 7. Show remaining triggers (should not include trg_refresh_account_balances)
\echo ''
\echo '📋 Remaining triggers on unified_journal_lines:'
SELECT 
    tgname as trigger_name,
    tgtype as trigger_type,
    tgenabled as enabled
FROM pg_trigger 
WHERE tgrelid = 'unified_journal_lines'::regclass 
AND tgname NOT LIKE 'RI_%'
ORDER BY tgname;

-- 8. Show balance sync triggers (these should exist)
\echo ''
\echo '📋 Balance sync triggers (should exist):'
SELECT 
    t.tgname as trigger_name,
    c.relname as table_name,
    p.proname as function_name
FROM pg_trigger t
JOIN pg_class c ON t.tgrelid = c.oid
JOIN pg_proc p ON t.tgfoid = p.oid
WHERE t.tgname LIKE '%sync_account_balance%'
ORDER BY c.relname, t.tgname;

-- Success summary
\echo ''
\echo '========================================================================='
\echo '✅ EMERGENCY FIX COMPLETED'
\echo '========================================================================='
\echo ''
\echo 'What was done:'
\echo '  1. Removed problematic trigger: trg_refresh_account_balances'
\echo '  2. Created manual refresh function: manual_refresh_account_balances()'
\echo '  3. Created freshness check function: check_account_balances_freshness()'
\echo ''
\echo 'Next steps:'
\echo '  1. Restart your backend application'
\echo '  2. Test creating transactions (deposit, sales, etc.)'
\echo '  3. Error SQLSTATE 55000 should be gone'
\echo ''
\echo 'For scheduled refresh, see: backend/FIX_CONCURRENT_REFRESH_README.md'
\echo '========================================================================='
