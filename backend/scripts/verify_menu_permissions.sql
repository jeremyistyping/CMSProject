-- Verify Menu Permissions Update
-- This script checks if CanMenu field has been properly set for all users

-- Check Finance user permissions (show all modules with Menu status)
SELECT 
    u.username,
    u.role,
    mp.module,
    mp.can_view,
    mp.can_create,
    mp.can_edit,
    mp.can_delete,
    mp.can_approve,
    mp.can_export,
    mp.can_menu
FROM users u
JOIN module_permissions mp ON u.id = mp.user_id
WHERE u.role = 'finance'
ORDER BY mp.module;

-- Summary by role showing which modules have CanMenu = true
SELECT 
    u.role,
    mp.module,
    mp.can_menu
FROM users u
JOIN module_permissions mp ON u.id = mp.user_id
WHERE mp.can_menu = true
ORDER BY u.role, mp.module;

-- Count of CanMenu permissions by role
SELECT 
    u.role,
    COUNT(CASE WHEN mp.can_menu = true THEN 1 END) as menu_enabled_count,
    COUNT(*) as total_modules
FROM users u
JOIN module_permissions mp ON u.id = mp.user_id
GROUP BY u.role
ORDER BY u.role;
