# Script Reset Data Transaksi - Sistem Akuntansi

## Deskripsi
Kumpulan script untuk melakukan reset data transaksi pada aplikasi sistem akuntansi berbasis Go + GORM + PostgreSQL.

## ⚠️ PERINGATAN PENTING
- **BACKUP database Anda sebelum menjalankan script apapun!**
- Script ini akan **MENGHAPUS DATA SECARA PERMANEN**
- Pastikan Anda memahami konsekuensi setiap script sebelum menjalankannya

## Script yang Tersedia

### 1. 🟢 Reset Data Transaksi (DISARANKAN)
**File:** `cmd/reset_transaction_data.go`
**Eksekusi:** `go run cmd/reset_transaction_data.go` atau `.\scripts\reset-transaksi.ps1`

#### Yang DIPERTAHANKAN:
- ✅ Chart of Accounts (COA) - struktur tetap ada
- ✅ Master data produk - data produk tetap ada  
- ✅ Data kontak/customer/vendor
- ✅ Data user dan permission
- ✅ Master data cash bank
- ✅ Master data kategori

#### Yang DIHAPUS:
- ❌ Semua transaksi penjualan (sales, sale_items, sale_payments)
- ❌ Semua transaksi pembelian (purchases, purchase_items, etc)
- ❌ Semua jurnal entry (journals, journal_entries)
- ❌ Semua payment records
- ❌ Semua inventory movements
- ❌ Semua expense records
- ❌ Semua approval data
- ❌ Balance accounts direset ke 0
- ❌ Stock produk direset ke 0

### 2. 🔴 Reset Database Total (BERBAHAYA!)
**File:** `cmd/reset_database_total.go`
**Eksekusi:** `go run cmd/reset_database_total.go`

#### Yang DIHAPUS:
- ❌ **SEMUA DATA** termasuk COA, master data, dll
- ❌ Database dikosongkan total

#### Yang DIBUAT ULANG:
- ✅ Struktur database kosong dari migration
- ✅ Data seed default (jika ada)

### 3. 🔵 Restore COA dari Backup
**File:** `cmd/restore_coa_from_backup.go`
**Eksekusi:** `go run cmd/restore_coa_from_backup.go`

Mengembalikan COA dari tabel backup yang dibuat saat reset transaksi.

## Cara Penggunaan

### Metode 1: PowerShell Script (Termudah)
```powershell
# Jalankan di PowerShell dari direktori backend
.\scripts\reset-transaksi.ps1
```

### Metode 2: Manual Go Command
```bash
# Reset data transaksi (COA dipertahankan)
go run cmd/reset_transaction_data.go

# Reset database total (BERBAHAYA!)
go run cmd/reset_database_total.go

# Restore COA dari backup
go run cmd/restore_coa_from_backup.go
```

### Metode 3: SQL Manual
```bash
# Backup COA terlebih dahulu
psql -d your_database -f scripts/backup_coa.sql

# Reset data transaksi
psql -d your_database -f scripts/reset_transaction_data.sql
```

## Urutan Proses Reset Transaksi

1. **Backup COA** → Buat tabel backup: `accounts_backup`, `accounts_hierarchy_backup`, `accounts_original_balances`
2. **Delete Transaction Data** → Hapus data transaksi dalam urutan yang aman (foreign key dependencies)
3. **Reset Balances** → Set balance accounts dan cash banks ke 0
4. **Reset Stock** → Set stock produk ke 0  
5. **Reset Sequences** → Reset auto increment ID ke 1
6. **Audit Log** → Catat aktivitas reset

## Tabel Backup yang Dibuat

Saat menjalankan reset data transaksi, akan dibuat tabel backup:

- `accounts_backup` - Backup semua data COA
- `accounts_hierarchy_backup` - Backup struktur hierarki COA
- `accounts_original_balances` - Backup balance asli sebelum direset

## Recovery

Jika terjadi kesalahan atau ingin mengembalikan COA:

```bash
go run cmd/restore_coa_from_backup.go
```

## Untuk "Mengulang dari 0" Seperti Prisma

Jika Anda ingin equivalent dengan `npx prisma db push` atau `prisma generate` untuk "mengulang dari 0":

### Opsi 1: Reset Data Transaksi Saja
```bash
go run cmd/reset_transaction_data.go
```
Ini akan membuat Anda bisa mulai input transaksi dari 0 dengan COA yang masih ada.

### Opsi 2: Total Reset (Seperti Drop + Recreate Database)
```bash
go run cmd/reset_database_total.go
```
Ini equivalent dengan drop semua tabel dan recreate dari migration.

### Opsi 3: Manual Database Recreation
```bash
# 1. Drop database (manual di PostgreSQL)
# 2. Create database baru
# 3. Jalankan aplikasi untuk auto migration
go run cmd/main.go
```

## Troubleshooting

### Error "table does not exist"
Beberapa tabel mungkin belum ada. Script sudah dibuat untuk handle ini dengan `IF EXISTS`.

### Error foreign key constraint
Script sudah diurutkan untuk menghapus child tables terlebih dahulu sebelum parent tables.

### Error sequence
Jika ada error saat reset sequence, coba manual:
```sql
ALTER SEQUENCE <table_name>_id_seq RESTART WITH 1;
```

## Keamanan

- Script memiliki multiple confirmation untuk mencegah eksekusi tidak sengaja
- Backup COA dibuat otomatis sebelum reset
- Audit log dicatat untuk tracking
- Transaction digunakan untuk rollback jika ada error

## Environment

Script ini menggunakan:
- Go 1.23+
- GORM v1.30+
- PostgreSQL
- File `.env` untuk konfigurasi database
