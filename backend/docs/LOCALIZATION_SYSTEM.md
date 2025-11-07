# Sistem Localization untuk PDF dan CSV Export

## Overview

Sistem localization ini telah diimplementasikan untuk memastikan semua generate PDF dan CSV mengikuti pengaturan bahasa user yang tersimpan di database settings. Sistem ini mendukung bahasa Indonesia (id) dan Inggris (en).

## Komponen Utama

### 1. Localization Utility (`utils/localization.go`)

File ini berisi:
- **Translation Map**: Menyimpan semua terjemahan untuk kedua bahasa
- **Helper Functions**: Fungsi untuk mengambil bahasa user dari database dan menerjemahkan text
- **CSV Headers**: Fungsi untuk generate header CSV yang sudah diterjemahkan

### 2. Key Functions

#### `GetUserLanguageFromDB(db *gorm.DB, userID uint) string`
Mengambil pengaturan bahasa user dari database. Jika tidak ditemukan, menggunakan pengaturan sistem.

#### `GetUserLanguageFromSettings(db *gorm.DB) string`
Mengambil pengaturan bahasa dari tabel settings sistem.

#### `T(key, language string) string`
Menerjemahkan key tertentu ke bahasa yang dipilih.

#### `GetCSVHeaders(reportType, language string) []string`
Menghasilkan header CSV yang sudah diterjemahkan berdasarkan jenis laporan.

## Implementasi di Service Layer

### 1. Export Service

```go
// Before localization
func (s *ExportServiceImpl) ExportAccountsPDF(ctx context.Context) ([]byte, error) {
    // ...
    pdf.Cell(190, 10, "Chart of Accounts")
    // ...
}

// After localization
func (s *ExportServiceImpl) ExportAccountsPDF(ctx context.Context, userID uint) ([]byte, error) {
    language := utils.GetUserLanguageFromDB(s.db, userID)
    // ...
    pdf.Cell(190, 10, utils.T("chart_of_accounts", language))
    // ...
}
```

### 2. Cash Flow Export Service

```go
// Before localization
writer.Write([]string{"Cash Flow Statement"})

// After localization
writer.Write([]string{utils.T("cash_flow_statement", language)})
```

### 3. Purchase Report Export Service

```go
// Before localization
w.Write([]string{"PURCHASE REPORT"})

// After localization
w.Write([]string{utils.T("purchase_report", language)})
```

## Translation Keys yang Tersedia

### Common Keys
- `company` - Perusahaan / Company
- `address` - Alamat / Address
- `phone` - Telepon / Phone
- `email` - Email / Email
- `generated_on` - Dibuat pada / Generated on
- `total` - Total / Total
- `date` - Tanggal / Date
- `amount` - Jumlah / Amount
- `status` - Status / Status

### Report Specific Keys
- `chart_of_accounts` - Daftar Akun / Chart of Accounts
- `cash_flow_statement` - Laporan Arus Kas / Cash Flow Statement
- `purchase_report` - Laporan Pembelian / Purchase Report
- `balance_sheet` - Neraca / Balance Sheet
- `profit_loss_statement` - Laporan Laba Rugi / Profit & Loss Statement

### Status Keys
- `active` - Aktif / Active
- `inactive` - Tidak Aktif / Inactive
- `paid` - Lunas / Paid
- `unpaid` - Belum Lunas / Unpaid

## Cara Menggunakan

### 1. Untuk Service Baru

```go
func (s *NewService) ExportReport(userID uint) ([]byte, error) {
    // Get user language preference
    language := utils.GetUserLanguageFromDB(s.db, userID)
    
    // Use translation
    title := utils.T("report_title", language)
    
    // For CSV headers
    headers := utils.GetCSVHeaders("report_type", language)
    
    // For PDF
    pdf.Cell(100, 10, utils.T("some_key", language))
    
    return data, nil
}
```

### 2. Menambah Translation Baru

Edit file `utils/localization.go` dan tambahkan di `PDFTranslations`:

```go
// Indonesian
"new_key": "Terjemahan Indonesia",

// English  
"new_key": "English Translation",
```

### 3. Untuk Report Type Baru di CSV Headers

Tambahkan case baru di function `GetCSVHeaders`:

```go
case "new_report_type":
    return []string{
        T("header1", language),
        T("header2", language),
        T("header3", language),
    }
```

## Service yang Sudah Diupdate

1. **ExportService** (`export_service.go`)
   - `ExportAccountsPDF(ctx context.Context, userID uint)`
   - `ExportAccountsExcel(ctx context.Context, userID uint)`

2. **CashFlowExportService** (`cash_flow_export_service.go`)
   - `ExportToCSV(data *SSOTCashFlowData, userID uint)`

3. **PurchaseReportExportService** (`purchase_report_export_service.go`)
   - `ExportToCSV(data *PurchaseReportData, userID uint)`
   - `ExportToPDF(data *PurchaseReportData, userID uint)`

## Testing

### 1. Mengubah Bahasa User di Settings

```sql
-- Set bahasa ke Indonesia
UPDATE settings SET language = 'id';

-- Set bahasa ke Inggris  
UPDATE settings SET language = 'en';
```

### 2. Test Export

```go
// Test dengan userID tertentu
pdfBytes, err := exportService.ExportAccountsPDF(ctx, userID)
csvBytes, err := cashFlowService.ExportToCSV(data, userID)
```

### 3. Validasi Output

- PDF harus menampilkan judul dan header dalam bahasa yang sesuai
- CSV harus memiliki header kolom dalam bahasa yang sesuai
- Format mata uang dan tanggal harus mengikuti locale

## Best Practices

1. **Selalu gunakan userID** saat memanggil export functions
2. **Fallback ke Indonesian** jika bahasa user tidak ditemukan
3. **Consistent naming** untuk translation keys
4. **Test kedua bahasa** setiap kali menambah fitur export baru
5. **Update translation** bersamaan dengan penambahan fitur

## Troubleshooting

### Translation Key Tidak Ditemukan
- Check log untuk pesan "Translation not found for key"  
- Pastikan key sudah ditambahkan di `PDFTranslations`
- Periksa spelling key dan language code

### Language Setting Tidak Terbaca
- Pastikan tabel `settings` memiliki column `language`
- Check koneksi database di service
- Validate language value ('id' atau 'en')

### CSV Headers Tidak Sesuai
- Pastikan report type sudah terdaftar di `GetCSVHeaders`
- Check case sensitivity untuk report type
- Validate translation keys untuk headers

## Future Improvements

1. **Dynamic language loading** dari file eksternal
2. **Caching translation** untuk performa
3. **Language per user** (saat ini per system)
4. **Date/Number formatting** sesuai locale
5. **RTL language support** jika diperlukan

---

*Dokumentasi ini akan diupdate seiring dengan penambahan fitur localization baru.*