# Database Seeding Strategy - Idempotent & Safe

## Overview

Backend menggunakan **idempotent seeding** - aman untuk dijalankan berulang kali tanpa membuat duplicate atau merusak data existing.

## Key Principles

### 1. **Idempotent Operations**
Setiap seed function dapat dijalankan berkali-kali dengan hasil yang sama:
- ✅ Run 1x: Create data
- ✅ Run 2x: Skip (data already exists)
- ✅ Run 100x: Still skip, no duplicates!

### 2. **Data Preservation**
- ✅ Existing balances **NEVER** overwritten
- ✅ User data preserved
- ✅ Transaction history untouched
- ✅ Only metadata (name, type) updated if needed

### 3. **Atomic Operations**
- ✅ Uses database transactions
- ✅ FOR UPDATE locks to prevent race conditions
- ✅ All-or-nothing approach

## Seed Functions & Protection Mechanisms

### 1. **Accounts (COA) - `SeedAccountsImproved()`**

**Protection Strategy:**
```go
// 1. Check for duplicates BEFORE seeding
checkExistingDuplicates(tx)

// 2. Upsert with balance preservation
upsertAccount(tx, account)
  -> If exists: Update metadata, PRESERVE balance
  -> If new: Create with initial balance 0

// 3. Validate no duplicates in seed data itself
verifyNoDuplicatesInSeed(accounts)
```

**Duplicate Detection:**
- Pre-check: Scan database for existing duplicate codes
- In-memory check: Validate seed data has no duplicates
- Database constraint: Unique index on `code` column

**Safe for:**
- ✅ Fresh database
- ✅ Existing database
- ✅ Multiple concurrent startups
- ✅ Run after manual account creation

**Example:**
```go
// First run
SeedAccountsImproved(db) 
// -> Creates 33 accounts

// Second run
SeedAccountsImproved(db)
// -> Skips all 33 (already exist)
// -> Logs: "🔒 Account exists: 1101 - KAS (preserving balance)"

// After manual edits
db.Exec("UPDATE accounts SET balance = 10000000 WHERE code = '1101'")
SeedAccountsImproved(db)
// -> Updates name/type if changed
// -> PRESERVES balance = 10000000 ✓
```

### 2. **Users - `seedUsers()`**

**Protection Strategy:**
```sql
INSERT INTO users (...) VALUES (...)
ON CONFLICT (username) DO UPDATE SET
  email = EXCLUDED.email,
  role = EXCLUDED.role,
  -- password NOT updated (preserve existing)
  updated_at = NOW()
```

**Safe for:**
- ✅ Multiple runs
- ✅ Doesn't reset passwords
- ✅ Updates role if changed in seed

### 3. **Products - `seedProducts()`**

**Protection Strategy:**
```go
// Check if product exists by code
var existingProduct models.Product
if err := db.Where("code = ?", product.Code).First(&existingProduct).Error; err == nil {
  log.Printf("Product %s already exists, skipping", product.Code)
  continue
}

// Only create if not exists
db.Create(&product)
```

**Safe for:**
- ✅ Multiple runs
- ✅ Preserves stock levels
- ✅ Preserves prices if modified

### 4. **Product Categories - `seedProductCategories()`**

**Protection Strategy:**
```go
var count int64
db.Model(&models.ProductCategory{}).Count(&count)
if count > 0 {
  return // Skip entirely if any category exists
}
```

**Safe for:**
- ✅ Multiple runs (all-or-nothing approach)
- ⚠️ Won't add new categories if ANY category exists

### 5. **Permissions - `seedPermissions()`**

**Protection Strategy:**
```sql
INSERT INTO permissions (name, resource, action, description, ...)
VALUES (?, ?, ?, ?, ...)
ON CONFLICT (name) DO NOTHING
```

**Safe for:**
- ✅ Multiple runs
- ✅ Silently skips existing permissions
- ✅ Adds new permissions if seed list grows

### 6. **Company Profile - `seedCompanyProfile()`**

**Protection Strategy:**
```go
var count int64
db.Model(&models.CompanyProfile{}).Count(&count)
if count > 0 {
  return // Skip if company profile exists
}
```

**Safe for:**
- ✅ Multiple runs
- ✅ Never overwrites existing company data

## Verification After Seeding

### Check All Seeds Completed

```sql
-- Accounts (should be >= 36 after adding new ones)
SELECT COUNT(*) FROM accounts WHERE deleted_at IS NULL;

-- Users (should be 5)
SELECT COUNT(*) FROM users WHERE deleted_at IS NULL;

-- Products (should be >= 3)
SELECT COUNT(*) FROM products WHERE deleted_at IS NULL;

-- Product Categories (should be 4)
SELECT COUNT(*) FROM product_categories WHERE deleted_at IS NULL;

-- Permissions (should be >= 36)
SELECT COUNT(*) FROM permissions;

-- Company Profile (should be 1)
SELECT COUNT(*) FROM company_profiles WHERE deleted_at IS NULL;
```

