# Role Pengguna & User Flow

Dokumen ini menjelaskan peran (RBAC), kemampuan, dan user journey umum.

## Peran Utama
- Admin: akses penuh, user/role management, security & settings
- Director: dashboard eksekutif, laporan, approval bernilai tinggi
- Finance / Finance Manager: transaksi keuangan, jurnal, posting/reversal, rekonsiliasi, laporan, approval pembelian
- Inventory Manager: produk, stok, opname/adjust, peringatan stok
- Employee / Operational: input transaksi dasar sesuai izin
- Auditor: read-only audit trail, jurnal, laporan (tanpa mutasi)

## User Journey (Contoh)
- Finance:
  1) Pastikan periode terbuka & COA siap
  2) Input transaksi penjualan/pembelian
  3) Review jurnal otomatis → posting
  4) Rekonsiliasi kas/bank & monitoring piutang/hutang
  5) Jalankan laporan keuangan
- Inventory Manager:
  1) Kelola master produk & kategori
  2) Pantau & update stok lewat receipt/penjualan
  3) Opname & penyesuaian stok
- Director:
  1) Lihat KPI & ringkasan
  2) Lakukan approval pembelian besar
  3) Review laporan periodik

## Keamanan & Audit (Ringkas)
- JWT + Refresh token, blacklisted tokens
- Rate limiting & session tracking
- Audit log semua aktivitas penting

Detail keamanan & deployment: 09_SECURITY_AND_DEPLOYMENT.md
