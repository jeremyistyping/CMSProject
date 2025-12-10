-- Test the new budget vs actual query
WITH budget_data AS (
    SELECT 
        pb.project_id,
        pb.account_id as coa_account_id,
        pb.estimated_amount as budget_estimation
    FROM project_budgets pb
    WHERE pb.project_id = 6
        AND pb.deleted_at IS NULL
),
actual_data AS (
    SELECT 
        et.project_id,
        et.coa_account_id,
        SUM(et.amount) as actual_amount
    FROM expense_transactions et
    WHERE et.project_id = 6
        AND et.deleted_at IS NULL
        AND et.transaction_date BETWEEN '2025-11-30' AND '2025-12-08'
    GROUP BY et.project_id, et.coa_account_id
),
all_coa AS (
    SELECT DISTINCT coa_account_id, project_id FROM budget_data
    UNION
    SELECT DISTINCT coa_account_id, project_id FROM actual_data
)
SELECT 
    ac.project_id,
    ac.coa_account_id,
    coa.code as coa_code,
    coa.name as coa_name,
    coa.budget_category,
    coa.work_package,
    COALESCE(bd.budget_estimation, 0) as budget_estimation,
    COALESCE(ad.actual_amount, 0) as actual_amount,
    COALESCE(bd.budget_estimation, 0) - COALESCE(ad.actual_amount, 0) as variance
FROM all_coa ac
JOIN coa_accounts coa ON ac.coa_account_id = coa.id
LEFT JOIN budget_data bd ON bd.coa_account_id = ac.coa_account_id
LEFT JOIN actual_data ad ON ad.coa_account_id = ac.coa_account_id
WHERE coa.deleted_at IS NULL
ORDER BY coa.budget_category, coa.code;
