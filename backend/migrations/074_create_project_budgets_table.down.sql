-- Rollback: Drop Project Budgets Table
-- Version: 074

-- Drop trigger
DROP TRIGGER IF EXISTS trigger_update_project_budgets_updated_at ON project_budgets;

-- Drop function
DROP FUNCTION IF EXISTS update_project_budgets_updated_at();

-- Drop table
DROP TABLE IF EXISTS project_budgets;
