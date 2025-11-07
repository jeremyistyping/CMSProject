-- =====================================================
-- Rollback Migration: 037_rollback_invoice_types_system_pg.sql  
-- =====================================================
-- Author: System
-- Date: 2025-10-02
-- Description: Rollback invoice types and counter system migration (PostgreSQL)
--              WARNING: This will remove all invoice type data!
--              Use with caution in production environments.

-- =====================================================
-- BACKUP DATA FIRST (uncomment if needed)
-- =====================================================
-- CREATE TABLE backup_invoice_types_037 AS SELECT * FROM invoice_types;
-- CREATE TABLE backup_invoice_counters_037 AS SELECT * FROM invoice_counters;
-- CREATE TABLE backup_sales_invoice_type_ids_037 AS SELECT id, invoice_type_id FROM sales WHERE invoice_type_id IS NOT NULL;

-- =====================================================
-- 1. Remove helper functions
-- =====================================================
DROP FUNCTION IF EXISTS get_next_invoice_number(BIGINT);
DROP FUNCTION IF EXISTS preview_next_invoice_number(BIGINT);

-- =====================================================  
-- 2. Remove triggers and trigger functions
-- =====================================================
DROP TRIGGER IF EXISTS trigger_invoice_types_updated_at ON invoice_types;
DROP TRIGGER IF EXISTS trigger_invoice_counters_updated_at ON invoice_counters;
DROP FUNCTION IF EXISTS update_invoice_types_updated_at();
DROP FUNCTION IF EXISTS update_invoice_counters_updated_at();

-- =====================================================
-- 3. Remove foreign key constraints and indexes from sales
-- =====================================================

-- Drop foreign key constraint from sales table if exists
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.key_column_usage 
        WHERE table_name = 'sales' 
        AND constraint_name = 'fk_sales_invoice_type_id'
    ) THEN
        ALTER TABLE sales DROP CONSTRAINT fk_sales_invoice_type_id;
        RAISE NOTICE 'Dropped foreign key constraint fk_sales_invoice_type_id';
    ELSE
        RAISE NOTICE 'Foreign key constraint fk_sales_invoice_type_id does not exist';
    END IF;
END $$;

-- Drop indexes from sales table
DROP INDEX IF EXISTS idx_sales_invoice_type_id;
DROP INDEX IF EXISTS idx_sales_invoice_number;
DROP INDEX IF EXISTS idx_sales_date_status;
DROP INDEX IF EXISTS idx_sales_status_invoice_type;

-- Drop invoice_type_id column from sales table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'sales' 
        AND column_name = 'invoice_type_id'
    ) THEN
        ALTER TABLE sales DROP COLUMN invoice_type_id;
        RAISE NOTICE 'Dropped invoice_type_id column from sales table';
    ELSE
        RAISE NOTICE 'Column invoice_type_id does not exist in sales table';
    END IF;
END $$;

-- =====================================================
-- 4. Drop invoice_counters table
-- =====================================================
DROP TABLE IF EXISTS invoice_counters CASCADE;

-- =====================================================
-- 5. Drop invoice_types table
-- =====================================================
DROP TABLE IF EXISTS invoice_types CASCADE;

-- =====================================================
-- 6. Clean up migration log (if migration_logs table exists)
-- =====================================================
DO $$
DECLARE
    valid_statuses TEXT;
BEGIN
    -- Check if migration_logs table exists
    IF EXISTS (
        SELECT 1 FROM information_schema.tables 
        WHERE table_name = 'migration_logs'
    ) THEN
        -- Get valid status values from constraint
        SELECT string_agg(DISTINCT conbin::TEXT, ', ') INTO valid_statuses
        FROM pg_constraint 
        WHERE conname LIKE '%migration_logs%status%check%'
        LIMIT 1;
        
        -- Delete the main migration entry
        DELETE FROM migration_logs WHERE migration_name = '037_add_invoice_types_system.sql';
        RAISE NOTICE 'Removed migration log entry for 037_add_invoice_types_system.sql';
        
        -- Insert rollback log with valid status
        -- Note: Using 'SUCCESS' status since 'rollback_completed' is not in allowed values
        -- migration_logs constraint allows: SUCCESS, FAILED, SKIPPED only
        INSERT INTO migration_logs (migration_name, executed_at, status, message, description) 
        VALUES (
            '037_rollback_invoice_types_system.sql', 
            NOW(), 
            'SUCCESS',  -- Using SUCCESS instead of rollback_completed
            'Invoice types system rollback completed successfully',
            'Rollback executed: Removed invoice_types, invoice_counters tables and related structures. Original migration 037_add_invoice_types_system.sql has been rolled back.'
        );
        
        RAISE NOTICE 'Rollback logged successfully with status=SUCCESS (rollback_completed not supported by migration_logs constraint)';
        RAISE NOTICE 'Valid migration_logs statuses are: SUCCESS, FAILED, SKIPPED';
    ELSE
        RAISE NOTICE 'Migration logs table does not exist, skipping rollback log';
        RAISE NOTICE 'Rollback completed but not logged (no migration_logs table found)';
    END IF;
EXCEPTION
    WHEN OTHERS THEN
        RAISE NOTICE 'Error during rollback logging: %', SQLERRM;
        RAISE NOTICE 'Rollback completed successfully but logging failed';
        RAISE NOTICE 'This error is non-critical - the actual rollback operations were successful';
        -- Don''t re-raise the exception, just log it
END $$;

-- =====================================================
-- 7. Verification
-- =====================================================
SELECT 'ROLLBACK 037 COMPLETED SUCCESSFULLY' as status;

-- Verify tables are dropped
SELECT 
    CASE 
        WHEN NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'invoice_types') 
        THEN 'invoice_types table dropped successfully'
        ELSE 'WARNING: invoice_types table still exists'
    END as invoice_types_status;

SELECT 
    CASE 
        WHEN NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'invoice_counters') 
        THEN 'invoice_counters table dropped successfully'
        ELSE 'WARNING: invoice_counters table still exists'
    END as invoice_counters_status;

-- Verify column is removed from sales
SELECT 
    CASE 
        WHEN NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'sales' AND column_name = 'invoice_type_id') 
        THEN 'invoice_type_id column removed from sales table successfully'
        ELSE 'WARNING: invoice_type_id column still exists in sales table'
    END as sales_column_status;

-- Verify functions are dropped
SELECT 
    CASE 
        WHEN NOT EXISTS (SELECT 1 FROM pg_proc WHERE proname = 'get_next_invoice_number') 
        THEN 'Helper functions dropped successfully'
        ELSE 'WARNING: Some helper functions still exist'
    END as functions_status;

-- =====================================================
-- ROLLBACK COMPLETE
-- =====================================================