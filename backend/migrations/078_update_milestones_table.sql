-- Migration: Update milestones table structure
-- Purpose: Add missing columns (work_area and progress)
-- Author: System
-- Date: 2025-12-08

DO $$
BEGIN
    -- Add work_area column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'milestones' AND column_name = 'work_area') THEN
        ALTER TABLE milestones ADD COLUMN work_area VARCHAR(100);
        COMMENT ON COLUMN milestones.work_area IS 'Work area or phase (e.g., Site Preparation, Foundation Work)';
        CREATE INDEX IF NOT EXISTS idx_milestones_work_area ON milestones(work_area);
    END IF;

    -- Add progress column if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name = 'milestones' AND column_name = 'progress') THEN
        ALTER TABLE milestones ADD COLUMN progress DECIMAL(5,2) DEFAULT 0;
        COMMENT ON COLUMN milestones.progress IS 'Progress percentage (0-100)';
    END IF;
END $$;
