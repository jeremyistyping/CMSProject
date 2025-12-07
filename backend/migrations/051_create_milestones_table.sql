-- Migration: Create milestones table for project milestone tracking
-- Version: 051
-- Description: Adds milestones table with full project milestone management support

-- Create milestones table
CREATE TABLE IF NOT EXISTS milestones (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    target_date TIMESTAMP NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    progress INTEGER DEFAULT 0 CHECK (progress >= 0 AND progress <= 100),
    order_number INTEGER DEFAULT 0,
    weight DECIMAL(5,2) DEFAULT 0 CHECK (weight >= 0 AND weight <= 100),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    deleted_at TIMESTAMP,
    CONSTRAINT fk_milestone_project FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_milestones_project_id ON milestones(project_id);
CREATE INDEX IF NOT EXISTS idx_milestones_status ON milestones(status);
CREATE INDEX IF NOT EXISTS idx_milestones_target_date ON milestones(target_date);
CREATE INDEX IF NOT EXISTS idx_milestones_deleted_at ON milestones(deleted_at);

-- Create composite index for common queries
CREATE INDEX IF NOT EXISTS idx_milestones_project_status ON milestones(project_id, status) WHERE deleted_at IS NULL;

-- Add comment to table
COMMENT ON TABLE milestones IS 'Project milestones for tracking project progress and deliverables';
