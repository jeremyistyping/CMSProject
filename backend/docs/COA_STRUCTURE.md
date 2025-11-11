# Chart of Accounts (COA) - Cost Control Padel Bandung
## Struktur Akuntansi untuk Project Construction Management

---

## 📊 Ringkasan COA

Total: **49 Accounts** (setelah cleanup)
- Aset Lancar: 7 accounts
- Kewajiban: 5 accounts  
- Ekuitas: 2 accounts
- Pendapatan Proyek: 3 accounts
- Beban Langsung Proyek: 20 accounts
- Overhead Kantor & Admin: 3 accounts

---

## 1️⃣ ASET LANCAR (1000-1999)

### 1000 - ASET LANCAR (Header)
Aset yang dapat dicairkan dalam waktu singkat untuk kebutuhan proyek.

#### Kas & Bank
- **1101** - KAS PROYEK
  - Deskripsi: Uang tunai site
  - Mapping: Operasional Budget
  
- **1102** - BANK  
  - Deskripsi: Transfer project
  - Mapping: Operasional Budget

- **1201** - DEPOSIT
  - Deskripsi: DP supplier / sewa
  - Mapping: Operasional Budget

#### Tax Prepaid Accounts (Pajak Dibayar Dimuka)
- **1114** - PPh 21 DIBAYAR DIMUKA
- **1115** - PPh 23 DIBAYAR DIMUKA  
- **1240** - PPN MASUKAN

---

## 2️⃣ KEWAJIBAN (2000-2999)

### 2000 - KEWAJIBAN (Header)

#### Current Liabilities
- **2101** - UTANG USAHA
- **2103** - PPN KELUARAN
- **2104** - PPh YANG DIPOTONG
- **2111** - UTANG PPh 21
- **2112** - UTANG PPh 23

---

## 3️⃣ EKUITAS (3000-3999 & 7000)

### 3000 - EKUITAS (Header)

- **3101** - MODAL PEMILIK
- **7000** - LABA / RUGI PROYEK
  - Deskripsi: Selisih income - cost
  - Mapping: Profitability Report

---

## 4️⃣ PENDAPATAN PROYEK - INCOME (4000-4999)

### 4000 - PENDAPATAN PROYEK (INCOME) (Header)
Mapping: **Contract / Income Budget**

- **4101** - PENDAPATAN TERMIN 1
  - Deskripsi: Termin proyek
  
- **4102** - PENDAPATAN TERMIN 2
  - Deskripsi: Pembayaran bertahap
  
- **4201** - RETENSI
  - Deskripsi: Potongan retensi

---

## 5️⃣ BEBAN LANGSUNG PROYEK - DIRECT COST (5000-5999)

### 5000 - BEBAN LANGSUNG PROYEK (DIRECT COST) (Header)
Biaya yang langsung terkait dengan proyek

### 5100 - MATERIAL BANGUNAN (Header)
Mapping: **Material Building Budget**

- **5101** - SEMEN & PASIR (Material utama)
- **5102** - BESI & BAJA (Struktur)
- **5103** - PLUMBING & FITTING (Instalasi)
- **5104** - KACA, ALUMINIUM (Finishing)
- **5105** - CAT & FINISHING (Pengecatan)

### 5200 - SEWA ALAT BERAT / EQUIPMENT HIRE (Header)
Mapping: **Budget Sewa (Equipment Hire)**

- **5201** - SEWA ALAT BERAT (Harian/Mingguan)
- **5202** - TRANSPORT & MOBILISASI (Mobilisasi alat)

### 5300 - TENAGA KERJA (LABOUR) (Header)
Mapping: **Labour Budget**

- **5301** - MANDOR (Gaji mandor)
- **5302** - TUKANG & HELPER (Upah pekerja lapangan)
- **5303** - OVERTIME & BONUS (Tambahan jam kerja)

### 5400 - BIAYA OPERASIONAL SITE (Header)
Mapping: **Operasional Budget**

- **5401** - AIR & LISTRIK KERJA (Utilitas proyek)
- **5402** - TRANSPORTASI & TOL (Perjalanan tim)
- **5403** - KONSUMSI & ENTERTAIN (Konsumsi lapangan)
- **5404** - AKOMODASI (KOSAN, HOTEL) (Tempat tinggal tim)
- **5405** - ATK & ALAT KECIL (Meteran, spidol, helm, dll)
- **5406** - KOMPENSASI & KEAMANAN (Gaji security, kompensasi)

