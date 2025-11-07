# Database Migrations

## Overview

This project uses [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations.

**Benefits:**
- ✅ Automatic migration on git pull + restart
- ✅ Version controlled SQL files
- ✅ Rollback support
- ✅ No manual SQL execution needed
- ✅ Safe production deployments

## How It Works

1. **Developer creates migration file** (e.g., `001_add_cash_bank_table.sql`)
2. **Git push** → Other developers pull
3. **Application restart** → Auto-runs pending migrations
4. **Database updated** automatically

---

## Installation

### Install migrate CLI tool:

**MacOS:**
```bash
brew install golang-migrate
```

**Linux:**
```bash
curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz
sudo mv migrate /usr/local/bin/
```

**Windows:**
```powershell
# Download from https://github.com/golang-migrate/migrate/releases
# Extract and add to PATH
```

### Install Go package:
```bash
go get -u github.com/golang-migrate/migrate/v4
go get -u github.com/golang-migrate/migrate/v4/database/postgres
go get -u github.com/golang-migrate/migrate/v4/source/iofs
```

---

## Creating Migrations

### Create new migration:
```bash
# Using CLI tool
cd backend/migrations
migrate create -ext sql -dir . -seq add_unique_gl_per_cashbank

# This creates:
# - 000001_add_unique_gl_per_cashbank.up.sql   (apply changes)
# - 000001_add_unique_gl_per_cashbank.down.sql (rollback changes)
```

### Example migration files:

**`000001_add_unique_gl_per_cashbank.up.sql`:**
```sql
-- Apply changes
BEGIN;

-- Add validation to prevent GL sharing
ALTER TABLE cash_banks ADD CONSTRAINT unique_account_id UNIQUE(account_id);

-- Add index for performance
CREATE INDEX idx_cash_banks_account_id ON cash_banks(account_id) WHERE deleted_at IS NULL;

COMMIT;
```

**`000001_add_unique_gl_per_cashbank.down.sql`:**
```sql
-- Rollback changes
BEGIN;

DROP INDEX IF EXISTS idx_cash_banks_account_id;
ALTER TABLE cash_banks DROP CONSTRAINT IF EXISTS unique_account_id;

COMMIT;
```

---

## Migration Naming Convention

```
<version>_<description>.up.sql
<version>_<description>.down.sql

Examples:
001_create_initial_schema.up.sql
001_create_initial_schema.down.sql

002_add_cash_bank_validation.up.sql
002_add_cash_bank_validation.down.sql

003_fix_opening_balance_sync.up.sql
003_fix_opening_balance_sync.down.sql
```

---

## Running Migrations

### Automatic (Recommended for Production):

Migrations run automatically when application starts:

```go
// In main.go
func main() {
    // ... setup database connection ...
    
    // Run migrations automatically
    migrationService := services.NewMigrationService(sqlDB)
    if err := migrationService.RunMigrations(migrationsFS, "migrations"); err != nil {
        log.Fatalf("Failed to run migrations: %v", err)
    }
    
    // ... start application ...
}
```

### Manual (For Development):

```bash
# Apply all pending migrations
migrate -path ./migrations -database "postgres://user:pass@localhost/dbname?sslmode=disable" up

# Rollback last migration
migrate -path ./migrations -database "postgres://user:pass@localhost/dbname?sslmode=disable" down 1

# Go to specific version
migrate -path ./migrations -database "postgres://user:pass@localhost/dbname?sslmode=disable" goto 5

# Force version (fix dirty state)
migrate -path ./migrations -database "postgres://user:pass@localhost/dbname?sslmode=disable" force 5
```

---

## Workflow

### Developer Workflow:

1. **Create migration:**
   ```bash
   migrate create -ext sql -dir ./migrations -seq my_feature
   ```

2. **Write SQL:**
   - Edit `.up.sql` with changes
   - Edit `.down.sql` with rollback

3. **Test locally:**
   ```bash
   # Restart backend - migrations run automatically
   # Or manually test:
   migrate -path ./migrations -database "postgres://..." up
   ```

4. **Commit & Push:**
   ```bash
   git add migrations/
   git commit -m "Add migration: my_feature"
   git push
   ```

### Team Member Workflow:

1. **Git pull:**
   ```bash
   git pull origin main
   ```

2. **Restart backend:**
   ```bash
   # Automatic migration runs on startup
   go run main.go
   ```

3. **Database updated!** ✅

---

## Best Practices

### DO ✅

1. **Always create both .up and .down files**
   - .up: Apply changes
   - .down: Rollback changes

2. **Make migrations idempotent:**
   ```sql
   -- Good
   CREATE TABLE IF NOT EXISTS my_table (...);
   ALTER TABLE my_table ADD COLUMN IF NOT EXISTS my_column VARCHAR(255);
   
   -- Bad
   CREATE TABLE my_table (...);  -- Fails if table exists
   ```

3. **Use transactions:**
   ```sql
   BEGIN;
   -- Your changes here
   COMMIT;
   ```

4. **Test rollback:**
   ```bash
   migrate up    # Apply
   migrate down 1 # Rollback
   migrate up    # Re-apply
   ```

5. **Sequential numbering:**
   - Let migrate CLI auto-generate numbers
   - Don't skip numbers

### DON'T ❌

1. **Don't modify existing migration files**
   - Once applied, they're history
   - Create new migration instead

2. **Don't break schema:**
   - Always have rollback path
   - Test before deploying

3. **Don't mix DDL and data changes:**
   ```sql
   -- Bad: Mix schema and data
   ALTER TABLE users ADD COLUMN status VARCHAR(50);
   UPDATE users SET status = 'active';
   
   -- Good: Separate migrations
   -- Migration 1: Schema
   ALTER TABLE users ADD COLUMN status VARCHAR(50) DEFAULT 'active';
   
   -- Migration 2: Data (if needed)
   UPDATE users SET status = 'active' WHERE status IS NULL;
   ```

---

## Troubleshooting

### Problem: "Dirty database version"

**Cause:** Migration failed halfway

**Solution:**
```bash
# Check current state
migrate -path ./migrations -database "postgres://..." version

# Force to last good version
migrate -path ./migrations -database "postgres://..." force <version>

# Re-run migrations
migrate -path ./migrations -database "postgres://..." up
```

### Problem: "File exists but not applied"

**Cause:** Migration file added after database was at higher version

**Solution:**
```bash
# Check version
migrate -path ./migrations -database "postgres://..." version

# Create new migration with higher number
migrate create -ext sql -dir ./migrations -seq fix_issue
```

### Problem: "Migration out of order"

**Cause:** Git merge with conflicting migration numbers

**Solution:**
- Rename newer migration to higher number
- Update internal version reference

---

## Migration vs AutoMigrate

### GORM AutoMigrate (NOT RECOMMENDED for Production):
```go
// ❌ DON'T USE IN PRODUCTION
db.AutoMigrate(&User{}, &Product{})
```
**Problems:**
- No version control
- No rollback
- Can't customize
- Risky for production
- No collaboration history

### golang-migrate (RECOMMENDED):
```go
// ✅ USE THIS
migrationService.RunMigrations(migrationsFS, "migrations")
```
**Benefits:**
- ✅ Version controlled
- ✅ Rollback support
- ✅ Team collaboration
- ✅ Production safe
- ✅ Audit trail

---

## Example: Converting Existing Fix Scripts

### Old Way (Manual SQL):
```bash
# Developer creates fix_cashbank.sql
# Sends to team: "Run this SQL script"
# Everyone manually executes
# Error-prone, inconsistent
```

### New Way (Migration):
```bash
# 1. Create migration
migrate create -ext sql -dir ./migrations -seq fix_cashbank_balance_sync

# 2. Put SQL in .up.sql file
# 3. Git commit & push
# 4. Team git pull + restart
# 5. Auto-applied! ✅
```

---

## CI/CD Integration

### GitHub Actions Example:

```yaml
name: Deploy

on:
  push:
    branches: [ main ]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      
      - name: Run Database Migrations
        run: |
          migrate -path ./migrations \
                  -database "${{ secrets.DATABASE_URL }}" \
                  up
      
      - name: Deploy Application
        run: ./deploy.sh
```

---

## Migration History Table

Migrations are tracked in `schema_migrations` table:

```sql
SELECT * FROM schema_migrations;

-- Output:
-- version | dirty
-- --------+-------
--    1    | false
--    2    | false
--    3    | false
```

---

## Summary

**Before:**
1. Create SQL script
2. Send to team
3. Everyone manually runs
4. ❌ Error-prone

**After:**
1. Create migration file
2. Git push
3. Team git pull + restart
4. ✅ Auto-applied!

**Perfect for:**
- Team collaboration
- Production deployments
- Safe schema changes
- Version control
- Audit compliance
