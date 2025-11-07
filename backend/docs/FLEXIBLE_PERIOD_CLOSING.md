# Flexible Period Closing Documentation

## Overview
Flexible Period Closing adalah fitur tutup buku periode akuntansi yang lebih fleksibel dibandingkan fiscal year-end closing tradisional. Fitur ini memungkinkan penutupan periode secara:
- **Bulanan** (monthly)
- **Triwulan** (quarterly) 
- **Semester** (semi-annual)
- **Tahunan** (annual)
- Atau periode custom sesuai kebutuhan bisnis

## Key Differences: Flexible vs Traditional Fiscal Year Closing

| Aspect | Traditional Fiscal Year | Flexible Period Closing |
|--------|------------------------|-------------------------|
| Frequency | Once per year | Flexible (monthly/quarterly/etc) |
| Start Date | Fixed (e.g., Jan 1) | Auto-detected from last closing |
| End Date | Fixed (e.g., Dec 31) | User-defined |
| Use Case | Annual reporting | Regular period management |
| Flexibility | Low | High |

## Core Concepts

### 1. **Accounting Period**
Setiap periode tutup buku dicatat sebagai `AccountingPeriod` dengan:
- `start_date`: Tanggal mulai periode
- `end_date`: Tanggal akhir periode
- `is_closed`: Status tutup buku
- `is_locked`: Hard lock (tidak bisa dibuka kembali)
- `net_income`: Laba/rugi periode
- `closing_journal_id`: Journal entry penutupan

### 2. **Auto-Detection Logic**
System automatically detects:
- **First Closing**: If no previous closing exists, start from earliest transaction
- **Subsequent Closing**: Start date = last closing end date + 1 day

### 3. **Temporary vs Permanent Accounts**
- **Temporary Accounts (RESET to 0):**
  - Revenue (4xxx)
  - Expense (5xxx)
  
- **Permanent Accounts (CARRY FORWARD):**
  - Assets (1xxx)
  - Liabilities (2xxx)
  - Equity (3xxx) - except Revenue & Expense

## Database Schema

### AccountingPeriod Model
```sql
CREATE TABLE accounting_periods (
    id BIGSERIAL PRIMARY KEY,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    description TEXT,
    is_closed BOOLEAN DEFAULT FALSE,
    is_locked BOOLEAN DEFAULT FALSE,
    closed_by BIGINT,
    closed_at TIMESTAMP,
    
    -- Closing summary
    total_revenue DECIMAL(20,2) DEFAULT 0,
    total_expense DECIMAL(20,2) DEFAULT 0,
    net_income DECIMAL(20,2) DEFAULT 0,
    closing_journal_id BIGINT,
    
    notes TEXT,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP
);
```

## API Endpoints

### 1. Get Last Closing Info
```http
GET /api/v1/period-closing/last-info
```

**Response:**
```json
{
  "success": true,
  "data": {
    "has_previous_closing": true,
    "last_closing_date": "2024-12-31T00:00:00Z",
    "next_start_date": "2025-01-01T00:00:00Z",
    "period_start_date": null
  }
}
```

### 2. Preview Period Closing
```http
GET /api/v1/period-closing/preview?start_date=2025-01-01&end_date=2025-01-31
```

**Response:**
```json
{
  "success": true,
  "data": {
    "start_date": "2025-01-01T00:00:00Z",
    "end_date": "2025-01-31T00:00:00Z",
    "total_revenue": 5000000,
    "total_expense": 3000000,
    "net_income": 2000000,
    "retained_earnings_id": 123,
    "revenue_accounts": [
      {
        "id": 401,
        "code": "4101",
        "name": "Pendapatan Penjualan",
        "balance": 5000000,
        "type": "REVENUE"
      }
    ],
    "expense_accounts": [
      {
        "id": 501,
        "code": "5101",
        "name": "Harga Pokok Penjualan",
        "balance": 3000000,
        "type": "EXPENSE"
      }
    ],
    "closing_entries": [
      {
        "description": "Close Revenue Accounts to Retained Earnings",
        "debit_account": "Revenue Accounts (Total)",
        "credit_account": "3201 - Laba Ditahan",
        "amount": 5000000
      },
      {
        "description": "Close Expense Accounts to Retained Earnings",
        "debit_account": "3201 - Laba Ditahan",
        "credit_account": "Expense Accounts (Total)",
        "amount": 3000000
      }
    ],
    "can_close": true,
    "validation_messages": [
      "✅ Found 45 posted transactions in this period",
      "📊 Will close 3 revenue accounts and 5 expense accounts",
      "💰 Net Income will be transferred to Retained Earnings"
    ],
    "transaction_count": 45,
    "period_days": 31
  }
}
```

