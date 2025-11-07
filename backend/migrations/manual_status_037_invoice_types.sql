-- =====================================================
-- Invoice Types System Status Check (Silent when OK)
-- =====================================================
-- File: status_037_invoice_types.sql
-- Purpose: Quick status check that only reports when action is needed

DO $$
DECLARE
    failed_entries INTEGER;
    system_operational BOOLEAN;
    invoice_types_count INTEGER;
    invoice_counters_count INTEGER;
    main_migration_status TEXT;
    show_details BOOLEAN := FALSE;
BEGIN
    -- Get system status
    SELECT COUNT(*) INTO failed_entries
    FROM migration_logs 
    WHERE migration_name = '037_rollback_invoice_types_system.sql' 
    AND status = 'FAILED';
    
    SELECT COUNT(*) INTO invoice_types_count FROM invoice_types WHERE 1=1; -- Handle table not exists
    SELECT COUNT(*) INTO invoice_counters_count FROM invoice_counters WHERE 1=1; -- Handle table not exists
    
    SELECT status INTO main_migration_status
    FROM migration_logs 
    WHERE migration_name = '037_add_invoice_types_system.sql'
    LIMIT 1;
    
    system_operational := EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'invoice_types')
                         AND EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'invoice_counters')
                         AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'sales' AND column_name = 'invoice_type_id');
    
    -- Determine if we need to show details
    show_details := (failed_entries > 0 OR NOT system_operational OR main_migration_status != 'SUCCESS');
    
    IF NOT show_details THEN
        -- System is OK, just show brief status
        RAISE NOTICE '✅ Invoice Types System: OPERATIONAL (%/% types/counters)', invoice_types_count, invoice_counters_count;
        RETURN;
    END IF;
    
    -- System needs attention, show details
    RAISE NOTICE '🔍 INVOICE TYPES SYSTEM STATUS REPORT';
    RAISE NOTICE '=====================================';
    RAISE NOTICE 'Main Migration: %', COALESCE(main_migration_status, 'NOT FOUND');
    RAISE NOTICE 'Failed Rollback Entries: %', failed_entries;
    RAISE NOTICE 'System Operational: %', CASE WHEN system_operational THEN 'YES' ELSE 'NO' END;
    RAISE NOTICE 'Invoice Types: %', invoice_types_count;
    RAISE NOTICE 'Invoice Counters: %', invoice_counters_count;
    RAISE NOTICE '';
    
    IF failed_entries > 0 THEN
        RAISE NOTICE '💡 ACTION: Run quick_fix_037_rollback_error.sql to clean up';
    END IF;
    
    IF NOT system_operational THEN
        RAISE NOTICE '❌ CRITICAL: System not properly installed';
        RAISE NOTICE '💡 ACTION: Re-run 037_add_invoice_types_system.sql';
    END IF;
    
EXCEPTION
    WHEN OTHERS THEN
        -- If tables don't exist, it's expected during fresh install
        IF SQLSTATE = '42P01' THEN -- undefined_table
            RAISE NOTICE 'ℹ️  Invoice types system not yet installed (expected during fresh setup)';
        ELSE
            RAISE NOTICE '⚠️  Status check error: %', SQLERRM;
        END IF;
END $$;

-- Show minimal table output for automated systems
SELECT 
    CASE 
        WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'invoice_types')
        THEN 'INSTALLED'
        ELSE 'NOT_INSTALLED'
    END as system_status,
    CASE 
        WHEN EXISTS (SELECT 1 FROM migration_logs WHERE migration_name = '037_rollback_invoice_types_system.sql' AND status = 'FAILED')
        THEN 'CLEANUP_NEEDED'
        ELSE 'CLEAN'
    END as cleanup_status,
    NOW() as checked_at;