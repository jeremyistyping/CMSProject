# Database Reset Tool

Script Go untuk mereset database aplikasi Unipro Cost Control Management System.

## 🎯 Fungsi

Script ini akan:
1. ✅ Membaca kredensial database dari file `.env`
2. 🗑️ Menghapus semua tabel di database
3. 🗑️ Menghapus semua sequences
4. 🗑️ Menghapus semua custom types (enums)
5. 🔄 Mempersiapkan database untuk migration ulang

## ⚠️ Peringatan

**SCRIPT INI AKAN MENGHAPUS SEMUA DATA DI DATABASE!**

Gunakan dengan hati-hati, terutama di environment production.

## 📋 Prasyarat

1. File `.env` sudah dikonfigurasi dengan benar di folder `backend/`
2. Database PostgreSQL sudah berjalan
3. Go sudah terinstall

## 🚀 Cara Penggunaan

### Opsi 1: Dari folder backend/cmd/reset_db

```bash
cd backend/cmd/reset_db
go run main.go
```

### Opsi 2: Dari folder backend

```bash
cd backend
go run cmd/reset_db/main.go
```

### Opsi 3: Build dan jalankan

```bash
cd backend/cmd/reset_db
go build -o reset_db.exe
./reset_db.exe
```

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
   ...
✅ All tables dropped successfully

🗑️  Step 2: Dropping all sequences...
   📋 Found 30 sequences to drop
   ✓ Dropped sequence: users_id_seq
   ...
✅ All sequences dropped successfully

🗑️  Step 3: Dropping all custom types...
   📋 Found 5 custom types to drop
   ✓ Dropped type: approval_status
   ...
✅ All custom types dropped successfully

✅ Database reset completed successfully!
📝 Next steps:
   1. Run your application to apply migrations
   2. Or run migrations manually using golang-migrate

💡 To start fresh, run: go run main.go
```

## 🔄 Setelah Reset

Setelah database direset, jalankan aplikasi utama untuk menerapkan migrations:

```bash
cd backend
go run main.go
```

Aplikasi akan otomatis:
1. Menjalankan semua SQL migrations
2. Menjalankan GORM AutoMigrate
3. Seed data awal (users, roles, permissions, master data)

## 🔐 Konfigurasi Database

Script ini membaca konfigurasi dari file `.env`:

```env
DATABASE_URL=postgres://username:password@localhost/database_name?sslmode=disable
```

### Contoh untuk berbagai environment:

**Development:**
```env
DATABASE_URL=postgres://postgres:postgres@localhost/sistem_akuntans_dev?sslmode=disable
```

**Testing:**
```env
DATABASE_URL=postgres://test_user:test_pass@localhost/sistem_akuntans_test?sslmode=disable
```

**Production (HATI-HATI!):**
```env
DATABASE_URL=postgres://prod_user:secure_password@prod-server.com/sistem_akuntans_prod?sslmode=require
```

## 🛡️ Keamanan

- Password di database URL akan di-mask saat ditampilkan di log
- Script meminta konfirmasi dengan mengetik 'YES' sebelum menghapus data
- Tidak ada hardcoded credentials

## 🐛 Troubleshooting

### Error: DATABASE_URL not found
```bash
# Pastikan file .env ada di folder backend/
ls backend/.env

# Atau set environment variable manual
export DATABASE_URL="postgres://user:pass@localhost/dbname?sslmode=disable"
```

### Error: Failed to connect to database
```bash
# Cek apakah PostgreSQL berjalan
psql -U postgres -c "SELECT version();"

# Cek koneksi
psql "postgres://user:pass@localhost/dbname"
```

### Error: Permission denied
```bash
# Pastikan user database memiliki permission untuk DROP
# Login sebagai superuser dan grant permission
```

## 📚 Dependencies

Script ini menggunakan:
- `github.com/joho/godotenv` - Untuk membaca .env file
- `github.com/lib/pq` - PostgreSQL driver untuk Go

Install dependencies:
```bash
cd backend
go mod download
```

## 💡 Tips

1. **Backup sebelum reset**: Selalu backup database sebelum reset
   ```bash
   pg_dump -U postgres sistem_akuntans_dev > backup.sql
   ```

2. **Testing environment**: Gunakan database terpisah untuk testing
   ```env
   DATABASE_URL=postgres://postgres:postgres@localhost/sistem_akuntans_test?sslmode=disable
   ```

3. **Quick reset workflow**:
   ```bash
   # Reset database
   cd backend/cmd/reset_db && go run main.go
   
   # Start aplikasi (akan auto-migrate dan seed)
   cd ../.. && go run main.go
   ```

## 📞 Support

Jika ada masalah, cek:
1. File `.env` sudah benar
2. PostgreSQL service berjalan
3. User database memiliki permission yang cukup
4. Database name sudah benar

---

**⚠️ INGAT: Jangan jalankan di production tanpa backup!**
