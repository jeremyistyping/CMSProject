# Flow Bisnis

Dokumen ini memetakan alur bisnis inti: Sales, Purchase, Inventory, Cash/Bank, dan Asset.

## Sales Management
- Tahapan: QUOTATION → ORDER → INVOICE → PAYMENT
- Poin penting:
  - Diskon multi-level, PPN, PPh, ongkir
  - Partial payment didukung
  - PDF invoice & analitik penjualan
- Output akuntansi: AR (piutang), revenue, PPN keluaran, kas/bank (jika dibayar)

## Purchase Management
- Tahapan: Request → Approval → Order/Invoice → Receipt (opsional) → Payment
- Poin penting:
  - Approval multi-level (threshold amount)
  - Three-way matching (PO-Receipt-Invoice) opsional
  - Metode bayar: CASH, CREDIT, TRANSFER
- Output akuntansi: Expense/Inventory, PPN masukan, AP (hutang), kas/bank

## Inventory Control
- Real-time movements, opname, adjust, low stock alerts
- Valuasi: FIFO/LIFO/Average (disesuaikan)
- Integrasi:
  - Purchase Receipt → stok bertambah
  - Sales Confirmed → stok berkurang

## Cash & Bank
- Multi akun kas/bank, transfer, rekonsiliasi
- Dashboard pembayaran & arus kas

## Asset Management
- Siklus: purchase → capitalization → depreciation → disposal
- Dokumen lampiran & jadwal depresiasi

## Monitoring & Notifications
- Audit log menyeluruh, security incidents, balance monitoring
- Notifikasi kontekstual (threshold, event-driven)

Lihat 04_ACCOUNTING_FLOW.md untuk detail jurnal otomatis dari setiap alur.
