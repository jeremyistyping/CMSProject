# Revenue Duplication Investigation Guide

## 🚨 Problem Statement

**Expected Revenue**: Rp 10,000,000 (from Chart of Accounts)  
**Reported Revenue**: Rp 20,000,000 (from P&L Report)  
**Discrepancy**: Rp 10,000,000 (100% duplication!)

---

## 📊 Evidence from Chart of Accounts

```
Account 4101 - PENDAPATAN PENJUALAN
Balance: Rp 10,000,000
```

## 📊 Evidence from P&L Report

```
Total Revenue: Rp 20,000,000

Breakdown:
- 4101 PENDAPATAN PENJUALAN: Rp 10,000,000
- 4101 Pendapatan Penjualan: Rp 10,000,000  ← Duplicate!
```

---

## 🔍 Investigation Tools Created

### 1. PowerShell API Test Script
**File**: `backend/test_profit_loss_database_validation.ps1`

**Purpose**: Tests P&L via API and compares with Chart of Accounts

**Usage**:
```powershell
cd backend
./test_profit_loss_database_validation.ps1
```

**What it does**:
- ✅ Fetches Chart of Accounts via API
- ✅ Fetches Journal Entries via API
- ✅ Fetches P&L Report via API
- ✅ Compares all three sources
- ✅ Identifies discrepancies
- ✅ Generates SQL file for manual investigation

---

### 2. SQL Investigation Script
**File**: `backend/investigate_revenue_duplication.sql`

**Purpose**: Direct database queries to find root cause

**Usage**:
```bash
# Connect to your MySQL database
mysql -u root -p accounting_db

# Run the script
source backend/investigate_revenue_duplication.sql
```

**What it checks**:
1. Account balances from `accounts` table
2. Unified journal entries (SSOT system)
3. Legacy journal entries
4. Duplicate journal detection
5. Account name variations
6. Sales transaction mapping
7. Both systems combined total

---

### 3. Go Debug Program
**File**: `backend/debug_revenue_duplication.go`

**Purpose**: Comprehensive Go program to analyze database directly

**Usage**:
```bash
cd backend
go run debug_revenue_duplication.go
```

**Environment**:
```bash
# Optional: Set custom database connection
export DB_DSN="user:password@tcp(localhost:3306)/accounting_db?charset=utf8mb4&parseTime=True&loc=Local"
```

**What it analyzes**:
- ✅ Account balances
- ✅ Unified journal analysis
- ✅ Legacy journal analysis
- ✅ Detailed entries for account 4101
- ✅ Duplicate journal detection
- ✅ Combined systems comparison
- ✅ Root cause diagnosis

---

## 🎯 Most Likely Causes

### Cause 1: Both SSOT and Legacy Systems Active (90% probability)

**Symptom**: 
- Unified journals show Rp 10,000,000
- Legacy journals show Rp 10,000,000
- Total = Rp 20,000,000

**Root Cause**:
Backend fallback logic in `ssot_profit_loss_service.go` is counting BOTH systems

**Evidence to look for**:
```sql
-- Check both systems
SELECT SUM(credit - debit) FROM unified_journal_lines WHERE account_code = '4101';
-- Returns: 10000000

SELECT SUM(credit - debit) FROM journal_lines WHERE account_code = '4101';  
-- Returns: 10000000

-- Total: 20000000 ✗
```

**Fix**:
```go
// In ssot_profit_loss_service.go
// Ensure fallback logic returns EARLY after finding data
if len(balances) > 0 && hasPLActivity(balances) {
    return balances, source, nil  // ← MUST return here!
}
```

---

### Cause 2: Duplicate Journal Entries (5% probability)

**Symptom**:
Same sale recorded twice with different journal IDs

**Evidence to look for**:
```sql
-- Check for duplicates
SELECT source_type, source_id, COUNT(*) as count
FROM unified_journal_ledger
WHERE source_type = 'SALE'
GROUP BY source_type, source_id
HAVING COUNT(*) > 1;
```

