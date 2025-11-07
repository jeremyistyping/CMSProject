-- Migration: Lock Critical Tax Accounts
-- Description: Add is_system_critical flag and trigger to prevent modification of critical accounts
-- Version: 040
-- Created: 2025-10-22
-- PostgreSQL Compatible Version

-- ============================================================================
-- STEP 1: Add is_system_critical column to accounts table
-- ============================================================================

DO $$ 
BEGIN
    -- Check if column doesn't exist, then add it
    IF NOT EXISTS (
        SELECT 1 
        FROM information_schema.columns 
        WHERE table_name = 'accounts' 
        AND column_name = 'is_system_critical'
    ) THEN
        ALTER TABLE accounts 
        ADD COLUMN is_system_critical BOOLEAN DEFAULT FALSE;
        
        RAISE NOTICE '✅ Added is_system_critical column to accounts table';
    ELSE
        RAISE NOTICE 'ℹ️  Column is_system_critical already exists';
    END IF;
END $$;

-- ============================================================================
-- STEP 2: Mark critical accounts (cannot be changed)
-- ============================================================================

DO $$
BEGIN
    -- Mark Level 1: ABSOLUTE CRITICAL accounts (4 accounts)
    -- These accounts are used in ALL sales/purchase transactions
    UPDATE accounts 
    SET is_system_critical = TRUE 
    WHERE code IN (
        '1201',  -- Piutang Usaha (Sales Receivable)
        '4101',  -- Pendapatan Penjualan (Sales Revenue)
        '2103',  -- PPN Keluaran (Output VAT)
        '2001'   -- Hutang Usaha (Purchase Payable)
    )
    AND deleted_at IS NULL;
    
    RAISE NOTICE '✅ Marked 4 critical accounts (1201, 4101, 2103, 2001) as system_critical';
    
    -- Also lock PPN Masukan (Input VAT) for tax compliance
    UPDATE accounts 
    SET is_system_critical = TRUE 
    WHERE code IN (
        '1240'   -- PPN Masukan (Input VAT) - Critical for tax remittance
    )
    AND deleted_at IS NULL;
    
    RAISE NOTICE '✅ Also marked account 1240 (PPN Masukan) as system_critical';
END $$;

-- ============================================================================
-- STEP 3: Create trigger function to prevent critical account modifications
-- ============================================================================

-- Drop existing trigger and function if exists
DROP TRIGGER IF EXISTS check_critical_account_update ON accounts;
DROP FUNCTION IF EXISTS prevent_critical_account_update();

-- Create trigger function
CREATE OR REPLACE FUNCTION prevent_critical_account_update()
RETURNS TRIGGER AS $$
BEGIN
    -- Only check if account is marked as critical
    IF OLD.is_system_critical = TRUE THEN
        
        -- Prevent changing critical fields
        IF NEW.code != OLD.code THEN
            RAISE EXCEPTION 
                '🔒 BLOCKED: Cannot change account code for system critical account: % (%) - This account is essential for system integrity',
                OLD.code, OLD.name
                USING HINT = 'Critical accounts are locked to prevent data corruption. Contact system administrator if changes are absolutely necessary.';
        END IF;
        
        IF NEW.type != OLD.type THEN
            RAISE EXCEPTION 
                '🔒 BLOCKED: Cannot change account type for system critical account: % (%) from % to %',
                OLD.code, OLD.name, OLD.type, NEW.type
                USING HINT = 'Changing account type would break journal entries and reports.';
        END IF;
        
        IF NEW.category != OLD.category THEN
            RAISE EXCEPTION 
                '🔒 BLOCKED: Cannot change account category for system critical account: % (%) from % to %',
                OLD.code, OLD.name, OLD.category, NEW.category
                USING HINT = 'Changing account category would affect financial statements.';
        END IF;
        
        -- Prevent deactivation
        IF NEW.is_active = FALSE AND OLD.is_active = TRUE THEN
            RAISE EXCEPTION 
                '🔒 BLOCKED: Cannot deactivate system critical account: % (%)',
                OLD.code, OLD.name
                USING HINT = 'This account is actively used in transactions and cannot be deactivated.';
        END IF;
        
        -- Prevent soft deletion
        IF NEW.deleted_at IS NOT NULL AND OLD.deleted_at IS NULL THEN
            RAISE EXCEPTION 
                '🔒 BLOCKED: Cannot delete system critical account: % (%)',
                OLD.code, OLD.name
                USING HINT = 'This account is essential for system operations and cannot be deleted.';
        END IF;
        
        -- Log warning if name is changed (allow but log)
        IF NEW.name != OLD.name THEN
            RAISE NOTICE '⚠️  WARNING: Changing name of critical account % from "%" to "%"', 
                OLD.code, OLD.name, NEW.name;
        END IF;
        
        -- Allow balance updates and other non-critical field changes
        RAISE NOTICE 'ℹ️  Updating critical account % (%): Non-critical fields changed', OLD.code, OLD.name;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger
CREATE TRIGGER check_critical_account_update
    BEFORE UPDATE ON accounts
    FOR EACH ROW
    EXECUTE FUNCTION prevent_critical_account_update();

DO $$
BEGIN
    RAISE NOTICE '✅ Created trigger to protect critical accounts';
END $$;

-- ============================================================================
-- STEP 4: Verification
-- ============================================================================

DO $$
DECLARE
    critical_count INTEGER;
    account_record RECORD;
BEGIN
    -- Count critical accounts
    SELECT COUNT(*) INTO critical_count
    FROM accounts
    WHERE is_system_critical = TRUE
    AND deleted_at IS NULL;
    
    RAISE NOTICE '';
    RAISE NOTICE '====================================';
    RAISE NOTICE '  CRITICAL ACCOUNTS SUMMARY';
    RAISE NOTICE '====================================';
    RAISE NOTICE 'Total critical accounts: %', critical_count;
    RAISE NOTICE '';
    
    -- List all critical accounts
    FOR account_record IN 
        SELECT code, name, type, category
        FROM accounts
        WHERE is_system_critical = TRUE
        AND deleted_at IS NULL
        ORDER BY code
    LOOP
        RAISE NOTICE '🔒 % - % (% / %)', 
            account_record.code, 
            account_record.name, 
            account_record.type,
            account_record.category;
    END LOOP;
    
    RAISE NOTICE '';
    RAISE NOTICE '✅ Migration 040 completed successfully!';
    RAISE NOTICE '';
    RAISE NOTICE 'ℹ️  What this migration does:';
    RAISE NOTICE '   1. Added is_system_critical column to accounts';
    RAISE NOTICE '   2. Marked 4 critical accounts (1201, 4101, 2103, 2001)';
    RAISE NOTICE '   3. Created trigger to prevent code/type/category changes';
    RAISE NOTICE '   4. Prevents deactivation and deletion of critical accounts';
    RAISE NOTICE '';
    RAISE NOTICE '⚠️  NOTE: Account name and balance CAN still be updated';
    RAISE NOTICE '⚠️  NOTE: To unlock, set is_system_critical = FALSE manually';
    RAISE NOTICE '====================================';
END $$;

