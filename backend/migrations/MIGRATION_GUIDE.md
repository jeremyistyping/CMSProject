# 📋 Migration Guide - COGS Journal Creation

## 🎯 Purpose

This migration creates missing COGS (Cost of Goods Sold) journal entries for existing sales transactions. This fixes the issue where P&L reports show `Total Expense = Rp 0` because purchases were recorded to Inventory but never transferred to COGS when goods were sold.

---

## ⚠️ **When to Run This Migration**

Run this migration **ONCE** after deploying the P&L categorization fixes if:

1. ✅ You have existing sales transactions (INVOICED or PAID)
2. ✅ Those sales don't have corresponding COGS journal entries
3. ✅ Your P&L report shows incorrect Gross Profit (100% margin)

**Safe to run multiple times** - script automatically skips sales that already have COGS journals.

---

## 🚀 **How to Run**

### **Option 1: Manual Run (Recommended for first time)**

```bash
# Navigate to backend directory
cd backend

# Run migration script
go run migrations/scripts/create_missing_cogs_journals.go
```

**Expected Output:**
```
================================================================================
MIGRATION: Create Missing COGS Journals
Version: 1.0
Date: 2025-10-17
================================================================================

✅ Database connected

📊 Step 1: Finding sales transactions...
   Found 10 sales transactions

📊 Step 2: Checking for missing COGS journals...
   ⚠️  Sales SOA-00016 (Rp 5550000.00) - Missing COGS journal
   ⚠️  Sales SOA-00017 (Rp 3000000.00) - Missing COGS journal
   
   2 sales need COGS journals

📊 Step 3: Loading chart of accounts...
   ✅ COGS Account: HARGA POKOK PENJUALAN (5101)
   ✅ Inventory Account: PERSEDIAAN BARANG DAGANGAN (1301)

📊 Step 4: Creating COGS journals...
--------------------------------------------------------------------------------

[1/2] Processing Sale SOA-00016...
   COGS Amount: Rp 500000.00
   ✅ COGS journal created successfully

[2/2] Processing Sale SOA-00017...
   COGS Amount: Rp 240000.00
   ✅ COGS journal created successfully

================================================================================
MIGRATION SUMMARY
================================================================================
Total sales processed:     2
✅ Successfully created:   2
⚠️  Skipped (zero amount): 0
❌ Errors:                 0
================================================================================

✅ MIGRATION COMPLETED SUCCESSFULLY!

📋 Next Steps:
   1. Restart backend server to apply code changes
   2. Re-generate P&L reports to see corrected values
   3. Verify Gross Profit and Net Income calculations
```

---

### **Option 2: Automated via Deployment Script**

Add to your deployment pipeline:

```bash
#!/bin/bash
# deploy.sh

echo "Pulling latest changes..."
git pull origin main

echo "Building backend..."
cd backend
go build -o app

echo "Running migrations..."
go run migrations/scripts/create_missing_cogs_journals.go

echo "Restarting server..."
systemctl restart accounting-backend
```

---

### **Option 3: Integrate with Existing Migration System**

If you have a migration table/system, add this:

```sql
-- Add to your migrations tracking table
INSERT INTO schema_migrations (version, name, executed_at) 
VALUES (
  '20251017_create_missing_cogs_journals',
  'Create missing COGS journals for existing sales',
  NOW()
);
```

Then modify the Go script to check if migration already ran:

```go
// Check if migration already executed
var count int64
db.Table("schema_migrations").
  Where("version = ?", "20251017_create_missing_cogs_journals").
  Count(&count)

if count > 0 {
  fmt.Println("✅ Migration already executed. Skipping.")
  return
}
```

---

## 🔧 **Configuration**

### **Database Connection**

The script uses hardcoded connection for simplicity. For production, use environment variables:

```go
// Read from environment
dbHost := os.Getenv("DB_HOST")
dbPort := os.Getenv("DB_PORT")
dbUser := os.Getenv("DB_USER")
dbPass := os.Getenv("DB_PASSWORD")
dbName := os.Getenv("DB_NAME")

dsn := fmt.Sprintf(
  "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
  dbHost, dbPort, dbUser, dbPass, dbName,
)
```

### **COGS Calculation Strategy**

The script tries 3 strategies in order:

1. **Actual Purchase Cost** - from purchase history
2. **Product Standard Cost** - from product master data
3. **Estimated Cost** - 80% of sales price (20% margin assumption)

You can adjust the estimate percentage in `calculateCOGSAmount()`:

