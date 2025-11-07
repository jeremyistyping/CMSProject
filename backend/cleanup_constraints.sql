-- Manual Constraint Cleanup Script
-- Run this SQL script directly in PostgreSQL before starting the application
-- This will clean up all problematic constraints that might interfere with migrations

-- ================================================
-- DROP ALL PROBLEMATIC CONSTRAINTS
-- ================================================

-- Drop constraint if it exists on journal_entries table (even if table doesn't exist)
DO $$
BEGIN
    BEGIN
        ALTER TABLE journal_entries DROP CONSTRAINT IF EXISTS uni_journal_entries_code CASCADE;
        RAISE NOTICE 'Dropped constraint uni_journal_entries_code from journal_entries';
    EXCEPTION
        WHEN undefined_table THEN
            RAISE NOTICE 'Table journal_entries does not exist, skipping constraint drop';
    END;
END $$;

-- Drop similar constraints from other tables
DO $$
BEGIN
    BEGIN
        ALTER TABLE products DROP CONSTRAINT IF EXISTS uni_products_code CASCADE;
        ALTER TABLE products DROP CONSTRAINT IF EXISTS products_code_key CASCADE;
        RAISE NOTICE 'Dropped product code constraints';
    EXCEPTION
        WHEN undefined_table THEN
            RAISE NOTICE 'Table products does not exist, skipping constraint drop';
    END;
END $$;

DO $$
BEGIN
    BEGIN
        ALTER TABLE contacts DROP CONSTRAINT IF EXISTS uni_contacts_code CASCADE;
        ALTER TABLE contacts DROP CONSTRAINT IF EXISTS contacts_code_key CASCADE;
        RAISE NOTICE 'Dropped contact code constraints';
    EXCEPTION
        WHEN undefined_table THEN
            RAISE NOTICE 'Table contacts does not exist, skipping constraint drop';
    END;
END $$;

DO $$
BEGIN
    BEGIN
        ALTER TABLE accounts DROP CONSTRAINT IF EXISTS uni_accounts_code CASCADE;
        ALTER TABLE accounts DROP CONSTRAINT IF EXISTS accounts_code_key CASCADE;
        RAISE NOTICE 'Dropped account code constraints';
    EXCEPTION
        WHEN undefined_table THEN
            RAISE NOTICE 'Table accounts does not exist, skipping constraint drop';
    END;
END $$;

DO $$
BEGIN
    BEGIN
        ALTER TABLE product_units DROP CONSTRAINT IF EXISTS uni_product_units_code CASCADE;
        ALTER TABLE product_units DROP CONSTRAINT IF EXISTS product_units_code_key CASCADE;
        RAISE NOTICE 'Dropped product unit code constraints';
    EXCEPTION
        WHEN undefined_table THEN
            RAISE NOTICE 'Table product_units does not exist, skipping constraint drop';
    END;
END $$;

DO $$
BEGIN
    BEGIN
        ALTER TABLE sales DROP CONSTRAINT IF EXISTS uni_sales_code CASCADE;
        ALTER TABLE sales DROP CONSTRAINT IF EXISTS sales_code_key CASCADE;
        RAISE NOTICE 'Dropped sales code constraints';
    EXCEPTION
        WHEN undefined_table THEN
            RAISE NOTICE 'Table sales does not exist, skipping constraint drop';
    END;
END $$;

DO $$
BEGIN
    BEGIN
        ALTER TABLE purchases DROP CONSTRAINT IF EXISTS uni_purchases_code CASCADE;
        ALTER TABLE purchases DROP CONSTRAINT IF EXISTS purchases_code_key CASCADE;
        RAISE NOTICE 'Dropped purchases code constraints';
    EXCEPTION
        WHEN undefined_table THEN
            RAISE NOTICE 'Table purchases does not exist, skipping constraint drop';
    END;
END $$;

-- ================================================
-- DROP ALL CODE-RELATED CONSTRAINTS FROM ANY TABLE
-- ================================================

-- Find and drop all unique constraints that contain 'code' in their name
DO $$
DECLARE
    constraint_record RECORD;
BEGIN
    FOR constraint_record IN 
        SELECT table_name, constraint_name 
        FROM information_schema.table_constraints 
        WHERE constraint_type = 'UNIQUE'
        AND constraint_name LIKE '%code%'
        AND table_schema = 'public'
    LOOP
        BEGIN
            EXECUTE format('ALTER TABLE %I DROP CONSTRAINT IF EXISTS %I CASCADE', 
                         constraint_record.table_name, 
                         constraint_record.constraint_name);
            RAISE NOTICE 'Dropped constraint % from table %', 
                       constraint_record.constraint_name, 
                       constraint_record.table_name;
        EXCEPTION
            WHEN OTHERS THEN
                RAISE NOTICE 'Failed to drop constraint % from table %: %', 
                           constraint_record.constraint_name, 
                           constraint_record.table_name, 
                           SQLERRM;
        END;
    END LOOP;
END $$;

-- ================================================
-- DROP ALL ORPHANED INDEXES RELATED TO CODE FIELDS
-- ================================================

-- Find and drop all indexes that contain 'code' in their name
DO $$
DECLARE
    index_record RECORD;
BEGIN
    FOR index_record IN 
        SELECT indexname, tablename
        FROM pg_indexes 
        WHERE schemaname = 'public'
        AND indexname LIKE '%code%'
    LOOP
        BEGIN
            EXECUTE format('DROP INDEX IF EXISTS %I CASCADE', index_record.indexname);
            RAISE NOTICE 'Dropped index % from table %', 
                       index_record.indexname, 
                       index_record.tablename;
        EXCEPTION
            WHEN OTHERS THEN
                RAISE NOTICE 'Failed to drop index %: %', 
                           index_record.indexname, 
                           SQLERRM;
        END;
    END LOOP;
END $$;

-- ================================================
-- SUMMARY REPORT
-- ================================================

-- Show remaining constraints (should be clean after this script)
SELECT 
    'REMAINING CONSTRAINTS' as check_type,
    table_name, 
    constraint_name,
    constraint_type
FROM information_schema.table_constraints 
WHERE constraint_name LIKE '%code%'
AND table_schema = 'public'
ORDER BY table_name, constraint_name;

-- Show remaining indexes
SELECT 
    'REMAINING INDEXES' as check_type,
    tablename as table_name, 
    indexname as index_name,
    'INDEX' as constraint_type
FROM pg_indexes 
WHERE schemaname = 'public'
AND indexname LIKE '%code%'
ORDER BY tablename, indexname;

-- Final message
SELECT 'CLEANUP COMPLETED - Ready for application migration' as status;