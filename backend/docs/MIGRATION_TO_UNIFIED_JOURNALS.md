# Migration to Unified Journals System (SSOT)

**Date:** 2025-11-07  
**Status:** Ready for Execution  
**Migration Type:** Full Migration to SSOT/Unified System

---

## Executive Summary

This migration converts the accounting system from using dual journal systems (old `journal_entries` + unified journals) to using **UNIFIED JOURNALS ONLY** (SSOT - Single Source of Truth).

### Current Problem:
- System has 2 journal systems running simultaneously
- `journal_entries` table: 2 entries (closing only)
- `unified_journal_ledger` table: 5 entries (actual transactions)
- `accounts.balance` is corrupted - mix of both systems
- **Balance Sheet NOT BALANCED: -Rp 68.300.000 difference**

### Solution:
✅ **OPTION A - Use Unified Journals Only**
- Delete old journal_entries (2 closing entries)
- Reset accounts.balance = 0
- Recalculate balances from unified_journal_ledger
- Balance sheet will use unified journals exclusively

---

## Pre-Migration Status

### Old Journal System (`journal_entries`):
```
Total Entries: 2
- CLO-2025-11-12-31 (Closing 2025)
- CLO-2026-01-12-31 (Closing 2026)
Status: Will be DELETED
```

### Unified Journal System (`unified_journal_ledger`):
```
Total Entries: 5
Total Lines: 18
Status: KEPT - This is the source of truth
```

### Account Balances (BEFORE Migration):
```
From accounts table:
- BANK (1102): Rp 108.780.000 (mixed source)
- BANK UOB (1104): Rp 16.700.000 (mixed source)
- PPN MASUKAN (1240): Rp 4.950.000 (mixed source)
- PERSEDIAAN (1301): -Rp 1.000.000 (negative - error!)
- UTANG USAHA (2101): Rp 33.300.000 (mixed source)
- PPN KELUARAN (2103): Rp 12.430.000 (mixed source)
- MODAL PEMILIK (3101): Rp 50.000.000 (mixed source)
- LABA DITAHAN (3201): Rp 102.000.000 (from old closings)

Balance Sheet: NOT BALANCED (-68.3jt diff)
```

### Expected After Migration:
```
From unified_journal_ledger only:
- BANK (1102): Rp 125.430.000
- BANK UOB (1104): Rp 50.000.000
- PPN MASUKAN (1240): Rp 4.950.000
- PERSEDIAAN (1301): -Rp 1.000.000 (still negative - need investigation)
- UTANG USAHA (2101): Rp 49.950.000
- PPN KELUARAN (2103): Rp 12.430.000
- MODAL PEMILIK (3101): Rp 50.000.000
- PENDAPATAN (4101): Rp 113.000.000 (NOT closed yet)
- HPP (5101): Rp 46.000.000 (NOT closed yet)
- LABA DITAHAN (3201): Rp 0 (not yet closed in unified)

Balance Sheet: BALANCED (with temp accounts)
Net Income: 113jt - 46jt = 67jt
```

---

## Migration Steps

### Step 1: Backup Database ✅

**CRITICAL: DO THIS FIRST!**

```bash
# Using PowerShell Script (Recommended):
.\run_migration_to_unified.ps1

# Or Manual:
pg_dump -U postgres sistem_akuntansi > backup_before_migration.sql
```

Backup location: `backend/backups/backup_before_unified_migration_[timestamp].sql`

### Step 2: Run Migration Script

#### Option A: Using PowerShell (Recommended) ✅

```powershell
# DRY RUN (test without committing)
.\run_migration_to_unified.ps1 -DryRun

# PRODUCTION (actual migration)
.\run_migration_to_unified.ps1
```

#### Option B: Manual SQL Execution

```bash
# If you have psql installed:
psql -U postgres -d sistem_akuntansi -f migrate_to_unified_journals.sql

# If psql not available, use alternative tool or install PostgreSQL client
```

### Step 3: Verify Migration

The script will automatically verify:
- ✅ Old journal entries deleted
- ✅ Account balances reset to 0
- ✅ Balances recalculated from unified journals
- ✅ Balance sheet equation validated

### Step 4: Run Period Closing

**IMPORTANT:** Revenue & Expense accounts still have balances after migration.
You need to close the period to transfer Net Income to Retained Earnings.

#### Via Application UI:
1. Go to Period Closing menu
2. Set date range: 2025-01-01 to 2025-12-31 (or your fiscal year)
3. Click "Preview" to see closing entries
4. If OK, click "Execute" to close the period

#### Via API:
```bash
# Preview
curl -X POST http://localhost:8080/api/period-closing/preview \
  -H "Content-Type: application/json" \
  -d '{
    "start_date": "2025-01-01",
    "end_date": "2025-12-31",
    "description": "Annual Closing 2025"
  }'

# Execute (if preview OK)
curl -X POST http://localhost:8080/api/period-closing/execute \
  -H "Content-Type: application/json" \
  -d '{
    "start_date": "2025-01-01",
    "end_date": "2025-12-31",
    "description": "Annual Closing 2025"
  }'
```

### Step 5: Final Verification

