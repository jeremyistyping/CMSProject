# Duplicate Journal Entry Fix - CRITICAL BUG

## Problem Description

**CRITICAL BUG FOUND**: Sales journals are being created TWICE, causing **DOUBLE POSTING** to COA accounts!

**Symptoms**:
- Bank account (1102) shows Rp 11.100.000 instead of Rp 5.550.000 (exactly 2x)
- Every sales transaction posts to COA **TWICE**
- This affects ALL account balances

## Root Cause Analysis

### Issue 1: No Duplicate Prevention

**File**: `backend/services/sales_journal_service_v2.go`
**Function**: `CreateSalesJournal()` (Line 46-95)

**Problem**: Function does NOT check if journal already exists before creating new one.

```go
// BEFORE (BROKEN):
func (s *SalesJournalServiceV2) CreateSalesJournal(sale *models.Sale, tx *gorm.DB) error {
    // ... validation ...
    
    // Directly creates journal WITHOUT checking if it already exists!
    ssotEntry := &models.SimpleSSOTJournal{
        EntryNumber:       fmt.Sprintf("SALES-%d", sale.ID),
        TransactionType:   "SALES",
        TransactionID:     sale.ID,
        ...
    }
    
    if err := dbToUse.Create(ssotEntry).Error; err != nil {
        return fmt.Errorf("failed to create SSOT journal: %v", err)
    }
    // ❌ NO CHECK! Creates duplicate if called twice!
}
```

### Issue 2: Same Problem in Payment Journal

**Function**: `CreateSalesPaymentJournal()` (Line 318-430)

Same issue - no duplicate prevention for payment journals.

### Why Duplicates Occur

Possible scenarios:
1. **Frontend calls API twice** (double-click on Invoice button)
2. **Retry logic** somewhere in the code
3. **Transaction rollback + retry** without proper idempotency check
4. **Multiple services** calling CreateSalesJournal()

## Solution Implemented

### Fix 1: Add Duplicate Check in CreateSalesJournal

**File**: `backend/services/sales_journal_service_v2.go` (Line 63-72)

```go
// ✅ CRITICAL FIX: Check if journal already exists for this sale
// Prevent duplicate journal entries which cause double posting to COA
var existingCount int64
if err := dbToUse.Model(&models.SimpleSSOTJournal{}).
    Where("transaction_type = ? AND transaction_id = ?", "SALES", sale.ID).
    Count(&existingCount).Error; err == nil && existingCount > 0 {
    log.Printf("⚠️ Journal already exists for Sale #%d (found %d entries), skipping creation to prevent duplicate", 
        sale.ID, existingCount)
    return nil // Don't create duplicate journal
}
```

### Fix 2: Add Duplicate Check in CreateSalesPaymentJournal

**File**: `backend/services/sales_journal_service_v2.go` (Line 331-340)

```go
// ✅ CRITICAL FIX: Check if payment journal already exists
// Prevent duplicate payment journal entries which cause double posting to COA
var existingCount int64
if err := dbToUse.Model(&models.SimpleSSOTJournal{}).
    Where("transaction_type = ? AND transaction_id = ?", "SALES_PAYMENT", payment.ID).
    Count(&existingCount).Error; err == nil && existingCount > 0 {
    log.Printf("⚠️ Payment journal already exists for Payment #%d (found %d entries), skipping creation to prevent duplicate", 
        payment.ID, existingCount)
    return nil // Don't create duplicate journal
}
```

## Cleanup Script for Existing Duplicates

**File**: `backend/scripts/cleanup_duplicate_sales_journals.go`

Script ini akan:
1. ✅ Find duplicate SALES journals
2. ✅ Keep oldest journal entry
3. ✅ Delete duplicate journals and their items
4. ✅ Find duplicate SALES_PAYMENT journals
5. ✅ Cleanup payment journal duplicates
6. ✅ Recalculate ALL COA balances from scratch
7. ✅ Update parent account balances

### How to Run Cleanup

```bash
cd backend
go run scripts/cleanup_duplicate_sales_journals.go
```

**Output Example**:
```
🔧 Starting cleanup of duplicate sales journals...
✅ Connected to database

📋 Checking for duplicate SALES journals...
⚠️ Found 3 sales with duplicate journals

🔍 Processing Sale ID 1 (2 duplicate entries)...
  ✅ Keeping journal ID 123 (created: 2025-10-16T10:00:00Z)
  🗑️ Deleting duplicate journal ID 124 (created: 2025-10-16T10:00:05Z)
    ✅ Deleted journal ID 124 and its items

📊 Recalculating COA balances from journal entries...
  🔄 Resetting all account balances to 0...
  📝 Processing 156 journal items...
  💾 Updating balances for 45 accounts...
  🔗 Updating parent account balances...
    ✅ Updated parent account 1000 (ASSETS) balance to 8325000.00
    ✅ Updated parent account 1100 (CURRENT ASSETS) balance to 8325000.00

✅ Cleanup completed successfully!
```

