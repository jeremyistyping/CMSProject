-- Migration: Create project_progress table and project cost views
-- Date: 2025-11-15
-- Purpose: Store physical progress history per project and expose derived actual costs for Budget vs Actual reports

BEGIN;

-- 1. Project progress history table
CREATE TABLE IF NOT EXISTS project_progress (
    id BIGSERIAL PRIMARY KEY,
    project_id BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    date DATE NOT NULL,
    physical_progress_percent DECIMAL(5,2) NOT NULL DEFAULT 0,
    volume_achieved DECIMAL(20,4),
    remarks TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMPTZ NULL,
    UNIQUE(project_id, date)
);

COMMENT ON TABLE project_progress IS 'Time-series snapshot of physical progress per project (for Budget vs Actual analysis)';
COMMENT ON COLUMN project_progress.physical_progress_percent IS 'Physical progress percentage (0-100) as of the given date';

CREATE INDEX IF NOT EXISTS idx_project_progress_project_id ON project_progress(project_id);
CREATE INDEX IF NOT EXISTS idx_project_progress_date ON project_progress(date);

-- 2. View: project_budget_items (normalized view over project_budgets + accounts)
DROP VIEW IF EXISTS project_budget_items;
CREATE VIEW project_budget_items AS
SELECT
    pb.id,
    pb.project_id,
    pb.account_id,
    a.code AS cost_code,
    a.name AS description,
    COALESCE(a.category, 'UNCATEGORIZED') AS category,
    NULL::VARCHAR AS unit,
    NULL::NUMERIC(20,4) AS qty_budget,
    NULL::NUMERIC(20,4) AS unit_price_budget,
    pb.estimated_amount AS amount_budget,
    pb.created_at,
    pb.updated_at
FROM project_budgets pb
JOIN accounts a ON a.id = pb.account_id
WHERE pb.deleted_at IS NULL
  AND a.deleted_at IS NULL;

COMMENT ON VIEW project_budget_items IS 'Logical budget_items view for projects, backed by project_budgets + accounts';

-- 3. View: project_actual_costs (derived from purchases & purchase_items)
DROP VIEW IF EXISTS project_actual_costs;
CREATE VIEW project_actual_costs AS
SELECT
    pi.id AS id,
    p.project_id,
    pb.id AS project_budget_id,
    'PURCHASE'::VARCHAR(30) AS source_type,
    p.id AS source_id,
    p.date::DATE AS date,
    pi.total_price AS amount,
    COALESCE(a.category, 'UNCATEGORIZED') AS category,
    CASE
        WHEN p.status IN ('APPROVED','COMPLETED','PAID')
          OR p.approval_status IN (
              'APPROVED',
              'APPROVED_GM',
              'APPROVED_COST_CONTROL'
          )
        THEN 'APPROVED'
        ELSE 'DRAFT'
    END AS status
FROM purchase_items pi
JOIN purchases p ON p.id = pi.purchase_id
LEFT JOIN accounts a ON a.id = pi.expense_account_id
LEFT JOIN project_budgets pb
    ON pb.project_id = p.project_id
   AND pb.account_id = pi.expense_account_id
   AND pb.deleted_at IS NULL
WHERE p.project_id IS NOT NULL
  AND p.deleted_at IS NULL
  AND pi.deleted_at IS NULL;

COMMENT ON VIEW project_actual_costs IS 'Derived actual_costs per project, based on approved purchases and their expense accounts';

COMMIT;
