# Migration to Unified Journals System - Quick Guide

## 🚀 Quick Start

### Option 1: AUTOMATIC - via main.go (EASIEST!) ✅✨

**Migration runs AUTOMATICALLY when you start the backend!**

```bash
# Just run the backend - migration happens automatically!
go run cmd/main.go
```

The migration will:
- ✅ Check if old journal_entries exist
- ✅ Auto-migrate if needed (with transaction safety)
- ✅ Skip if already migrated
- ✅ Show balance sheet summary

**That's it!** No manual steps needed.

---

### Option 2: Manual - Standalone Script

```bash
# 1. DRY RUN (test tanpa commit)
go run cmd/migrate_to_unified.go --dry-run

# 2. PRODUCTION (apply changes)
go run cmd/migrate_to_unified.go
```

### Option 3: Using PowerShell

```powershell
# DRY RUN
.\run_migration_to_unified.ps1 -DryRun

# PRODUCTION
.\run_migration_to_unified.ps1
```

### Option 4: Manual SQL (Advanced)

```bash
psql -U postgres -d sistem_akuntansi -f migrate_to_unified_journals.sql
```

---

## 📋 What It Does

1. ✅ **Deletes** old `journal_entries` and `journal_lines` (2 closing entries)
2. ✅ **Resets** all `accounts.balance` to 0
3. ✅ **Recalculates** balances from `unified_journal_ledger`
4. ✅ **Validates** balance sheet equation

---

## 🎯 Expected Result

### BEFORE Migration:
```
Assets: Rp 129.430.000
Liabilities: Rp 45.730.000
Equity: Rp 152.000.000
Difference: -Rp 68.300.000 ❌ NOT BALANCED
```

### AFTER Migration:
```
Assets: Rp 179.380.000
Liabilities: Rp 62.380.000
Equity: Rp 50.000.000
Revenue: Rp 113.000.000 (not closed)
Expense: Rp 46.000.000 (not closed)
Net Income: Rp 67.000.000
Difference: Rp 0 ✅ BALANCED (with temp accounts)
```

### AFTER Period Closing:
```
Assets: Rp 179.380.000
Liabilities: Rp 62.380.000
Equity: Rp 117.000.000 (50jt + 67jt)
Revenue: Rp 0
Expense: Rp 0
Difference: Rp 0 ✅ FULLY BALANCED
```

---

## ⚠️ Important Notes

1. **Backup First!** - Go program uses transaction (auto-rollback on error), but backup is still recommended
2. **DRY RUN First!** - Always test with `--dry-run` before production
3. **Period Closing Needed!** - After migration, run period closing to close Revenue/Expense
4. **No Downtime!** - Transaction-based, either all succeed or all rollback

---

## 🔍 Verification

```bash
# Check balance sheet
go run cmd/complete_balance_check.go

# Check unified journals
go run cmd/check_unified_journals.go
```

---

## 🆘 Rollback (if needed)

Go program uses transaction - if error occurs, changes are automatically rolled back.

But if you need manual rollback:
```bash
# Restore from backup (if you created one)
psql -U postgres -d sistem_akuntansi -f backups/backup_[timestamp].sql
```

---

## 📖 Full Documentation

See: `docs/MIGRATION_TO_UNIFIED_JOURNALS.md`

---

## 🎯 RECOMMENDED: Just Start The Backend!

```bash
go run cmd/main.go
```

Migration runs automatically! Check the logs for:
```
⚡ Checking migration to Unified Journals...
📝 Step 1/4: Deleting old journal_lines...
📝 Step 2/4: Deleting old journal_entries...
📝 Step 3/4: Deleting old accounting_periods...
📝 Step 4/4: Resetting account balances...
📊 Recalculating balances from unified_journal_ledger...
✅ Migration to Unified Journals completed successfully!
📊 Balance Sheet Summary:
```