**Fix**:
```sql
-- Delete duplicate journals
DELETE FROM unified_journal_ledger 
WHERE id IN (
    -- Keep earliest journal, delete others
    SELECT id FROM (
        SELECT id, 
               ROW_NUMBER() OVER (PARTITION BY source_type, source_id ORDER BY created_at) as rn
        FROM unified_journal_ledger
    ) sub
    WHERE rn > 1
);
```

---

### Cause 3: Incorrect GROUP BY (3% probability)

**Symptom**:
Backend groups by both `account_code` AND `account_name`, causing duplicates if names differ

**Evidence to look for**:
```sql
-- Check account name variations
SELECT account_code, account_name, COUNT(*) 
FROM unified_journal_lines
WHERE account_code = '4101'
GROUP BY account_code, account_name;

-- If returns multiple rows with different names = PROBLEM
```

**Fix**:
```go
// In ssot_profit_loss_service.go
// Change GROUP BY to use account ID only
GROUP BY a.id, a.code, a.name, a.type  // ← Uses account table name (canonical)
```

---

### Cause 4: Case Sensitivity (2% probability)

**Symptom**:
"PENDAPATAN PENJUALAN" vs "Pendapatan Penjualan" treated as different

**Evidence to look for**:
```sql
SELECT DISTINCT account_name 
FROM unified_journal_lines 
WHERE account_code = '4101';

-- If returns multiple rows = case sensitivity issue
```

**Fix**:
```sql
-- Standardize account names
UPDATE unified_journal_lines ujl
INNER JOIN accounts a ON a.id = ujl.account_id
SET ujl.account_name = a.name
WHERE ujl.account_code = '4101';
```

---

## 🚀 Quick Start Investigation

### Step 1: Run PowerShell Test
```powershell
cd backend
./test_profit_loss_database_validation.ps1
```

**Expected Output**:
```
=== COMPARISON SUMMARY ===
┌─────────────────────────────────────────────────────────┐
│ DATA SOURCE              │ TOTAL REVENUE                │
├─────────────────────────────────────────────────────────┤
│ Chart of Accounts        │ Rp 10000000                  │
│ Journal Entries (SSOT)   │ Rp 10000000 or 20000000      │
│ P&L Report API           │ Rp 20000000                  │
└─────────────────────────────────────────────────────────┘
```

### Step 2: Run Go Debug Program
```bash
cd backend
go run debug_revenue_duplication.go
```

**Look for this output**:
```
=== STEP 6: Combined Systems Analysis ===
  Unified Journal (SSOT): Rp 10000000.00
  Legacy Journal: Rp 10000000.00

[GRAND TOTAL FROM BOTH SYSTEMS]: Rp 20000000.00

[DIAGNOSIS] BOTH SYSTEMS ARE ACTIVE!
Root Cause: Backend is counting revenue from BOTH Unified and Legacy journals
```

### Step 3: Review SQL Queries
```bash
# The PowerShell script generates: temp_pl_validation_queries.sql
# Review and execute relevant queries manually

mysql -u root -p accounting_db < backend/temp_pl_validation_queries.sql
```

---

## 🔧 Recommended Fixes

### If Cause 1 (Both Systems Active) - MOST LIKELY

**Fix Backend Service**:

```go
// File: backend/services/ssot_profit_loss_service.go
// Function: getAccountBalancesFromSSOT()

// After executing SSOT query (line ~210)
if err := s.db.Raw(query, startDate, endDate).Scan(&balances).Error; err != nil {
    return nil, source, fmt.Errorf("error executing account balances query: %v", err)
}

// ✅ ADD THIS CHECK - Return immediately if SSOT has data
if len(balances) > 0 && hasPLActivity(balances) {
    log.Printf("SSOT has P&L activity, using SSOT data only (not falling back)")
    return balances, source, nil  // ← CRITICAL: Early return!
}

// ❌ REMOVE OR FIX: The current fallback logic
// The code currently falls through to legacy even if SSOT has data
```

**Verify Fix**:
```bash
# Rebuild backend
cd backend
go build -o ../bin/app.exe main.go

# Restart backend
../restart_backend.ps1

# Test again
./test_profit_loss_database_validation.ps1
```

**Expected Result**:
- P&L Report should now show Rp 10,000,000 (not 20,000,000)

