-- +goose Up
-- Create weekly_reports table
CREATE TABLE IF NOT EXISTS weekly_reports (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL,
    week INTEGER NOT NULL CHECK (week >= 1 AND week <= 53),
    year INTEGER NOT NULL CHECK (year >= 2000),
    project_manager VARCHAR(200),
    total_work_days INTEGER DEFAULT 0,
    weather_delays INTEGER DEFAULT 0,
    team_size INTEGER DEFAULT 0,
    accomplishments TEXT,
    challenges TEXT,
    next_week_priorities TEXT,
    generated_date TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by VARCHAR(100),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    CONSTRAINT fk_weekly_reports_project
        FOREIGN KEY (project_id)
        REFERENCES projects(id)
        ON DELETE CASCADE,
    
    -- Ensure only one report per project per week/year
    CONSTRAINT unique_project_week_year
        UNIQUE (project_id, week, year)
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_weekly_reports_project_id ON weekly_reports(project_id);
CREATE INDEX IF NOT EXISTS idx_weekly_reports_year ON weekly_reports(year);
CREATE INDEX IF NOT EXISTS idx_weekly_reports_deleted_at ON weekly_reports(deleted_at);
CREATE INDEX IF NOT EXISTS idx_weekly_reports_project_year ON weekly_reports(project_id, year);

-- Add trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_weekly_reports_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_weekly_reports_updated_at
    BEFORE UPDATE ON weekly_reports
    FOR EACH ROW
    EXECUTE FUNCTION update_weekly_reports_updated_at();

-- +goose Down
-- Drop weekly_reports table and related objects
DROP TRIGGER IF EXISTS trigger_update_weekly_reports_updated_at ON weekly_reports;
DROP FUNCTION IF EXISTS update_weekly_reports_updated_at();
DROP TABLE IF EXISTS weekly_reports;

