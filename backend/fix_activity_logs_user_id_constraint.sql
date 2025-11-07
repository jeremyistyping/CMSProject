-- Fix user_id NOT NULL constraint in activity_logs table
-- This allows logging for anonymous/unauthenticated users

BEGIN;

-- Drop existing foreign key constraint if it exists
ALTER TABLE activity_logs 
    DROP CONSTRAINT IF EXISTS fk_activity_logs_user;

-- Make user_id nullable
ALTER TABLE activity_logs 
    ALTER COLUMN user_id DROP NOT NULL;

-- Re-add foreign key constraint that allows NULL values
ALTER TABLE activity_logs 
    ADD CONSTRAINT fk_activity_logs_user 
    FOREIGN KEY (user_id) 
    REFERENCES users(id) 
    ON DELETE CASCADE;

-- Add comment
COMMENT ON COLUMN activity_logs.user_id IS 'ID of the user who performed the action (NULL for anonymous/unauthenticated users)';

COMMIT;
