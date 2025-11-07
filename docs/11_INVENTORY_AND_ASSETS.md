# Inventory & Asset Management

## Inventory
- Master Produk & Kategori
- Pergerakan Stok: penerimaan (purchase), pengeluaran (sales), transfer, adjust, opname
- Valuasi (konfigurasi): FIFO/LIFO/Average
- Monitoring: low stock alerts, laporan pergerakan
- Endpoint contoh: /api/v1/products, /inventory/movements, /products/adjust-stock, /products/opname

## Asset Management
- Siklus Aset: pembelian → kapitalisasi → depresiasi → pelepasan
- Depresiasi: penjadwalan otomatis, jurnal debit beban (6201) kredit akumulasi (1502)
- Dokumen: lampiran foto/dokumen aset, riwayat perawatan (opsional)
- Endpoint contoh: asset controller (backend), laporan terkait di reporting

## Integrasi ke Akuntansi
- Purchase → Persediaan/Aset bertambah (debit)
- Sales → Persediaan berkurang & COGS (jika diaktifkan) — sesuaikan modul COGS Anda
- Depresiasi → pengakuan biaya periodik sesuai metode
