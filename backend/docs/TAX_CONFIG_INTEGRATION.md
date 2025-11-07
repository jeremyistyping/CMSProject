# Integrasi Tax Configuration dengan Sales & Purchase Management

## Overview
Sistem akuntansi kini memiliki integrasi penuh antara **TaxConfig** (pengaturan rate pajak) dan **TaxAccountSettings** (pengaturan akun pajak) untuk sales management dan purchase management.

## Struktur Integrasi

### 1. TaxConfig (models/tax_config.go)
Mengatur **rate pajak default** dan **account mapping** untuk:
- **Sales Taxes:**
  - PPN Keluaran (Sales PPN Rate & Account)
  - PPh 21 dipotong customer (Sales PPh21 Rate & Account)
  - PPh 23 dipotong customer (Sales PPh23 Rate & Account)
  - Pajak lainnya (Sales Other Tax Rate & Account)

- **Purchase Taxes:**
  - PPN Masukan (Purchase PPN Rate & Account)
  - PPh 21 yang kita potong (Purchase PPh21 Rate & Account)
  - PPh 23 yang kita potong (Purchase PPh23 Rate & Account)
  - PPh 25 (Purchase PPh25 Rate & Account)
  - Pajak lainnya (Purchase Other Tax Rate & Account)

- **Tax Calculation Rules:**
  - `shipping_taxable`: Apakah ongkir kena pajak
  - `discount_before_tax`: Diskon sebelum atau sesudah pajak
  - `rounding_method`: ROUND_UP, ROUND_DOWN, ROUND_HALF_UP

### 2. TaxAccountSettings (models/tax_account_settings.go)
Mengatur **default accounts** untuk transaksi:
- **Sales:** Receivable, Cash, Bank, Revenue, Output VAT
- **Purchase:** Payable, Cash, Bank, Input VAT, Expense
- **Tax:** Withholding Tax 21, 23, 25, Tax Payable
- **Inventory:** Inventory, COGS

### 3. TaxAccountService (services/tax_account_service.go)
Service yang mengintegrasikan kedua model di atas dengan fungsi:

#### Fungsi Baru untuk TaxConfig Integration:

```go
// Mendapatkan tax config aktif
GetActiveTaxConfig() (*models.TaxConfig, error)

// Mendapatkan tax rate
GetSalesTaxRate(taxType string) (float64, error)
GetPurchaseTaxRate(taxType string) (float64, error)

// Mendapatkan tax account ID dari TaxConfig
GetSalesTaxAccountID(taxType string) (uint, error)
GetPurchaseTaxAccountID(taxType string) (uint, error)

// Menghitung pajak dengan aturan dari TaxConfig
CalculateTax(amount, taxRate float64, discountBeforeTax bool, discount float64) float64

// Menghitung semua pajak sales sekaligus
CalculateSalesTax(subtotal, discount float64, enablePPN, enablePPh21, enablePPh23, enableOther bool) (*SalesTaxCalculation, error)

// Menghitung semua pajak purchase sekaligus
CalculatePurchaseTax(subtotal, discount float64, enablePPN, enablePPh21, enablePPh23, enablePPh25, enableOther bool) (*PurchaseTaxCalculation, error)
```

## Cara Penggunaan

### A. Mengatur Tax Configuration

1. **Buat atau Update TaxConfig:**
```go
// Create default tax config
defaultConfig := models.GetDefaultTaxConfig()
db.Create(defaultConfig)

// Update tax rates
db.Model(&models.TaxConfig{}).
  Where("id = ?", 1).
  Updates(map[string]interface{}{
    "sales_ppn_rate": 11.0,      // PPN 11%
    "purchase_pph23_rate": 2.0,  // PPh23 2%
  })
```

2. **Set Active Config:**
```sql
UPDATE tax_configs SET is_active = true, is_default = true WHERE id = 1;
```

### B. Menggunakan di Sales Management

