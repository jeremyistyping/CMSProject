# Master Data Implementation - Unipro

## Overview
Implementasi fitur Master Data untuk aplikasi Unipro yang mencakup Chart of Accounts (COA), Master Material, dan Master Vendor dengan integrasi penuh ke modul Purchase Request dan Cost Breakdown Structure (CBS).

## Fitur yang Ditambahkan

### 1. Chart of Accounts (COA)
**Backend:**
- Model: `backend/models/coa.go`
- Repository: `backend/repositories/coa_repository.go`
- Service: `backend/services/coa_service.go`
- Controller: `backend/controllers/coa_controller.go`

**Frontend:**
- Page: `frontend/app/master-data/coa/page.tsx`
- Component: `frontend/src/components/master-data/COAList.tsx`

**Fitur:**
- Hierarchical COA structure dengan parent-child relationship
- 5 tipe akun: Asset, Liability, Equity, Revenue, Expense
- Kategori khusus konstruksi: Material, Labor, Equipment, Overhead, Subcontractor
- Tree view dan list view
- Header account (tidak bisa digunakan untuk transaksi)
- Default COA untuk proyek konstruksi sudah di-seed otomatis

**API Endpoints:**
```
GET    /api/v1/master-data/coa
GET    /api/v1/master-data/coa/tree
GET    /api/v1/master-data/coa/type/:type
GET    /api/v1/master-data/coa/category/:category
GET    /api/v1/master-data/coa/:id
POST   /api/v1/master-data/coa
PUT    /api/v1/master-data/coa/:id
DELETE /api/v1/master-data/coa/:id
```

### 2. Master Material
**Backend:**
- Model: `backend/models/material.go`
- Repository: `backend/repositories/material_repository.go`
- Service: `backend/services/material_service.go`
- Controller: `backend/controllers/material_controller.go`

**Frontend:**
- Page: `frontend/app/master-data/materials/page.tsx`
- Component: `frontend/src/components/master-data/MaterialList.tsx`

**Fitur:**
- Material dengan kode unik, nama, deskripsi
- Kategori material (hierarchical)
- Unit of Measure (UoM) - 16 satuan default
- Harga satuan, min/max stock, current stock
- Link ke COA Account untuk kategorisasi biaya
- Summary statistics (total material, low stock, stock value)
- Filter by category, search, low stock alert

**Default Categories:**
- STR - Material Struktural (Besi, beton, semen)
- FIN - Material Finishing (Cat, keramik, plafon)
- MEP - Material MEP (Pipa, kabel, fitting)
- WOD - Material Kayu (Kayu, triplek)
- SAN - Sanitary (Closet, wastafel, kran)
- ELC - Elektrikal (Kabel, saklar, lampu)
- OTH - Lain-lain

**Default UoM:**
- Length: m, cm, mm
- Area: m²
- Volume: m³, liter
- Weight: kg, ton, zak
- Quantity: pcs, unit, set, batang, lembar, roll, dus

**API Endpoints:**
```
GET    /api/v1/master-data/materials
GET    /api/v1/master-data/materials/summary
GET    /api/v1/master-data/materials/:id
GET    /api/v1/master-data/materials/:id/vendors
POST   /api/v1/master-data/materials
PUT    /api/v1/master-data/materials/:id
DELETE /api/v1/master-data/materials/:id
GET    /api/v1/master-data/material-categories
GET    /api/v1/master-data/material-categories/tree
POST   /api/v1/master-data/material-categories
PUT    /api/v1/master-data/material-categories/:id
DELETE /api/v1/master-data/material-categories/:id
GET    /api/v1/master-data/uom
```

### 3. Master Vendor
**Backend:**
- Model: `backend/models/vendor.go`
- Repository: `backend/repositories/vendor_repository.go`
- Service: `backend/services/vendor_service.go`
- Controller: `backend/controllers/vendor_controller.go`

**Frontend:**
- Page: `frontend/app/master-data/vendors/page.tsx`
- Component: `frontend/src/components/master-data/VendorList.tsx`

**Fitur:**
- Vendor dengan kode unik, nama, contact person
- Informasi kontak lengkap (email, phone, address)
- Informasi pajak (NPWP)
- Informasi bank (nama bank, no rekening, cabang)
- Payment terms (termin pembayaran dalam hari)
- Rating vendor (0-5 stars)
- Kategori vendor
- Vendor-Material relationship (many-to-many)
- Summary statistics (total vendor, active vendor, average rating)

