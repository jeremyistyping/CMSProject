-- Migration: Create Expense Transactions Table
-- Version: 072
-- Description: Create table for tracking all project expenses (labour, material, operational)

CREATE TABLE IF NOT EXISTS expense_transactions (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id),
    transaction_date DATE NOT NULL,
    coa_account_id INTEGER NOT NULL REFERENCES coa_accounts(id),
    description VARCHAR(500) NOT NULL,
    amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    unit VARCHAR(20) DEFAULT 'ls',
    quantity DECIMAL(10,2) DEFAULT 1,
    transaction_type VARCHAR(30), -- LABOUR, MATERIAL, OPERATIONAL, OTHER
    reference_type VARCHAR(30), -- PR, PO, MANUAL, CBS
    reference_id INTEGER, -- ID of PR, PO, or other reference
    reference_no VARCHAR(50), -- PR number, PO number, or manual ref
    notes TEXT,
    created_by INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_expense_project ON expense_transactions(project_id);
CREATE INDEX IF NOT EXISTS idx_expense_coa ON expense_transactions(coa_account_id);
CREATE INDEX IF NOT EXISTS idx_expense_date ON expense_transactions(transaction_date);
CREATE INDEX IF NOT EXISTS idx_expense_type ON expense_transactions(transaction_type);
CREATE INDEX IF NOT EXISTS idx_expense_reference ON expense_transactions(reference_type, reference_id);
CREATE INDEX IF NOT EXISTS idx_expense_deleted ON expense_transactions(deleted_at);

-- Add COA link to purchase_request_items if not exists
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.columns 
                   WHERE table_name = 'purchase_request_items' AND column_name = 'coa_account_id') THEN
        ALTER TABLE purchase_request_items ADD COLUMN coa_account_id INTEGER REFERENCES coa_accounts(id);
        CREATE INDEX IF NOT EXISTS idx_pr_items_coa ON purchase_request_items(coa_account_id);
    END IF;
END $$;

-- Note: View budget_vs_actual_summary will be created later after project_budgets structure is updated
-- The current project_budgets table uses 'category' field instead of 'account_id' (COA link)
-- For now, budget vs actual will be calculated in the application layer
