# Seed Error Fixes - 2025-10-28

## Problem Summary
After update, 3 errors appeared in the backend during account seeding:

### Error 1 & 2: "record not found" logs
```
2025/10/28 07:24:12 /Users/.../account_seed_improved.go:183 record not found
[0.125ms] [rows:0] SELECT * FROM "accounts" WHERE code = '1116' ...
```
- **Cause**: GORM logs every `ErrRecordNotFound` at INFO level when checking if account exists
- **Impact**: Clutters logs with expected behavior (checking before creating new accounts)

### Error 3: Cannot change account category
```
2025/10/28 07:24:12 /Users/.../account_seed_improved.go:219 ERROR: 
🔒 BLOCKED: Cannot change account category for system critical account: 4101 
(PENDAPATAN PENJUALAN) from  to OPERATING_REVENUE (SQLSTATE P0001)
```
- **Cause**: Migration 040 added `is_system_critical` flag to protect critical accounts (1201, 4101, 2103, 2001)
- **Issue**: Seed function tried to update `type`, `category`, and `is_active` fields on existing critical accounts
- **Trigger**: Database trigger `prevent_critical_account_update()` blocks changes to critical accounts

## Solution Implemented

### 1. Suppress Expected "Record Not Found" Logs
```go
// Use Silent mode for expected not-found queries
err := tx.Session(&gorm.Session{Logger: tx.Logger.LogMode(4)}). // 4 = Silent mode
    Clauses(clause.Locking{Strength: "UPDATE"}).
    Where("code = ?", account.Code).
    Where("deleted_at IS NULL").
    First(&existingAccount).Error
```
- Only query for checking existence uses silent mode
- Other errors still logged normally

### 2. Skip Protected Fields for Critical Accounts
```go
// Build update map with safe fields
updates := map[string]interface{}{
    "name":        normalizedName,
    "level":       account.Level,
    "is_header":   account.IsHeader,
    "description": account.Description,
}

// Only update protected fields if NOT system critical
if !existingAccount.IsSystemCritical {
    updates["type"] = account.Type
    updates["category"] = account.Category
    updates["is_active"] = account.IsActive
}
```
- Safe fields (name, level, is_header, description) always updated
- Protected fields (type, category, is_active) only updated if NOT critical
- Preserves database integrity without triggering protection trigger

## Critical Accounts (Protected)
These accounts are marked as `is_system_critical = TRUE` and have restricted updates:
- **1201** - PIUTANG USAHA (Sales Receivable)
- **4101** - PENDAPATAN PENJUALAN (Sales Revenue) 
- **2103** - PPN KELUARAN (Output VAT)
- **2001** - HUTANG USAHA (Purchase Payable)

## Result
✅ All seed errors resolved
✅ Backend starts cleanly without errors
✅ Critical accounts protected from accidental changes
✅ Non-critical accounts still fully updateable
✅ Clean logs without spam

## Migration Reference
- Migration: `040_lock_critical_tax_accounts.sql`
- Trigger: `prevent_critical_account_update()`
- Protection: Prevents changes to `code`, `type`, `category`, `is_active`, and deletion