```go
// Di controller atau service
taxAccountService := NewTaxAccountService(db)

// 1. Mendapatkan tax rate untuk sales
ppnRate, err := taxAccountService.GetSalesTaxRate("ppn")
if err != nil {
  // Handle error atau gunakan default rate 11%
  ppnRate = 11.0
}

// 2. Menghitung PPN dengan aturan dari config
subtotal := 1000000.0
discount := 50000.0
ppnAmount := taxAccountService.CalculateTax(subtotal, ppnRate, true, discount)
// Hasil: (1000000 - 50000) * 11% = 104500

// 3. Menghitung semua pajak sekaligus
taxCalc, err := taxAccountService.CalculateSalesTax(
  subtotal,
  discount,
  true,  // enable PPN
  false, // disable PPh21
  false, // disable PPh23
  false, // disable other
)
if err == nil {
  sale.PPN = taxCalc.PPNAmount
  sale.TotalAmount = subtotal - discount + taxCalc.TotalTaxAmount
}

// 4. Mendapatkan account untuk posting journal
ppnAccountID, err := taxAccountService.GetSalesTaxAccountID("ppn")
// atau fallback ke TaxAccountSettings
if err != nil {
  ppnAccountID, err = taxAccountService.GetAccountID("sales_output_vat")
}
```

### C. Menggunakan di Purchase Management

```go
// Di controller atau service
taxAccountService := NewTaxAccountService(db)

// 1. Mendapatkan tax rates untuk purchase
ppnRate, _ := taxAccountService.GetPurchaseTaxRate("ppn")
pph23Rate, _ := taxAccountService.GetPurchaseTaxRate("pph23")

// 2. Menghitung pajak purchase
subtotal := 500000.0
discount := 0.0

// Hitung PPN Masukan
ppnAmount := taxAccountService.CalculateTax(subtotal, ppnRate, true, discount)

// Hitung PPh23 yang kita potong
pph23Amount := taxAccountService.CalculateTax(subtotal, pph23Rate, true, discount)

// 3. Atau hitung semua sekaligus
taxCalc, err := taxAccountService.CalculatePurchaseTax(
  subtotal,
  discount,
  true,  // enable PPN
  false, // disable PPh21
  true,  // enable PPh23
  false, // disable PPh25
  false, // disable other
)
if err == nil {
  purchase.PPNAmount = taxCalc.PPNAmount
  purchase.PPh23Amount = taxCalc.PPh23Amount
  purchase.TotalAmount = subtotal - discount + taxCalc.PPNAmount - taxCalc.PPh23Amount
}

// 4. Mendapatkan accounts
ppnAccountID, _ := taxAccountService.GetPurchaseTaxAccountID("ppn")
pph23AccountID, _ := taxAccountService.GetPurchaseTaxAccountID("pph23")
```

### D. Integration dengan Journal Services

Kedua journal service (`sales_journal_service_enhanced.go` dan `purchase_journal_service_enhanced.go`) sudah terintegrasi dengan TaxAccountService:

```go
// Di sales_journal_service_enhanced.go
// PPN account dari TaxAccountSettings
accountID, err := s.taxAccountService.GetAccountID("sales_output_vat")

// Bisa juga menggunakan TaxConfig:
// ppnRate, _ := s.taxAccountService.GetSalesTaxRate("ppn")
// calculatedPPN := s.taxAccountService.CalculateTax(sale.Subtotal, ppnRate, true, discount)
```

## Keuntungan Integrasi

1. **Centralized Tax Configuration** - Semua tax rates dan accounts di satu tempat
2. **Flexible Tax Calculation** - Support berbagai jenis pajak dan aturan perhitungan
3. **Consistent Accounting** - Journal entries menggunakan accounts yang konsisten
4. **Easy Maintenance** - Update tax rate cukup di TaxConfig, tidak perlu hardcode
5. **Audit Trail** - Semua perubahan tax config tercatat dengan user dan timestamp

## Migration Path

Sistem saat ini sudah support TaxConfig, tapi masih fallback ke hardcoded values jika config tidak ada:

1. **Phase 1 (Current)**: TaxAccountSettings untuk accounts, hardcoded rates
2. **Phase 2 (Implemented)**: TaxConfig untuk rates + accounts, TaxAccountSettings untuk default accounts
3. **Phase 3 (Future)**: Full migration ke TaxConfig untuk semua tax calculations

