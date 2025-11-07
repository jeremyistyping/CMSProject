-- Fix audit_logs schema to match both middleware and SQL triggers
-- This migration adds missing 'notes' column and increases 'action' size

-- Add notes column if not exists
ALTER TABLE audit_logs 
ADD COLUMN IF NOT EXISTS notes TEXT;

-- Increase action column size from VARCHAR(20) to VARCHAR(50)
ALTER TABLE audit_logs 
ALTER COLUMN action TYPE VARCHAR(50);

-- Add comment for documentation
COMMENT ON COLUMN audit_logs.notes IS 'Additional notes for audit events, used by SQL triggers';
COMMENT ON COLUMN audit_logs.action IS 'Action type: CREATE, UPDATE, DELETE, etc. (max 50 chars)';

-- Verify changes
SELECT 
    column_name, 
    data_type, 
    character_maximum_length,
    is_nullable
FROM information_schema.columns 
WHERE table_name = 'audit_logs' 
ORDER BY ordinal_position;
