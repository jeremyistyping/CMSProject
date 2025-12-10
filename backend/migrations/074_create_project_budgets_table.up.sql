-- Migration: Create Project Budgets Table
-- Version: 074
-- Description: Create project_budgets table to store budget allocation per project per COA account

-- Create project_budgets table
CREATE TABLE IF NOT EXISTS project_budgets (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    account_id INTEGER NOT NULL REFERENCES coa_accounts(id) ON DELETE RESTRICT,
    estimated_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Constraints
    CONSTRAINT project_budgets_unique_project_account UNIQUE (project_id, account_id),
    CONSTRAINT project_budgets_positive_amount CHECK (estimated_amount >= 0)
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_project_budgets_project_id ON project_budgets(project_id);
CREATE INDEX IF NOT EXISTS idx_project_budgets_account_id ON project_budgets(account_id);
CREATE INDEX IF NOT EXISTS idx_project_budgets_deleted_at ON project_budgets(deleted_at);

-- Add comments
COMMENT ON TABLE project_budgets IS 'Budget allocation per project per COA account';
COMMENT ON COLUMN project_budgets.project_id IS 'Reference to project';
COMMENT ON COLUMN project_budgets.account_id IS 'Reference to COA account';
COMMENT ON COLUMN project_budgets.estimated_amount IS 'Estimated budget amount for this project-account combination';
COMMENT ON COLUMN project_budgets.deleted_at IS 'Soft delete timestamp';

-- Create trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_project_budgets_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_project_budgets_updated_at
    BEFORE UPDATE ON project_budgets
    FOR EACH ROW
    EXECUTE FUNCTION update_project_budgets_updated_at();