---

## 6️⃣ OVERHEAD KANTOR & ADMIN (6000-6999)

### 6000 - OVERHEAD KANTOR & ADMIN (Header)
Mapping: **Operasional Budget**

- **6101** - ADMIN FEE (Biaya transfer, admin bank)
- **6102** - PAJAK PROYEK (PPH, PPN) (Biaya pajak)
- **6103** - FEE MARKETING (Fee rekanan / marketing)

---

## 📈 Mapping Budget ke COA

| Budget Group | COA Range | Accounts |
|-------------|-----------|----------|
| **Income Budget** | 4000 - 4201 | Pendapatan Termin 1, Pendapatan Termin 2, Retensi |
| **Material Building Budget** | 5100 - 5105 | Semen & Pasir, Besi & Baja, Plumbing, Kaca, Cat |
| **Budget Sewa** | 5200 - 5202 | Sewa Alat Berat, Transport & Mobilisasi |
| **Labour Budget** | 5300 - 5303 | Mandor, Tukang & Helper, Overtime & Bonus |
| **Operasional Budget** | 5400 - 6103 | Utilitas, Transportasi, Konsumsi, Akomodasi, ATK, Admin, Pajak, Fee Marketing |

---

## 📊 Output Laporan yang Dapat Dibentuk

### 1. Budget vs Actual by COA Group
Menampilkan total estimasi vs realisasi per akun.

### 2. Profitability Report per Project
Formula: (Pendapatan 4000-4201) – (Total Beban 5000-6103)
Result: 7000 - LABA / RUGI PROYEK

### 3. Cash Flow per Project
Dari kas masuk & kas keluar sesuai COA tipe Asset/Expense:
- Kas Masuk: 1101, 1102
- Kas Keluar: 5000-6103

### 4. Cost Summary Report
Rekap per kategori:
- Material (5100-5105)
- Sewa (5200-5202)
- Labour (5300-5303)
- Operasional (5400-5406)
- Overhead (6000-6103)

---

## ❌ Account yang Dihapus (Tidak Relevan)

### Fixed Assets (1500-1509) - DIHAPUS
Alasan: Sistem ini untuk cost control proyek, bukan mengelola aset tetap. Alat berat disewa, bukan dimiliki.
- ~~1500 - FIXED ASSETS~~
- ~~1501 - PERALATAN KANTOR~~
- ~~1502 - KENDARAAN~~
- ~~1503 - BANGUNAN~~
- ~~1509 - TRUK~~

### Old Liability Structure - DIHAPUS
- ~~2100 - CURRENT LIABILITIES~~ (header lama)
- ~~2107 - PEMOTONGAN PAJAK LAINNYA~~ (duplikat dengan 2104)
- ~~2108 - PENAMBAHAN PAJAK LAINNYA~~ (tidak sesuai)

### Old Equity - DIHAPUS
- ~~3201 - LABA DITAHAN~~ (diganti dengan 7000)

### Old Revenue - DIHAPUS
- ~~4900 - OTHER INCOME~~ (tidak sesuai untuk proyek konstruksi)

### Old Expenses - DIHAPUS
- ~~5203 - BEBAN TELEPON~~ (masuk ke 5401)
- ~~5204 - BEBAN TRANSPORTASI~~ (duplikat dengan 5402)
- ~~5900 - GENERAL EXPENSE~~ (terlalu umum)

---

## 🔧 Maintenance

### File Terkait
- **Seed File:** `database/account_seed_improved.go`
- **Cleanup Script:** `cmd/scripts/cleanup_unused_accounts.go`
- **SQL Backup:** `database/cleanup_unused_accounts.sql`

### Cara Update COA
1. Edit `database/account_seed_improved.go`
2. Run `go run main.go` untuk apply perubahan
3. COA akan otomatis di-update dengan UPSERT (tidak reset balance)

### Proteksi
- COA menggunakan **soft delete** (deleted_at)
- Balance existing accounts **TIDAK akan direset** saat seed
- System critical accounts **dilindungi** dari perubahan

---

**Last Updated:** 2025-11-11
**Version:** 1.0 - Cost Control Padel Bandung
