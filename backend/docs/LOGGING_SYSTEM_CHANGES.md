# Perubahan Sistem Logging

## Ringkasan Perubahan

Dokumen ini menjelaskan perubahan yang telah dilakukan pada sistem logging untuk mengurangi duplikasi dan memfokuskan setiap log pada fungsinya masing-masing.

## 1. Activity Log (activity_logs table)

### Sebelum Perubahan
- Mencatat SEMUA aktivitas API termasuk:
  - Request user yang sebenarnya (create, update, delete)
  - Notifikasi otomatis sistem (polling, auto-generated)
  - Request internal sistem

### Setelah Perubahan
**Fokus: User-Initiated Actions Only**

Activity log sekarang HANYA mencatat aktivitas yang:
- Dilakukan langsung oleh user (user-requested actions)
- Bukan aktivitas sistem otomatis

#### Path yang di-skip (tidak masuk activity log):
```go
// Tidak dicatat di activity log:
- /api/v1/notifications  // Notification polling otomatis
- /health
- /favicon.ico
- /uploads/
- /templates/
- /swagger/
```

#### Manfaat:
- ✅ Activity log lebih bersih dan fokus ke aksi user
- ✅ Mudah tracking apa yang dilakukan user
- ✅ Menampilkan full log untuk aksi yang direquest user
- ✅ Tidak ada spam dari notifikasi otomatis

### File yang Diubah:
- `backend/middleware/activity_logger_middleware.go`
  - Fungsi `shouldSkipLogging()` - ditambahkan `/api/v1/notifications`

## 2. Audit Log (audit_logs table)

### Sebelum Perubahan
- Mencatat SEMUA transaksi API
- Duplikasi dengan activity log untuk banyak endpoint
- File audit.log terlalu besar dan susah dibaca

### Setelah Perubahan
**Fokus: Financial Transactions Only**

Audit log sekarang HANYA mencatat transaksi finansial/akuntansi:

#### Endpoint yang di-audit:
```go
✅ /api/v1/sales          // Transaksi penjualan
✅ /api/v1/purchases      // Transaksi pembelian
✅ /api/v1/payments       // Pembayaran piutang/hutang
✅ /api/v1/payment        // Pembayaran
✅ /api/v1/cashbank       // Transaksi kas/bank
✅ /api/v1/cash-bank      // Alternatif endpoint kas/bank
```

#### Yang TIDAK di-audit:
```go
❌ /api/v1/users
❌ /api/v1/products
❌ /api/v1/contacts
❌ /api/v1/settings
❌ /api/v1/reports
❌ /api/v1/dashboard
❌ /api/v1/approvals (sudah di activity log)
❌ /api/v1/notifications
```

#### Manfaat:
- ✅ Audit log fokus ke transaksi finansial saja
- ✅ Tidak ada duplikasi dengan activity log
- ✅ File audit.log lebih ringkas dan mudah dibaca
- ✅ Memudahkan audit keuangan dan compliance

### File yang Diubah:
- `backend/middleware/audit_logger.go`
  - Tambah fungsi `isFinancialTransaction()` 
  - Update `AuditMiddleware()` untuk filter hanya transaksi finansial

## 3. Pemisahan Fungsi Kedua Log

| Aspek | Activity Log | Audit Log |
|-------|-------------|-----------|
| **Tujuan** | Tracking user actions | Financial transaction audit trail |
| **Scope** | Semua user activities (kecuali notif otomatis) | Hanya transaksi finansial (sales, purchase, payment, kas/bank) |
| **Target User** | Admin untuk monitoring user behavior | Finance/Accounting untuk audit keuangan |
| **File Output** | Database: `activity_logs` | Database: `audit_logs` + File: `logs/audit.log` |
| **Contoh Use Case** | "Siapa yang update product X?" | "Siapa yang approve payment Y senilai Rp 100jt?" |

## 4. Testing & Verifikasi

### Test Scenario 1: Notification Polling
**Expected:** Tidak masuk activity_logs
```bash
# Request: GET /api/v1/notifications
# Result: ✅ Tidak tercatat di activity_logs
# Result: ✅ Tidak tercatat di audit_logs
```

### Test Scenario 2: Create Sales
**Expected:** Masuk activity_logs DAN audit_logs
```bash
# Request: POST /api/v1/sales
# Result: ✅ Tercatat di activity_logs (user action)
# Result: ✅ Tercatat di audit_logs (financial transaction)
```

### Test Scenario 3: Update Product
**Expected:** Masuk activity_logs saja, TIDAK masuk audit_logs
```bash
# Request: PUT /api/v1/products/123
# Result: ✅ Tercatat di activity_logs (user action)
# Result: ✅ TIDAK tercatat di audit_logs (bukan transaksi finansial)
```

### Test Scenario 4: Create Payment
**Expected:** Masuk activity_logs DAN audit_logs
```bash
# Request: POST /api/v1/payments
# Result: ✅ Tercatat di activity_logs (user action)
# Result: ✅ Tercatat di audit_logs (financial transaction)
```

## 5. Migration Notes

### Tidak Ada Breaking Changes
- Struktur table tidak berubah
- Existing logs tetap utuh
- Hanya filtering yang ditambahkan

### Cleanup Recommendations (Optional)
Jika ingin membersihkan log lama yang redundant:

```sql
-- OPTIONAL: Hapus notification logs dari activity_logs (jika mau cleanup)
DELETE FROM activity_logs 
WHERE path LIKE '/api/v1/notifications%';

-- OPTIONAL: Hapus non-financial dari audit_logs (jika mau cleanup)
DELETE FROM audit_logs 
WHERE table_name NOT IN ('sales', 'purchases', 'payments', 'cash_banks');
```

## 6. Future Improvements

### Potential Enhancements:
1. **Log Rotation**: Auto-delete logs older than X months
2. **Log Analytics**: Dashboard untuk visualisasi activity & audit logs
3. **Alert System**: Email notification untuk suspicious activities
4. **Export Feature**: Export logs to CSV/Excel untuk reporting

## 7. Rollback Instructions

Jika perlu rollback ke sistem lama:

### Rollback Activity Log:
```go
// Di activity_logger_middleware.go, remove line:
"/api/v1/notifications", // Skip automatic notification polling
```

### Rollback Audit Log:
```go
// Di audit_logger.go, remove block:
// Only audit financial transactions
if !isFinancialTransaction(c.Request.URL.Path) {
    c.Next()
    return
}

// Dan hapus fungsi isFinancialTransaction()
```

---

**Tanggal Implementasi:** 2025-10-24  
**Developer:** System Admin  
**Status:** ✅ Implemented & Ready for Testing