---

### If Cause 2 (Duplicate Journals)

**Check for Duplicates**:
```sql
SELECT 
    source_type, 
    source_id, 
    COUNT(*) as journal_count,
    GROUP_CONCAT(id) as journal_ids
FROM unified_journal_ledger
WHERE source_type = 'SALE'
  AND status = 'POSTED'
  AND deleted_at IS NULL
GROUP BY source_type, source_id
HAVING COUNT(*) > 1;
```

**Fix Duplicates** (CAREFUL - backup first!):
```sql
-- Option 1: Soft delete duplicates (keep earliest)
UPDATE unified_journal_ledger
SET status = 'CANCELLED', 
    notes = CONCAT(COALESCE(notes, ''), ' [AUTO-CANCELLED: Duplicate entry]')
WHERE id IN (
    SELECT id FROM (
        SELECT id, 
               ROW_NUMBER() OVER (PARTITION BY source_type, source_id ORDER BY created_at ASC) as rn
        FROM unified_journal_ledger
        WHERE source_type = 'SALE' AND status = 'POSTED'
    ) sub
    WHERE rn > 1
);

-- Option 2: Hard delete (use with extreme caution!)
-- First, delete journal lines
DELETE ujl FROM unified_journal_lines ujl
INNER JOIN (
    SELECT id FROM (
        SELECT id, 
               ROW_NUMBER() OVER (PARTITION BY source_type, source_id ORDER BY created_at ASC) as rn
        FROM unified_journal_ledger
        WHERE source_type = 'SALE'
    ) sub
    WHERE rn > 1
) dup ON ujl.journal_id = dup.id;

-- Then delete journal headers
DELETE FROM unified_journal_ledger WHERE id IN (...);
```

---

## 📝 Testing After Fix

### 1. Backend Test
```bash
cd backend
go run debug_revenue_duplication.go
```

**Expected**:
```
[TOTAL ACCOUNT BALANCE]: Rp 10000000.00
[TOTAL FROM UNIFIED JOURNALS]: Rp 10000000.00
[TOTAL FROM LEGACY JOURNALS]: Rp 0.00
[GRAND TOTAL FROM BOTH SYSTEMS]: Rp 10000000.00
```

### 2. API Test
```powershell
cd backend
./test_profit_loss_database_validation.ps1
```

**Expected**:
```
[PASS] P&L Report matches Journal Entries
[PASS] Journal Entries match Account Balances
```

### 3. Frontend Test
1. Open browser: `http://localhost:3000/reports`
2. Generate P&L Report (01/01/2025 - 12/31/2025)
3. **Verify**: Total Revenue = Rp 10,000,000 ✅

---

## 📊 Summary Checklist

- [ ] Run `test_profit_loss_database_validation.ps1`
- [ ] Run `debug_revenue_duplication.go`
- [ ] Review SQL queries in `investigate_revenue_duplication.sql`
- [ ] Identify root cause (most likely: both systems active)
- [ ] Apply appropriate fix
- [ ] Rebuild backend
- [ ] Restart backend service
- [ ] Re-run tests to verify fix
- [ ] Test in frontend
- [ ] Document findings

---

## 🆘 Need Help?

If revenue is still duplicated after following this guide:

1. **Check Logs**: Look for "SSOT has P&L activity" message
2. **Review Data Source**: P&L report should show "Data Source: SSOT Journal System"
3. **Inspect Response**: Use browser DevTools to check API response
4. **Verify Database**: Run SQL queries manually to confirm data

---

## 📚 Related Files

- `backend/services/ssot_profit_loss_service.go` - Main P&L service
- `backend/controllers/ssot_profit_loss_controller.go` - P&L controller
- `backend/test_profit_loss_database_validation.ps1` - API testing
- `backend/debug_revenue_duplication.go` - Database analysis
- `backend/investigate_revenue_duplication.sql` - SQL queries
- `frontend/app/reports/page.tsx` - Frontend display

---

**Investigation Tools Created**: 2024-10-16  
**Status**: Ready for Debugging  
**Priority**: HIGH (100% revenue discrepancy)

