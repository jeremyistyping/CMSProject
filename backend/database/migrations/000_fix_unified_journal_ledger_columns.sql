-- Fix unified_journal_ledger table missing columns
-- Some installations have the table but missing transaction_uuid column

DO $$
BEGIN
    -- Check if table exists but missing transaction_uuid column
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'unified_journal_ledger')
       AND NOT EXISTS (SELECT 1 FROM information_schema.columns 
                       WHERE table_name = 'unified_journal_ledger' 
                       AND column_name = 'transaction_uuid') THEN
        
        RAISE NOTICE '🔧 Adding missing transaction_uuid column to unified_journal_ledger';
        
        -- Add missing column
        ALTER TABLE unified_journal_ledger 
        ADD COLUMN transaction_uuid UUID UNIQUE DEFAULT uuid_generate_v4();
        
        RAISE NOTICE '✅ Added transaction_uuid column';
        
        -- Update migration log to mark as success
        UPDATE migration_logs 
        SET status = 'SUCCESS', 
            message = 'Fixed missing columns in unified_journal_ledger',
            updated_at = NOW()
        WHERE migration_name = '020_create_unified_journal_ssot.sql'
          AND status = 'FAILED';
          
    ELSIF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'unified_journal_ledger') THEN
        RAISE NOTICE 'ℹ️  Table unified_journal_ledger does not exist - will be created by main migration';
    ELSE
        RAISE NOTICE '✅ Table unified_journal_ledger already has all required columns';
    END IF;
END $$;
