-- Fix audit_logs table to handle NULL user_id values
-- This script ensures the table is compatible with anonymous/system operations
-- Date: 2025-10-22

BEGIN;

-- Step 1: Check current state
SELECT 
    COUNT(*) as total_rows,
    COUNT(user_id) as non_null_user_id,
    COUNT(*) - COUNT(user_id) as null_user_id
FROM audit_logs;

-- Step 2: Drop NOT NULL constraint if exists
ALTER TABLE audit_logs 
    ALTER COLUMN user_id DROP NOT NULL;

-- Step 3: Drop and recreate foreign key constraint to allow NULL
ALTER TABLE audit_logs 
    DROP CONSTRAINT IF EXISTS fk_audit_logs_user,
    DROP CONSTRAINT IF EXISTS fk_users_created_audit_logs,
    DROP CONSTRAINT IF EXISTS audit_logs_user_id_fkey;

-- Step 4: Add foreign key that allows NULL values
ALTER TABLE audit_logs 
    ADD CONSTRAINT fk_audit_logs_user 
    FOREIGN KEY (user_id) 
    REFERENCES users(id) 
    ON DELETE SET NULL;

-- Step 5: Add comment
COMMENT ON COLUMN audit_logs.user_id IS 'ID of the user who performed the action (NULL for anonymous/system users)';

-- Step 6: Verify the fix
SELECT 
    column_name, 
    data_type, 
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_name = 'audit_logs' 
    AND column_name = 'user_id';

COMMIT;

-- Summary
SELECT 'Audit logs table fixed - user_id is now nullable for anonymous/system operations' as status;

