# Proposal Pengembangan Aplikasi Cost Control & Project Management

**Klien** : PT Unipro  
**Penyedia** : [Nama Perusahaan/Tim Anda]  
**Tanggal** : [Isi Tanggal]

---

## 1. Latar Belakang

Dalam rangka meningkatkan efisiensi pengelolaan proyek dan pengendalian biaya, PT Unipro membutuhkan sebuah aplikasi terintegrasi yang dapat:

- Memantau progres fisik proyek secara real-time.
- Mengelola laporan harian dan mingguan proyek sesuai format RKS PT Unipro.
- Mengendalikan biaya proyek (budget vs actual) secara terstruktur.
- Mengelola permintaan pembelian (Purchase Request/PR) dengan alur approval bertingkat.
- Memantau penggunaan material dan stok yang terhubung dengan BOM (Bill of Material).

Aplikasi ini akan dikembangkan secara **custom** untuk menyesuaikan alur kerja dan struktur organisasi PT Unipro, serta menggantikan proses manual (misalnya penggunaan file PDF/PowerPoint) menjadi sistem digital terpusat.

---

## 2. Tujuan Pengembangan

1. Membangun satu platform terintegrasi untuk **Project Management**, **Cost Control**, dan **Purchasing** di PT Unipro.
2. Meningkatkan transparansi dan akurasi data melalui sistem **approval bertingkat** dan **audit trail**.
3. Mengotomatisasi pembuatan **weekly report** dari laporan harian, sesuai template RKS PT Unipro.
4. Menyediakan laporan **Budget vs Actual** dan **Cost Breakdown Structure (CBS)** per proyek, termasuk estimasi profit margin.
5. Menyediakan **Role-Based Access Control (RBAC)** yang menyesuaikan struktur organisasi dan pembagian kewenangan di PT Unipro.

---

## 3. Ruang Lingkup Fitur

### 3.1. Dashboard Proyek & Monitoring Progress

- Tampilan progres proyek dalam persen (%), per proyek.
- Perbandingan **progress fisik** dengan **penggunaan material**.
- Ringkasan indikator utama (key metrics): jumlah proyek aktif, status PR, budget vs actual, dan status material.

### 3.2. Modul Project Management & Laporan Lapangan

- Master data proyek (nama proyek, lokasi, periode, PIC, dan parameter lain yang dibutuhkan).
- **Daily Report**:
  - Input aktivitas harian.
  - Progress harian (%) dan keterangan.
  - Kondisi cuaca.
  - Upload dokumentasi foto (multi-upload).
- **Weekly Report Otomatis**:
  - Rekap otomatis dari daily report per minggu.
  - Format disesuaikan dengan **template RKS PT Unipro**.
  - Export dalam bentuk **PDF** (PDF generator).

### 3.3. Modul Cost Control & Cost Breakdown Structure (CBS)

- Input dan pengelolaan **budget** per proyek berdasarkan struktur CBS:
  - Direct Cost: Material, Labor, Equipment.
  - Indirect Cost: Overhead.
- Pencatatan **actual cost** yang terhubung dengan transaksi PR/pembelian.
- Laporan **Budget vs Actual** per proyek dan per kategori biaya.
- Perhitungan **Total Cost** dan **Profit Margin** per proyek.

### 3.4. Modul Purchasing & Purchase Request (PR)

#### 3.4.1. Create PR

- Pemilihan proyek (project dropdown).
- Input daftar item (material/alat) per PR.
- Field: Quantity, Unit, Unit Price, Total Amount.
- Pemilihan vendor.
- Notes/description tambahan.

#### 3.4.2. PR Approval Dashboard

- Daftar PR yang **pending approval** per user/role.
- Halaman detail PR dengan breakdown item.
- Tombol aksi: **Approve / Reject / Request Revision**.
- Kolom komentar antar divisi (comment section).
- Tracking status PR dan riwayat perubahan status.

#### 3.4.3. PR History & Reporting

- Daftar seluruh PR dengan filter:
  - Project
  - Status
  - Date range
