-- ============================================================================
-- Migration: 023_create_purchase_approval_workflows.sql
-- Purpose: Create simplified purchase approval workflows for employee workflow
-- Author: System
-- Date: 2025-01-21
-- ============================================================================

-- Check if approval workflows already exist for PURCHASE module
DO $$
DECLARE
    workflow_count INTEGER;
    step_count INTEGER;
    standard_workflow_id INTEGER;
BEGIN
    -- Initialize variables
    workflow_count := 0;
    step_count := 0;
    
    -- Check existing workflows
    SELECT COUNT(*) INTO workflow_count FROM approval_workflows WHERE module = 'PURCHASE';
    
    -- Log current status
    RAISE NOTICE '🔍 Found % existing PURCHASE approval workflows', workflow_count;
    
    -- Check if our specific workflow already exists
    SELECT COUNT(*) INTO workflow_count FROM approval_workflows WHERE module = 'PURCHASE' AND name = 'Standard Purchase Approval';
    
    IF workflow_count > 0 THEN
        RAISE NOTICE '✅ Standard Purchase Approval workflow already exists - skipping creation';
        -- Still check if steps exist
        SELECT id INTO standard_workflow_id FROM approval_workflows WHERE module = 'PURCHASE' AND name = 'Standard Purchase Approval' LIMIT 1;
        SELECT COUNT(*) INTO step_count FROM approval_workflow_steps WHERE workflow_id = standard_workflow_id;
        RAISE NOTICE '📋 Existing workflow has % steps', step_count;
        
        -- If steps missing, create them
        IF step_count < 3 THEN
            RAISE NOTICE '🔧 Adding missing workflow steps...';
            
            -- Insert Employee Submission step (if not exists)
            INSERT INTO approval_workflow_steps (workflow_id, step_order, step_name, approver_role, is_optional, time_limit, created_at, updated_at)
            SELECT standard_workflow_id, 1, 'Employee Submission', 'employee', false, 24, NOW(), NOW()
            WHERE NOT EXISTS (SELECT 1 FROM approval_workflow_steps WHERE workflow_id = standard_workflow_id AND step_order = 1);
            
            -- Insert Finance Approval step (if not exists)
            INSERT INTO approval_workflow_steps (workflow_id, step_order, step_name, approver_role, is_optional, time_limit, created_at, updated_at)
            SELECT standard_workflow_id, 2, 'Finance Approval', 'finance', false, 48, NOW(), NOW()
            WHERE NOT EXISTS (SELECT 1 FROM approval_workflow_steps WHERE workflow_id = standard_workflow_id AND step_order = 2);
            
            -- Insert Director Approval step (if not exists)
            INSERT INTO approval_workflow_steps (workflow_id, step_order, step_name, approver_role, is_optional, time_limit, created_at, updated_at)
            SELECT standard_workflow_id, 3, 'Director Approval', 'director', true, 72, NOW(), NOW()
            WHERE NOT EXISTS (SELECT 1 FROM approval_workflow_steps WHERE workflow_id = standard_workflow_id AND step_order = 3);
            
            RAISE NOTICE '✅ Missing workflow steps added';
        END IF;
    ELSE
        RAISE NOTICE '📋 Creating Standard Purchase Approval workflow...';
        
        -- Create the main workflow
        INSERT INTO approval_workflows (name, module, min_amount, max_amount, require_director, require_finance, is_active, created_at, updated_at)
        VALUES ('Standard Purchase Approval', 'PURCHASE', 0, 999999999999, false, true, true, NOW(), NOW())
        RETURNING id INTO standard_workflow_id;
        
        RAISE NOTICE '✅ Created workflow with ID: %', standard_workflow_id;
        
        -- Create workflow steps
        INSERT INTO approval_workflow_steps (workflow_id, step_order, step_name, approver_role, is_optional, time_limit, created_at, updated_at)
        VALUES 
            (standard_workflow_id, 1, 'Employee Submission', 'employee', false, 24, NOW(), NOW()),
            (standard_workflow_id, 2, 'Finance Approval', 'finance', false, 48, NOW(), NOW()),
            (standard_workflow_id, 3, 'Director Approval', 'director', true, 72, NOW(), NOW());
        
        RAISE NOTICE '✅ Created 3 workflow steps for Standard Purchase Approval';
    END IF;
    
    -- Final verification
    SELECT COUNT(*) INTO workflow_count FROM approval_workflows WHERE module = 'PURCHASE';
    SELECT COUNT(*) INTO step_count FROM approval_workflow_steps aws 
    JOIN approval_workflows aw ON aws.workflow_id = aw.id 
    WHERE aw.module = 'PURCHASE';
    
    RAISE NOTICE '📊 Final Status: % PURCHASE workflows with % total steps', workflow_count, step_count;
    RAISE NOTICE '🎯 Purchase approval workflow system ready for employee submissions!';
    
END $$;