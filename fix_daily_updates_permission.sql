-- Fix Daily Updates Permission Issue
-- Run this on your LOCAL database first to test

-- Step 1: Check current permissions for admin user (user_id = 1)
SELECT 
    id, 
    user_id, 
    module, 
    can_view, 
    can_create, 
    can_edit, 
    can_delete, 
    can_approve,
    created_at
FROM module_permission_records 
WHERE user_id = 1 
ORDER BY module;

-- Step 2: Add daily_updates permission if missing
-- This uses ON CONFLICT to update if exists, insert if not
INSERT INTO module_permission_records 
    (user_id, module, can_view, can_create, can_edit, can_delete, can_approve, can_export, can_menu, created_at, updated_at)
VALUES 
    (1, 'daily_updates', true, true, true, true, true, true, true, NOW(), NOW())
ON CONFLICT (user_id, module) 
DO UPDATE SET
    can_view = true,
    can_create = true,
    can_edit = true,
    can_delete = true,
    can_approve = true,
    can_export = true,
    can_menu = true,
    updated_at = NOW();

-- Step 3: Verify the permission was added/updated
SELECT 
    id, 
    user_id, 
    module, 
    can_view, 
    can_create, 
    can_edit, 
    can_delete, 
    can_approve,
    updated_at
FROM module_permission_records 
WHERE user_id = 1 AND module = 'daily_updates';

-- Step 4: Check all permissions again to confirm
SELECT 
    module, 
    can_view, 
    can_create, 
    can_edit, 
    can_delete, 
    can_approve
FROM module_permission_records 
WHERE user_id = 1 
ORDER BY module;

-- Expected result: You should see 'daily_updates' with all permissions set to true