- Export laporan PR ke **Excel/PDF**.
- Ringkasan (summary) per periode: total PR, jumlah approved, rejected, pending.

### 3.5. Flow Approval PR Bertingkat

Flow approval PR akan dilakukan penuh melalui sistem (tanpa PDF manual), dengan alur:

1. **Purchasing** menginput kebutuhan & harga material (Create PR).
2. **Cost Control (Pak Patrick)** melakukan verifikasi biaya.
3. **GM (Pak Marlin)** melakukan approval.
4. **Direktur Proyek (Pak Christopher)** melakukan approval lanjutan.
5. **Direktur Utama (Pak Jason)** melakukan approval akhir.

Sistem akan menyimpan **riwayat approval** lengkap (siapa, kapan, aksi apa, dan komentar), serta mengirimkan notifikasi kepada pihak terkait pada setiap tahapan.

### 3.6. Modul Material Tracking & Integrasi BOM

- Master data material dan stok per proyek/gudang (bila diperlukan).
- Pencatatan material yang sudah dibeli (PR approved).
- Pencatatan material yang sudah terpakai di lapangan (terhubung dengan daily report atau form penggunaan material).
- Perhitungan **material sisa (inventory)** per proyek.
- Perbandingan **budget material vs actual material usage**.
- Integrasi dengan **BOM (Bill of Material)** per proyek untuk monitoring kebutuhan vs realisasi.

### 3.7. Role-Based Access Control (RBAC) & Audit Trail

- Penerapan role & permission, antara lain:
  - **Director**: full access ke seluruh modul dan laporan.
  - **GM/Manager**: akses proyek dan approval sesuai kewenangan.
  - **Cost Control**: akses verifikasi biaya dan laporan cost.
  - **Purchasing**: akses modul PR dan purchasing.
  - **Tim Lapangan (Site Team)**: akses input data proyek harian (tanpa akses angka finansial).
- Pembatasan akses menu dan data berdasarkan role.
- Pencatatan **audit trail** untuk aktivitas penting (create, update, approval, perubahan status).

### 3.8. Infrastruktur, Deployment & Maintenance

- Penyediaan **server** dan **domain name** untuk aplikasi (durasi 1 tahun pertama).
- Konfigurasi hosting, database, dan SSL.
- Proses deployment ke lingkungan produksi.
- Dokumentasi dasar penggunaan dan sesi training untuk user kunci.
- Maintenance dan support teknis selama **6 bulan** setelah go-live.

---

## 4. Estimasi Biaya Pengembangan

Total biaya pengembangan aplikasi, termasuk server, domain, deployment, free revisi dalam scope, dan 6 bulan maintenance adalah:

> **Rp 95.000.000,- (Sembilan Puluh Lima Juta Rupiah)**

Rincian alokasi biaya per modul/komponen adalah sebagai berikut:

| No | Modul / Komponen                                                                 | Deskripsi Singkat                                                                                                                      | Biaya (IDR)       |
|----|----------------------------------------------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------|-------------------|
| 1  | Modul Project Management & Laporan Lapangan                                      | Project, daily report (aktivitas, cuaca, foto), weekly report otomatis, PDF generator sesuai template RKS PT Unipro                   | Rp 25.000.000,-   |
| 2  | Modul Cost Control & CBS (Budget vs Actual, Profit Margin)                       | Struktur CBS, input budget, perhitungan actual cost, laporan budget vs actual per proyek dan per kategori                             | Rp 20.000.000,-   |
| 3  | Modul Purchasing & PR Workflow                                                   | Create PR, PR Approval Dashboard, multi-level approval (Purchasing → Cost Control → GM → Direktur Proyek → Direktur Utama), komentar & riwayat approval | Rp 20.000.000,-   |
| 4  | Modul Material Tracking & BOM Integration                                        | Master material, stok, integrasi dengan BOM proyek, perhitungan material dibeli/terpakai/sisa, comparison budget vs actual           | Rp 10.000.000,-   |
| 5  | Modul RBAC, Notifikasi & Audit Trail                                             | Role & permission, pembatasan akses per divisi/level, notifikasi dasar, audit trail aktivitas penting                                | Rp 7.500.000,-    |
| 6  | Deployment, Testing, Training & Dokumentasi                                      | UAT, perbaikan, deployment ke server produksi, training user, dokumentasi/manual penggunaan dasar                                     | Rp 5.000.000,-    |
| 7  | Server & Domain (1 tahun) + Backup & Monitoring Dasar                            | VPS + domain, konfigurasi SSL, backup berkala dan pemantauan dasar selama masa maintenance                                            | Rp 7.500.000,-    |
|    | **Total Investasi**                                                              |                                                                                                                                        | **Rp 95.000.000,-** |

