# Balance Monitoring & Verifikasi

Dokumen ini memandu cara memantau dan memverifikasi keseimbangan pembukuan.

## Endpoint Monitoring
- GET /api/v1/balance-monitor/status
- POST /api/v1/balance-monitor/sync
- GET /api/v1/balance-monitor/anomalies

## Verifikasi via Tools
- `backend/tools/verify_accounting.go` melakukan:
  - List transaksi purchase & status approval
  - Tampilkan jurnal per referensi dan cek balance
  - Tampilkan saldo akun (Assets, Liabilities, Equity)
  - Verifikasi persamaan: Assets = Liabilities + Equity
  - Validasi PPN per transaksi

## Prosedur Rutin
1) Jalankan sinkronisasi balance bila perlu
2) Tinjau anomali dari endpoint
3) Jalankan tool verifikasi untuk audit menyeluruh
4) Perbaiki jurnal/konfigurasi yang tidak konsisten
