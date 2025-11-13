-- ============================================================================
-- MANUAL MIGRATION FOR TIMELINE SCHEDULES
-- Copy and paste this entire script into pgAdmin Query Tool
-- Database: CMSNew
-- ============================================================================

-- Drop table if exists (for re-running migration)
DROP TABLE IF EXISTS timeline_schedules CASCADE;

-- Create timeline_schedules table
CREATE TABLE timeline_schedules (
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
CREATE INDEX idx_timeline_schedules_project_id ON timeline_schedules(project_id);
CREATE INDEX idx_timeline_schedules_work_area ON timeline_schedules(work_area);
CREATE INDEX idx_timeline_schedules_start_date ON timeline_schedules(start_date);
CREATE INDEX idx_timeline_schedules_end_date ON timeline_schedules(end_date);
CREATE INDEX idx_timeline_schedules_status ON timeline_schedules(status);
CREATE INDEX idx_timeline_schedules_deleted_at ON timeline_schedules(deleted_at);

-- Create composite indexes
CREATE INDEX idx_timeline_schedules_project_dates ON timeline_schedules(project_id, start_date, end_date) WHERE deleted_at IS NULL;
CREATE INDEX idx_timeline_schedules_project_status ON timeline_schedules(project_id, status) WHERE deleted_at IS NULL;

-- Add table comment
COMMENT ON TABLE timeline_schedules IS 'Project timeline schedules for tracking work areas and their time allocation';

-- Add column comments
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

-- Insert sample data (optional)
INSERT INTO timeline_schedules (project_id, work_area, assigned_team, start_date, end_date, start_time, end_time, notes, status) VALUES
(1, 'Site Preparation', 'Team Hardy', '2025-07-28', '2025-07-30', '08:00', '17:00', 'Initial site clearing and preparation', 'not-started');

-- Verify table was created
SELECT 
    table_name, 
    column_name, 
    data_type, 
    is_nullable
FROM 
    information_schema.columns
WHERE 
    table_name = 'timeline_schedules'
ORDER BY 
    ordinal_position;

-- Success message
DO $$
BEGIN
    RAISE NOTICE '✅ Timeline Schedules table created successfully!';
    RAISE NOTICE '✅ Sample data inserted.';
    RAISE NOTICE '✅ Ready to use!';
END $$;