Angka di atas sudah disusun agar alokasi biaya per modul realistis dan dapat menjadi acuan internal dalam pengembangan serta pengelolaan infrastruktur.

---

## 5. Free Revisi, Deployment & Maintenance

### 5.1. Free Revisi

- Termasuk **free revisi** untuk perubahan **minor** (tampilan, label, penyesuaian kecil alur kerja) selama masa pengembangan hingga maksimal **1 (satu) bulan setelah go-live**, selama masih dalam ruang lingkup fitur yang tercantum di proposal ini.
- Permintaan fitur baru atau perubahan besar di luar scope awal akan dibahas dan ditawarkan dalam proposal terpisah.

### 5.2. Free Deployment

- Deployment aplikasi ke **server produksi** yang disediakan.
- Setup domain dan konfigurasi **SSL** sehingga aplikasi dapat diakses dengan aman (HTTPS).
- Konfigurasi dasar backup database dan file penting.

### 5.3. Free Maintenance 6 Bulan

- Masa maintenance: **6 (enam) bulan** sejak tanggal go-live.
- Cakupan maintenance:
  - Bug fixing untuk masalah yang muncul dalam penggunaan normal sistem.
  - Penyesuaian minor (misalnya penyesuaian teks, tampilan sederhana, atau report minor).
- Contoh SLA yang dapat diterapkan (dapat disesuaikan):
  - Bug kritis (sistem tidak bisa digunakan): respon awal 1–2 hari kerja.
  - Bug non-kritis: respon 3–5 hari kerja.

Perpanjangan maintenance dan perpanjangan server/domain setelah periode awal akan dibahas secara terpisah sesuai kebutuhan PT Unipro.

---

## 6. Estimasi Timeline Pengerjaan

Estimasi durasi pengembangan (dapat disesuaikan dengan prioritas modul dan ketersediaan tim PT Unipro):

1. **Analisis Detail & Finalisasi Requirement**: 2 minggu  
2. **Desain UI/UX & Arsitektur Teknis**: 2 minggu  
3. **Pengembangan Modul Utama** (Project Management, Cost Control, PR, Material): 8–10 minggu  
4. **Integrasi, Testing, dan User Acceptance Test (UAT)**: 3–4 minggu  
5. **Deployment & Training User**: 1 minggu  

Perkiraan total durasi: **±4 bulan**.

---

## 7. Skema Pembayaran (Opsional)

Skema pembayaran yang diusulkan (dapat dinegosiasikan kembali):

1. **30%** di awal, setelah proposal disetujui dan kontrak kerja ditandatangani.
2. **30%** setelah penyelesaian pengembangan modul utama dan demo internal.
3. **30%** setelah UAT selesai dan perbaikan utama diselesaikan.
4. **10%** setelah sistem go-live.

---

## 8. Penutup

Proposal ini disusun untuk memberikan gambaran menyeluruh mengenai fitur, ruang lingkup pekerjaan, estimasi biaya, dan komitmen dukungan dalam pengembangan aplikasi Cost Control & Project Management untuk PT Unipro.

Kami siap melakukan penyesuaian lebih lanjut terkait detail teknis, prioritas modul, maupun skema kerja sama sesuai kebutuhan PT Unipro.

Hormat kami,

**[Nama Perusahaan/Tim Anda]**  
**[Nama Penanggung Jawab]**  
**[Jabatan]**  
**[Kontak/Email/Telepon]**
