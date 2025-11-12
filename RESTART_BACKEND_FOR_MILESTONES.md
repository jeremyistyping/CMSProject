# 🔄 Restart Backend untuk Apply Milestone Migration

## ✅ Fix yang Sudah Diterapkan

1. ✅ Menambahkan `&models.Milestone{}` ke AutoMigrate di `backend/database/init.go`
2. ✅ Membuat SQL migration `backend/migrations/051_create_milestones_table.sql`
3. ✅ Kode berhasil compile tanpa error

## 📋 Langkah-langkah Restart Backend

### Opsi 1: Restart Manual (Recommended)

1. **Stop Backend Server yang sedang Running**
   - Tekan `Ctrl+C` di terminal yang menjalankan backend
   - Atau tutup terminal/command prompt yang menjalankan backend

2. **Buka Terminal Baru di Backend Directory**
   ```bash
   cd "C:\Users\jeremia.kaligis\Desktop\CMS New\backend"
   ```

3. **Jalankan Backend Server**
   ```bash
   go run main.go
   ```

4. **Perhatikan Log Startup**
   - Cari baris yang menunjukkan "Database migrations completed successfully"
   - Pastikan tidak ada error terkait milestone
   - Tabel `milestones` akan dibuat otomatis

### Opsi 2: Restart dengan Build Binary

```bash
cd "C:\Users\jeremia.kaligis\Desktop\CMS New\backend"
go build -o cms-backend.exe .
.\cms-backend.exe
```

## ✅ Verifikasi Setelah Restart

### 1. Cek Log Backend
Pastikan ada baris seperti ini di log:
```
Running database migrations...
Database migrations completed successfully
```

### 2. Test API Endpoint

**Menggunakan Browser atau Postman:**
1. Login dulu untuk dapatkan token:
   ```
   POST http://localhost:8080/api/v1/auth/login
   Body: {
     "email": "admin@company.com",
     "password": "password123"
   }
   ```

2. Test milestone endpoint:
   ```
   GET http://localhost:8080/api/v1/projects/1/milestones
   Headers: Authorization: Bearer <your-token>
   ```

**Atau gunakan test script:**
```powershell
cd backend
powershell -ExecutionPolicy Bypass -File test_milestone_api.ps1
```

### 3. Test di Frontend
1. Buka browser: http://localhost:3000/projects/1
2. Klik tab "Milestones"
3. Seharusnya tidak ada error "Failed to fetch milestones" lagi
4. Coba klik "Add Milestone" dan buat milestone baru

## 🔍 Troubleshooting

### Jika masih error 500:
1. **Cek apakah tabel sudah dibuat:**
   ```sql
   -- Connect ke database
   psql -U postgres -d CMSNew
   
   -- Check tabel milestones
   \dt milestones
   
   -- Lihat struktur tabel
   \d milestones
   ```

2. **Jika tabel belum ada, jalankan migration manual:**
   ```bash
   cd backend
   psql -U postgres -d CMSNew -f migrations/051_create_milestones_table.sql
   ```

3. **Restart backend lagi setelah migration manual**

### Jika tabel sudah ada tapi masih error:
1. Cek log backend untuk error spesifik
2. Pastikan tidak ada typo di nama kolom
3. Verify bahwa GORM AutoMigrate berjalan

### Jika error "column does not exist":
Kemungkinan ada mismatch antara struct dan database. Jalankan:
```sql
-- Hapus tabel (HATI-HATI: ini akan menghapus data!)
DROP TABLE IF EXISTS milestones CASCADE;

-- Restart backend (akan create ulang dengan struktur yang benar)
```

## 📊 Expected Result

Setelah restart berhasil, Anda harus bisa:
- ✅ Melihat tab Milestones tanpa error
- ✅ Menambah milestone baru
- ✅ Edit milestone
- ✅ Delete milestone  
- ✅ Mark milestone as complete
- ✅ Switch antara Grid dan Timeline view
- ✅ Search dan filter milestones

## 🎯 Quick Test Checklist

- [ ] Backend restart tanpa error
- [ ] Log menunjukkan "Database migrations completed"
- [ ] Endpoint `/api/v1/projects/1/milestones` return 200 (bukan 404 atau 500)
- [ ] Frontend tab Milestones terbuka tanpa error popup
- [ ] Bisa create milestone baru
- [ ] Milestone muncul di list
- [ ] Grid view dan Timeline view keduanya bekerja

---

**Estimated Time:** 2-3 menit  
**Risk Level:** Rendah (hanya restart server, tidak ada perubahan data)

