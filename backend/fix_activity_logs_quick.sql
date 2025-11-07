-- Quick fix for activity_logs user_id constraint
-- Makes user_id nullable to allow anonymous user logging

-- Check current state
SELECT column_name, is_nullable, data_type 
FROM information_schema.columns 
WHERE table_name = 'activity_logs' 
  AND column_name = 'user_id';

-- Make user_id nullable
ALTER TABLE activity_logs 
ALTER COLUMN user_id DROP NOT NULL;

-- Verify the change
SELECT column_name, is_nullable, data_type 
FROM information_schema.columns 
WHERE table_name = 'activity_logs' 
  AND column_name = 'user_id';

-- Test insert with NULL user_id
INSERT INTO activity_logs 
(user_id, username, role, method, path, action, resource, status_code, ip_address, duration, created_at)
VALUES 
(NULL, 'anonymous', 'guest', 'GET', '/test', 'test', 'test', 200, '127.0.0.1', 0, NOW());

-- Clean up test data
DELETE FROM activity_logs WHERE username = 'anonymous' AND path = '/test';
