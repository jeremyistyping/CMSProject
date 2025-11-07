# Overview

Dokumen ini merangkum arsitektur, modul utama, dan teknologi yang digunakan.

## Arsitektur Tingkat Tinggi
- Backend: Go (Gin, GORM), PostgreSQL/MySQL, JWT, RBAC, Audit Log, Balance Monitoring
- Frontend: Next.js 15 (TypeScript), Chakra UI + Tailwind, Dark/Light Theme, i18n (ID/EN)
- Integrasi: Export PDF/Excel, reporting engine, monitoring & notifications

## Struktur Direktori (Inti)
- backend/
  - controllers/: HTTP handlers (sales, purchase, inventory, reports, security)
  - models/: Struktur data dan aturan bisnis (Sale, Purchase, JournalEntry, dll)
  - services/, repositories/, middleware/, routes/: arsitektur clean
  - config/: konfigurasi aplikasi & Swagger updater
  - docs/: dokumentasi teknis backend
  - tools/: utilitas verifikasi & pengujian akuntansi
- frontend/
  - app/, src/components, src/services: UI dan integrasi API

## Teknologi Utama
- Backend: Go 1.23+, Gin, GORM, jwt-go, excelize, gofpdf
- Frontend: Next.js 15, Turbopack, React Context, Axios, Chakra UI, Tailwind
- Database: PostgreSQL (disarankan), MySQL (opsional)

## Modul Bisnis
- Sales Management: Quotation → Order → Invoice → Payment
- Purchase Management: Request → Approval → Order/Receipt → Payment
- Inventory Control: movements, opname, alerts, costing
- Cash & Bank: multi-account, transfer, reconciliation
- Asset Management: purchase, depreciation, disposal
- Reporting: Financial statements, dashboard, ratios
- Security & Monitoring: auth, audit, incidents, balance monitoring

## Prinsip Akuntansi
- Double-entry: setiap transaksi menghasilkan jurnal seimbang (debit=credit)
- Validasi periode: tidak boleh posting ke periode yang ditutup
- Validasi akun: tidak boleh posting ke akun header/nonaktif

Lanjutkan ke 02_USAGE_GUIDE.md untuk menjalankan aplikasi dan contoh alur penggunaan.
