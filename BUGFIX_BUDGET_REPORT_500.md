# Bugfix: Budget Report 500 Error - Missing project_budgets Table

## Masalah

Error 500 saat mengakses Budget vs Actual Report:
```
ERROR: column pb.account_id does not exist (SQLSTATE 42703)
```

### Root Cause

Tabel `project_budgets` tidak pernah dibuat di database. Query SQL di `expense_transaction_repository.go` mencoba mengakses tabel ini tapi tabel tidak ada.

### Error Log
```
2025/12/08 01:59:38 ERROR: column pb.account_id does not exist (SQLSTATE 42703)
SELECT
    pb.project_id,
    pb.account_id as coa_account_id,
    ...
FROM project_budgets pb
...
```

## Solusi

### ✅ OTOMATIS - Restart Backend (RECOMMENDED)

Tabel `project_budgets` sekarang akan **otomatis dibuat** saat backend dijalankan melalui AutoMigration.

```bash
cd backend
go run main.go
```

Backend akan:
1. Cek apakah tabel `project_budgets` sudah ada
2. Jika belum ada, otomatis membuat tabel dengan struktur lengkap
3. Membuat indexes untuk performa
4. Membuat trigger untuk auto-update `updated_at`

Log yang akan muncul:
```
🔄 Creating project_budgets table...
✅ project_budgets table created successfully
```

### Opsi Manual (Jika Diperlukan)

Jika ingin membuat tabel secara manual tanpa restart backend:

```bash
# Masuk ke direktori migrations
cd backend/migrations

# Jalankan migration
psql -U postgres -d db_unipro -f run_074_manually.sql
```

Atau copy-paste SQL berikut langsung ke pgAdmin atau psql:

```sql
-- Create project_budgets table
CREATE TABLE IF NOT EXISTS project_budgets (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    account_id INTEGER NOT NULL REFERENCES coa_accounts(id) ON DELETE RESTRICT,
    estimated_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    
    CONSTRAINT project_budgets_unique_project_account UNIQUE (project_id, account_id),
    CONSTRAINT project_budgets_positive_amount CHECK (estimated_amount >= 0)
);

CREATE INDEX IF NOT EXISTS idx_project_budgets_project_id ON project_budgets(project_id);
CREATE INDEX IF NOT EXISTS idx_project_budgets_account_id ON project_budgets(account_id);
CREATE INDEX IF NOT EXISTS idx_project_budgets_deleted_at ON project_budgets(deleted_at);

CREATE OR REPLACE FUNCTION update_project_budgets_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_project_budgets_updated_at
    BEFORE UPDATE ON project_budgets
    FOR EACH ROW
    EXECUTE FUNCTION update_project_budgets_updated_at();
```

## Verifikasi

Setelah menjalankan migration, verifikasi tabel sudah dibuat:

```sql
-- Cek struktur tabel
\d project_budgets

-- Atau
SELECT column_name, data_type 
FROM information_schema.columns 
WHERE table_name = 'project_budgets';
```

Expected output:
```
 column_name     | data_type
-----------------+------------------------
 id              | integer
 project_id      | integer
 account_id      | integer
 estimated_amount| numeric
 created_at      | timestamp without time zone
 updated_at      | timestamp without time zone
 deleted_at      | timestamp without time zone
```

## Testing

Setelah tabel dibuat, test Budget Report:

1. Buka frontend: http://localhost:3000/cost-control/budget-vs-actual
2. Pilih project (contoh: Project ID 6)
3. Pilih date range
4. Klik "Generate Report"
5. Seharusnya tidak ada error 500 lagi

**CATATAN**: Report mungkin kosong jika belum ada data budget. Anda perlu menambahkan budget data terlebih dahulu melalui API atau UI.

## Files Changed

### Modified:
1. `backend/database/auto_migrations.go` - **Ditambahkan fungsi `ensureProjectBudgetsTable()`** untuk auto-create tabel saat backend start

### Created:
1. `backend/migrations/074_create_project_budgets_table.up.sql` - Migration untuk create table
2. `backend/migrations/074_create_project_budgets_table.down.sql` - Migration untuk rollback
3. `backend/migrations/run_074_manually.sql` - Script untuk manual execution
4. `BUGFIX_BUDGET_REPORT_500.md` - Dokumentasi ini

### Existing Files (No Changes):
- `backend/repositories/expense_transaction_repository.go` - Query sudah benar
- `backend/models/project_budget.go` - Model sudah benar
- `backend/services/expense_transaction_service.go` - Service sudah benar
- `backend/database/init.go` - Model ProjectBudget sudah ada di AutoMigrate

## Next Steps

Setelah tabel dibuat, Anda perlu:

1. **Seed Budget Data**: Tambahkan data budget untuk project yang ada
   ```sql
   -- Contoh: Tambah budget untuk project 6
   INSERT INTO project_budgets (project_id, account_id, estimated_amount)
   VALUES 
       (6, 1, 50000000),  -- Ganti dengan COA account ID yang sesuai
       (6, 2, 30000000);
   ```

2. **Atau gunakan API**: Buat endpoint untuk manage project budgets
   - Endpoint sudah ada di `backend/controllers/project_budget_controller.go`
   - Route: `/api/v1/projects/:id/budgets`

3. **Test Budget Report**: Setelah ada data budget, test report lagi

## Technical Details

### Tabel Structure

```sql
project_budgets
├── id (SERIAL PRIMARY KEY)
├── project_id (FK to projects)
├── account_id (FK to coa_accounts)
├── estimated_amount (DECIMAL(15,2))
├── created_at (TIMESTAMP)
├── updated_at (TIMESTAMP)
└── deleted_at (TIMESTAMP) -- Soft delete
```

### Relationships

- `project_id` → `projects.id` (CASCADE DELETE)
- `account_id` → `coa_accounts.id` (RESTRICT DELETE)

### Constraints

- UNIQUE: `(project_id, account_id)` - Satu project hanya bisa punya satu budget per COA account
- CHECK: `estimated_amount >= 0` - Budget tidak boleh negatif

### Indexes

- `idx_project_budgets_project_id` - Query by project
- `idx_project_budgets_account_id` - Query by COA account
- `idx_project_budgets_deleted_at` - Soft delete queries

## Status

✅ Migration files created
✅ AutoMigration function added to `auto_migrations.go`
✅ Tabel `project_budgets` sudah t backend restart
⏳ Waiting for backend restart
⏳ Waiting for budget data seeding
⏳ Waiting for testing

## Author

Kiro AI Assistant
Date: 2025-12-08
