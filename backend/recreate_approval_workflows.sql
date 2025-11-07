-- Recreate Approval Workflows setelah HARD DELETE reset
-- Script ini membuat ulang approval workflows untuk modul purchases

-- 1. Bersihkan data lama jika ada
DELETE FROM approval_actions WHERE request_id IN (SELECT id FROM approval_requests);
DELETE FROM approval_history WHERE request_id IN (SELECT id FROM approval_requests);
DELETE FROM approval_requests;
DELETE FROM approval_workflow_steps;
DELETE FROM approval_workflows WHERE module = 'PURCHASE';

-- 2. Insert approval workflows untuk purchases
INSERT INTO approval_workflows (name, module, min_amount, max_amount, require_director, require_finance, is_active, created_at, updated_at) VALUES
('Small Purchase Approval', 'PURCHASE', 0, 5000000, false, true, true, NOW(), NOW()),
('Medium Purchase Approval', 'PURCHASE', 5000000, 25000000, false, true, true, NOW(), NOW()),
('Large Purchase Approval', 'PURCHASE', 25000000, 100000000, true, true, true, NOW(), NOW()),
('Very Large Purchase Approval', 'PURCHASE', 100000000, 0, true, true, true, NOW(), NOW());

-- 3. Insert approval workflow steps
-- Small Purchase Approval Steps
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
FROM approval_workflows w WHERE w.name = 'Small Purchase Approval';

INSERT INTO approval_workflow_steps (workflow_id, step_order, step_name, approver_role, is_optional, time_limit, created_at, updated_at)
SELECT 
    w.id,
    2,
    'Finance Approval',
    'finance',
    false,
    24,
    NOW(),
    NOW()
FROM approval_workflows w WHERE w.name = 'Small Purchase Approval';

INSERT INTO approval_workflow_steps (workflow_id, step_order, step_name, approver_role, is_optional, time_limit, created_at, updated_at)
SELECT 
    w.id,
    3,
    'Director Approval (Optional)',
    'director',
    true,
    48,
    NOW(),
    NOW()
FROM approval_workflows w WHERE w.name = 'Small Purchase Approval';

-- Medium Purchase Approval Steps
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
FROM approval_workflows w WHERE w.name = 'Medium Purchase Approval';

INSERT INTO approval_workflow_steps (workflow_id, step_order, step_name, approver_role, is_optional, time_limit, created_at, updated_at)
SELECT 
    w.id,
    2,
    'Finance Approval',
    'finance',
    false,
    24,
    NOW(),
    NOW()
FROM approval_workflows w WHERE w.name = 'Medium Purchase Approval';

INSERT INTO approval_workflow_steps (workflow_id, step_order, step_name, approver_role, is_optional, time_limit, created_at, updated_at)
SELECT 
    w.id,
    3,
    'Director Approval (Optional)',
    'director',
    true,
    48,
    NOW(),
    NOW()
FROM approval_workflows w WHERE w.name = 'Medium Purchase Approval';

-- Large Purchase Approval Steps
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
FROM approval_workflows w WHERE w.name = 'Large Purchase Approval';

INSERT INTO approval_workflow_steps (workflow_id, step_order, step_name, approver_role, is_optional, time_limit, created_at, updated_at)
SELECT 
    w.id,
    2,
    'Finance Approval',
    'finance',
    false,
    24,
    NOW(),
    NOW()
FROM approval_workflows w WHERE w.name = 'Large Purchase Approval';

INSERT INTO approval_workflow_steps (workflow_id, step_order, step_name, approver_role, is_optional, time_limit, created_at, updated_at)
SELECT 
    w.id,
    3,
    'Director Approval',
    'director',
    false,
    48,
    NOW(),
    NOW()
FROM approval_workflows w WHERE w.name = 'Large Purchase Approval';

-- Very Large Purchase Approval Steps  
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
FROM approval_workflows w WHERE w.name = 'Very Large Purchase Approval';

INSERT INTO approval_workflow_steps (workflow_id, step_order, step_name, approver_role, is_optional, time_limit, created_at, updated_at)
SELECT 
    w.id,
    2,
    'Finance Approval',
    'finance',
    false,
    24,
    NOW(),
    NOW()
FROM approval_workflows w WHERE w.name = 'Very Large Purchase Approval';

INSERT INTO approval_workflow_steps (workflow_id, step_order, step_name, approver_role, is_optional, time_limit, created_at, updated_at)
SELECT 
    w.id,
    3,
    'Director Approval',
    'director',
    false,
    48,
    NOW(),
    NOW()
FROM approval_workflows w WHERE w.name = 'Very Large Purchase Approval';

INSERT INTO approval_workflow_steps (workflow_id, step_order, step_name, approver_role, is_optional, time_limit, created_at, updated_at)
SELECT 
    w.id,
    4,
    'Admin Final Approval',
    'admin',
    false,
    48,
    NOW(),
    NOW()
FROM approval_workflows w WHERE w.name = 'Very Large Purchase Approval';

-- 4. Verify created data
SELECT 'WORKFLOWS CREATED:' as status;
SELECT id, name, module, min_amount, max_amount FROM approval_workflows WHERE module = 'PURCHASE';

SELECT 'WORKFLOW STEPS CREATED:' as status;
SELECT 
    w.name as workflow_name,
    aws.step_order,
    aws.step_name,
    aws.approver_role,
    aws.is_optional
FROM approval_workflows w
JOIN approval_workflow_steps aws ON w.id = aws.workflow_id
WHERE w.module = 'PURCHASE'
ORDER BY w.name, aws.step_order;

-- 5. Show summary
SELECT 'SUMMARY:' as status;
SELECT 
    COUNT(*) as total_workflows 
FROM approval_workflows 
WHERE module = 'PURCHASE';

SELECT 
    COUNT(*) as total_steps 
FROM approval_workflow_steps aws
JOIN approval_workflows w ON aws.workflow_id = w.id
WHERE w.module = 'PURCHASE';