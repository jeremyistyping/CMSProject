-- Migration: Create daily_updates table for construction project daily tracking
-- Purpose: Track daily construction activities, weather, workers, and progress
-- Author: System
-- Date: 2025-11-11

-- Create daily_updates table
CREATE TABLE IF NOT EXISTS daily_updates (
    id SERIAL PRIMARY KEY,
    
    -- Project Information
    project_id INTEGER NOT NULL,
    
    -- Date and Weather
    date TIMESTAMP NOT NULL,
    weather VARCHAR(50) NOT NULL DEFAULT 'Sunny',
    
    -- Work Information
    workers_present INTEGER NOT NULL DEFAULT 0,
    work_description TEXT NOT NULL,
    materials_used TEXT,
    issues TEXT,
    
    -- Photos/Attachments
    photos TEXT[], -- Array of photo URLs/paths
    
    -- Audit Information
    created_by VARCHAR(100) NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    -- Foreign Key Constraint
    CONSTRAINT fk_daily_updates_project FOREIGN KEY (project_id) 
        REFERENCES projects(id) ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_daily_updates_project_id ON daily_updates(project_id);
CREATE INDEX IF NOT EXISTS idx_daily_updates_date ON daily_updates(date);
CREATE INDEX IF NOT EXISTS idx_daily_updates_created_at ON daily_updates(created_at);
CREATE INDEX IF NOT EXISTS idx_daily_updates_deleted_at ON daily_updates(deleted_at);

-- Create composite index for common query patterns
CREATE INDEX IF NOT EXISTS idx_daily_updates_project_date 
    ON daily_updates(project_id, date DESC);

-- Add comments to table and columns
COMMENT ON TABLE daily_updates IS 'Daily construction project updates tracking work, weather, and progress';
COMMENT ON COLUMN daily_updates.project_id IS 'Reference to the project';
COMMENT ON COLUMN daily_updates.date IS 'Date of the daily update';
COMMENT ON COLUMN daily_updates.weather IS 'Weather condition (Sunny, Cloudy, Rainy, Stormy, Partly Cloudy)';
COMMENT ON COLUMN daily_updates.workers_present IS 'Number of workers present on site';
COMMENT ON COLUMN daily_updates.work_description IS 'Description of work completed';
COMMENT ON COLUMN daily_updates.materials_used IS 'Materials used or delivered';
COMMENT ON COLUMN daily_updates.issues IS 'Issues or problems encountered';
COMMENT ON COLUMN daily_updates.photos IS 'Array of photo URLs/paths';
COMMENT ON COLUMN daily_updates.created_by IS 'User who created the daily update';

-- Create a view for daily updates summary
CREATE OR REPLACE VIEW daily_updates_summary AS
SELECT 
    du.project_id,
    p.project_name,
    DATE(du.date) as update_date,
    COUNT(*) as total_updates,
    AVG(du.workers_present) as avg_workers,
    STRING_AGG(DISTINCT du.weather, ', ') as weather_conditions
FROM daily_updates du
LEFT JOIN projects p ON du.project_id = p.id
WHERE du.deleted_at IS NULL
GROUP BY du.project_id, p.project_name, DATE(du.date)
ORDER BY update_date DESC;

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_daily_updates_timestamp()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_daily_updates_timestamp
    BEFORE UPDATE ON daily_updates
    FOR EACH ROW
    EXECUTE FUNCTION update_daily_updates_timestamp();

COMMIT;

