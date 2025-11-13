-- Migration: Add work_area, priority, and assigned_team to milestones table
-- Version: 052
-- Description: Adds new columns for enhanced milestone management

-- Add new columns to milestones table
ALTER TABLE milestones ADD COLUMN IF NOT EXISTS work_area VARCHAR(100);
ALTER TABLE milestones ADD COLUMN IF NOT EXISTS priority VARCHAR(20) DEFAULT 'medium';
ALTER TABLE milestones ADD COLUMN IF NOT EXISTS assigned_team VARCHAR(200);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_milestones_work_area ON milestones(work_area) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_milestones_priority ON milestones(priority) WHERE deleted_at IS NULL;

-- Add comments to new columns
COMMENT ON COLUMN milestones.work_area IS 'Work area/phase (e.g., Site Preparation, Foundation Work, Electrical Installation)';
COMMENT ON COLUMN milestones.priority IS 'Milestone priority: low, medium, high';
COMMENT ON COLUMN milestones.assigned_team IS 'Team assigned to this milestone';

