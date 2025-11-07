# Settings History Table Migration

## Overview
File `038_create_settings_history.sql` membuat tabel untuk mencatat audit trail perubahan settings sistem.

## Auto Migration
Tabel ini akan **otomatis dibuat** saat aplikasi backend dijalankan pertama kali melalui sistem auto-migration.

## Cara Kerja

### 1. Git Pull di PC Lain
```bash
git pull origin main
```

### 2. Jalankan Backend
```bash
cd backend
go run main.go
```

### 3. Auto Migration Berjalan
Sistem akan:
- ✅ Membaca file `migrations/038_create_settings_history.sql`
- ✅ Memeriksa apakah tabel `settings_history` sudah ada
- ✅ Jika belum ada, otomatis membuat tabel beserta indexes
- ✅ Mencatat status migrasi di tabel `migration_logs`

## Struktur Tabel

```sql
CREATE TABLE settings_history (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    -- Reference
    settings_id BIGINT NOT NULL,
    
    -- Change tracking
    field VARCHAR(255) NOT NULL,
    old_value TEXT,
    new_value TEXT,
    action VARCHAR(50) DEFAULT 'UPDATE',
    
    -- User tracking
    changed_by BIGINT NOT NULL,
    
    -- Context
    ip_address VARCHAR(255),
    user_agent TEXT,
    reason TEXT,
    
    CONSTRAINT fk_settings_history_settings 
        FOREIGN KEY (settings_id) REFERENCES settings(id) ON DELETE CASCADE
);
```

## Fitur

### Audit Trail
Setiap perubahan pada settings dicatat dengan informasi:
- Field yang berubah
- Nilai lama dan baru
- User yang melakukan perubahan
- Timestamp perubahan
- IP Address dan User Agent (opsional)
- Alasan perubahan (opsional)

### Contoh Penggunaan
```go
// Service otomatis mencatat perubahan
service.UpdateSettings(settingsID, updates, userID)
// Log akan tercatat di settings_history

// Query history
history, err := service.GetSettingsHistory(filter)
```

## Verifikasi

### Cek Tabel Sudah Ada
```sql
SELECT EXISTS (
    SELECT 1 FROM information_schema.tables 
    WHERE table_name = 'settings_history'
);
```

### Cek Status Migrasi
```sql
SELECT * FROM migration_logs 
WHERE migration_name = '038_create_settings_history.sql';
```

### Cek Indexes
```sql
SELECT indexname, indexdef 
FROM pg_indexes 
WHERE tablename = 'settings_history';
```

## Manual Migration (Jika Diperlukan)

Jika auto-migration gagal, jalankan manual:

```bash
cd backend
go run cmd/create_settings_history_table.go
```

Atau jalankan SQL langsung:
```bash
psql -U postgres -d sistem_akuntansi -f migrations/038_create_settings_history.sql
```

## Troubleshooting

### Error: "relation settings_history does not exist"
**Solusi:**
1. Restart backend untuk trigger auto-migration
2. Atau jalankan manual migration script
3. Check database connection string di .env

### Error: "foreign key constraint fails"
**Penyebab:** Tabel `settings` belum ada
**Solusi:**
```sql
-- Cek apakah tabel settings ada
SELECT * FROM information_schema.tables WHERE table_name = 'settings';

-- Jika tidak ada, jalankan migrasi settings terlebih dahulu
```

### Auto Migration Tidak Berjalan
**Periksa:**
1. File migrasi ada di `backend/migrations/`
2. File berformat `.sql`
3. Log startup backend untuk error messages
4. Database connection berhasil

## Best Practices

### 1. Selalu Git Pull Sebelum Develop
```bash
git pull origin main
go run main.go  # Auto-migration akan berjalan
```

### 2. Jangan Edit Tabel Manual
Biarkan auto-migration yang mengelola schema

### 3. Monitor Migration Logs
```sql
SELECT * FROM migration_logs 
ORDER BY executed_at DESC 
LIMIT 10;
```

### 4. Backup Database Sebelum Migration Besar
```bash
pg_dump -U postgres sistem_akuntansi > backup_$(date +%Y%m%d).sql
```

## FAQ

**Q: Apakah safe untuk menjalankan migration berkali-kali?**
A: Ya, migration menggunakan `CREATE TABLE IF NOT EXISTS` sehingga idempotent.

**Q: Bagaimana jika migrasi gagal di tengah jalan?**
A: Auto-migration akan skip migrasi yang sudah berhasil (dicek dari `migration_logs`) dan hanya retry yang gagal.

**Q: Apakah data existing akan hilang?**
A: Tidak, migration hanya CREATE TABLE baru, tidak mengubah data existing.

**Q: Bagaimana cara rollback migration?**
A: Drop table manual jika diperlukan:
```sql
DROP TABLE IF EXISTS settings_history CASCADE;
DELETE FROM migration_logs WHERE migration_name = '038_create_settings_history.sql';
```

## Support

Jika mengalami masalah:
1. Cek log di console backend
2. Cek tabel `migration_logs` untuk status detail
3. Jalankan manual migration script
4. Contact team lead jika masalah persist
