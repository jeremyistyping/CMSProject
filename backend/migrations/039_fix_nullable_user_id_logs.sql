-- Migration: Fix nullable user_id in activity_logs and audit_logs
-- Purpose: Allow logging for unauthenticated users without FK violations
-- Date: 2025-10-20

-- Fix activity_logs table
ALTER TABLE activity_logs 
    DROP CONSTRAINT IF EXISTS fk_activity_logs_user,
    ALTER COLUMN user_id DROP NOT NULL;

-- Re-add foreign key constraint that allows NULL
ALTER TABLE activity_logs 
    ADD CONSTRAINT fk_activity_logs_user 
    FOREIGN KEY (user_id) 
    REFERENCES users(id) 
    ON DELETE CASCADE;

-- Fix audit_logs table
ALTER TABLE audit_logs 
    DROP CONSTRAINT IF EXISTS fk_users_created_audit_logs,
    ALTER COLUMN user_id DROP NOT NULL;

-- Re-add foreign key constraint that allows NULL
ALTER TABLE audit_logs 
    ADD CONSTRAINT fk_users_created_audit_logs 
    FOREIGN KEY (user_id) 
    REFERENCES users(id) 
    ON DELETE CASCADE;

COMMENT ON COLUMN activity_logs.user_id IS 'ID of the user who performed the action (NULL for anonymous/unauthenticated users)';
COMMENT ON COLUMN audit_logs.user_id IS 'ID of the user who performed the action (NULL for anonymous/unauthenticated users)';

COMMIT;
