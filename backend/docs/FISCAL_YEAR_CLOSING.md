# Fiscal Year Closing Process

## Overview
Fiscal year closing adalah proses akuntansi untuk menutup periode tahun fiskal dan memindahkan saldo akun temporary ke akun permanent sesuai prinsip akuntansi yang berlaku umum.

## Account Classification

### 1. Temporary Accounts (Nominal Accounts)
Akun yang **direset menjadi 0** setiap akhir tahun fiskal:
- ✅ **Revenue Accounts** (4xxx) - Pendapatan
- ✅ **Expense Accounts** (5xxx) - Beban

**Karakteristik:**
- Saldo direset ke nol pada fiscal year-end
- Melacak aktivitas untuk satu periode saja
- Digunakan untuk menghitung Net Income

**Contoh di sistem:**
- `4101` - Pendapatan Penjualan
- `4102` - Pendapatan Jasa/Ongkir
- `5101` - Harga Pokok Penjualan
- `5201` - Beban Gaji
- `5202` - Beban Listrik

### 2. Permanent Accounts (Real Accounts)
Akun yang **TIDAK direset** dan saldonya terus berlanjut:
- ✅ **Asset Accounts** (1xxx) - Aset
- ✅ **Liability Accounts** (2xxx) - Kewajiban
- ✅ **Equity Accounts** (3xxx) - Ekuitas

**Karakteristik:**
- Saldo terus berlanjut dari tahun ke tahun
- Melacak posisi keuangan kumulatif
- Muncul di Balance Sheet

**Contoh di sistem:**
- `1101` - Kas
- `1102` - Bank
- `2101` - Utang Usaha
- `3101` - Modal Pemilik
- `3201` - **Laba Ditahan** (Retained Earnings) ← **Kunci untuk closing!**

## Closing Entries Process

### Step 1: Close Revenue Accounts
```
Debit: Revenue Accounts (untuk menjadi 0)
Credit: Retained Earnings (menambah equity)
```

**Tujuan:** Transfer semua pendapatan ke Laba Ditahan

### Step 2: Close Expense Accounts
```
Debit: Retained Earnings (mengurangi equity)
Credit: Expense Accounts (untuk menjadi 0)
```

**Tujuan:** Transfer semua beban ke Laba Ditahan

### Step 3: Calculate Net Income Effect
```
Net Income = Total Revenue - Total Expense
Retained Earnings Balance = Previous Balance + Net Income
```

**Hasil Akhir:**
- ✅ Revenue accounts = 0
- ✅ Expense accounts = 0
- ✅ Retained Earnings += Net Income
- ✅ Balance Sheet tetap balance

## Configuration

### Retained Earnings Account
File: `config/accounting_config.json`

```json
{
  "default_accounts": {
    "retained_earnings": 3201  // Akun LABA DITAHAN (Code: 3201)
  }
}
```

**Penting:** 
- Retained Earnings harus menunjuk ke akun dengan **Type = Equity**
- Default: Account Code `3201` - LABA DITAHAN
- Bukan `3101` (Modal Pemilik) ❌

## Implementation Details

### Service: `FiscalYearClosingService`
Location: `services/fiscal_year_closing_service.go`

**Key Functions:**

#### 1. PreviewFiscalYearClosing
- Menampilkan preview closing entries
- Validasi retained earnings account
- Cek unbalanced journal entries
- Hitung net income

#### 2. ExecuteFiscalYearClosing
- Jalankan closing entries dalam transaction
- Reset temporary accounts (Revenue & Expense) ke 0
- Update retained earnings dengan net income
- Lock accounting period (hard lock)
- Generate closing journal entry

#### 3. GetFiscalYearClosingHistory
- Tampilkan riwayat fiscal year closing
- Filter by `reference_type = "CLOSING"`

## Validation Rules

### Pre-Closing Checks
1. ✅ Retained earnings account must be configured
2. ✅ Retained earnings must exist in database
3. ✅ Retained earnings must be type EQUITY
4. ✅ No unbalanced journal entries in fiscal year
5. ✅ Only non-header accounts will be closed

### Post-Closing Verification
1. ✅ All revenue accounts balance = 0
2. ✅ All expense accounts balance = 0
3. ✅ Retained earnings increased by net income
4. ✅ Balance sheet equation still balanced:
   ```
   Assets = Liabilities + Equity
   ```

## Journal Entry Format

### Closing Entry Code Pattern
```
CLO-YYYY-12-31
```
Example: `CLO-2024-12-31`

### Reference Type
```
reference_type = "CLOSING"
```

### Journal Lines Example
```
Line 1: Debit Revenue Account 4101 (100,000)
Line 2: Debit Revenue Account 4102 (50,000)
Line 3: Credit Retained Earnings 3201 (150,000)
Line 4: Debit Retained Earnings 3201 (80,000)
Line 5: Credit Expense Account 5101 (50,000)
Line 6: Credit Expense Account 5201 (30,000)

Net Effect: Retained Earnings += 70,000 (Net Income)
```

## API Endpoints

### 1. Preview Closing
```
GET /api/v1/fiscal-closing/preview?fiscal_year_end=2024-12-31
```

**Response:**
```json
{
  "success": true,
  "data": {
    "fiscal_year_end": "2024-12-31",
    "total_revenue": 1000000,
    "total_expense": 700000,
    "net_income": 300000,
    "retained_earnings_id": 3201,
    "revenue_accounts": [...],
    "expense_accounts": [...],
    "closing_entries": [...],
    "can_close": true,
    "validation_messages": [...]
  }
}
```

### 2. Execute Closing
```
POST /api/v1/fiscal-closing/execute
Content-Type: application/json

{
  "fiscal_year_end": "2024-12-31",
  "notes": "Fiscal year 2024 closing"
}
```

### 3. Get History
```
GET /api/v1/fiscal-closing/history
```

## Period Locking

After successful closing:
- Period is marked as `is_closed = true`
- Period is marked as `is_locked = true` (hard lock)
- No further transactions can be posted to closed periods
- Only admin can reopen locked periods

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

## Troubleshooting

### Error: "Retained Earnings account not configured"
**Solution:** Set retained earnings in `accounting_config.json` to account ID with code 3201

### Error: "Retained Earnings must be an Equity account"
**Solution:** Ensure retained earnings points to account with `type = "EQUITY"`

### Error: "Found X unbalanced journal entries"
**Solution:** Review and fix unbalanced entries before closing

### Warning: "No revenue or expense to close"
**Solution:** Verify transactions were properly posted during the fiscal year

## References

- Accounting Period Model: `models/accounting_period.go`
- Journal Entry Model: `models/journal_entry.go`
- Account Model: `models/account.go`
- Config: `config/accounting_config.go`
- Service: `services/fiscal_year_closing_service.go`
- Controller: `controllers/fiscal_year_closing_controller.go`

## Version History

- **v1.0** - Initial fiscal year closing implementation
- **v1.1** - Added temporary vs permanent account documentation
- **v1.2** - Fixed retained earnings config (3201 instead of 3101)