```go
// Strategy 3: Estimate (adjust percentage here)
totalCOGS += item.TotalPrice * 0.8  // 80% = 20% margin
```

---

## ✅ **Verification**

After running migration:

### **1. Check COGS Journals Created**

```sql
-- Check how many COGS journals were created
SELECT COUNT(*) 
FROM simple_ssot_journals 
WHERE transaction_type = 'COGS';

-- List all COGS journals
SELECT 
  id,
  entry_number,
  transaction_number,
  date,
  total_amount,
  description
FROM simple_ssot_journals 
WHERE transaction_type = 'COGS'
ORDER BY date DESC;
```

### **2. Verify Journal Items**

```sql
-- Check journal items for specific COGS entry
SELECT 
  ji.journal_id,
  a.code,
  a.name,
  ji.debit,
  ji.credit,
  ji.description
FROM simple_ssot_journal_items ji
JOIN accounts a ON ji.account_id = a.id
WHERE ji.journal_id IN (
  SELECT id FROM simple_ssot_journals WHERE transaction_type = 'COGS'
)
ORDER BY ji.journal_id, ji.debit DESC;
```

### **3. Test P&L Report**

```bash
# Call P&L API
curl "http://localhost:8080/api/reports/ssot-profit-loss?start_date=2025-10-01&end_date=2025-10-31&format=json"
```

**Expected results:**
- `total_expenses` > 0 (was 0 before)
- `grossProfit` < total_revenue (was equal before)
- `grossProfitMargin` < 100% (was 100% before)

---

## 🔄 **Rollback**

If you need to rollback the migration:

```sql
-- Delete COGS journals created by migration
DELETE FROM simple_ssot_journal_items 
WHERE journal_id IN (
  SELECT id FROM simple_ssot_journals 
  WHERE transaction_type = 'COGS' 
    AND description LIKE '%Auto-migrated%'
);

DELETE FROM simple_ssot_journals 
WHERE transaction_type = 'COGS' 
  AND description LIKE '%Auto-migrated%';
```

---

## 📊 **Impact on Reports**

After migration, P&L reports will show:

### **Before Migration:**
```
Revenue:        Rp 5,000,000
COGS:           Rp 0           ❌
------------------------------
Gross Profit:   Rp 5,000,000  ❌ (100% margin)
Net Income:     Rp 3,750,000  ❌
```

### **After Migration:**
```
Revenue:        Rp 5,000,000
COGS:           Rp 500,000    ✅
------------------------------
Gross Profit:   Rp 4,500,000  ✅ (90% margin)
Net Income:     Rp 3,375,000  ✅
```

---

## ⚠️ **Important Notes**

1. **Backup Database First**
   ```bash
   pg_dump sistem_akuntansi > backup_before_cogs_migration.sql
   ```

2. **Required Accounts Must Exist**
   - 5101 or 5001: HARGA POKOK PENJUALAN (COGS)
   - 1301: PERSEDIAAN BARANG DAGANGAN (Inventory)

3. **Safe to Re-run**
   - Script checks for existing COGS journals
   - Only creates missing ones
   - No duplicates will be created

4. **COGS Estimation**
   - If no purchase history found, uses 80% estimate
   - Review and adjust if your business has different margins

---

## 🆘 **Troubleshooting**

### **Error: "COGS account not found"**

Create the required account:

```sql
INSERT INTO accounts (code, name, type, is_active) 
VALUES ('5101', 'HARGA POKOK PENJUALAN', 'EXPENSE', true);
```

### **Error: "Inventory account not found"**

Create the required account:

```sql
INSERT INTO accounts (code, name, type, is_active) 
VALUES ('1301', 'PERSEDIAAN BARANG DAGANGAN', 'ASSET', true);
```

### **COGS Amount Seems Wrong**

Check your purchase history and product costs:

```sql
-- Check purchase prices for specific product
SELECT 
  p.code,
  pi.product_id,
  pr.name as product_name,
  pi.unit_price,
  pi.quantity,
  p.date
FROM purchase_items pi
JOIN purchases p ON pi.purchase_id = p.id
JOIN products pr ON pi.product_id = pr.id
WHERE pi.product_id = X  -- replace with your product ID
ORDER BY p.date DESC;
```

---

## 📞 **Support**

If you encounter issues:

1. Check migration logs for error messages
2. Verify database schema is up to date
3. Ensure required accounts (5101, 1301) exist
4. Review COGS calculation logic for your business model

---

**Migration Version:** 1.0  
**Created:** 2025-10-17  
**Status:** Production Ready

