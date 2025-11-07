-- =======================================================================================
-- Fix audit_logs Table Structure for CashBank Sync Compatibility
-- =======================================================================================
-- This script adds missing columns and updates the audit_logs table structure
-- to be compatible with CashBank sync triggers

-- Check and add missing 'notes' column if it doesn't exist
DO $$
BEGIN
    -- Add notes column if it doesn't exist
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns 
        WHERE table_name = 'audit_logs' AND column_name = 'notes'
    ) THEN
        ALTER TABLE audit_logs ADD COLUMN notes TEXT;
        RAISE NOTICE 'Added notes column to audit_logs table';
    ELSE
        RAISE NOTICE 'notes column already exists in audit_logs table';
    END IF;
END $$;

-- Update existing audit_logs table to ensure all required columns exist
-- and have correct data types for our sync triggers

-- Update column types to match what our triggers expect
DO $$
BEGIN
    -- Ensure old_values and new_values are JSONB (not just TEXT)
    BEGIN
        ALTER TABLE audit_logs ALTER COLUMN old_values TYPE JSONB USING old_values::JSONB;
        RAISE NOTICE 'Updated old_values column to JSONB type';
    EXCEPTION
        WHEN OTHERS THEN
            RAISE NOTICE 'old_values column already JSONB or conversion failed: %', SQLERRM;
    END;
    
    BEGIN
        ALTER TABLE audit_logs ALTER COLUMN new_values TYPE JSONB USING new_values::JSONB;
        RAISE NOTICE 'Updated new_values column to JSONB type';
    EXCEPTION
        WHEN OTHERS THEN
            RAISE NOTICE 'new_values column already JSONB or conversion failed: %', SQLERRM;
    END;
END $$;

-- Create indexes for better performance if they don't exist
CREATE INDEX IF NOT EXISTS idx_audit_logs_table_action_v2 ON audit_logs (table_name, action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at_v2 ON audit_logs (created_at);

-- Test insert to verify the structure works
DO $$
BEGIN
    INSERT INTO audit_logs (
        table_name, 
        action, 
        record_id, 
        old_values, 
        new_values,
        created_at,
        notes
    ) VALUES (
        'test_audit_structure',
        'STRUCTURE_TEST',
        1,
        '{"test": "old_value"}',
        '{"test": "new_value"}',
        NOW(),
        'Testing audit_logs structure compatibility'
    );
    
    RAISE NOTICE 'Successfully inserted test record into audit_logs';
    
    -- Clean up test record
    DELETE FROM audit_logs WHERE table_name = 'test_audit_structure';
    RAISE NOTICE 'Test record cleaned up';
    
EXCEPTION
    WHEN OTHERS THEN
        RAISE EXCEPTION 'Failed to test audit_logs structure: %', SQLERRM;
END $$;

-- =======================================================================================
-- Update CashBank Sync Triggers to Handle Missing Notes Column Gracefully
-- =======================================================================================

-- Update the sync function to be more robust with audit logging
CREATE OR REPLACE FUNCTION sync_coa_to_cashbank_robust()
RETURNS TRIGGER AS $$
BEGIN
    -- Only process if balance actually changed
    IF OLD.balance IS DISTINCT FROM NEW.balance THEN
        
        -- Update all linked CashBanks
        UPDATE cash_banks 
        SET balance = NEW.balance, updated_at = NOW()
        WHERE account_id = NEW.id 
        AND deleted_at IS NULL 
        AND is_active = true;
        
        -- Robust audit log insert - try with notes first, fallback without
        BEGIN
            INSERT INTO audit_logs (
                table_name, 
                action, 
                record_id, 
                old_values, 
                new_values,
                created_at,
                notes
            ) VALUES (
                'coa_to_cashbank_sync',
                'COA_BALANCE_CHANGED',
                NEW.id,
                json_build_object('old_balance', OLD.balance),
                json_build_object('new_balance', NEW.balance, 'account_code', NEW.code),
                NOW(),
                'COA balance changed - syncing to CashBank'
            );
        EXCEPTION
            WHEN undefined_column THEN
                -- Fallback: insert without notes column
                INSERT INTO audit_logs (
                    table_name, 
                    action, 
                    record_id, 
                    old_values, 
                    new_values,
                    created_at
                ) VALUES (
                    'coa_to_cashbank_sync',
                    'COA_BALANCE_CHANGED',
                    NEW.id,
                    json_build_object('old_balance', OLD.balance),
                    json_build_object('new_balance', NEW.balance, 'account_code', NEW.code),
                    NOW()
                );
        END;
        
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Replace the trigger with the robust version
DROP TRIGGER IF EXISTS trg_sync_coa_to_cashbank ON accounts;
CREATE TRIGGER trg_sync_coa_to_cashbank
    AFTER UPDATE ON accounts
    FOR EACH ROW
    EXECUTE FUNCTION sync_coa_to_cashbank_robust();

-- =======================================================================================
-- Verify Table Structure
-- =======================================================================================

-- Show current audit_logs structure
SELECT 
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_name = 'audit_logs'
ORDER BY ordinal_position;
