# Konfigurasi

Dokumen ini menjelaskan konfigurasi environment dan accounting_config.json.

## Environment Variables (Backend)
- DB_HOST, DB_PORT, DB_USER, DB_PASS, DB_NAME
- JWT_SECRET, GIN_MODE (release untuk produksi)
- Pengaturan lainnya sesuai kebutuhan middleware/security

## Environment Variables (Frontend)
- NEXT_PUBLIC_API_URL, contoh: `http://localhost:8080/api/v1`

## accounting_config.json
- default_accounts: mapping akun utama (kas, bank, AR, AP, PPN, pendapatan, beban, dll)
- tax_rates: PPN/PPh default & type-based rates
- currency_settings: base currency, simbol, decimal places
- journal_settings: prefix, auto code, balanced entry, future date policy
- period_settings: awal tahun fiskal, aturan close period
- audit_settings: audit trail, retensi, sensitive fields

Catatan penting:
- Selaraskan mapping PPN keluaran/masukan pada helper jurnal agar konsisten dengan file ini.
- Sesuaikan `allow_post_to_old_period`, `max_old_period_months` sesuai kebijakan perusahaan.
