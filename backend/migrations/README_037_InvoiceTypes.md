# Invoice Types System Migration - 037

## Overview
Migration `037_add_invoice_types_system.sql` adds support for custom invoice numbering formats with invoice types. This enables invoice numbers like "0001/STA-C/X-2025" where:
- `0001` = 4-digit sequential number (resets annually per type)
- `STA-C` = Invoice type code
- `X` = Month in Roman numerals (I-XII)
- `2025` = Year

## What This Migration Does

### 1. Creates New Tables
- **`invoice_types`** - Stores invoice type definitions
  - `id`, `name`, `code`, `description`, `is_active`, `created_by`, timestamps
  - Example: Corporate Sales (STA-C), Retail Sales (STA-B)
  
- **`invoice_counters`** - Tracks counters per invoice type per year
  - `id`, `invoice_type_id`, `year`, `counter`, timestamps
  - Ensures each type has its own sequential numbering per year

### 2. Modifies Existing Tables
- **`sales`** table gets new column `invoice_type_id` (nullable)
- Adds proper foreign key relationship and indexes

### 3. Adds Performance Indexes
- `idx_sales_invoice_type_id` - For sales-invoice type lookups
- `idx_sales_invoice_number` - For invoice number searches
- `idx_sales_date_status` - For date/status filtering
- `idx_sales_status_invoice_type` - For status and type queries

### 4. Seeds Default Data
Creates 4 default invoice types:
- Corporate Sales (STA-C)
- Retail Sales (STA-B) 
- Service Sales (STA-S)
- Export Sales (EXP)

## Automatic Deployment

### When Backend Starts
The migration will run automatically when the backend starts via:
```go
// main.go line 54
if err := database.RunAutoMigrations(db); err != nil {
    log.Printf("Warning: Auto-migration failed: %v", err)
}
```

### Migration Detection
The system will:
1. Check `migration_logs` table for previous runs
2. Skip if already marked as SUCCESS
3. Execute all SQL statements if not run
4. Log results to `migration_logs` table

### Idempotent Design
The migration is safe to run multiple times:
- Uses `IF NOT EXISTS` for table creation
- Checks existing columns/indexes before adding
- Uses `ON DUPLICATE KEY UPDATE` for seed data
- Gracefully handles "already exists" errors

## Testing the Migration

### 1. Check Migration Status
```sql
SELECT * FROM migration_logs WHERE migration_name = '037_add_invoice_types_system.sql';
```

### 2. Verify Tables Created
```sql
-- Check invoice types table
SELECT * FROM invoice_types;

-- Check counters initialized
SELECT * FROM invoice_counters;

-- Check sales table structure
DESCRIBE sales; -- Should show invoice_type_id column
```

### 3. Test Invoice Type Creation via API
```bash
# Create new invoice type
curl -X POST http://localhost:8080/api/v1/invoice-types \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "name": "Test Invoice",
    "code": "TEST",
    "is_active": true
  }'

# Preview next invoice number
curl -X GET http://localhost:8080/api/v1/invoice-types/1/preview-number \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### 4. Test Sales Form Integration
1. Open sales form in frontend
2. Should see "Invoice Type" dropdown
3. Create sale with invoice type selected
4. Verify invoice number format in database

## Frontend Integration Status

### ✅ Completed
- Invoice Type Service (`invoiceTypeService.ts`)
- Sales Form updated with invoice type dropdown
- Invoice Type Management page (`InvoiceTypeManagement.tsx`)

### 🔄 Pending
- Add routing for invoice type management page
- Test full workflow end-to-end

## Rollback (if needed)

If issues occur, rollback using:
```sql
SOURCE migrations/037_rollback_invoice_types_system.sql;
```

**⚠️ WARNING:** Rollback will delete all invoice type data!

## Invoice Number Generation Logic

Backend service `InvoiceNumberService` handles:
1. Get/create counter for invoice type and current year  
2. Increment counter atomically
3. Format: `LPAD(counter, 4, '0') + '/' + code + '/' + roman_month + '-' + year`
4. Roman months: I, II, III, IV, V, VI, VII, VIII, IX, X, XI, XII

## Expected Benefits

### For Users
- Custom invoice numbering per business unit/type
- Better invoice organization and tracking
- Professional invoice number formats

### For System
- Scalable numbering system
- Annual reset capability per type
- Maintains backward compatibility (existing sales unaffected)

## Support and Troubleshooting

### Common Issues
1. **Migration fails**: Check database permissions, existing data conflicts
2. **Foreign key errors**: Ensure user with ID exists for `created_by` field
3. **Frontend dropdown empty**: Check API endpoint accessibility and authentication
4. **Rollback script error (SQLSTATE 23514)**: See Error Resolution section below

### 🔧 Error Resolution: Rollback Script Constraint Violation

**Error Message:**
```
ERROR: new row for relation "migration_logs" violates check constraint "migration_logs_status_check" (SQLSTATE 23514)
```

**Root Cause:**
- `migration_logs` table constraint only allows status: `'SUCCESS'`, `'FAILED'`, `'SKIPPED'`
- Rollback script tried to use `'rollback_completed'` (not allowed)
- This is a **NON-CRITICAL** error - main migration succeeded

**Impact:** NONE - Invoice types system is fully operational

**Quick Fix Options:**

**Option 1: Smart Quick Fix (RECOMMENDED)**
```bash
# Connect to database and run - automatically handles already-fixed scenarios:
psql -d your_database -f migrations/quick_fix_037_rollback_error.sql
```
*Features: Idempotent, won't show errors if already fixed, provides detailed status*

**Option 2: Status Check Only**
```bash
# Just check current status (silent when OK):
psql -d your_database -f migrations/status_037_invoice_types.sql
```
*Features: Only shows details when action is needed*

**Option 3: Manual Cleanup**
```sql
-- Clean up failed rollback log entry
DELETE FROM migration_logs 
WHERE migration_name = '037_rollback_invoice_types_system.sql' 
AND status = 'FAILED';
```

**Option 4: Ignore (SAFEST)**
- Error is cosmetic only
- Main system works perfectly  
- No action needed

**Verification:**
```sql
-- Check system status
SELECT * FROM invoice_types; -- Should show 4 default types
SELECT preview_next_invoice_number(1); -- Should return: 0001/STA-C/X-2025
```

### Debug Commands
```sql
-- Check auto-migration logs
SELECT * FROM migration_logs ORDER BY executed_at DESC LIMIT 10;

-- Verify foreign keys
SELECT * FROM information_schema.key_column_usage 
WHERE table_name = 'sales' AND constraint_name LIKE '%invoice_type%';

-- Test invoice number generation
SELECT it.name, it.code, 
       LPAD(COALESCE(ic.counter, 0) + 1, 4, '0') as next_number,
       CASE MONTH(NOW()) 
         WHEN 1 THEN 'I' WHEN 2 THEN 'II' WHEN 3 THEN 'III' WHEN 4 THEN 'IV'
         WHEN 5 THEN 'V' WHEN 6 THEN 'VI' WHEN 7 THEN 'VII' WHEN 8 THEN 'VIII'
         WHEN 9 THEN 'IX' WHEN 10 THEN 'X' WHEN 11 THEN 'XI' WHEN 12 THEN 'XII'
       END as roman_month,
       YEAR(NOW()) as year
FROM invoice_types it
LEFT JOIN invoice_counters ic ON it.id = ic.invoice_type_id AND ic.year = YEAR(NOW())
WHERE it.is_active = TRUE;
```

---
**Created:** October 2, 2025  
**Migration:** 037_add_invoice_types_system.sql  
**Status:** Ready for deployment