-- Migration: Make user_id nullable in activity_logs table
-- Date: 2025-10-23
-- Description: Fix error where anonymous users cannot be logged because user_id is NOT NULL

-- Alter the column to allow NULL values
ALTER TABLE activity_logs 
ALTER COLUMN user_id DROP NOT NULL;

-- Verify the change
SELECT column_name, is_nullable, data_type 
FROM information_schema.columns 
WHERE table_name = 'activity_logs' 
  AND column_name = 'user_id';
