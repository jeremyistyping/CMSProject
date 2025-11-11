-- Migration: Integrate Project Management with Purchase for Cost Control
-- Date: 2025-11-11
-- Purpose: Link purchases to projects and add cost tracking fields

-- Add project_id to purchases table
ALTER TABLE purchases 
ADD COLUMN IF NOT EXISTS project_id INTEGER REFERENCES projects(id);

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_purchases_project_id ON purchases(project_id);

-- Add cost tracking fields to projects table
ALTER TABLE projects
ADD COLUMN IF NOT EXISTS actual_cost DECIMAL(20,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS material_cost DECIMAL(20,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS labor_cost DECIMAL(20,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS equipment_cost DECIMAL(20,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS overhead_cost DECIMAL(20,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS variance DECIMAL(20,2) DEFAULT 0,
ADD COLUMN IF NOT EXISTS variance_percent DECIMAL(5,2) DEFAULT 0;

-- Add comments
COMMENT ON COLUMN purchases.project_id IS 'Link to project for cost tracking and budget monitoring';
COMMENT ON COLUMN projects.actual_cost IS 'Total actual cost spent on project (sum of all purchases)';
COMMENT ON COLUMN projects.material_cost IS 'Total material cost';
COMMENT ON COLUMN projects.labor_cost IS 'Total labor cost';
COMMENT ON COLUMN projects.equipment_cost IS 'Total equipment cost';
COMMENT ON COLUMN projects.overhead_cost IS 'Total overhead cost';
COMMENT ON COLUMN projects.variance IS 'Budget variance (Budget - Actual Cost)';
COMMENT ON COLUMN projects.variance_percent IS 'Budget variance percentage ((Variance/Budget)*100)';

-- Update variance for existing projects
UPDATE projects 
SET variance = budget - actual_cost,
    variance_percent = CASE 
        WHEN budget > 0 THEN ((budget - actual_cost) / budget) * 100
        ELSE 0 
    END
WHERE budget > 0;
