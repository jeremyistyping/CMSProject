-- Migration: Add tomorrows_plan column to daily_updates table
-- Purpose: Add field for planning tomorrow's work activities
-- Author: System
-- Date: 2025-11-12

-- Add tomorrows_plan column to daily_updates table
ALTER TABLE daily_updates 
ADD COLUMN IF NOT EXISTS tomorrows_plan TEXT;

-- Add comment to the new column
COMMENT ON COLUMN daily_updates.tomorrows_plan IS 'Planned activities and tasks for tomorrow';

COMMIT;

