# Chart of Accounts (COA) Requirements untuk Sales Module

## Overview
Sales module memerlukan beberapa account COA spesifik untuk mencatat transaksi penjualan dengan benar. Semua account ini **sudah otomatis dibuat** saat backend pertama kali dijalankan melalui seeding.

## Accounts yang Diperlukan

### 1. Asset Accounts - Tax Prepaid

| Kode | Nama | Tipe | Fungsi |
|------|------|------|--------|
| 1114 | PPh 21 DIBAYAR DIMUKA | ASSET | Mencatat PPh 21 yang dipotong customer (prepaid tax) |
| 1115 | PPh 23 DIBAYAR DIMUKA | ASSET | Mencatat PPh 23 yang dipotong customer (prepaid tax) |
| 1116 | POTONGAN PAJAK LAINNYA DIBAYAR DIMUKA | ASSET | Mencatat potongan pajak lainnya (prepaid tax) |

**Catatan**: Saat customer memotong pajak dari invoice, perusahaan mencatatnya sebagai asset (prepaid) karena pajak tersebut dapat dikreditkan saat lapor pajak.

### 2. Liability Accounts - Tax Payable

| Kode | Nama | Tipe | Fungsi |
|------|------|------|--------|
| 2103 | PPN KELUARAN | LIABILITY | Mencatat PPN yang dibebankan ke customer (harus dibayar ke negara) |
| 2108 | PENAMBAHAN PAJAK LAINNYA | LIABILITY | Mencatat penambahan pajak lainnya |
| 292 | PENAMBAHAN PAJAK LAINNYA (SALES) | LIABILITY | Mencatat penambahan pajak spesifik untuk sales |

### 3. Revenue Accounts

| Kode | Nama | Tipe | Fungsi |
|------|------|------|--------|
| 4101 | PENDAPATAN PENJUALAN | REVENUE | Mencatat pendapatan dari penjualan barang |
| 4102 | PENDAPATAN JASA/ONGKIR | REVENUE | Mencatat pendapatan dari jasa dan ongkir |
| 293 | PENDAPATAN ONGKIR (SHIPPING) | REVENUE | Mencatat pendapatan spesifik dari biaya kirim |

### 4. Expense Accounts

| Kode | Nama | Tipe | Fungsi |
|------|------|------|--------|
| 5101 | HARGA POKOK PENJUALAN | EXPENSE | Mencatat cost of goods sold (COGS) |

### 5. Asset Accounts - Core

| Kode | Nama | Tipe | Fungsi |
|------|------|------|--------|
| 1101 | KAS | ASSET | Untuk pembayaran tunai |
| 1102 | BANK | ASSET | Untuk pembayaran transfer/bank |
| 1201 | PIUTANG USAHA | ASSET | Untuk penjualan kredit |
| 1301 | PERSEDIAAN BARANG DAGANGAN | ASSET | Inventory yang berkurang saat penjualan |

## Contoh Journal Entry - Sales dengan Tax

Contoh penjualan senilai Rp 1,237,500 (setelah diskon) dengan PPN 11%:

```
Tanggal: 2025-10-27
Ref: 0003/STAA,<9/X-2025

DEBIT  | 1102 - BANK                              | Rp 1,347,881.25
       |                                           |
CREDIT | 4101 - PENDAPATAN PENJUALAN              |                | Rp 1,237,500.00
CREDIT | 2103 - PPN KELUARAN (11%)                |                | Rp   134,763.75
CREDIT | 292 - PENAMBAHAN PAJAK LAINNYA           |                | Rp       120.00
CREDIT | 293 - PENDAPATAN ONGKIR                  |                | Rp       120.00
DEBIT  | 1114 - PPh 21 DIBAYAR DIMUKA             | Rp    12,251.25|
DEBIT  | 1115 - PPh 23 DIBAYAR DIMUKA             | Rp    12,251.25|
DEBIT  | 1116 - POTONGAN PAJAK LAINNYA            | Rp       120.00|

COGS Entry:
DEBIT  | 5101 - HARGA POKOK PENJUALAN             | Rp   400,000.00|
CREDIT | 1301 - PERSEDIAAN BARANG DAGANGAN        |                | Rp   400,000.00
-----------------------------------------------------------------------
TOTAL DEBIT:  Rp 1,772,503.75   |   TOTAL CREDIT: Rp 1,772,503.75 ✓
```

## Auto-Seeding

Semua account di atas **otomatis dibuat** saat:
1. Backend pertama kali dijalankan (`go run main.go`)
2. Function `SeedAccountsImproved()` di `database/account_seed_improved.go` dipanggil

### Cara Kerja Seeding:
- ✅ Jika account sudah ada: **SKIP** (tidak overwrite balance existing)
- ✅ Jika account belum ada: **CREATE** account baru
- ✅ Thread-safe: Menggunakan database transaction dan FOR UPDATE lock
- ✅ Duplicate-proof: Validasi tidak ada kode account duplicate

## Verifikasi Account

Untuk memastikan semua account sudah ada, jalankan query:

```sql
SELECT code, name, type 
FROM accounts 
WHERE code IN ('1114', '1115', '1116', '292', '293', '4101', '4102', '5101', '1301', '2103')
  AND deleted_at IS NULL
ORDER BY code;
```

Harus return 10 rows.

## Troubleshooting

### Log Warning: "account code XXX not found"

**Solusi**: Restart backend untuk trigger seeding ulang, atau manual insert account yang missing.

### Error: "TotalAmount mismatch"

**Bukan masalah COA**, ini calculation issue di sales creation. Total amount harus dihitung dengan formula:
```
TotalAmount = Subtotal + PPN + OtherTaxAdd + Shipping - PPh21 - PPh23 - OtherTaxDed
```

### Inventory Balance Negative

**Bukan error!** Ini normal untuk perpetual inventory system:
- Saat beli: DEBIT 1301 (balance naik)
- Saat jual: CREDIT 1301 (balance turun)
- Balance negative = sudah jual lebih banyak dari beli (perlu restock)

## References

- Account seeding: `backend/database/account_seed_improved.go`
- Sales journal entries: `backend/services/sales_journal_service_ssot.go`
- Account resolution logic: Lines 113-186 in sales_journal_service_ssot.go