## Database Schema

### TaxConfig Table
```sql
CREATE TABLE tax_configs (
  id SERIAL PRIMARY KEY,
  config_name VARCHAR(100) NOT NULL UNIQUE,
  description TEXT,
  
  -- Sales tax rates
  sales_ppn_rate DECIMAL(5,2) DEFAULT 11.0,
  sales_pph21_rate DECIMAL(5,2) DEFAULT 0.0,
  sales_pph23_rate DECIMAL(5,2) DEFAULT 0.0,
  sales_other_tax_rate DECIMAL(5,2) DEFAULT 0.0,
  
  -- Sales tax accounts
  sales_ppn_account_id INT,
  sales_pph21_account_id INT,
  sales_pph23_account_id INT,
  sales_other_tax_account_id INT,
  
  -- Purchase tax rates
  purchase_ppn_rate DECIMAL(5,2) DEFAULT 11.0,
  purchase_pph21_rate DECIMAL(5,2) DEFAULT 0.0,
  purchase_pph23_rate DECIMAL(5,2) DEFAULT 2.0,
  purchase_pph25_rate DECIMAL(5,2) DEFAULT 0.0,
  purchase_other_tax_rate DECIMAL(5,2) DEFAULT 0.0,
  
  -- Purchase tax accounts
  purchase_ppn_account_id INT,
  purchase_pph21_account_id INT,
  purchase_pph23_account_id INT,
  purchase_pph25_account_id INT,
  purchase_other_tax_account_id INT,
  
  -- Additional settings
  shipping_taxable BOOLEAN DEFAULT true,
  discount_before_tax BOOLEAN DEFAULT true,
  rounding_method VARCHAR(20) DEFAULT 'ROUND_HALF_UP',
  
  is_active BOOLEAN DEFAULT true,
  is_default BOOLEAN DEFAULT false,
  
  updated_by INT NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);
```

## API Endpoints

### TaxConfig APIs
- `GET /api/settings/tax-configs` - List all tax configs
- `GET /api/settings/tax-configs/:id` - Get specific config
- `POST /api/settings/tax-configs` - Create new config
- `PUT /api/settings/tax-configs/:id` - Update config
- `DELETE /api/settings/tax-configs/:id` - Delete config
- `POST /api/settings/tax-configs/:id/activate` - Set as active/default

### TaxAccountSettings APIs
- `GET /api/settings/tax-accounts` - Get active settings
- `POST /api/settings/tax-accounts` - Create new settings
- `PUT /api/settings/tax-accounts/:id` - Update settings

## Contoh Use Case

### Scenario: Perusahaan mengubah PPN dari 11% ke 12%

**Cara Lama (Hardcoded):**
```go
// Harus update di banyak file
const PPN_RATE = 12.0 // Di semua controller/service
```

**Cara Baru (TaxConfig):**
```sql
-- Cukup update di database
UPDATE tax_configs 
SET sales_ppn_rate = 12.0, purchase_ppn_rate = 12.0 
WHERE is_active = true;
```

Atau via API:
```bash
curl -X PUT http://localhost:8080/api/settings/tax-configs/1 \
  -H "Content-Type: application/json" \
  -d '{"sales_ppn_rate": 12.0, "purchase_ppn_rate": 12.0}'
```

Semua transaksi baru otomatis menggunakan rate 12%!

## Testing

Untuk testing integrasi:

```bash
# Run backend
cd backend
go run cmd/scripts/test_tax_account_api.go
```

## Kesimpulan

Integrasi TaxConfig dengan TaxAccountSettings memberikan fleksibilitas penuh dalam mengatur:
- ✅ **Tax rates** - Dapat diubah kapan saja dari UI/API
- ✅ **Tax accounts** - Mapping akun pajak yang konsisten
- ✅ **Tax calculation rules** - Discount, rounding, shipping taxable
- ✅ **Sales & Purchase management** - Otomatis menggunakan konfigurasi terbaru
- ✅ **Journal entries** - Posting ke akun yang tepat sesuai konfigurasi
