-- Migration: Add account_id column to project_budgets table
-- Version: 075
-- Description: Add account_id column to link project_budgets with coa_accounts

-- Add account_id column if not exists
ALTER TABLE project_budgets 
ADD COLUMN IF NOT EXISTS account_id INTEGER REFERENCES coa_accounts(id) ON DELETE RESTRICT;

-- Make category column nullable (for backward compatibility)
ALTER TABLE project_budgets 
ALTER COLUMN category DROP NOT NULL;

-- Create index for performance
CREATE INDEX IF NOT EXISTS idx_project_budgets_account_id ON project_budgets(account_id);

-- Add comment
COMMENT ON COLUMN project_budgets.account_id IS 'Reference to COA account for budget allocation';

-- Note: The old 'category' column is kept for backward compatibility
-- New records should use account_id instead of category
-- Future migrations may migrate data from category to account_id and drop category column
