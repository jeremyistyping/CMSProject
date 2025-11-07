# Auto-Update Fiscal Year Start After Period Closing

## Feature Overview
Setelah period closing berhasil dilakukan, sistem akan **otomatis update** `fiscal_year_start` di settings menjadi tanggal 1 hari setelah end date period closing.

## Motivation
Fitur ini memastikan bahwa **Fiscal Year Start** selalu sinkron dengan periode akuntansi yang aktif, sehingga:
- User tidak perlu manual update fiscal year start setelah tutup buku
- Fiscal period range selalu akurat dan up-to-date
- Menghindari konfusi antara fiscal year yang lama dan periode aktif yang baru

## Implementation

### File Modified
`services/period_closing_service.go`

### Code Changes

#### 1. Helper Function `formatFiscalYearStart` (lines 577-586)
Converts a date to fiscal year start format (e.g., "January 1"):

```go
func formatFiscalYearStart(date time.Time) string {
	months := []string{
		"January", "February", "March", "April", "May", "June",
		"July", "August", "September", "October", "November", "December",
	}
	month := months[date.Month()-1]
	day := date.Day()
	return fmt.Sprintf("%s %d", month, day)
}
```

#### 2. Auto-Update Logic in `ExecutePeriodClosing` (lines 431-453)
After creating the accounting period record, system auto-updates fiscal year start:

```go
// Auto-update fiscal year start in settings to next period start date
// This ensures fiscal year is always synchronized with the active accounting period
nextPeriodStart := endDate.AddDate(0, 0, 1)
newFiscalYearStart := formatFiscalYearStart(nextPeriodStart)

log.Printf("[PERIOD CLOSING] Auto-updating fiscal_year_start to: %s (from %s)", 
    newFiscalYearStart, nextPeriodStart.Format("2006-01-02"))

var settings models.Settings
if err := tx.First(&settings).Error; err != nil {
	log.Printf("[PERIOD CLOSING] Warning: failed to get settings for fiscal year update: %v", err)
	// Don't fail the entire closing if settings update fails
} else {
	oldFiscalYearStart := settings.FiscalYearStart
	settings.FiscalYearStart = newFiscalYearStart
	settings.UpdatedBy = userID
	
	if err := tx.Save(&settings).Error; err != nil {
		log.Printf("[PERIOD CLOSING] Warning: failed to update fiscal_year_start: %v", err)
		// Don't fail the entire closing if settings update fails
	} else {
		log.Printf("[PERIOD CLOSING] ✅ Fiscal year start updated: %s → %s", 
		    oldFiscalYearStart, newFiscalYearStart)
	}
}
```

### Key Design Decisions

#### 1. **Non-Breaking Error Handling**
Jika update fiscal year start gagal, **period closing tetap berhasil**.
- Reason: Fiscal year start adalah metadata, bukan critical data
- Period closing data (journal entries, balances) lebih penting
- User masih bisa manual update fiscal year start jika diperlukan

#### 2. **Transaction Safety**
Update fiscal year start dilakukan **di dalam transaksi** yang sama dengan period closing.
- Jika transaksi rollback, fiscal year start tidak akan ter-update
- Ensures data consistency

#### 3. **Logging**
Extensive logging untuk debugging dan audit trail:
- Log before update (target value)
- Log warning jika gagal
- Log success dengan old → new value

## Behavior Examples

### Example 1: Monthly Closing
```
Period: 2024-12-01 to 2024-12-31
Action: Execute period closing

Result:
✅ Period closed successfully
✅ fiscal_year_start updated: "December 1" → "January 1"
✅ Next period start date: 2025-01-01
```

### Example 2: Quarterly Closing
```
Period: 2024-10-01 to 2024-12-31 (Q4)
Action: Execute period closing

Result:
✅ Period closed successfully
✅ fiscal_year_start updated: "October 1" → "January 1"
✅ Next period start date: 2025-01-01
```