### Check No Duplicates

```sql
-- Check for duplicate account codes
SELECT code, COUNT(*) as count
FROM accounts
WHERE deleted_at IS NULL
GROUP BY code
HAVING COUNT(*) > 1;
-- Should return 0 rows

-- Check for duplicate usernames
SELECT username, COUNT(*) as count
FROM users
WHERE deleted_at IS NULL
GROUP BY username
HAVING COUNT(*) > 1;
-- Should return 0 rows

-- Check for duplicate product codes
SELECT code, COUNT(*) as count
FROM products
WHERE deleted_at IS NULL
GROUP BY code
HAVING COUNT(*) > 1;
-- Should return 0 rows
```

## Testing Idempotency

### Test Script

```bash
# Run 1: Fresh database
go run main.go

# Run 2: Immediate re-run
go run main.go

# Run 3: After some data changes
psql -U user -d accounting_db -c "UPDATE accounts SET balance = 5000000 WHERE code = '1101'"
go run main.go

# Verify balance preserved
psql -U user -d accounting_db -c "SELECT code, name, balance FROM accounts WHERE code = '1101'"
# Should show balance = 5000000 (NOT 0!)
```

### Expected Log Output

**First Run:**
```
✅ Created new account: 1101 - KAS
✅ Created new account: 1102 - BANK
... (33 more)
✅ Account seeding completed successfully
```

**Second Run:**
```
🔒 Account exists: 1101 - KAS (preserving balance)
🔒 Account exists: 1102 - BANK (preserving balance)
... (33 more)
✅ Account seeding completed successfully
```

**After Balance Change:**
```
🔒 Account exists: 1101 - KAS (preserving balance)
Balance preserved: 5000000.00 ✓
```

## Troubleshooting

### "Database has duplicate account codes"

**Error:**
```
⚠️  WARNING: Found duplicate accounts in database:
   - Code 1101 has 2 instances
database has 1 duplicate account codes - please clean up first
```

**Solution:**
```sql
-- Find duplicates
SELECT code, id, name, balance, created_at
FROM accounts
WHERE code IN (
  SELECT code FROM accounts
  WHERE deleted_at IS NULL
  GROUP BY code HAVING COUNT(*) > 1
)
ORDER BY code, created_at;

-- Keep oldest, soft-delete others
-- (Manual review recommended!)
UPDATE accounts 
SET deleted_at = NOW() 
WHERE id IN (select_ids_to_delete);
```

### "Seed data contains duplicate codes"

**Error:**
```
seed data contains duplicate codes: [1101, 2103]
```

**Cause:** Bug in seed data definition (same code listed twice)

**Solution:** Fix `account_seed_improved.go` - remove duplicate entries

### Seeding Takes Too Long

**If seeding > 5 seconds:**

1. Check for missing indexes:
```sql
-- Module permissions should have index
\d module_permission_records
-- Should show: idx_module_permission_user_module
```

2. Check for deadlocks (multiple instances starting simultaneously):
```sql
SELECT * FROM pg_stat_activity 
WHERE state = 'active' AND query LIKE '%accounts%';
```

3. Use `FOR UPDATE SKIP LOCKED` if needed (advanced)

## Best Practices

### ✅ DO:
- Run seeding on every startup (it's safe!)
- Add new seed data to existing functions
- Use `ON CONFLICT DO NOTHING/UPDATE` for SQL inserts
- Check existence before creating
- Preserve user data (balances, passwords, etc.)
- Use transactions for atomic operations

### ❌ DON'T:
- Don't assume seed data doesn't exist
- Don't overwrite balances
- Don't use `db.Save()` without checking existence
- Don't skip duplicate checks
- Don't assume single-server environment

## Future Improvements

### 1. **Migration-Based Seeding**
Move from startup seeding to versioned migrations:
```
migrations/
  001_create_tables.sql
  002_seed_initial_accounts.sql  
  003_seed_users.sql
  004_add_tax_accounts.sql  ← New accounts here
```

### 2. **Seed Versioning**
Track which seeds have run:
```sql
CREATE TABLE seed_versions (
  name VARCHAR(255) PRIMARY KEY,
  run_at TIMESTAMP,
  success BOOLEAN
);
```

### 3. **Rollback Support**
Add ability to rollback seed changes if needed.

## References

- Account seeding: `database/account_seed_improved.go`
- Main seed function: `database/seed.go`
- Duplicate cleanup: `account_seed_improved.go:279-350`
