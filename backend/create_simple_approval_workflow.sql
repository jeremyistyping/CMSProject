-- SIMPLIFIED APPROVAL WORKFLOW untuk Purchases
-- Tidak tergantung amount, pakai manual escalation checkbox

-- 1. Bersihkan data lama
DELETE FROM approval_actions WHERE request_id IN (SELECT id FROM approval_requests);
DELETE FROM approval_history WHERE request_id IN (SELECT id FROM approval_requests);
DELETE FROM approval_requests;
DELETE FROM approval_workflow_steps;
DELETE FROM approval_workflows WHERE module = 'PURCHASE';

-- 2. Buat SATU workflow saja untuk semua purchases
INSERT INTO approval_workflows (name, module, min_amount, max_amount, require_director, require_finance, is_active, created_at, updated_at) VALUES
('Standard Purchase Approval', 'PURCHASE', 0, 999999999999, false, true, true, NOW(), NOW());

-- 3. Buat steps untuk workflow tersebut
-- Step 1: Employee Submission
INSERT INTO approval_workflow_steps (workflow_id, step_order, step_name, approver_role, is_optional, time_limit, created_at, updated_at)
SELECT 
    w.id,
    1,
    'Employee Submission',
    'employee',
    false,
    24,
    NOW(),
    NOW()
FROM approval_workflows w WHERE w.name = 'Standard Purchase Approval';

-- Step 2: Finance Approval (dengan opsi escalate ke Director)
INSERT INTO approval_workflow_steps (workflow_id, step_order, step_name, approver_role, is_optional, time_limit, created_at, updated_at)
SELECT 
    w.id,
    2,
    'Finance Approval',
    'finance',
    false,
    48,
    NOW(),
    NOW()
FROM approval_workflows w WHERE w.name = 'Standard Purchase Approval';

-- Step 3: Director Approval (optional, activated via checkbox)
INSERT INTO approval_workflow_steps (workflow_id, step_order, step_name, approver_role, is_optional, time_limit, created_at, updated_at)
SELECT 
    w.id,
    3,
    'Director Approval',
    'director',
    true,
    72,
    NOW(),
    NOW()
FROM approval_workflows w WHERE w.name = 'Standard Purchase Approval';

-- 4. Verify workflow created
SELECT 'CREATED WORKFLOW:' as status;
SELECT id, name, module, min_amount, max_amount, require_finance FROM approval_workflows WHERE module = 'PURCHASE';

SELECT 'CREATED STEPS:' as status;
SELECT 
    w.name as workflow_name,
    aws.step_order,
    aws.step_name,
    aws.approver_role,
    aws.is_optional,
    aws.time_limit
FROM approval_workflows w
JOIN approval_workflow_steps aws ON w.id = aws.workflow_id
WHERE w.module = 'PURCHASE'
ORDER BY aws.step_order;

-- 5. Summary
SELECT 'SUMMARY:' as result;
SELECT 'Single workflow created for all purchase amounts' as description;
SELECT 'Finance can manually escalate to Director via checkbox' as escalation_method;
SELECT 'No more complex amount-based rules' as benefit;