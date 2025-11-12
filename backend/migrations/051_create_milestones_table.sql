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
    actual_completion_date TIMESTAMP,
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

-- Add comments to columns
COMMENT ON COLUMN milestones.id IS 'Primary key';
COMMENT ON COLUMN milestones.project_id IS 'Foreign key to projects table';
COMMENT ON COLUMN milestones.title IS 'Milestone title/name';
COMMENT ON COLUMN milestones.description IS 'Detailed description of the milestone';
COMMENT ON COLUMN milestones.target_date IS 'Target completion date';
COMMENT ON COLUMN milestones.actual_completion_date IS 'Actual completion date (NULL if not completed)';
COMMENT ON COLUMN milestones.status IS 'Milestone status: pending, in_progress, completed, delayed';
COMMENT ON COLUMN milestones.progress IS 'Progress percentage (0-100)';
COMMENT ON COLUMN milestones.order_number IS 'Order/sequence number for display';
COMMENT ON COLUMN milestones.weight IS 'Weight/importance percentage for weighted progress calculation';
COMMENT ON COLUMN milestones.created_at IS 'Timestamp when milestone was created';
COMMENT ON COLUMN milestones.updated_at IS 'Timestamp when milestone was last updated';
COMMENT ON COLUMN milestones.deleted_at IS 'Soft delete timestamp';

