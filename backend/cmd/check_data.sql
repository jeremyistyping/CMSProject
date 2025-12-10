-- Check CBS nodes for Padel Bandung (project_id = 6)
SELECT 'CBS Nodes' as table_name, COUNT(*) as count FROM cbs_nodes WHERE project_id = 6 AND deleted_at IS NULL;

-- Check project budgets for Padel Bandung
SELECT 'Project Budgets' as table_name, COUNT(*) as count FROM project_budgets WHERE project_id = 6 AND deleted_at IS NULL;

-- Check expense transactions for Padel Bandung
SELECT 'Expense Transactions' as table_name, COUNT(*) as count FROM expense_transactions WHERE project_id = 6 AND deleted_at IS NULL;

-- Check CBS nodes detail
SELECT id, code, name, budget_amount, coa_account_id FROM cbs_nodes WHERE project_id = 6 AND deleted_at IS NULL ORDER BY code;

-- Check project budgets detail
SELECT pb.id, pb.account_id, coa.code, coa.name, pb.estimated_amount 
FROM project_budgets pb
LEFT JOIN coa_accounts coa ON pb.account_id = coa.id
WHERE pb.project_id = 6 AND pb.deleted_at IS NULL;

-- Check if CBS nodes have COA mapping
SELECT 
    COUNT(*) as total_nodes,
    COUNT(coa_account_id) as nodes_with_coa,
    SUM(budget_amount) as total_budget
FROM cbs_nodes 
WHERE project_id = 6 AND deleted_at IS NULL;