After period closing:
```bash
# Run balance check
go run cmd/complete_balance_check.go
```

Expected result:
- ✅ Revenue accounts = 0
- ✅ Expense accounts = 0
- ✅ Retained Earnings = 50jt (Modal) + 67jt (Net Income) = 117jt
- ✅ Balance Sheet BALANCED (difference = 0)

---

## What Changes

### Tables Affected:

| Table | Action | Impact |
|-------|--------|--------|
| `journal_entries` | DELETE ALL | Old journal system removed |
| `journal_lines` | DELETE ALL | Old journal lines removed |
| `accounting_periods` | DELETE ALL | Old closing records removed |
| `accounts` | UPDATE balance = 0, then recalc | Reset then recalculate from unified |
| `unified_journal_ledger` | NO CHANGE | Source of truth (kept as-is) |
| `unified_journal_lines` | NO CHANGE | Source of truth (kept as-is) |

### Services Using Unified Journals:

✅ **Already using unified journals:**
- Trial Balance Report
- General Ledger Report
- Balance Sheet Report (SSOT version)
- Profit & Loss Report
- Cash Flow Report
- All SSOT reports

❌ **Old system (will be obsolete after migration):**
- `period_closing_service.go` - uses old journal_entries
- Need to update or use new unified closing

---

## Rollback Plan

If something goes wrong:

### Immediate Rollback:
```bash
# Restore from backup
psql -U postgres -d sistem_akuntansi -f backups/backup_before_unified_migration_[timestamp].sql
```

### Validate After Rollback:
```bash
# Check data restored
go run cmd/complete_balance_check.go
go run cmd/check_all_journals.go
```

---

## Post-Migration Validation Checklist

- [ ] Old journal_entries table is empty
- [ ] Old journal_lines table is empty
- [ ] accounts.balance matches unified journal calculations
- [ ] Balance sheet shows Revenue & Expense with balances (before closing)
- [ ] Total Assets = Liabilities + Equity + (Revenue - Expense)
- [ ] All reports show data from unified journals
- [ ] Period closing executed successfully
- [ ] After closing: Revenue = 0, Expense = 0
- [ ] After closing: Balance Sheet fully balanced
- [ ] Retained Earnings = Opening + Net Income

---

## Known Issues & Fixes

### Issue 1: Persediaan Negatif (-1jt)
**Status:** Exists in unified journals  
**Impact:** Asset should not be negative  
**Action:** Investigate inventory transactions after migration

### Issue 2: Duplicate Account Codes
**Status:** Some accounts have duplicate codes with different IDs  
**Impact:** Trial balance may show duplicates  
**Action:** Consolidate duplicate accounts after migration

### Issue 3: Period Closing in Old System
**Status:** Old period_closing_service uses journal_entries  
**Impact:** Won't work after migration  
**Action:** Use unified journal closing or update service

---

## FAQ

### Q: Can I undo the migration?
**A:** Yes, restore from backup. But you'll lose any new transactions created after migration.

### Q: What happens to new transactions?
**A:** After migration, all transactions must be created in unified_journal_ledger. Old journal_entries is no longer used.

### Q: Will reports still work?
**A:** Yes, all SSOT reports already use unified journals. They will work better after migration.

### Q: Do I need to re-enter data?
**A:** No, all transaction data is preserved in unified_journal_ledger. Only corrupt closing entries are removed.

### Q: How long does migration take?
**A:** < 1 minute for backup, < 5 seconds for migration script execution.

### Q: Can I run in production?
**A:** Yes, but:
1. Schedule during off-hours
2. Backup first (MANDATORY)
3. Test with -DryRun first
4. Have rollback plan ready

---

## Technical Details

### Migration SQL Logic:

```sql
-- 1. Delete old journals
DELETE FROM journal_lines;
DELETE FROM journal_entries;
DELETE FROM accounting_periods;

-- 2. Reset balances
UPDATE accounts SET balance = 0;

-- 3. Recalculate from unified
UPDATE accounts a
SET balance = (
    -- Calculate from unified_journal_lines
    -- Based on account type (Debit normal vs Credit normal)
)
FROM unified_journal_lines ujl
JOIN unified_journal_ledger uj ON uj.id = ujl.journal_id
WHERE uj.status = 'POSTED';
```

### Balance Calculation Logic:

**Asset & Expense (Debit Normal):**
```
Balance = SUM(debit_amount) - SUM(credit_amount)
```

**Liability, Equity, Revenue (Credit Normal):**
```
Balance = SUM(credit_amount) - SUM(debit_amount)
```

---

## Support & Contact

If issues occur:
1. Check backup exists
2. Review migration logs
3. Run diagnostic scripts:
   - `go run cmd/check_unified_journals.go`
   - `go run cmd/complete_balance_check.go`
4. Restore from backup if needed
5. Report issues with logs

---

## Files Created

1. **migrate_to_unified_journals.sql** - Main migration script
2. **run_migration_to_unified.ps1** - PowerShell execution script
3. **This document** - Complete migration guide

---

**READY TO EXECUTE**  
Review this document, then run: `.\run_migration_to_unified.ps1`
