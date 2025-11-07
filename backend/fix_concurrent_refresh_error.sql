-- ========================================================================
-- FIX: CONCURRENT MATERIALIZED VIEW REFRESH ERROR (SQLSTATE 55000)
-- ========================================================================
-- Problem: trg_refresh_account_balances trigger causes concurrent refresh conflicts
-- Solution: Drop the trigger and use manual refresh strategy
-- Balance sync is handled by setup_automatic_balance_sync.sql triggers
-- ========================================================================

BEGIN;

-- 1. Drop the problematic trigger
DROP TRIGGER IF EXISTS trg_refresh_account_balances ON unified_journal_lines;

-- 2. Verify trigger is dropped
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM pg_trigger 
        WHERE tgname = 'trg_refresh_account_balances'
    ) THEN
        RAISE EXCEPTION 'Failed to drop trigger trg_refresh_account_balances';
    ELSE
        RAISE NOTICE '✅ Trigger trg_refresh_account_balances dropped successfully';
    END IF;
END $$;

-- 3. Create manual refresh function (for scheduled jobs or API calls)
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

COMMENT ON FUNCTION manual_refresh_account_balances() IS 
'Manually refresh account_balances materialized view. Use this for scheduled jobs or API calls.';

-- 4. Create function to check if refresh is needed
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
        age_mins > 60 as needs_refresh; -- Suggest refresh if older than 1 hour
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION check_account_balances_freshness() IS 
'Check how old the account_balances materialized view is and if it needs refresh.';

COMMIT;

-- Log the fix
INSERT INTO migration_logs (migration_name, status, executed_at, message, description) 
VALUES (
    'fix_concurrent_refresh_error', 
    'SUCCESS', 
    NOW(), 
    'Fixed SQLSTATE 55000 error by removing auto-refresh trigger. Use manual_refresh_account_balances() for scheduled refresh.',
    'Removed trg_refresh_account_balances trigger. Added manual_refresh_account_balances() and check_account_balances_freshness() functions.'
) ON CONFLICT (migration_name) DO UPDATE SET
    status = 'SUCCESS',
    executed_at = NOW(),
    message = 'Fixed SQLSTATE 55000 error by removing auto-refresh trigger. Use manual_refresh_account_balances() for scheduled refresh.';

-- Display success message and usage instructions
DO $$
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '=========================================================================';
    RAISE NOTICE '✅ CONCURRENT REFRESH ERROR FIX APPLIED SUCCESSFULLY';
    RAISE NOTICE '=========================================================================';
    RAISE NOTICE '';
    RAISE NOTICE 'What was fixed:';
    RAISE NOTICE '  - Removed trigger: trg_refresh_account_balances';
    RAISE NOTICE '  - This eliminates SQLSTATE 55000 concurrent refresh conflicts';
    RAISE NOTICE '';
    RAISE NOTICE 'Balance synchronization:';
    RAISE NOTICE '  - Real-time sync: Handled by setup_automatic_balance_sync.sql triggers';
    RAISE NOTICE '  - accounts.balance field is always up-to-date';
    RAISE NOTICE '';
    RAISE NOTICE 'Materialized view refresh (for reporting):';
    RAISE NOTICE '  - Manual refresh: SELECT * FROM manual_refresh_account_balances();';
    RAISE NOTICE '  - Check freshness: SELECT * FROM check_account_balances_freshness();';
    RAISE NOTICE '  - Recommended: Schedule refresh every 1 hour via cron or API';
    RAISE NOTICE '';
    RAISE NOTICE 'Example usage:';
    RAISE NOTICE '  -- Refresh now';
    RAISE NOTICE '  SELECT * FROM manual_refresh_account_balances();';
    RAISE NOTICE '';
    RAISE NOTICE '  -- Check if refresh needed';
    RAISE NOTICE '  SELECT * FROM check_account_balances_freshness();';
    RAISE NOTICE '=========================================================================';
    RAISE NOTICE '';
END $$;

-- Verify the fix
SELECT 
    '✅ Fix Complete' as status,
    'Concurrent refresh error resolved' as result,
    'Use manual_refresh_account_balances() for scheduled refresh' as next_step;
