-- Migration: Create project_budgets table for project-based Budget vs Actual reports
-- Date: 2025-11-14
-- Purpose: Store budget per project per COA account, used by ProjectReportService.GenerateBudgetVsActualReport

CREATE TABLE IF NOT EXISTS project_budgets (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id),
    account_id INTEGER NOT NULL REFERENCES accounts(id),
    estimated_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    UNIQUE(project_id, account_id, deleted_at)
);

-- Indexes to speed up queries
CREATE INDEX IF NOT EXISTS idx_project_budgets_project_id
    ON project_budgets(project_id)
    WHERE deleted_at IS NULL;

CREATE INDEX IF NOT EXISTS idx_project_budgets_account_id
    ON project_budgets(account_id)
    WHERE deleted_at IS NULL;

