# 🔄 Database Reset Guide

Panduan lengkap untuk mereset database aplikasi Unipro Cost Control Management System.

## 🎯 Kapan Menggunakan Reset Database?

Gunakan reset database ketika:
- ✅ Ingin testing dari awal dengan data bersih
- ✅ Struktur database berubah dan perlu migration ulang
- ✅ Data testing sudah berantakan
- ✅ Ingin memulai development dari nol
- ⚠️ **JANGAN** gunakan di production tanpa backup!

## 🚀 Quick Start

### Windows (Paling Mudah)

```cmd
cd backend\cmd\reset_db
reset_db.bat
```

### Linux/Mac

```bash
cd backend/cmd/reset_db
chmod +x reset_db.sh
./reset_db.sh
```

### Menggunakan Make

```bash
cd backend/cmd/reset_db
make reset
```

## 🛡️ Reset dengan Backup (Recommended)

### Windows

```cmd
cd backend\cmd\reset_db
backup_and_reset.bat
```

### Linux/Mac

```bash
cd backend/cmd/reset_db
chmod +x backup_and_reset.sh
./backup_and_reset.sh
```

## 📋 Langkah-langkah Detail

### 1. Persiapan

Pastikan file `.env` sudah dikonfigurasi:

```env
DATABASE_URL=postgres://postgres:postgres@localhost/sistem_akuntans_dev?sslmode=disable
```

### 2. Backup (Optional tapi Recommended)

```bash
# Manual backup
pg_dump -U postgres sistem_akuntans_dev > backup.sql

# Atau gunakan script backup_and_reset
```

### 3. Reset Database

```bash
cd backend/cmd/reset_db
go run main.go
```

Script akan:
1. Membaca DATABASE_URL dari .env
2. Meminta konfirmasi (ketik 'YES')
3. Drop semua tables
4. Drop semua sequences
5. Drop semua custom types

### 4. Jalankan Aplikasi

Setelah reset, jalankan aplikasi untuk apply migrations:

```bash
cd backend
go run main.go
```

Aplikasi akan otomatis:
- ✅ Menjalankan SQL migrations
- ✅ Menjalankan GORM AutoMigrate
- ✅ Seed data awal (users, roles, permissions)
- ✅ Seed master data (COA, materials, vendors)

## 📝 Contoh Output

```
🔄 Starting database reset process...
⚠️  WARNING: This will DELETE ALL DATA in the database!
📍 Database: postgres://postgres:****@localhost/sistem_akuntans_dev

❓ Are you sure you want to reset the database? Type 'YES' to confirm: YES
✅ Connected to database successfully

🗑️  Step 1: Dropping all tables...
   📋 Found 45 tables to drop
   ✓ Dropped table: users
   ✓ Dropped table: projects
   ✓ Dropped table: purchase_requests
   ✓ Dropped table: coa_accounts
   ✓ Dropped table: materials
   ... (dan seterusnya)
✅ All tables dropped successfully

🗑️  Step 2: Dropping all sequences...
   📋 Found 30 sequences to drop
   ✓ Dropped sequence: users_id_seq
   ... (dan seterusnya)
✅ All sequences dropped successfully

🗑️  Step 3: Dropping all custom types...
   📋 Found 5 custom types to drop
   ✓ Dropped type: approval_status
   ✓ Dropped type: pr_status
   ... (dan seterusnya)
✅ All custom types dropped successfully

✅ Database reset completed successfully!
```

## 🔐 Keamanan

- ✅ Tidak ada hardcoded credentials
- ✅ Password di-mask di log output
- ✅ Meminta konfirmasi sebelum delete
- ✅ Membaca dari .env file

## 🐛 Troubleshooting

### Error: DATABASE_URL not found

```bash
# Cek apakah .env ada
ls backend/.env

# Atau copy dari example
cp backend/.env.example backend/.env
# Edit .env dan sesuaikan DATABASE_URL
```

### Error: Failed to connect to database

```bash
# Cek PostgreSQL berjalan
psql -U postgres -c "SELECT version();"

# Cek koneksi manual
psql "postgres://postgres:postgres@localhost/sistem_akuntans_dev"
```

### Error: Permission denied

```bash
# Pastikan user memiliki permission DROP
# Login sebagai postgres superuser
psql -U postgres

# Grant permission
GRANT ALL PRIVILEGES ON DATABASE sistem_akuntans_dev TO your_user;
```

## 💡 Tips & Best Practices

### 1. Gunakan Database Terpisah untuk Testing

```env
# Development
DATABASE_URL=postgres://postgres:postgres@localhost/sistem_akuntans_dev?sslmode=disable

# Testing
DATABASE_URL=postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable
```

### 2. Backup Sebelum Reset

```bash
# Backup manual
pg_dump -U postgres sistem_akuntans_dev > backups/backup_$(date +%Y%m%d_%H%M%S).sql

# Restore jika perlu
psql -U postgres sistem_akuntans_dev < backups/backup_20231208_120000.sql
```

### 3. Quick Reset Workflow

```bash
# 1. Reset database
cd backend/cmd/reset_db && go run main.go

# 2. Start aplikasi (auto-migrate dan seed)
cd ../.. && go run main.go

# 3. Test di browser
# http://localhost:3000/login
```

### 4. Automated Testing Script

Buat script untuk automated testing:

```bash
#!/bin/bash
# test_workflow.sh

echo "Resetting database..."
cd backend/cmd/reset_db && go run main.go

echo "Starting application..."
cd ../.. && go run main.go &
APP_PID=$!

echo "Waiting for app to start..."
sleep 5

echo "Running tests..."
# Jalankan test Anda di sini

echo "Stopping application..."
kill $APP_PID
```

## 📚 File Structure

```
backend/
├── cmd/
│   └── reset_db/
│       ├── main.go                 # Main reset script
│       ├── README.md               # Detailed documentation
│       ├── Makefile                # Make commands
│       ├── reset_db.bat            # Windows script
│       ├── reset_db.sh             # Linux/Mac script
│       ├── backup_and_reset.bat    # Windows backup + reset
│       └── backup_and_reset.sh     # Linux/Mac backup + reset
├── .env                            # Database credentials
└── RESET_DATABASE.md               # This file
```

## 🔗 Related Commands

```bash
# Build reset tool
cd backend/cmd/reset_db
go build -o reset_db

# Run built executable
./reset_db

# Clean build
make clean

# Install dependencies
cd backend
go mod download
```

## ⚠️ Production Warning

**JANGAN PERNAH** jalankan reset di production database!

Jika benar-benar perlu reset production:
1. ✅ Backup database terlebih dahulu
2. ✅ Informasikan ke semua stakeholder
3. ✅ Jadwalkan downtime
4. ✅ Test restore procedure
5. ✅ Dokumentasikan semua langkah

## 📞 Support

Jika mengalami masalah:
1. Cek file `.env` sudah benar
2. Cek PostgreSQL service berjalan
3. Cek user database memiliki permission
4. Lihat log error untuk detail

---

**Happy Testing! 🚀**
