-- Check workflow data setelah reset
-- Script untuk verifikasi data workflow yang hilang

-- 1. Check users (should still exist)
SELECT 'USERS' as table_name, COUNT(*) as count FROM users;

-- 2. Check module_permissions (should still exist)
SELECT 'MODULE_PERMISSIONS' as table_name, COUNT(*) as count FROM module_permissions;

-- 3. Check approval workflows (might be missing)
SELECT 'APPROVAL_WORKFLOWS' as table_name, COUNT(*) as count FROM approval_workflows;

-- 4. Check approval workflow steps (might be missing)
SELECT 'APPROVAL_WORKFLOW_STEPS' as table_name, COUNT(*) as count FROM approval_workflow_steps;

-- 5. Check approval requests (should be 0 after reset)
SELECT 'APPROVAL_REQUESTS' as table_name, COUNT(*) as count FROM approval_requests;

-- 6. Check approval actions (should be 0 after reset)  
SELECT 'APPROVAL_ACTIONS' as table_name, COUNT(*) as count FROM approval_actions;

-- 7. Check approval history (should be 0 after reset)
SELECT 'APPROVAL_HISTORY' as table_name, COUNT(*) as count FROM approval_history;

-- 8. Check specific user permissions
SELECT 
    u.username,
    u.role,
    mp.module,
    mp.can_view,
    mp.can_create,
    mp.can_edit,
    mp.can_approve
FROM users u
LEFT JOIN module_permissions mp ON u.id = mp.user_id
WHERE mp.module = 'purchases'
ORDER BY u.username;

-- 9. Check approval workflows for purchases
SELECT * FROM approval_workflows WHERE module = 'purchases' OR module = 'purchase';

-- 10. Check workflow steps
SELECT 
    aw.name as workflow_name,
    aw.module,
    aws.step_order,
    aws.step_name,
    aws.approver_role
FROM approval_workflows aw
LEFT JOIN approval_workflow_steps aws ON aw.id = aws.workflow_id
WHERE aw.module IN ('purchases', 'purchase')
ORDER BY aw.name, aws.step_order;