**Default Categories:**
- MAT - Supplier Material
- EQP - Rental Alat
- SUB - Subkontraktor
- SVC - Jasa
- OTH - Lainnya

**API Endpoints:**
```
GET    /api/v1/master-data/vendors
GET    /api/v1/master-data/vendors/summary
GET    /api/v1/master-data/vendors/:id
GET    /api/v1/master-data/vendors/:id/materials
POST   /api/v1/master-data/vendors
PUT    /api/v1/master-data/vendors/:id
DELETE /api/v1/master-data/vendors/:id
POST   /api/v1/master-data/vendors/:id/materials
DELETE /api/v1/master-data/vendors/:id/materials/:materialId
GET    /api/v1/master-data/vendor-categories
POST   /api/v1/master-data/vendor-categories
PUT    /api/v1/master-data/vendor-categories/:id
DELETE /api/v1/master-data/vendor-categories/:id
```

## Integrasi dengan Modul Lain

### Purchase Request
- Form PR sekarang bisa memilih Material dari master data
- Auto-fill item name, unit, dan estimated price dari material
- Bisa memilih Vendor untuk PR
- Field `material_id` ditambahkan ke `purchase_request_items`
- Field `vendor_id` sudah ada di `purchase_requests`

### Cost Breakdown Structure (CBS)
- CBS Node sekarang bisa di-link ke COA Account
- Field `coa_account_id` ditambahkan ke `cbs_nodes`
- Memudahkan kategorisasi biaya per CBS node

## Database Migration

Migration file: `backend/migrations/070_create_master_data_tables.sql`

**Tables Created:**
1. `coa_accounts` - Chart of Accounts
2. `material_categories` - Kategori material
3. `unit_of_measures` - Satuan ukuran
4. `materials` - Master material
5. `vendor_categories` - Kategori vendor
6. `vendors` - Master vendor
7. `vendor_materials` - Relasi vendor-material (many-to-many)

**Columns Added:**
- `purchase_request_items.material_id` - Link ke master material
- `cbs_nodes.coa_account_id` - Link ke COA (jika belum ada)

## Auto-Seeding

Saat aplikasi pertama kali dijalankan, data berikut akan di-seed otomatis:

1. **COA**: 27 akun default untuk proyek konstruksi
2. **Material Categories**: 7 kategori default
3. **Unit of Measure**: 16 satuan default
4. **Vendor Categories**: 5 kategori default

## Permissions & Access Control

**Roles yang bisa akses Master Data:**
- **ADMIN**: Full access ke semua master data
- **COST_CONTROL**: Full access ke COA, Material, Vendor
- **PURCHASING**: Access ke Material dan Vendor (tidak bisa COA)

## UI/UX Features

### COA List
- Tree view untuk melihat hierarchy
- List view untuk melihat flat list
- Filter by type (Asset, Liability, etc.)
- Search by code atau name
- Badge untuk status (Active/Inactive)
- Badge untuk tipe dan kategori

### Material List
- Summary cards (Total Material, Active, Low Stock, Stock Value)
- Filter by category
- Search by code atau name
- Toggle untuk show low stock only
- Alert icon untuk material dengan stok rendah
- Form dengan tabs untuk info lengkap

### Vendor List
- Summary cards (Total Vendor, Active, Categories, Average Rating)
- Filter by category
- Search by code, name, atau contact
- Star rating display
- Form dengan tabs (Info Dasar, Alamat, Bank & Pajak)
- Contact info display (phone, email)

## Testing

Untuk testing implementasi:

1. **Backend**: Jalankan `go run main.go` - auto-migration dan seeding akan berjalan
2. **Frontend**: Jalankan `npm run dev`
3. **Login**: Gunakan user dengan role ADMIN atau COST_CONTROL
4. **Navigate**: Sidebar > Master Data > COA/Material/Vendor

## Future Enhancements

Potential improvements:
1. Import/Export Excel untuk bulk data entry
2. Material stock movement tracking
3. Vendor performance tracking
4. Price history untuk material
5. Multi-currency support untuk vendor
6. Material substitution/alternative
7. Vendor approval workflow
8. Material requisition planning

## Notes

- Semua master data menggunakan soft delete (deleted_at)
- Validation untuk prevent delete jika data sedang digunakan
- Unique constraint pada code untuk semua master data
- Hierarchical structure support untuk COA dan Material Category
- Many-to-many relationship antara Vendor dan Material
