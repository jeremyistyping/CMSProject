# Bugfix: Budget vs Actual Menampilkan Rp 0

## Masalah

Pada halaman Budget vs Actual untuk project Padel Bandung:
- CBS Summary menampilkan Total Budget: **Rp 3.5 miliar** ✅
- Budget vs Actual menampilkan: **Rp 0** untuk semua kategori ❌

## Root Cause Analysis

### 1. Struktur Data
Ada 3 tabel yang terlibat:
- **`cbs_nodes`**: Cost Breakdown Structure (hierarki biaya proyek)
- **`project_budgets`**: Budget per COA account per project
- **`expense_transactions`**: Transaksi pengeluaran aktual

### 2. Data di Database

**CBS Nodes (Project ID = 6):**
```sql
SELECT COUNT(*), COUNT(coa_account_id), SUM(budget_amount) 
FROM cbs_nodes WHERE project_id = 6;
-- Result: 24 nodes, 0 with COA mapping, Rp 3.5 miliar total budget
```

**Project Budgets (Project ID = 6):**
```sql
SELECT COUNT(*), SUM(estimated_amount) 
FROM project_budgets WHERE project_id = 6;
-- Result: 3 entries, Rp 105 juta total
```

**Expense Transactions (Project ID = 6):**
```sql
SELECT COUNT(*), SUM(amount) 
FROM expense_transactions WHERE project_id = 6;
-- Result: 2 transactions, Rp 1.7 juta total
```

### 3. Masalah Utama

1. **CBS nodes tidak memiliki mapping ke COA accounts** (`coa_account_id` = NULL)
2. **Budget vs Actual query** menggunakan `project_budgets` sebagai sumber budget
3. **CBS Summary** menggunakan `cbs_nodes` sebagai sumber budget
4. Kedua sumber data tidak sinkron!

## Solusi Implementasi

### Perubahan pada `expense_transaction_repository.go`

Mengubah query `GetBudgetVsActualSummary` untuk:
1. Mengambil budget dari `project_budgets`
2. Mengambil actual dari `expense_transactions`
3. **Menampilkan semua COA yang memiliki budget ATAU transaksi**

```go
// Query baru menggunakan CTE (Common Table Expression)
WITH budget_data AS (
    -- Ambil budget dari project_budgets
    SELECT project_id, account_id, estimated_amount
    FROM project_budgets
    WHERE project_id = ? AND deleted_at IS NULL
),
actual_data AS (
    -- Ambil actual dari expense_transactions
    SELECT project_id, coa_account_id, SUM(amount) as actual_amount
    FROM expense_transactions
    WHERE project_id = ? AND deleted_at IS NULL
        AND transaction_date BETWEEN ? AND ?
    GROUP BY project_id, coa_account_id
),
all_coa AS (
    -- Gabungkan semua COA yang ada di budget atau actual
    SELECT DISTINCT coa_account_id, project_id FROM budget_data
    UNION
    SELECT DISTINCT coa_account_id, project_id FROM actual_data
)
SELECT 
    ac.project_id,
    ac.coa_account_id,
    coa.code, coa.name, coa.budget_category, coa.work_package,
    COALESCE(bd.budget_estimation, 0) as budget_estimation,
    COALESCE(ad.actual_amount, 0) as actual_amount,
    COALESCE(bd.budget_estimation, 0) - COALESCE(ad.actual_amount, 0) as variance
FROM all_coa ac
JOIN coa_accounts coa ON ac.coa_account_id = coa.id
LEFT JOIN budget_data bd ON bd.coa_account_id = ac.coa_account_id
LEFT JOIN actual_data ad ON ad.coa_account_id = ac.coa_account_id
WHERE coa.deleted_at IS NULL
ORDER BY coa.budget_category, coa.code
```

### Keuntungan Solusi Ini

1. ✅ **Menampilkan actual cost** meskipun tidak ada budget
2. ✅ **Menampilkan budget** meskipun belum ada transaksi
3. ✅ **Tidak perlu mengubah data existing** (CBS nodes)
4. ✅ **Backward compatible** dengan project yang sudah punya project_budgets
5. ✅ **Fleksibel** untuk project yang menggunakan CBS atau project_budgets

## Hasil

Setelah perubahan:
- Budget vs Actual akan menampilkan:
  - Budget dari `project_budgets` (jika ada)
  - Actual dari `expense_transactions` (selalu ditampilkan)
  - COA yang memiliki transaksi akan muncul meskipun tidak ada budget

## Rekomendasi Jangka Panjang

Untuk konsistensi data, disarankan:

1. **Tambahkan COA mapping ke CBS nodes**
   - Update seeder untuk include `coa_account_id`
   - Buat UI untuk mapping CBS node ke COA

2. **Sinkronisasi otomatis CBS ke Project Budgets**
   - Sudah ada method `syncToProjectBudget` di `cbs_service.go`
   - Perlu dijalankan untuk data existing

3. **Pilih satu source of truth untuk budget**
   - Opsi A: Gunakan `project_budgets` (current)
   - Opsi B: Gunakan `cbs_nodes` (lebih detail, hierarkis)
   - Opsi C: Hybrid (CBS untuk planning, project_budgets untuk execution)

## Files Changed

- `backend/repositories/expense_transaction_repository.go`
  - Modified: `GetBudgetVsActualSummary()` method

## Testing

```bash
# Test Budget vs Actual endpoint
curl -X GET "http://localhost:8080/api/v1/projects/6/reports/budget-vs-actual?start_date=2025-01-01&end_date=2025-12-31" \
  -H "Authorization: Bearer <token>"
```

Expected result:
- Menampilkan COA dengan actual cost
- Budget = Rp 105 juta (dari project_budgets)
- Actual = Rp 1.7 juta (dari expense_transactions)
