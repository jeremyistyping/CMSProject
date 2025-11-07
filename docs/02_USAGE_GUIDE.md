# Panduan Penggunaan (Usage Guide)

Panduan ini menjelaskan cara menjalankan aplikasi (backend & frontend) dan skenario penggunaan umum.

## Prasyarat
- Node.js 18+
- Go 1.23+
- PostgreSQL 12+ (disarankan)
- Git

## Menjalankan Backend
1. Masuk ke folder backend
2. Install dependency: `go mod tidy`
3. Siapkan database (contoh PostgreSQL): `createdb sistem_akuntansi`
4. Salin env: `cp .env.example .env` dan isi DB_HOST, DB_PORT, DB_USER, DB_PASS, DB_NAME, JWT_SECRET
5. Jalankan Balance Protection (WAJIB untuk PC baru)
6. Run server: `go run cmd/main.go`

Server otomatis migrasi schema, seed data awal, inisialisasi monitoring. API: `http://localhost:8080`.

## Menjalankan Frontend
1. Masuk ke folder frontend
2. Install dependency: `npm install`
3. Set `NEXT_PUBLIC_API_URL` (opsional) di `.env.local` ke `http://localhost:8080/api/v1`
4. Jalankan: `npm run dev` → `http://localhost:3000`

## Login Default
- admin@company.com / password123
- finance@company.com / password123
- director@company.com / password123
- employee@company.com / password123
- inventory@company.com / password123
- auditor@company.com / password123

## Alur Dasar (Contoh Cepat)
### Penjualan
1) Buat Sales (Quotation/Order/Invoice) → Confirm → Generate Invoice
2) Sistem hitung diskon, PPN, PPh (jika ada)
3) Record Payment (parsial atau penuh)
4) Jurnal otomatis: Piutang/Kas vs Pendapatan dan PPN Keluaran
5) Cek laporan (P&L, AR aging, dsb)

### Pembelian
1) Buat Purchase → Submit approval (jika perlu) → Approve
2) Terima barang (Receipt) → Matching/Status
3) Pembayaran (Cash/Transfer/Kredit)
4) Jurnal otomatis: Beban/Persediaan + PPN Masukan vs Hutang/Kas
5) Cek laporan (AP aging, Cash Flow, dsb)

### Inventory
- Update stok via penerimaan/penjualan, opname, atau penyesuaian
- Pantau low stock alerts dan laporan pergerakan stok

### Asset & Depresiasi
- Catat pembelian aset → jurnal otomatis
- Jalankan depresiasi periodik → jurnal otomatis

Lanjutkan ke 03_BUSINESS_FLOW.md untuk detail flow bisnis per modul.
