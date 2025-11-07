# Reporting & Analytics

Dokumen ini menjelaskan laporan keuangan dan cara mengaksesnya.

## Laporan Keuangan Utama
- Balance Sheet (Neraca)
- Profit & Loss (Laba Rugi)
- Cash Flow (Arus Kas)
- Trial Balance (Neraca Saldo)
- General Ledger (Buku Besar per akun)

## Akses via API
- GET /api/v1/enhanced-reports/balance-sheet
- GET /api/v1/enhanced-reports/profit-loss
- GET /api/v1/enhanced-reports/cash-flow
- POST /api/v1/financial-reports/trial-balance
- GET /api/v1/financial-reports/general-ledger/{account_id}

## Dashboard & Analitik
- Sales, Purchase, Inventory, Financial metrics
- Export PDF/Excel dengan formatting profesional

## Tips Validasi
- Pastikan semua jurnal POSTED untuk periode laporan
- Gunakan balance monitoring untuk deteksi selisih lebih awal
