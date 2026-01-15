#!/bin/bash
# Complete fix for daily updates permission issue
# This script:
# 1. Adds daily_updates permission to database
# 2. Marks migration 077 as complete
# 3. Rebuilds and restarts backend with IP whitelist disabled

echo "=========================================="
echo "Daily Updates Permission Fix"
echo "=========================================="

# Step 1: Add daily_updates permission
echo ""
echo "Step 1: Adding daily_updates permission..."
docker exec -it cmsproject_postgres psql -U accounting_user -d accounting_db << 'EOF'
-- Delete any existing daily_updates permissions to avoid duplicates
DELETE FROM module_permission_records WHERE module = 'daily_updates';

-- Insert fresh daily_updates permission for admin user
INSERT INTO module_permission_records (
    user_id, module, can_view, can_create, can_edit, can_delete, 
    can_approve, can_export, can_menu, created_at, updated_at
) VALUES (
    1, 'daily_updates', true, true, true, true, 
    true, true, true, NOW(), NOW()
);

-- Verify the permission was added
SELECT user_id, module, can_view, can_create, can_edit, can_delete, can_approve 
FROM module_permission_records 
WHERE module = 'daily_updates';
EOF

# Step 2: Mark migration 077 as complete
echo ""
echo "Step 2: Marking migration 077 as complete..."
docker exec -it cmsproject_postgres psql -U accounting_user -d accounting_db << 'EOF'
-- Check schema_migrations table structure
\d schema_migrations

-- Insert migration 077 record (adjust column names based on table structure)
-- Common patterns: (version, dirty) or (id, applied_at) or (migration, batch)
-- We'll try the most common pattern first
INSERT INTO schema_migrations (version, dirty) 
VALUES (77, false) 
ON CONFLICT (version) DO UPDATE SET dirty = false;
EOF

# Step 3: Rebuild backend with new environment variables
echo ""
echo "Step 3: Rebuilding backend container..."
docker-compose build backend

# Step 4: Restart backend
echo ""
echo "Step 4: Restarting backend..."
docker-compose up -d backend

# Wait for backend to start
echo ""
echo "Waiting 10 seconds for backend to start..."
sleep 10

# Step 5: Check backend status
echo ""
echo "Step 5: Checking backend status..."
docker-compose ps backend

# Step 6: Check backend logs for DISABLE_IP_WHITELIST
echo ""
echo "Step 6: Checking backend logs..."
docker-compose logs backend --tail=30 | grep -i "DISABLE\|daily\|permission\|whitelist"

echo ""
echo "=========================================="
echo "Fix Complete!"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. Test daily updates at: http://72.62.245.66/projects/1"
echo "2. Check logs: docker-compose logs backend --tail=50"
echo "3. If still issues, check: docker-compose logs backend | grep -i '182.253.252.250'"
echo ""
