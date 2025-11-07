# Perbaikan Masalah Double Posting Laba Ditahan

## Masalah

Pada periode closing kedua, terjadi **double posting** ke akun Laba Ditahan (3201) yang menyebabkan Balance Sheet **tidak balance** dengan selisih Rp 35.000.000.

### Gejala
- Balance Sheet menunjukkan: `NOT BALANCED (DIFF: -RP 35.000.000)`
- Total Assets: Rp 174.430.000
- Total Liabilities: Rp 57.430.000
- Total Equity: Rp 152.000.000
- **Diff: -Rp 35.000.000** (seharusnya Assets = Liabilities + Equity)

### Akar Masalah

**Periode closing kedua menutup akun Revenue/Expense yang sudah balance 0** dari closing sebelumnya, tetapi tetap posting journal entry ke Laba Ditahan, mengakibatkan **double posting** nilai yang sama.

#### Skenario Masalah:

**Periode 1 Closing (1 Jan - 31 Jan):**
```
Revenue: Rp 100.000.000
Expense: Rp 65.000.000
Net Income: Rp 35.000.000

Journal Entry:
DR Revenue         100.000.000
   CR Laba Ditahan             100.000.000
DR Laba Ditahan     65.000.000
   CR Expense                   65.000.000

Hasil: Laba Ditahan = +35.000.000 ✓
```

**Periode 2 Closing (1 Feb - 28 Feb):**
```
Revenue: Rp 0 (sudah di-close)
Expense: Rp 0 (sudah di-close)

TAPI kode lama masih membuat journal entry:
DR Revenue         0
   CR Laba Ditahan             0
DR Laba Ditahan    0
   CR Expense                  0

NAMUN: Karena bug di logika, total_revenue dan total_expense
masih menggunakan nilai dari periode sebelumnya!

Hasil: Laba Ditahan = +35.000.000 (DOUBLE!) ✗
```

### Bug di Kode

#### File: `services/period_closing_service.go`

**Baris 156-166 (SEBELUM):**
```go
var revenueAccounts []models.Account
if err := pcs.db.Where("type = ? AND balance != 0 AND is_active = true AND is_header = false", models.AccountTypeRevenue).
    Find(&revenueAccounts).Error; err != nil {
    return nil, fmt.Errorf("failed to get revenue accounts: %v", err)
}

var expenseAccounts []models.Account
if err := pcs.db.Where("type = ? AND balance != 0 AND is_active = true AND is_header = false", models.AccountTypeExpense).
    Find(&expenseAccounts).Error; err != nil {
    return nil, fmt.Errorf("failed to get expense accounts: %v", err)
}
```

**Masalah:** Query `balance != 0` akan match akun dengan balance `0.0000000001` (floating point precision issue) atau nilai yang sangat kecil hasil pembulatan.

## Solusi

### 1. Immediate Fix (SQL Script)

Jalankan script untuk memperbaiki data yang sudah rusak:

```bash
# 1. Analisis masalah
psql -U postgres -d sistem_akuntansi -f analyze_double_posting_issue.sql

# 2. Backup database
pg_dump -U postgres sistem_akuntansi > backup_before_fix.sql

# 3. Jalankan fix (setelah review hasil analisis)
psql -U postgres -d sistem_akuntansi -f fix_double_posting_issue.sql
```

**Script akan:**
1. Mengidentifikasi closing entries yang problematik
2. Me-reverse balance changes dari journal entries tersebut
3. Menandai journal entries sebagai `VOIDED`
4. Update `accounting_periods` untuk remove reference ke voided journals
5. Verify balance sheet sudah balanced

### 2. Code Fix (Preventive)

#### File: `services/period_closing_service.go`

**Perubahan:**

```go
// SEBELUM:
if err := pcs.db.Where("type = ? AND balance != 0 AND is_active = true AND is_header = false", models.AccountTypeRevenue).

// SESUDAH:
if err := pcs.db.Where("type = ? AND ABS(balance) > 0.01 AND is_active = true AND is_header = false", models.AccountTypeRevenue).
```

**Alasan:**
- `ABS(balance) > 0.01`: Hanya ambil akun dengan balance **absolut lebih dari 1 sen**
- Mencegah closing akun yang sudah 0 atau nilai floating point yang sangat kecil
- Threshold 0.01 sudah cukup aman untuk currency dengan 2 desimal

#### Perubahan Juga di Execute:

```go
// SEBELUM:
for _, revAccount := range preview.RevenueAccounts {
    if revAccount.Balance > 0 {
        // create journal line
    }
}

// SESUDAH:
for _, revAccount := range preview.RevenueAccounts {
    if revAccount.Balance > 0.01 {  // Double check dengan threshold
        // create journal line
    }
}
```

### 3. Rebuild & Test

```bash
# Rebuild aplikasi
cd D:\Project\clone_app_akuntansi\accounting_proj\backend
go build -o app-sistem-akuntansi.exe

# Restart service
# Stop aplikasi yang sedang running
# Jalankan ulang aplikasi
```

### 4. Testing Scenario

