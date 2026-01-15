# Direct Commands to Fix Daily Updates Issue

## Problem Summary
1. Migration 077 has syntax error (`DO $` should be `DO $$`)
2. Environment variables `DISABLE_IP_WHITELIST` and `SECURITY_ALLOWED_IPS` not passed to backend container
3. Daily updates permission needs to be added to database

## Solution - Run These Commands in Order

### Step 1: Add daily_updates permission to database
```bash
docker exec -it cmsproject_postgres psql -U accounting_user -d accounting_db -c "DELETE FROM module_permission_records WHERE module = 'daily_updates';"

docker exec -it cmsproject_postgres psql -U accounting_user -d accounting_db -c "INSERT INTO module_permission_records (user_id, module, can_view, can_create, can_edit, can_delete, can_approve, can_export, can_menu, created_at, updated_at) VALUES (1, 'daily_updates', true, true, true, true, true, true, true, NOW(), NOW());"

docker exec -it cmsproject_postgres psql -U accounting_user -d accounting_db -c "SELECT user_id, module, can_view, can_create, can_edit, can_delete, can_approve FROM module_permission_records WHERE module = 'daily_updates';"
```

### Step 2: Check schema_migrations table structure
```bash
docker exec -it cmsproject_postgres psql -U accounting_user -d accounting_db -c "\d schema_migrations"
```

### Step 3: Mark migration 077 as complete (adjust based on table structure from Step 2)

**If table has columns (version, dirty):**
```bash
docker exec -it cmsproject_postgres psql -U accounting_user -d accounting_db -c "INSERT INTO schema_migrations (version, dirty) VALUES (77, false) ON CONFLICT (version) DO UPDATE SET dirty = false;"
```

**If table has columns (id, applied_at) or similar:**
```bash
docker exec -it cmsproject_postgres psql -U accounting_user -d accounting_db -c "INSERT INTO schema_migrations (id, applied_at) VALUES (77, NOW()) ON CONFLICT (id) DO NOTHING;"
```

**If table has columns (migration, batch):**
```bash
docker exec -it cmsproject_postgres psql -U accounting_user -d accounting_db -c "INSERT INTO schema_migrations (migration, batch) VALUES ('077_add_approval_fields_to_daily_updates', 1) ON CONFLICT (migration) DO NOTHING;"
```

### Step 4: Rebuild backend with new environment variables
```bash
docker-compose build backend
```

### Step 5: Restart backend
```bash
docker-compose up -d backend
```

### Step 6: Wait and check status
```bash
sleep 10
docker-compose ps backend
```

### Step 7: Check logs for DISABLE_IP_WHITELIST
```bash
docker-compose logs backend --tail=50 | grep -i "DISABLE"
```

### Step 8: Check logs for IP whitelist checks
```bash
docker-compose logs backend --tail=100 | grep -i "182.253.252.250"
```

### Step 9: Test daily updates
Open browser and go to: http://72.62.245.66/projects/1
Click on "Daily Updates" tab and check if it loads without 403 error.

## What Was Fixed

1. **Migration 077**: Fixed syntax error `DO $` → `DO $$`
2. **docker-compose.yml**: Added environment variables:
   - `DISABLE_IP_WHITELIST: ${DISABLE_IP_WHITELIST:-false}`
   - `SECURITY_ALLOWED_IPS: ${SECURITY_ALLOWED_IPS:-127.0.0.1,::1,localhost}`
3. **.env**: Already has `DISABLE_IP_WHITELIST=true`
4. **Database**: Will add `daily_updates` permission for user_id=1

## Verification

After running all commands, you should see:
- Backend logs showing environment variables are loaded
- No more "record not found" for IP 182.253.252.250
- Daily Updates tab loads successfully
- Permission debug logs showing access granted for daily_updates module

## Troubleshooting

If still getting 403:
```bash
# Check if environment variable is loaded
docker exec -it cmsproject_backend env | grep DISABLE_IP_WHITELIST

# Check backend logs for permission checks
docker-compose logs backend --tail=100 | grep -i "daily_updates\|permission"

# Check if permission exists in database
docker exec -it cmsproject_postgres psql -U accounting_user -d accounting_db -c "SELECT * FROM module_permission_records WHERE module = 'daily_updates';"
```