## Expected Results After Fix

### Before Fix (BROKEN):
```
Sales: Rp 5,550,000
Invoice button clicked → Journal created
(Accidentally called twice)
→ Journal created AGAIN!

COA Balances:
- Bank (1102):     Rp 11,100,000 ❌ DOUBLE!
- Piutang (1201):  Rp 11,100,000 ❌ DOUBLE!
- Revenue (4101):  Rp 10,000,000 ❌ DOUBLE!
- PPN (2103):      Rp 1,100,000  ❌ DOUBLE!
```

### After Fix (CORRECT):
```
Sales: Rp 5,550,000
Invoice button clicked → Check if journal exists
→ Journal doesn't exist → Create journal
If clicked again → Journal exists → SKIP creation!

COA Balances:
- Piutang (1201):  Rp 5,550,000  ✅ CORRECT
- Revenue (4101):  Rp 5,000,000  ✅ CORRECT
- PPN (2103):      Rp 550,000    ✅ CORRECT
- Bank (1102):     Rp 0          ✅ CORRECT (no payment yet)

After partial payment Rp 2,775,000:
- Bank (1102):     Rp 2,775,000  ✅ CORRECT
- Piutang (1201):  Rp 2,775,000  ✅ CORRECT
```

## Testing Instructions

### Step 1: Run Cleanup Script
```bash
cd backend
go run scripts/cleanup_duplicate_sales_journals.go
```

### Step 2: Restart Backend
```powershell
.\restart_backend.ps1
```

### Step 3: Verify COA Balances
```bash
GET http://localhost:8080/api/v1/accounts
```

Expected after cleanup:
- Bank (1102): Rp 5,550,000 (not 11,100,000)
- Piutang (1201): Rp 5,550,000 (if no payment)
- OR if ada partial payment Rp 2,775,000:
  - Bank (1102): Rp 2,775,000
  - Piutang (1201): Rp 2,775,000

### Step 4: Test New Sale
1. Create new sale dengan CREDIT payment method
2. Click "Invoice" button
3. Check database:
   ```sql
   SELECT COUNT(*) as count, transaction_id
   FROM simple_ssot_journals
   WHERE transaction_type = 'SALES'
   GROUP BY transaction_id
   HAVING COUNT(*) > 1;
   ```
   Expected: NO RESULTS (no duplicates)

4. Try clicking "Invoice" again (should be idempotent)
5. Check database again - still NO duplicates

### Step 5: Test Partial Payment
1. Record partial payment Rp 2,775,000
2. Check database:
   ```sql
   SELECT COUNT(*) as count, transaction_id
   FROM simple_ssot_journals
   WHERE transaction_type = 'SALES_PAYMENT'
   GROUP BY transaction_id
   HAVING COUNT(*) > 1;
   ```
   Expected: NO RESULTS

3. Verify Bank balance: Rp 2,775,000 ✅

## Files Modified

### Prevention Code
1. ✅ `backend/services/sales_journal_service_v2.go`
   - Added duplicate check in `CreateSalesJournal()` (Line 63-72)
   - Added duplicate check in `CreateSalesPaymentJournal()` (Line 331-340)

### Cleanup Tool
2. ✅ `backend/scripts/cleanup_duplicate_sales_journals.go`
   - Complete cleanup script for existing duplicates
   - Recalculates all COA balances from scratch

### Documentation
3. ✅ `backend/docs/DUPLICATE_JOURNAL_FIX.md` - This document
4. ✅ `backend/docs/SALES_PARTIAL_PAYMENT_FIX.md` - Related fix

## Impact

✅ **Prevents**: Duplicate journal entries from being created
✅ **Fixes**: Double posting to COA accounts
✅ **Cleanup**: Removes existing duplicate journals
✅ **Recalculates**: All COA balances correctly
✅ **Idempotent**: Safe to call CreateSalesJournal() multiple times
✅ **No Breaking Changes**: Existing functionality preserved

## Prevention for Future

1. ✅ **Idempotency checks** added to all journal creation functions
2. ✅ **Cleanup script** available for maintenance
3. ✅ **Logging** added to detect duplicate attempts
4. 🔄 **TODO**: Add integration tests for duplicate prevention
5. 🔄 **TODO**: Add database unique constraint on (transaction_type, transaction_id)

## Version

- **Fixed Date**: 2025-10-16
- **Severity**: CRITICAL (affects all financial data)
- **Fixed By**: AI Assistant
- **Tested**: Pending user testing
- **Status**: Ready for cleanup + deployment

## URGENT ACTION REQUIRED

1. ✅ **Run cleanup script FIRST**: `go run scripts/cleanup_duplicate_sales_journals.go`
2. ✅ **Restart backend**: `.\restart_backend.ps1`
3. ✅ **Verify balances** are now correct
4. ✅ **Test new transactions** to ensure no more duplicates

⚠️ **WARNING**: Do NOT skip the cleanup step! Existing duplicates will cause incorrect balances!