### 3. Execute Period Closing
```http
POST /api/v1/period-closing/execute
Content-Type: application/json

{
  "start_date": "2025-01-01",
  "end_date": "2025-01-31",
  "description": "January 2025 Monthly Closing",
  "notes": "Monthly closing for January 2025"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Period closed successfully. All revenue and expense accounts have been reset and transferred to retained earnings."
}
```

### 4. Get Closing History
```http
GET /api/v1/period-closing/history?limit=20
```

**Response:**
```json
{
  "success": true,
  "data": [
    {
      "id": 5,
      "start_date": "2025-01-01T00:00:00Z",
      "end_date": "2025-01-31T00:00:00Z",
      "description": "January 2025 Monthly Closing",
      "is_closed": true,
      "is_locked": true,
      "total_revenue": 5000000,
      "total_expense": 3000000,
      "net_income": 2000000,
      "closed_at": "2025-02-01T10:30:00Z",
      "closed_by_user": {
        "id": 1,
        "name": "Admin User"
      }
    }
  ]
}
```

### 5. Check if Date is in Closed Period
```http
GET /api/v1/period-closing/check-date?date=2025-01-15
```

**Response:**
```json
{
  "success": true,
  "is_closed": true,
  "date": "2025-01-15"
}
```

## Business Logic

### Validation Rules

#### Pre-Closing Validation:
1. ✅ **Date Range Valid**: End date must be after start date
2. ✅ **No Duplicate Periods**: Cannot close the same period twice
3. ✅ **No Overlapping Periods**: Cannot close overlapping date ranges
4. ✅ **Period Continuity** (Warning): Start date should match expected start after last closing
5. ✅ **Retained Earnings Exists**: Account 3201 must exist and be active
6. ✅ **No Unbalanced Entries**: All journal entries must be balanced
7. ✅ **No Draft Entries**: All entries must be posted

### Closing Process Flow

```
1. VALIDATE
   ├── Check date range validity
   ├── Check for duplicates/overlaps
   ├── Validate retained earnings account
   ├── Check for unbalanced entries
   └── Check for draft entries

2. CALCULATE
   ├── Get all revenue accounts with balance ≠ 0
   ├── Get all expense accounts with balance ≠ 0
   ├── Calculate totals
   └── Calculate net income

3. EXECUTE (Transaction)
   ├── Create closing journal entry
   │   ├── Debit all Revenue accounts (reset to 0)
   │   ├── Credit Retained Earnings (total revenue)
   │   ├── Debit Retained Earnings (total expense)
   │   └── Credit all Expense accounts (reset to 0)
   ├── Update Retained Earnings balance
   ├── Create AccountingPeriod record
   └── Log activity

4. LOCK PERIOD
   ├── Set is_closed = true
   └── Set is_locked = true
```

## Frontend Implementation

### UI Components

#### 1. Period Selection Form
```typescript
<FormControl>
  <FormLabel>Dari Tanggal (Start Date)</FormLabel>
  <Input
    type="date"
    value={periodStartDate}
    onChange={(e) => setPeriodStartDate(e.target.value)}
    isReadOnly={lastClosingInfo?.has_previous_closing}
  />
  <FormHelperText>
    {lastClosingInfo?.has_previous_closing 
      ? 'Auto-filled from last closing date' 
      : 'Start date of the period to close'}
  </FormHelperText>
</FormControl>

<FormControl>
  <FormLabel>Sampai Tanggal (End Date)</FormLabel>
  <Input
    type="date"
    value={periodEndDate}
    onChange={(e) => setPeriodEndDate(e.target.value)}
  />
</FormControl>
```

