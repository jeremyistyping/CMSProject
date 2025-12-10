# CBS Seeder for Padel Bandung

Script untuk menambahkan Cost Breakdown Structure (CBS) dummy untuk project Padel Bandung.

## Struktur CBS yang Dibuat

### Level 1 - Main Categories (5 nodes)
1. **1.0 - Site Preparation & Foundation** (Rp 250 juta)
2. **2.0 - Court Construction** (Rp 800 juta)
3. **3.0 - Facilities & Amenities** (Rp 350 juta)
4. **4.0 - Equipment & Furnishing** (Rp 200 juta)
5. **5.0 - Utilities & Systems** (Rp 150 juta)

### Level 2 - Sub-categories (19 nodes)

#### 1.0 Site Preparation & Foundation
- 1.1 Land Clearing & Grading (Rp 50 juta)
- 1.2 Excavation & Earthwork (Rp 80 juta)
- 1.3 Foundation Work (Rp 120 juta)

#### 2.0 Court Construction
- 2.1 Court Surface & Base (Rp 300 juta)
- 2.2 Glass Walls (Rp 250 juta)
- 2.3 Metal Structure & Fencing (Rp 150 juta)
- 2.4 Court Lighting (Rp 100 juta)

#### 3.0 Facilities & Amenities
- 3.1 Clubhouse Building (Rp 150 juta)
- 3.2 Locker Rooms & Showers (Rp 80 juta)
- 3.3 Cafe & Lounge Area (Rp 70 juta)
- 3.4 Parking Area (Rp 50 juta)

#### 4.0 Equipment & Furnishing
- 4.1 Sports Equipment (Rp 50 juta)
- 4.2 Furniture & Fixtures (Rp 80 juta)
- 4.3 Audio Visual System (Rp 40 juta)
- 4.4 Security System (Rp 30 juta)

#### 5.0 Utilities & Systems
- 5.1 Electrical System (Rp 60 juta)
- 5.2 Plumbing & Water System (Rp 40 juta)
- 5.3 HVAC System (Rp 35 juta)
- 5.4 Internet & Network (Rp 15 juta)

**Total Budget: Rp 1.750.000.000 (1,75 Miliar)**

## Cara Menjalankan

### Dari direktori backend:
```bash
go run cmd/seed_cbs/main.go
```

### Atau compile dulu:
```bash
go build -o seed_cbs.exe cmd/seed_cbs/main.go
./seed_cbs.exe
```

## Catatan

- Script akan otomatis skip jika CBS sudah ada untuk project Padel Bandung
- Pastikan project "Padel Bandung" sudah ada di database
- Script menggunakan konfigurasi database dari file `.env`

## Verifikasi

Setelah menjalankan seeder, verifikasi dengan query:

```sql
SELECT code, name, budget_amount, parent_id 
FROM cbs_nodes 
WHERE project_id = (SELECT id FROM projects WHERE project_name = 'Padel Bandung')
ORDER BY code;
```

Atau cek di frontend:
1. Login ke aplikasi
2. Buka menu Cost Control > Cost Breakdown Structure
3. Pilih project "Padel Bandung"
4. Lihat struktur CBS yang sudah dibuat
