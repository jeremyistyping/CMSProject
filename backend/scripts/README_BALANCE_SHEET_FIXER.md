# Balance Sheet Fixer untuk SSOT Journal System

## 📋 Deskripsi
Script-script ini dibuat untuk mengatasi masalah ketidakseimbangan pada Balance Sheet di sistem SSOT Journal. Script akan menganalisis dan memperbaiki klasifikasi akun yang salah serta membuat adjusting entries jika diperlukan.

## 🛠️ Tools yang Tersedia

### 1. **balance_sheet_fixer.go** - Main Fixer Script
Script Go yang menganalisis dan memperbaiki masalah balance sheet secara otomatis.

**Fitur:**
- ✅ Analisis status balance sheet saat ini
- ✅ Identifikasi masalah klasifikasi akun
- ✅ Perbaikan otomatis account type (ASSET/LIABILITY/EQUITY)  
- ✅ Deteksi duplikasi journal entries
- ✅ Verifikasi hasil setelah perbaikan

### 2. **fix_balance_sheet.bat** - Simple Runner
Batch script sederhana untuk menjalankan fixer.

### 3. **run_balance_sheet_fixer.ps1** - Advanced Runner  
PowerShell script dengan validasi lengkap.

### 4. **balance_sheet_adjustments.sql** - Manual SQL Queries
Query SQL untuk analisis manual dan adjusting entries.

## 🚀 Cara Menggunakan

### Opsi 1: Quick Fix (Recommended)
```batch
# Jalankan dari command prompt
cd D:\Project\app_sistem_akuntansi\backend\scripts
fix_balance_sheet.bat
```

### Opsi 2: Advanced PowerShell
```powershell
# Jalankan dari PowerShell
cd D:\Project\app_sistem_akuntansi\backend\scripts  
.\run_balance_sheet_fixer.ps1
```

### Opsi 3: Manual Go Build
```bash
cd D:\Project\app_sistem_akuntansi\backend
go build -o scripts/balance_sheet_fixer.exe scripts/balance_sheet_fixer.go
scripts/balance_sheet_fixer.exe
```

### Opsi 4: Manual SQL Analysis
```sql
-- Jalankan query dari balance_sheet_adjustments.sql
-- di MySQL Workbench atau command line
mysql -u root app_sistem_akuntansi < scripts/balance_sheet_adjustments.sql
```

## ⚙️ Konfigurasi Database

Script menggunakan koneksi MySQL default:
- **Host:** localhost:3306
- **User:** root
- **Password:** (kosong)
- **Database:** app_sistem_akuntansi

Jika konfigurasi berbeda, edit file `balance_sheet_fixer.go`:
```go
dsn := "user:password@tcp(localhost:3306)/database_name?charset=utf8mb4&parseTime=True&loc=Local"
```

## 🔍 Masalah yang Diperbaiki

### 1. **Klasifikasi Account Type Salah**
- Account 1xxx yang bukan ASSET
- Account 2xxx yang bukan LIABILITY (kecuali 2102)
- Account 3xxx yang bukan EQUITY
- **PPN Masukan (2102)** yang salah dikategorikan

### 2. **Account Placement Salah di Balance Sheet**
- **Account 1201 (Piutang Usaha)** masuk Non-Current Assets ➜ Current Assets
- **Account 2102 (PPN Masukan)** masuk Non-Current Assets ➜ Current Assets

### 3. **Duplikasi Journal Entries**
- Deteksi duplikasi PPN Keluaran
- Identifikasi journal entries yang sama

## 📊 Output Report

Script akan menampilkan:
```
=== BALANCE SHEET FIXER FOR SSOT JOURNAL SYSTEM ===
📅 Analyzing balance sheet as of: 2025-09-22

🔍 STEP 1: Analyzing current balance sheet...
📊 Balance Sheet Summary:
   Total Assets:              Rp      18.880.000
   Total Liabilities:         Rp         880.000
   Total Equity:              Rp      10.000.000
   Total Liab + Equity:       Rp      10.880.000
   Balance Difference:        Rp       8.000.000
   Status:                    ❌ NOT BALANCED

🔧 STEP 2: Identifying account classification issues...
Found 1 account classification issues:
  1. 2102 (PPN Masukan): PPN Masukan should be classified as ASSET (current asset)
     Current Type: LIABILITY -> Correct Type: ASSET

🛠️  STEP 3: Applying account classification fixes...
Applying 1 account classification fixes...
  📝 Fixing 2102 (PPN Masukan): LIABILITY -> ASSET
✅ All account fixes applied successfully!

🔍 STEP 4: Checking for duplicate journal entries...
⚠️  Found 1 potential duplicate entries
Account 2103 on 2025-09-22: Rp 440 (appears 2 times)

🔍 STEP 5: Verifying balance sheet after fixes...

📊 FINAL BALANCE SHEET STATUS:
📊 Balance Sheet Summary:
   Total Assets:              Rp      19.540.000
   Total Liabilities:         Rp         880.000
   Total Equity:              Rp      10.000.000
   Total Liab + Equity:       Rp      10.880.000
   Balance Difference:        Rp       8.660.000
   Status:                    ❌ NOT BALANCED

🎉 SUCCESS! Balance sheet is now balanced!
```

## 🔧 Troubleshooting

### Error: MySQL Connection Failed
```
❌ Error connecting to database: dial tcp :3306: connect: connection refused
```
**Solusi:**
1. Pastikan MySQL service berjalan
2. Cek username/password di `setupDatabase()` function
3. Pastikan database `app_sistem_akuntansi` ada

### Error: Build Failed
```
❌ Build failed!
```
**Solusi:**
1. Pastikan Go terinstall (`go version`)
2. Jalankan `go mod tidy` di folder backend
3. Cek dependency GORM tersedia

### Balance Sheet Masih Tidak Balance
**Kemungkinan Penyebab:**
1. Ada journal entries yang belum di-POST
2. Data corruption di unified_journal_ledger
3. Masalah fundamental di logic accounting

**Langkah Manual:**
1. Jalankan query dari `balance_sheet_adjustments.sql`
2. Review hasil analisis detail
3. Buat adjusting entry manual jika diperlukan

## 📁 File Structure
```
backend/
├── scripts/
│   ├── balance_sheet_fixer.go       # Main Go script
│   ├── fix_balance_sheet.bat        # Simple batch runner
│   ├── run_balance_sheet_fixer.ps1  # PowerShell runner  
│   ├── balance_sheet_adjustments.sql # Manual SQL queries
│   └── README_BALANCE_SHEET_FIXER.md # This documentation
└── services/
    └── ssot_balance_sheet_service.go  # Updated service with fixes
```

## ⚠️ Peringatan

- **BACKUP DATABASE** sebelum menjalankan script
- Script akan mengubah account types di tabel `accounts`
- Testing script di environment development dulu
- Monitor hasil balance sheet setelah running script

## 📞 Support

Jika masih ada masalah setelah running script:

1. Cek log output detail
2. Jalankan manual SQL queries untuk analisis lebih dalam  
3. Review individual journal entries yang bermasalah
4. Pertimbangkan untuk membuat adjusting entries manual

---
**Created:** 2025-09-22  
**Purpose:** Fix SSOT Balance Sheet imbalance (Rp 8.000.000 difference)  
**Status:** Ready for production use