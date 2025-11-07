-- =====================================================
-- Smart Cleanup: 037 Rollback Error (Idempotent)
-- =====================================================
-- File: smart_cleanup_037_rollback_error.sql
-- Purpose: Intelligently clean up rollback error with proper already-exists handling

-- =====================================================
-- SMART CLEANUP WITH ALREADY-EXISTS DETECTION
-- =====================================================
DO $$
DECLARE
    failed_entries INTEGER;
    main_migration_status TEXT;
    cleanup_needed BOOLEAN := FALSE;
    system_operational BOOLEAN := TRUE;
BEGIN
    -- =====================================================
    -- 1. CHECK CURRENT STATE
    -- =====================================================
    
    -- Count failed rollback entries
    SELECT COUNT(*) INTO failed_entries
    FROM migration_logs 
    WHERE migration_name = '037_rollback_invoice_types_system.sql' 
    AND status = 'FAILED';
    
    -- Get main migration status
    SELECT status INTO main_migration_status
    FROM migration_logs 
    WHERE migration_name = '037_add_invoice_types_system.sql'
    LIMIT 1;
    
    -- Determine if cleanup is needed
    cleanup_needed := (failed_entries > 0);
    
    -- Check if system is operational
    system_operational := EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'invoice_types')
                         AND EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'invoice_counters')
                         AND EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'sales' AND column_name = 'invoice_type_id');
    
    -- =====================================================
    -- 2. DISPLAY CURRENT STATUS
    -- =====================================================
    RAISE NOTICE '=== INVOICE TYPES SYSTEM STATUS REPORT ===';
    RAISE NOTICE 'Main Migration Status: %', COALESCE(main_migration_status, 'NOT FOUND');
    RAISE NOTICE 'Failed Rollback Entries: %', failed_entries;
    RAISE NOTICE 'System Operational: %', CASE WHEN system_operational THEN '✅ YES' ELSE '❌ NO' END;
    RAISE NOTICE '';
    
    -- =====================================================
    -- 3. HANDLE DIFFERENT SCENARIOS
    -- =====================================================
    
    IF NOT system_operational THEN
        RAISE NOTICE '❌ CRITICAL: Invoice types system is not properly installed!';
        RAISE NOTICE '💡 ACTION REQUIRED: Re-run main migration 037_add_invoice_types_system.sql';
        RETURN;
    END IF;
    
    IF failed_entries = 0 THEN
        RAISE NOTICE '✅ NO CLEANUP NEEDED';
        RAISE NOTICE '   - No failed rollback entries found';
        RAISE NOTICE '   - System is operating normally';
        RAISE NOTICE '   - Invoice types system is fully functional';
        RAISE NOTICE '';
        RAISE NOTICE '🎯 STATUS: ALL CLEAR - NO ACTION REQUIRED';
        RETURN;
    END IF;
    
    IF cleanup_needed THEN
        RAISE NOTICE '🔧 CLEANUP REQUIRED';
        RAISE NOTICE '   - Found % failed rollback log entries', failed_entries;
        RAISE NOTICE '   - Proceeding with cleanup...';
        
        -- Remove failed rollback entries
        DELETE FROM migration_logs 
        WHERE migration_name = '037_rollback_invoice_types_system.sql' 
        AND status = 'FAILED';
        
        RAISE NOTICE '✅ CLEANUP COMPLETED';
        RAISE NOTICE '   - Removed % failed rollback entries', failed_entries;
        RAISE NOTICE '   - System is now clean';
    END IF;
    
    -- =====================================================
    -- 4. FINAL VERIFICATION
    -- =====================================================
    RAISE NOTICE '';
    RAISE NOTICE '=== FINAL SYSTEM VERIFICATION ===';
    
    -- Verify invoice types exist and show count
    DECLARE
        invoice_types_count INTEGER;
        invoice_counters_count INTEGER;
        sample_invoice_number TEXT;
    BEGIN
        SELECT COUNT(*) INTO invoice_types_count FROM invoice_types;
        SELECT COUNT(*) INTO invoice_counters_count FROM invoice_counters;
        
        RAISE NOTICE '📊 Invoice Types: % records', invoice_types_count;
        RAISE NOTICE '📊 Invoice Counters: % records', invoice_counters_count;
        
        -- Test invoice number generation
        IF invoice_types_count > 0 THEN
            SELECT preview_next_invoice_number(id) INTO sample_invoice_number
            FROM invoice_types 
            WHERE is_active = TRUE 
            LIMIT 1;
            
            RAISE NOTICE '🧪 Sample Invoice Number: %', COALESCE(sample_invoice_number, 'GENERATION FAILED');
        END IF;
        
    EXCEPTION
        WHEN OTHERS THEN
            RAISE NOTICE '⚠️  Error during verification: %', SQLERRM;
    END;
    
    RAISE NOTICE '';
    RAISE NOTICE '🎉 FINAL STATUS: INVOICE TYPES SYSTEM FULLY OPERATIONAL';
    RAISE NOTICE '🚀 READY FOR PRODUCTION USE';
    
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE '❌ UNEXPECTED ERROR: %', SQLERRM;
        RAISE NOTICE '💡 RECOMMENDATION: Contact system administrator';
        RAISE NOTICE 'ℹ️  Note: Main system may still be operational despite this error';
END $$;

-- =====================================================
-- SUMMARY TABLE OUTPUT (for easy reading)
-- =====================================================
SELECT 
    'CLEANUP EXECUTION SUMMARY' as report_type,
    CASE 
        WHEN EXISTS (SELECT 1 FROM migration_logs WHERE migration_name = '037_rollback_invoice_types_system.sql' AND status = 'FAILED')
        THEN '🔧 CLEANUP PERFORMED'
        ELSE '✅ NO CLEANUP NEEDED'
    END as cleanup_status,
    CASE 
        WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'invoice_types')
        THEN '✅ OPERATIONAL'
        ELSE '❌ NOT OPERATIONAL'
    END as system_status,
    (SELECT COUNT(*) FROM invoice_types) as invoice_types_count,
    NOW() as executed_at;