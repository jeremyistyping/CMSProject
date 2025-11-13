-- Migration: Create timeline_schedules table for project timeline management
-- Version: 054
-- Description: Adds timeline_schedules table with full project schedule tracking support

-- Create timeline_schedules table
CREATE TABLE IF NOT EXISTS timeline_schedules (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL,
    work_area VARCHAR(200) NOT NULL,
    assigned_team VARCHAR(200),
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    start_time VARCHAR(10) DEFAULT '08:00',
    end_time VARCHAR(10) DEFAULT '17:00',
    notes TEXT,
    status VARCHAR(20) NOT NULL DEFAULT 'not-started',
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT fk_timeline_schedule_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE,
    CONSTRAINT chk_timeline_dates CHECK (end_date >= start_date),
    CONSTRAINT chk_timeline_status CHECK (status IN ('not-started', 'in-progress', 'completed'))
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_timeline_schedules_project_id ON timeline_schedules(project_id);
CREATE INDEX IF NOT EXISTS idx_timeline_schedules_work_area ON timeline_schedules(work_area);
CREATE INDEX IF NOT EXISTS idx_timeline_schedules_start_date ON timeline_schedules(start_date);
CREATE INDEX IF NOT EXISTS idx_timeline_schedules_end_date ON timeline_schedules(end_date);
CREATE INDEX IF NOT EXISTS idx_timeline_schedules_status ON timeline_schedules(status);
CREATE INDEX IF NOT EXISTS idx_timeline_schedules_deleted_at ON timeline_schedules(deleted_at);

-- Create composite index for common queries
CREATE INDEX IF NOT EXISTS idx_timeline_schedules_project_dates ON timeline_schedules(project_id, start_date, end_date) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_timeline_schedules_project_status ON timeline_schedules(project_id, status) WHERE deleted_at IS NULL;

-- Add comment to table
COMMENT ON TABLE timeline_schedules IS 'Project timeline schedules for tracking work areas and their time allocation';

-- Add comments to columns
COMMENT ON COLUMN timeline_schedules.id IS 'Primary key';
COMMENT ON COLUMN timeline_schedules.project_id IS 'Foreign key to projects table';
COMMENT ON COLUMN timeline_schedules.work_area IS 'Work area/phase name (e.g., Site Preparation, Foundation Work)';
COMMENT ON COLUMN timeline_schedules.assigned_team IS 'Team or contractor assigned to this work area';
COMMENT ON COLUMN timeline_schedules.start_date IS 'Schedule start date';
COMMENT ON COLUMN timeline_schedules.end_date IS 'Schedule end date';
COMMENT ON COLUMN timeline_schedules.start_time IS 'Daily work start time (HH:MM format)';
COMMENT ON COLUMN timeline_schedules.end_time IS 'Daily work end time (HH:MM format)';
COMMENT ON COLUMN timeline_schedules.notes IS 'Additional notes or requirements';
COMMENT ON COLUMN timeline_schedules.status IS 'Schedule status: not-started, in-progress, completed';
COMMENT ON COLUMN timeline_schedules.created_at IS 'Timestamp when schedule was created';
COMMENT ON COLUMN timeline_schedules.updated_at IS 'Timestamp when schedule was last updated';
COMMENT ON COLUMN timeline_schedules.deleted_at IS 'Soft delete timestamp';

