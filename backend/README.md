# Accounting Backend

Backend API untuk sistem akuntansi dengan fitur lengkap termasuk SSOT (Single Source of Truth) Journal System.

## ⚡ SUPER QUICK START (PC Baru)

**Setelah `git pull`, hanya 3 langkah:**

```bash
# 1. Edit .env sesuai database Anda
# 2. Jalankan fix (WAJIB!)
go run apply_database_fixes.go
# 3. Start backend
go run main.go
```

✅ **Script otomatis baca .env Anda - no hardcode!**

📖 Detail lengkap: [SETUP_INSTRUCTIONS.md](./SETUP_INSTRUCTIONS.md)

---

## 🚀 Quick Start (Full Setup)

### Prerequisites
- Go 1.19+
- PostgreSQL 13+
- Database `sistem_akuntans_test` sudah dibuat

### 1. Setup Environment (Untuk PC Baru)

Setelah `git clone` atau `git pull` di PC baru, **WAJIB** jalankan setup berikut:

#### 🛡️ Balance Protection Setup (CRITICAL)

**⚠️ PENTING:** Sistem ini mencegah balance mismatch yang bisa merusak laporan keuangan!

**Windows:**
```bash
# Masuk ke direktori backend
cd backend

# Setup balance protection system
setup_balance_protection.bat
```

**Linux/Mac:**
```bash
# Masuk ke direktori backend
cd backend

# Setup balance protection system
chmod +x setup_balance_protection.sh
./setup_balance_protection.sh
```

**Manual Alternative:**
```bash
# Jika script di atas tidak bisa jalan
go run cmd/scripts/setup_balance_sync_auto.go
```

#### ⚙️ Environment Setup (Optional)

```bash
# Jalankan migration fixes jika diperlukan
go run cmd/fix_migrations.go
go run cmd/fix_remaining_migrations.go

# Verifikasi setup berhasil
go run cmd/final_verification.go
```

### 2. Jalankan Backend

```bash
go run cmd/main.go
```

Backend akan berjalan di:
- **API**: http://localhost:8080/api/v1
- **Swagger Docs**: http://localhost:8080/swagger/index.html
- **Health Check**: http://localhost:8080/api/v1/health

## 🔧 Database Configuration

Pastikan PostgreSQL connection string sudah benar:
```
postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable
```

## 📝 Migration Scripts

### Apa itu Migration Fixes?

Migration fixes adalah script untuk mengatasi masalah kompatibilitas database dan memastikan SSOT Journal System berjalan dengan baik. Script ini:

- ✅ Membuat tabel `purchase_payments` yang missing
- ✅ Membuat materialized view `account_balances` untuk SSOT
- ✅ Membuat functions untuk sync balance (`sync_account_balance_from_ssot`)
- ✅ Memperbaiki index dan constraint yang bermasalah

### Kapan Perlu Menjalankan?

**WAJIB dijalankan di:**
- ✅ PC baru setelah git clone
- ✅ Environment baru (development/staging/production)
- ✅ Setelah database reset/restore
- ✅ Jika muncul error SSOT Journal System

**TIDAK perlu dijalankan jika:**
- ❌ Sudah pernah dijalankan di PC yang sama
- ❌ Backend sudah berjalan normal tanpa error

### Troubleshooting

Jika backend masih error setelah migration fixes:

```bash
# Cek status database
go run cmd/final_verification.go

# Jika masih ada masalah, coba jalankan ulang
go run cmd/fix_remaining_migrations.go
```

## 🏗️ Build Backend

```bash
docker build --push --platform linux/amd64 -t registry.digitalocean.com/registry-tigapilar/dbm/account-backend:latest .
```

## 📚 API Documentation

Setelah backend running, akses dokumentasi lengkap di:
- **Swagger UI**: http://localhost:8080/swagger/index.html
- **API Endpoints**: 400+ endpoint tersedia
- **Authentication**: JWT-based dengan role permission

## 🛡️ Features

- ✅ **Balance Protection System** - **NEW!** Auto-prevent balance mismatch issues
- ✅ **SSOT Journal System** - Single source of truth untuk semua transaksi
- ✅ **Account Balance Sync** - Real-time automatic balance synchronization
- ✅ **Purchase Payment Integration** - Complete purchase-to-payment workflow
- ✅ **Sales Management** - Full sales cycle management
- ✅ **Financial Reporting** - Trial balance, P&L, Balance sheet
- ✅ **Asset Management** - Fixed asset tracking dengan depreciation
- ✅ **Cash Bank Management** - Multi-currency, multi-account
- ✅ **Approval Workflow** - Configurable approval processes
- ✅ **Audit Trail** - Complete transaction logging

## 🛡️ Balance Protection System

**⚠️ CRITICAL:** Sistem ini **WAJIB** di-setup di setiap PC untuk mencegah masalah balance yang bisa merusak data keuangan!

### ❓ Apa itu Balance Protection?

Sistem otomatis yang:
- 🔄 **Auto-sync** balance saat ada transaksi baru
- 🔍 **Monitor** konsistensi balance real-time  
- 🚑 **Fix** masalah balance secara otomatis
- 📈 **Prevent** laporan keuangan yang salah

### 🚀 Quick Setup

```bash
# Windows
setup_balance_protection.bat

# Linux/Mac  
./setup_balance_protection.sh

# Manual
go run cmd/scripts/setup_balance_sync_auto.go
```

### ✅ Verification

```sql
-- Check sistem sudah installed
SELECT COUNT(*) FROM information_schema.triggers WHERE trigger_name = 'balance_sync_trigger';
-- Hasil harus: 1

-- Check balance health
SELECT * FROM account_balance_monitoring WHERE status='MISMATCH';
-- Hasil harus: empty (no mismatches)
```

### 📁 More Info

- **Setup Guide**: `README_BALANCE_SETUP.md`
- **Full Documentation**: `BALANCE_PREVENTION_GUIDE.md`
- **Migration File**: `migrations/balance_sync_system.sql`

---

## 🔍 Troubleshooting

### Balance Issues

```sql
-- Health check
SELECT * FROM account_balance_monitoring WHERE status='MISMATCH';

-- Manual fix
SELECT * FROM sync_account_balances();
```

### General Issues

```bash
# Cek status database
go run cmd/final_verification.go

# Jika masih ada masalah, coba jalankan ulang
go run cmd/fix_remaining_migrations.go
```

---

> **💡 Tips**: Jika mengalami masalah, pastikan Balance Protection sudah di-setup dengan menjalankan `setup_balance_protection.bat` (Windows) atau `./setup_balance_protection.sh` (Linux/Mac).