### Example 3: Mid-Year Fiscal Year (April to March)
```
Period: 2024-03-01 to 2024-03-31
Action: Execute period closing

Result:
✅ Period closed successfully
✅ fiscal_year_start updated: "March 1" → "April 1"
✅ Next period start date: 2024-04-01
```

## UI Impact

### Before Closing
```
System Configuration
├── Fiscal Year Start: 01/01/2024
└── Current fiscal period: 01/01/2024 — 31/12/2024
```

### After Closing (Period 01/01/2024 - 31/12/2024)
```
System Configuration
├── Fiscal Year Start: 01/01/2025  ⬅️ AUTO-UPDATED (refreshed automatically)
└── Current fiscal period: 01/01/2025 — 31/12/2025

Period Closing
├── 🔄 Auto-Update Fiscal Year alert shown
├── 📅 Last Closing: 12/31/2024
└── Dari Tanggal: 01/01/2025 🔒 Locked - Auto-filled
```

### Frontend Auto-Refresh
Setelah period closing berhasil, frontend akan:
1. 🔄 Auto-refresh settings (`fetchSettings()`)
2. 🔄 Auto-refresh last closing info (`fetchLastClosingInfo()`)
3. ✅ Console log: "Settings and closing info refreshed"

User akan langsung melihat fiscal year start yang ter-update tanpa perlu manual refresh page.

## Database Changes
No schema changes required. Uses existing `settings.fiscal_year_start` field.

## API Response Enhancement
The `period-closing/execute` endpoint response now includes fiscal year update info in logs:

```json
{
  "success": true,
  "message": "Period closed successfully. All revenue and expense accounts have been reset and transferred to retained earnings."
}
```

Internal log includes:
```json
{
  "start_date": "2024-01-01T00:00:00Z",
  "end_date": "2024-12-31T00:00:00Z",
  "net_income": 50000000,
  "new_fiscal_year_start": "January 1"  ⬅️ NEW
}
```

## Testing Scenarios

### Test 1: Normal Monthly Closing
1. Set fiscal_year_start: "January 1"
2. Execute closing: 2024-01-01 to 2024-01-31
3. **Expected**: fiscal_year_start → "February 1"

### Test 2: Year-End Closing
1. Set fiscal_year_start: "December 1"
2. Execute closing: 2024-12-01 to 2024-12-31
3. **Expected**: fiscal_year_start → "January 1"

### Test 3: Custom Fiscal Year (July to June)
1. Set fiscal_year_start: "July 1"
2. Execute closing: 2024-07-01 to 2025-06-30
3. **Expected**: fiscal_year_start → "July 1" (next year)

### Test 4: Reopen Period
1. Execute closing (fiscal year updated)
2. Reopen period
3. **Expected**: fiscal_year_start **TIDAK** otomatis dikembalikan
   - Reason: Reopen adalah exceptional case
   - Admin should manually adjust if needed

## Backward Compatibility
✅ Fully backward compatible
- Existing settings tidak terpengaruh
- Tidak ada migration required
- Feature baru hanya aktif saat period closing

## Limitations & Future Enhancements

### Current Limitations
1. **One-way update**: Fiscal year start hanya update forward, tidak rollback saat reopen
2. **No manual override**: User tidak bisa disable auto-update (always on)

### Future Enhancements
1. **Optional flag**: Add setting `auto_update_fiscal_year_start` (true/false)
2. **Reopen support**: Restore old fiscal year start when reopening last period
3. **Audit trail**: Track fiscal year start changes in settings history

## Related Documentation
- [PERIOD_CLOSING_DATE_FIX.md](./PERIOD_CLOSING_DATE_FIX.md) - Period start date fix
- [FLEXIBLE_PERIOD_CLOSING.md](./FLEXIBLE_PERIOD_CLOSING.md) - Period closing overview

---

**Created**: 2025-01-07
**Author**: AI Assistant
**Status**: Implemented & Tested
**Version**: 1.0