#### 2. Preview Modal
- Shows period info (dates, duration, transaction count)
- Displays financial summary (revenue, expense, net income)
- Lists validation messages
- Previews closing journal entries
- Shows what will happen when executed

#### 3. Execute Button
- Disabled if validation fails
- Shows loading state during execution
- Displays success/error toast

### State Management
```typescript
const [periodStartDate, setPeriodStartDate] = useState('');
const [periodEndDate, setPeriodEndDate] = useState('');
const [periodClosingPreview, setPeriodClosingPreview] = useState(null);
const [loadingPeriodPreview, setLoadingPeriodPreview] = useState(false);
const [showPeriodClosingModal, setShowPeriodClosingModal] = useState(false);
const [executingPeriodClosing, setExecutingPeriodClosing] = useState(false);
const [lastClosingInfo, setLastClosingInfo] = useState(null);
```

## Example Use Cases

### Monthly Closing
```
Period: Jan 1, 2025 - Jan 31, 2025
Use Case: Close January monthly books
```

### Quarterly Closing
```
Period: Jan 1, 2025 - Mar 31, 2025
Use Case: Close Q1 2025 quarterly books
```

### Semi-Annual Closing
```
Period: Jan 1, 2025 - Jun 30, 2025
Use Case: Close H1 2025 semester books
```

### Annual Closing
```
Period: Jan 1, 2025 - Dec 31, 2025
Use Case: Close full year 2025 books
```

## Security & Permissions

### Required Roles
- `admin`
- `director`
- `finance`

### Audit Logging
All period closings are logged with:
- User ID who executed
- Timestamp
- Period details
- Financial summary
- Journal entry reference

## Best Practices

### Before Closing
1. ✅ Reconcile all accounts
2. ✅ Post all adjusting entries
3. ✅ Review trial balance
4. ✅ Verify all transactions are balanced
5. ✅ Backup database

### After Closing
1. ✅ Verify all temporary accounts = 0
2. ✅ Verify retained earnings balance
3. ✅ Generate financial statements
4. ✅ Archive closing journal entry
5. ✅ Document any adjustments made

### Frequency Recommendations
- **Monthly**: Best for active businesses
- **Quarterly**: Good balance for most companies
- **Semi-Annual**: Suitable for smaller businesses
- **Annual**: Minimum required for compliance

## Troubleshooting

### Error: "Retained Earnings account not found"
**Solution:** Create account with code 3201 (LABA DITAHAN) as EQUITY type

### Error: "Found X unbalanced journal entries"
**Solution:** Review and fix unbalanced entries before closing

### Error: "Found X draft journal entries"
**Solution:** Post or delete draft entries before closing

### Error: "Period already closed"
**Solution:** Cannot close the same period twice

### Warning: "Period start does not match expected start"
**Solution:** This is a warning - you can still close, but verify dates are correct

## Migration Path from Fiscal Year Closing

For existing systems using fiscal year closing:
1. Both systems can coexist (fiscal year closing marked as LEGACY)
2. Gradually migrate to flexible period closing
3. Frontend now uses flexible period closing by default
4. Old fiscal year closing routes still available at `/api/v1/fiscal-closing/*`
5. New flexible closing routes at `/api/v1/period-closing/*`

## Version History

- **v1.0** (2025-01-04) - Initial flexible period closing implementation
- **v0.9** (2024) - Legacy fiscal year closing (deprecated)

## References

- Model: `backend/models/accounting_period.go`
- Service: `backend/services/period_closing_service.go`
- Controller: `backend/controllers/period_closing_controller.go`
- Routes: `backend/routes/routes.go` (lines 924-933)
- Frontend: `frontend/app/settings/page.tsx` (lines 778-890)
- Migration: `backend/migrations/041_create_accounting_periods_table.sql`
