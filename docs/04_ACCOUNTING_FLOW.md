# Flow Akuntansi (Double-Entry)

Dokumen ini merinci jurnal otomatis, aturan validasi, dan kontrol periode.

## Prinsip Umum
- Setiap transaksi menghasilkan jurnal seimbang: total debit = total kredit > 0
- Validasi: akun aktif & bukan header, periode terbuka, tanggal tidak terlalu jauh ke depan
- Kode jurnal: otomatis berdasarkan tanggal (JE-YYYY-MM-DD-####)

## Mapping Akun Utama (Default)
- Kas: 1101
- Bank: 1102
- Piutang Usaha: 1201
- Persediaan: 1301
- PPN Masukan (klaim): 1105
- Hutang Usaha: 2001
- PPN Keluaran (kewajiban): 2102 (catatan: sebagian helper historis menggunakan 2101)
- Pendapatan: 4101
- Beban Pokok/Pembelian/Biaya: 5101/6001 sesuai item

Selaraskan mapping ini di `backend/config/accounting_config.json` dan pastikan helper otomatis konsisten.

## Penjualan (Invoice)
- Debit: Piutang (1201) sebesar Total Invoice (jika belum dibayar)
- Debit: Kas/Bank (1101/1102) sebesar pembayaran yang diterima saat penjualan
- Kredit: Pendapatan (4xxx) per item (nilai sebelum PPN)
- Kredit: PPN Keluaran (2102) sebesar PPN

## Pembayaran Penjualan
- Debit: Kas/Bank (1101/1102) sebesar jumlah bayar
- Kredit: Piutang (1201) sebesar jumlah bayar

## Pembelian (Kredit)
- Debit: Persediaan/Beban (1xxx/6xxx) per item (nilai sebelum PPN)
- Debit: PPN Masukan (1105) jika ada
- Kredit: Hutang (2001) sebesar total tagihan belum dibayar

## Pembelian (Tunai/Transfer)
- Debit: Persediaan/Beban (1xxx/6xxx)
- Debit: PPN Masukan (1105) jika ada
- Kredit: Kas/Bank (1101/1102) sebesar jumlah dibayar

## Pembelian Aset
- Debit: Akun Aset Tetap (15xx) sebesar harga perolehan
- Kredit: Kas/Bank (1101/1102) atau Hutang (2001) sesuai metode bayar

## Depresiasi Aset
- Debit: Beban Depresiasi (6201)
- Kredit: Akumulasi Depresiasi (1502)

## Kontrol & Validasi
- Periode:
  - Tidak boleh posting ke periode yang ditutup
  - Batas tanggal masa depan (default max 7 hari)
  - Batas umur transaksi historis (default 2 tahun, dapat diubah)
- Validasi baris jurnal:
  - Tidak boleh debit & kredit bersamaan pada satu baris
  - Nilai tidak boleh negatif; salah satu sisi wajib > 0

Untuk pemeriksaan konsistensi menyeluruh, gunakan panduan di 10_BALANCE_MONITORING.md.