**Test 1: Closing Pertama**
```
1. Buat transaksi Revenue: Rp 100.000.000
2. Buat transaksi Expense: Rp 65.000.000
3. Period Closing 1 Jan - 31 Jan
4. Verify: Revenue = 0, Expense = 0, Laba Ditahan = +35.000.000
5. Verify: Balance Sheet BALANCED
```

**Test 2: Closing Kedua (Critical)**
```
1. Pastikan Revenue = 0, Expense = 0 (dari closing sebelumnya)
2. Buat transaksi baru di Feb (jika ada)
3. Period Closing 1 Feb - 28 Feb
4. Verify: Preview hanya show accounts dengan balance > 0.01
5. Verify: Journal entry TIDAK include akun dengan balance 0
6. Verify: Balance Sheet tetap BALANCED
7. Verify: Laba Ditahan TIDAK double posting
```

**Test 3: Edge Cases**
```
1. Closing dengan semua akun sudah 0 (seharusnya no-op)
2. Closing dengan balance sangat kecil (< 0.01) seharusnya skip
3. Multiple closings in same day (should prevent)
```

## Validation

### Pre-Fix Check
```sql
-- 1. Cek balance sheet
SELECT 
    SUM(CASE WHEN type = 'ASSET' THEN balance ELSE 0 END) AS assets,
    SUM(CASE WHEN type = 'LIABILITY' THEN balance ELSE 0 END) AS liabilities,
    SUM(CASE WHEN type = 'EQUITY' THEN balance ELSE 0 END) AS equity,
    (SUM(CASE WHEN type = 'ASSET' THEN balance ELSE 0 END) - 
     SUM(CASE WHEN type = 'LIABILITY' THEN balance ELSE 0 END) - 
     SUM(CASE WHEN type = 'EQUITY' THEN balance ELSE 0 END)) AS difference
FROM accounts WHERE is_active = true AND is_header = false;

-- Expected: difference should be != 0 (e.g., -35000000)
```

### Post-Fix Check
```sql
-- Same query as above
-- Expected: difference should be ~0 (< 0.01)
```

### Verify Laba Ditahan
```sql
SELECT 
    je.entry_date,
    je.code,
    je.description,
    jl.debit_amount,
    jl.credit_amount,
    (jl.credit_amount - jl.debit_amount) AS net_effect,
    je.status
FROM journal_lines jl
JOIN journal_entries je ON jl.journal_entry_id = je.id
JOIN accounts a ON jl.account_id = a.id
WHERE a.code = '3201'
ORDER BY je.entry_date;

-- Expected: 
-- 1. Only ONE closing entry per period
-- 2. No voided entries affecting balance
-- 3. Sum of net_effect = current Laba Ditahan balance
```

## Prevention Checklist

✅ **Code Level:**
- [x] Use `ABS(balance) > 0.01` threshold in queries
- [x] Double-check balance in loops before creating journal lines
- [x] Add validation in preview to warn if no accounts to close
- [x] Apply same fix to both `period_closing_service.go` and `fiscal_year_closing_service.go`

✅ **Database Level:**
- [x] Create SQL scripts for analysis
- [x] Create SQL scripts for fixing
- [x] Add validation queries

✅ **Testing:**
- [ ] Unit test for closing with all zero balances
- [ ] Integration test for multiple consecutive closings
- [ ] Test with floating point precision edge cases

✅ **Documentation:**
- [x] Document root cause
- [x] Document solution
- [x] Create runbook for similar issues

## Files Changed

### Modified:
1. `services/period_closing_service.go`
   - Lines 156-166: Query dengan ABS(balance) > 0.01
   - Lines 332-344: Check balance > 0.01 sebelum create journal line (revenue)
   - Lines 372-384: Check balance > 0.01 sebelum create journal line (expense)

2. `services/fiscal_year_closing_service.go`
   - Lines 97-108: Query dengan ABS(balance) > 0.01
   - Lines 229-241: Check balance > 0.01 (revenue)
   - Lines 271-283: Check balance > 0.01 (expense)

### Created:
1. `analyze_double_posting_issue.sql` - Script untuk analisis
2. `fix_double_posting_issue.sql` - Script untuk perbaikan
3. `docs/FIX_DOUBLE_POSTING_LABA_DITAHAN.md` - Dokumentasi ini

## Timeline

- **Issue Detected:** [Tanggal issue ditemukan]
- **Root Cause Found:** [Tanggal]
- **Fix Applied:** [Tanggal]
- **Verified:** [Tanggal]

## Contact

Jika masalah serupa terjadi lagi:
1. Jangan panic
2. Jalankan `analyze_double_posting_issue.sql` untuk diagnosis
3. Review hasil analisis
4. Backup database
5. Jalankan `fix_double_posting_issue.sql` jika diperlukan
6. Verify dengan balance sheet check

## Related Issues

- Floating point precision in balance calculations
- Period closing logic
- Journal entry balance validation

## References

- PSAK (Standar Akuntansi): Temporary vs Permanent Accounts
- Period Closing Best Practices
- Double-Entry Bookkeeping Principles
