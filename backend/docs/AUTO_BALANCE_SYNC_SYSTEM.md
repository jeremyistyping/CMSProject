# Auto Balance Sync System

## Overview

The Auto Balance Sync System is a comprehensive solution that automatically synchronizes cash bank balances with their linked Chart of Accounts (COA) accounts and maintains parent account balance rollup. This system runs automatically when the backend starts up, ensuring balance consistency across all environments.

## Features

✅ **Automatic Cash Bank to COA Synchronization**
- Cash bank balances are automatically synced with their linked COA accounts
- Triggers fire on every cash bank transaction insert, update, or delete

✅ **Parent Account Balance Rollup**
- Parent accounts automatically aggregate child account balances
- Recursive balance updates maintain hierarchy consistency

✅ **Unified Trigger Approach**
- Single trigger handles both cash bank balance calculation and COA sync
- Avoids trigger execution order issues

✅ **BIGINT Parameter Compatibility**
- All database functions use BIGINT for account IDs
- Ensures compatibility with large datasets

✅ **Account Hierarchy Validation**
- Cash bank accounts are automatically configured as non-header accounts
- Prevents balance calculation issues

✅ **Manual Sync Functions**
- Manual sync functions available for maintenance and debugging
- Comprehensive sync for all cash banks or individual accounts

## How It Works

### 1. Auto Migration Integration

The balance sync system is integrated into the auto migration system and runs automatically on every backend startup:

```bash
cd backend
go run cmd/scripts/run_auto_migrations.go
```

### 2. Database Components

**Functions:**
- `update_parent_account_balances(BIGINT)` - Recursively updates parent balances
- `recalculate_cashbank_balance()` - Unified balance calculation and COA sync
- `validate_account_balance_consistency()` - Validates account balance consistency  
- `manual_sync_cashbank_coa(BIGINT)` - Manual sync for specific cash bank
- `manual_sync_all_cashbank_coa()` - Manual sync for all cash banks
- `ensure_cashbank_not_header()` - Ensures cash bank accounts are non-header

**Triggers:**
- `trigger_recalc_cashbank_balance_insert` - On cash bank transaction insert
- `trigger_recalc_cashbank_balance_update` - On cash bank transaction update
- `trigger_recalc_cashbank_balance_delete` - On cash bank transaction delete
- `trigger_validate_account_balance` - On account balance update

### 3. Migration Files

The system is installed via comprehensive migration files:
- `20250930_comprehensive_auto_balance_sync.sql` - Main installation migration
- Includes all fixes and improvements from previous versions

## Usage on Different PCs

### When Git Pulling to a New PC

When you `git pull` the project to a different PC, the auto balance sync system will be automatically installed and configured when you start the backend:

1. **Git pull the project:**
   ```bash
   git pull origin main
   ```

2. **Start the backend (any method):**
   ```bash
   # Method 1: Direct run
   go run main.go
   
   # Method 2: Build and run
   go build -o app.exe
   ./app.exe
   
   # Method 3: Run auto migrations manually first (optional)
   go run cmd/scripts/run_auto_migrations.go
   ```

3. **System automatically installs:**
   - The auto migration system runs on startup
   - Checks if balance sync system is installed
   - Installs/updates if missing or outdated
   - Configures cash bank accounts
   - Performs initial balance synchronization

### Database Requirements

The system works with your existing PostgreSQL database. No manual intervention required.

**Compatible with:**
- PostgreSQL 12+
- Existing accounting database schemas
- Multiple cash bank accounts
- Complex account hierarchies

## System Status Verification

To verify the system is working correctly, check the logs during backend startup. You should see:

```
============================================
🔧 STARTING COMPREHENSIVE BALANCE SYNC SYSTEM SETUP
============================================
🔍 Checking balance sync system status...
📊 Current status -> Triggers: true, Functions: true
✅ Balance sync system is already installed and up-to-date
🏦 Ensuring cash bank account configuration...
   📊 Found X cash banks to configure
✅ Cash bank account configuration completed
⚖️  Performing initial balance synchronization...
✅ Initial sync completed: Synced X cash bank accounts with their COA accounts
✅ Balance sync system setup completed successfully
✅ BALANCE SYNC SYSTEM SETUP COMPLETED SUCCESSFULLY
============================================
```

## Manual Operations

### Manual Sync All Cash Banks
```sql
SELECT manual_sync_all_cashbank_coa();
```

### Manual Sync Specific Cash Bank
```sql
SELECT manual_sync_cashbank_coa(123); -- Replace 123 with cash bank ID
```

### Ensure Non-Header Status
```sql
SELECT ensure_cashbank_not_header();
```

## Troubleshooting

### Issue: Balance sync system not installing
**Solution:** Check database connection and permissions

### Issue: Triggers not firing
**Solution:** Verify triggers exist:
```sql
SELECT trigger_name FROM information_schema.triggers 
WHERE trigger_name LIKE 'trigger_recalc_cashbank%';
```

### Issue: Functions missing
**Solution:** Verify functions exist:
```sql
SELECT proname FROM pg_proc WHERE proname LIKE '%cashbank%';
```

### Issue: Balance discrepancies
**Solution:** Run manual sync:
```sql
SELECT manual_sync_all_cashbank_coa();
```

## Performance Considerations

- **Indexes Created:** Performance indexes are automatically created
- **Trigger Efficiency:** Triggers only fire on actual balance changes
- **Recursive Limits:** Parent rollup has safety limits to prevent infinite loops
- **Transaction Safety:** All operations are transaction-safe

## Migration History

The system maintains migration logs for audit and troubleshooting:

```sql
SELECT migration_name, status, executed_at 
FROM migration_logs 
WHERE migration_name LIKE '%balance%' 
ORDER BY executed_at DESC;
```

## Support

This system is designed to be zero-maintenance and self-healing. It automatically:
- Detects and fixes configuration issues
- Handles database schema changes
- Maintains balance consistency
- Provides comprehensive logging

For any issues, check the backend logs during startup for detailed diagnostics.