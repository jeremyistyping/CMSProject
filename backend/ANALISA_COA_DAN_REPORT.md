# ANALISA APLIKASI - BAGIAN COA DAN REPORT

## 📋 RINGKASAN EKSEKUTIF

Aplikasi sistem akuntansi ini memiliki dua komponen utama yang dianalisa:
1. **COA (Chart of Accounts)** - Sistem manajemen akun/rekening akuntansi
2. **Report System** - Sistem pelaporan keuangan yang komprehensif

Kedua sistem telah diintegrasikan dengan **SSOT (Single Source of Truth)** menggunakan **Unified Journal Ledger** sebagai sumber data tunggal untuk memastikan konsistensi dan akurasi data.

---

## 🏗️ ARSITEKTUR SISTEM

### Struktur Umum
```
Backend
├── controllers/     # HTTP request handlers
│   ├── coa_controller_v2.go
│   ├── coa_posted_controller.go
│   ├── enhanced_report_controller.go
│   ├── psak_compliant_report_controller.go
│   ├── ssot_purchase_report_controller.go
│   └── ssot_report_integration_controller.go
├── services/       # Business logic
│   ├── coa_service.go
│   ├── coa_display_service.go
│   ├── coa_display_service_v2.go
│   ├── enhanced_report_service.go
│   ├── ssot_purchase_report_service.go
│   └── report_cache_service.go
├── models/         # Data models
│   └── financial_report.go
└── routes/         # API routing
    ├── enhanced_report_routes.go
    └── ssot_report_routes.go
```

---

## 📊 BAGIAN 1: COA (CHART OF ACCOUNTS)

### 1.1 Konsep & Tujuan

COA adalah **master data akuntansi** yang menyimpan:
- Daftar akun/rekening yang digunakan dalam sistem akuntansi
- Struktur hierarkis akun (header dan detail)
- Tipe akun: ASSET, LIABILITY, EQUITY, REVENUE, EXPENSE
- Balance/saldo setiap akun

### 1.2 Komponen Utama

#### A. COAControllerV2 (`controllers/coa_controller_v2.go`)

**Endpoint yang tersedia:**

1. **GetCOAWithDisplay** - Mendapatkan semua COA dengan format display
   - Menampilkan akun aktif dengan saldo non-zero
   - Balance ditampilkan sesuai konvensi akuntansi
   
2. **GetCOAByID** - Mendapatkan single COA berdasarkan ID
   
3. **GetCOABalancesByType** - Mendapatkan saldo yang dikelompokkan berdasarkan tipe akun
   
4. **GetSpecificAccounts** - Mendapatkan akun spesifik untuk sales display
   - Akun yang di-hardcode: Kas (1101), Bank (1102), Piutang (1201), PPN Keluaran (2103), Pendapatan (4101)
   
5. **GetSalesRelatedAccounts** - Mendapatkan akun untuk transaksi penjualan
   - Termasuk: Kas, Bank, Piutang, PPN Keluaran, Pendapatan Penjualan, Pendapatan Jasa

#### B. COADisplayService (`services/coa_display_service.go`)

**Fungsi Utama:**

```go
type COADisplayAccount struct {
    ID             uint    
    Code           string  
    Name           string  
    Type           string  
    Category       string  
    RawBalance     float64  // Balance asli dari database
    DisplayBalance float64  // Balance untuk display (sudah dikoreksi)
    IsActive       bool     
    ParentID       *uint    
    HasChildren    bool     
    Level          int      
}
```

**Logika Display Balance:**

```go
func getDisplayBalance(rawBalance float64, accountType string) float64 {
    switch accountType {
    case "ASSET", "EXPENSE":
        // Tampilkan as-is
        return rawBalance
        
    case "REVENUE", "LIABILITY", "EQUITY":
        // Flip sign: negative balance → positive display
        // Karena di sistem accounting, credit increases these accounts
        return rawBalance * -1
        
    default:
        return rawBalance
    }
}
```

**⚠️ MASALAH POTENSIAL:**
- `RawBalance` vs `DisplayBalance` bisa membingungkan user
- Logika flip sign harus dipahami dengan baik oleh frontend
- Tidak ada dokumentasi mengapa flip sign diperlukan

#### C. COAService (`services/coa_service.go`)

Service dasar untuk operasi CRUD COA:
- `GetByID` - Get COA by ID
- `GetByCode` - Get COA by code
- `UpdateBalance` - Update balance
- `GetAll` - Get semua COA
- `GetAccountsWithFilter` - Get dengan filter

### 1.3 Isu & Rekomendasi COA

#### ✅ KELEBIHAN:
1. **Dual Balance System** - `RawBalance` dan `DisplayBalance` memisahkan logika penyimpanan dan tampilan
2. **Hierarchy Support** - Mendukung parent-child relationship
3. **Flexible Filtering** - Berbagai endpoint untuk kebutuhan berbeda

#### ⚠️ MASALAH & PERBAIKAN:

1. **Hard-coded Account Codes**
   ```go
   // Di GetSpecificAccounts
   accountCodes := []string{
       "1101", // Kas
       "1102", // Bank
       "1201", // Piutang Usaha
       // ...
   }
   ```
   **REKOMENDASI:** Pindahkan ke configuration table atau constants file

2. **Tidak Ada Validasi Balance**
   - Tidak ada pengecekan apakah total debit = total credit
   - Tidak ada audit trail untuk perubahan balance
   
   **REKOMENDASI:** 
   ```go
   type COABalanceHistory struct {
       AccountID   uint
       OldBalance  float64
       NewBalance  float64
       ChangedBy   uint
       ChangedAt   time.Time
       Reason      string
   }
   ```

3. **Sync dengan Journal**
   - Ada file `cashbank_coa_sync_service.go` yang mengindikasikan ada masalah sinkronisasi
   - Banyak script perbaikan: `fix_coa_balance_sync.go`, `fix_coa_parent_balance_calculation.go`
   
   **REKOMENDASI:** 
   - Gunakan SSOT journal sebagai sumber tunggal kebenaran
   - COA balance harus SELALU dihitung dari journal entries
   - Jangan simpan balance di COA table (calculated field)

---

## 📈 BAGIAN 2: REPORT SYSTEM

### 2.1 Arsitektur Report

Sistem report memiliki **dua layer**:

1. **Enhanced Report System** (Legacy)
   - Service: `EnhancedReportService`
   - Controller: `EnhancedReportController`
   - Mengambil data dari berbagai tabel (sales, purchases, products, dll)
   
2. **SSOT Report System** (Current/Recommended)
   - Menggunakan Unified Journal Ledger sebagai sumber data tunggal
   - Controllers: 
     - `SSOTProfitLossController`
     - `SSOTBalanceSheetController`
     - `SSOTPurchaseReportController`
     - `SSOTReportIntegrationController`

### 2.2 Jenis-jenis Report

#### A. Financial Reports

**1. Balance Sheet (Neraca)**
```go
type BalanceSheetData struct {
    Company      CompanyInfo
    AsOfDate     time.Time
    Assets       BalanceSheetSection
    Liabilities  BalanceSheetSection
    Equity       BalanceSheetSection
    TotalAssets  float64
    TotalEquity  float64
    IsBalanced   bool
    Difference   float64
}
```

**Cara Kerja:**
- Mengambil semua account berdasarkan type (ASSET, LIABILITY, EQUITY)
- Menghitung balance dari journal entries
- Memeriksa apakah Assets = Liabilities + Equity

**⚠️ ISSUE:**
```go
// Di enhanced_report_controller.go line 61-72
func GetComprehensiveBalanceSheet() {
    // Delegate ke SSOT Balance Sheet controller
    ssotController := NewSSOTBalanceSheetController(erc.db)
    ssotController.GenerateSSOTBalanceSheet(c)
}
```
**MASALAH:** Enhanced Report hanya mendelegasikan ke SSOT tanpa menggunakan service sendiri

---

**2. Profit & Loss (Laba Rugi)**
```go
type ProfitLossData struct {
    Company               CompanyInfo
    StartDate             time.Time
    EndDate               time.Time
    Revenue               PLSection
    CostOfGoodsSold       PLSection
    GrossProfit           float64
    GrossProfitMargin     float64
    OperatingExpenses     PLSection
    OperatingIncome       float64
    NetIncome             float64
    NetIncomeMargin       float64
    ValidationReport      *ValidationReport
    DataQualityScore      float64
}
```

**Perhitungan:**
```
Gross Profit = Total Revenue - COGS
Operating Income = Gross Profit - Operating Expenses
Net Income = Operating Income + Other Income - Other Expenses - Tax
```

**✅ KELEBIHAN:**
- Ada **ValidationReport** untuk memvalidasi data quality
- Ada **DataQualityScore** untuk menilai keandalan data
- Support caching melalui `ReportCacheService`

**🔍 Fitur Validasi:**
```go
type ValidationReport struct {
    HealthScore       float64
    Recommendations   []string
    DataQualityIssues []DataQualityIssue
}
```

---

**3. Cash Flow Statement**
```go
type CashFlowData struct {
    OperatingActivities  CashFlowSection
    InvestingActivities  CashFlowSection
    FinancingActivities  CashFlowSection
    NetCashFlow          float64
    BeginningCash        float64
    EndingCash           float64
}
```

**Metode:** Indirect method
- Operating: Dari aktivitas operasional
- Investing: Pembelian/penjualan aset
- Financing: Hutang, modal, dividen

---

**4. Trial Balance**
```go
type TrialBalanceData struct {
    Accounts       []TrialBalanceItem
    TotalDebits    float64
    TotalCredits   float64
    IsBalanced     bool
    Difference     float64
    AssetSummary   AccountTypeSummary
    LiabilitySummary AccountTypeSummary
    // ...
}
```

**Validasi:**
```go
Difference = TotalDebits - TotalCredits
IsBalanced = abs(Difference) < 0.01  // Toleransi rounding
```

---

**5. General Ledger**
```go
type GeneralLedgerData struct {
    Account              Account
    OpeningBalance       float64
    ClosingBalance       float64
    TotalDebits          float64
    TotalCredits         float64
    NetPositionChange    float64
    Transactions         []GeneralLedgerEntry
    MonthlySummary       []MonthlyLedgerSummary
}
```

**Enhanced Features:**
- Monthly summary breakdown
- Running balance untuk setiap transaksi
- Net position status (increasing/decreasing)

---

#### B. Operational Reports

**1. Sales Summary**
```go
type SalesSummaryData struct {
    TotalRevenue           float64
    TotalTransactions      int64
    AverageOrderValue      float64
    TotalCustomers         int64
    NewCustomers           int64
    ReturningCustomers     int64
    SalesByPeriod          []PeriodData
    SalesByCustomer        []CustomerSalesData
    SalesByProduct         []ProductSalesData
    TopPerformers          TopPerformersData
    GrowthAnalysis         GrowthAnalysisData
    DataQualityScore       float64
}
```

**✅ FITUR UNGGULAN:**
- **Timezone-aware** (WIB/Asia Jakarta)
- **Data Quality Analysis** - Validasi kelengkapan dan kebenaran data
- **Growth Analysis** - Analisis pertumbuhan penjualan
- **Debug Info** - Information untuk troubleshooting

**Contoh Data Quality Check:**
```go
func analyzeDataQuality(sales []Sale) []string {
    var issues []string
    for _, sale := range sales {
        if sale.Code == "" {
            issues = append(issues, "Missing sale code")
        }
        if sale.TotalAmount < 0 {
            issues = append(issues, "Negative amount")
        }
        if sale.Date.After(time.Now()) {
            issues = append(issues, "Future date")
        }
    }
    return issues
}
```

---

**2. Purchase Report**

File: `ssot_purchase_report_service.go`

```go
type PurchaseReportData struct {
    TotalPurchases       int64
    CompletedPurchases   int64
    TotalAmount          float64
    TotalPaid            float64
    OutstandingPayables  float64
    PurchasesByVendor    []VendorPurchaseSummary
    PurchasesByMonth     []MonthlyPurchaseSummary
    PurchasesByCategory  []CategoryPurchaseSummary
    PaymentAnalysis      PurchasePaymentAnalysis
    TaxAnalysis          PurchaseTaxAnalysis
}
```

**✅ SSOT INTEGRATION:**
```sql
-- Query dari purchases table, BUKAN dari journal
SELECT 
    COUNT(DISTINCT p.id) as total_count,
    COALESCE(SUM(p.total_amount), 0) as total_amount,
    COALESCE(SUM(CASE 
        WHEN p.payment_method = 'CASH' 
        THEN p.total_amount
        ELSE 0 
    END), 0) as total_paid
FROM purchases p
WHERE p.date BETWEEN ? AND ?
  AND p.status = 'APPROVED'
```

**Payment Analysis:**
```go
type PurchasePaymentAnalysis struct {
    CashPurchases     int64
    CreditPurchases   int64
    CashAmount        float64
    CreditAmount      float64
    CashPercentage    float64
    CreditPercentage  float64
}
```

**Tax Analysis:**
```go
type PurchaseTaxAnalysis struct {
    TotalTaxableAmount     float64
    TotalTaxAmount         float64
    AverageTaxRate         float64  // (TotalTax / TotalTaxable) * 100
    TaxReclaimableAmount   float64  // Input tax is reclaimable
    TaxByMonth             []MonthlyTaxSummary
}
```

---

### 2.3 Report Validation & Quality

#### A. ReportValidationService

File: `report_validation_service.go`

**Fungsi:**
- Validasi integritas data report
- Detect anomali atau inkonsistensi
- Memberikan rekomendasi perbaikan

**Contoh Validasi:**
```go
type ValidationReport struct {
    HealthScore      float64
    Recommendations  []string
    DataQualityIssues []DataQualityIssue
    
    // Validation checks
    HasBalancedBooks     bool
    HasMissingAccounts   bool
    HasNegativeBalances  bool
    HasFutureDates       bool
}
```

#### B. Report Caching

File: `report_cache_service.go`

**Tujuan:**
- Mempercepat loading report yang sering diakses
- Mengurangi load database

**Implementation:**
```go
type ReportCacheService struct {
    cache map[string]CacheItem
    mutex sync.RWMutex
}

func (rcs *ReportCacheService) Get(key string) (interface{}, bool) {
    rcs.mutex.RLock()
    defer rcs.mutex.RUnlock()
    
    item, exists := rcs.cache[key]
    if !exists || item.IsExpired() {
        return nil, false
    }
    return item.Data, true
}
```

---

### 2.4 Export Functionality

#### A. PDF Export

File: `purchase_report_export_service.go`, `sales_report_export_service.go`

**Format Output:**
- PDF dengan header company info
- Table dengan styling
- Summary dan total

#### B. CSV Export

**Format:**
```csv
Date,Invoice,Vendor,Amount,Payment Method,Status
2025-01-01,INV-001,PT ABC,1000000,CASH,COMPLETED
```

#### C. Excel Export (Not yet implemented)

---

## 🔍 BAGIAN 3: MASALAH YANG DITEMUKAN

### 3.1 Masalah COA

#### ❌ CRITICAL ISSUES:

1. **Balance Sync Problems**
   
   **Evidence:**
   ```
   scripts/fix_coa_balance_sync.go
   scripts/fix_coa_parent_balance_calculation.go
   scripts/fix_cashbank_coa_balance_sync.go
   tools/diagnose_coa_balances.go
   ```
   
   **Root Cause:** COA balance disimpan terpisah dari journal entries
   
   **Solution:** 
   - COA balance harus **CALCULATED**, bukan stored
   - Gunakan view atau computed column
   ```sql
   CREATE VIEW vw_coa_balances AS
   SELECT 
       a.id,
       a.code,
       a.name,
       COALESCE(SUM(jl.debit_amount - jl.credit_amount), 0) as balance
   FROM accounts a
   LEFT JOIN journal_lines jl ON jl.account_id = a.id
   GROUP BY a.id, a.code, a.name;
   ```

2. **Multiple COA Services**
   
   Ada 3 service yang overlap:
   - `coa_service.go`
   - `coa_display_service.go`
   - `coa_display_service_v2.go`
   
   **Problem:** Duplicate logic, hard to maintain
   
   **Solution:** Consolidate menjadi satu service dengan proper versioning

3. **Hard-coded Account Codes**
   
   ```go
   // Di berbagai tempat
   accountCodes := []string{"1101", "1102", "1201", "2103", "4101"}
   ```
   
   **Solution:** 
   ```go
   const (
       ACCOUNT_KAS             = "1101"
       ACCOUNT_BANK            = "1102"
       ACCOUNT_PIUTANG         = "1201"
       ACCOUNT_PPN_KELUARAN    = "2103"
       ACCOUNT_PENDAPATAN      = "4101"
   )
   ```

---

### 3.2 Masalah Report

#### ❌ CRITICAL ISSUES:

1. **Enhanced Report Controller Tidak Digunakan**
   
   ```go
   // enhanced_report_controller.go line 60-72
   func (erc *EnhancedReportController) GetComprehensiveBalanceSheet(c *gin.Context) {
       // Just delegates to SSOT controller
       ssotController := NewSSOTBalanceSheetController(erc.db)
       ssotController.GenerateSSOTBalanceSheet(c)
   }
   ```
   
   **Problem:**
   - EnhancedReportService memiliki logic lengkap tapi tidak digunakan
   - Code duplication
   - Confusing untuk developer baru
   
   **Solution:** 
   - Hapus enhanced_report_controller atau
   - Gunakan sebagai facade pattern yang proper

2. **Inconsistent Date Handling**
   
   **Evidence di sales report:**
   ```go
   // Timezone-aware dengan logging detail
   utils.ReportLog.WithFields(utils.Fields{
       "input_start_date_utc": startDate.UTC().Format(...),
       "input_end_date_utc":   endDate.UTC().Format(...),
       "input_start_date_jkt": startDate.In(utils.JakartaTZ).Format(...),
   })
   ```
   
   **Tapi di purchase report:**
   ```go
   // Simple parsing tanpa timezone awareness
   startDate, err := time.Parse("2006-01-02", startDateStr)
   endDate, err := time.Parse("2006-01-02", endDateStr)
   ```
   
   **Solution:** Standardisasi timezone handling di semua report

3. **Purchase Report Menggunakan Purchase Table Langsung**
   
   ```sql
   -- Di ssot_purchase_report_service.go
   SELECT ... FROM purchases p
   LEFT JOIN unified_journal_ledger ujl ON ...
   ```
   
   **Problem:**
   - Seharusnya 100% dari unified_journal_ledger (SSOT)
   - LEFT JOIN membuat bisa ada data purchase tanpa journal
   - Inconsistent dengan prinsip SSOT
   
   **Solution:**
   ```sql
   -- Harus dari journal sebagai sumber tunggal
   SELECT ... FROM unified_journal_ledger ujl
   INNER JOIN purchases p ON ujl.source_id = p.id
   WHERE ujl.source_type = 'PURCHASE'
     AND ujl.status = 'POSTED'
   ```

4. **Company Info Scattered**
   
   Ada di berbagai tempat:
   - `CompanyProfile` model
   - `Settings` table
   - Hard-coded defaults
   
   **Solution:** Single source of truth untuk company info

---

### 3.3 Masalah Validasi

#### ⚠️ WARNINGS:

1. **Validation Purchase Report**
   
   ```go
   // ValidatePurchaseReport endpoint
   expectedOutstanding := report.TotalAmount - report.TotalPaid
   outstandingValid := abs(report.OutstandingPayables-expectedOutstanding) <= 0.01
   ```
   
   **Good:** Ada validasi
   **Problem:** Validasi dilakukan di controller, harusnya di service

2. **Data Quality Score**
   
   ```go
   func calculateDataQualityScore(sales []Sale) float64 {
       totalChecks := len(sales) * 5  // 5 checks per sale
       // Check: code, customer, amount, date, status
   }
   ```
   
   **Good:** Comprehensive checks
   **Problem:** 
   - Hard-coded weights
   - Tidak configurable
   - Tidak ada threshold untuk warning/error

---

## 💡 BAGIAN 4: REKOMENDASI PERBAIKAN

### 4.1 Short-term Fixes (1-2 minggu)

#### 1. Consolidate COA Services
```go
// NEW: coa_unified_service.go
type COAUnifiedService struct {
    db *gorm.DB
}

func (s *COAUnifiedService) GetWithDisplayFormat(activeOnly, nonZeroOnly bool) ([]COADisplayAccount, error) {
    // Single method dengan parameter
}
```

#### 2. Standardize Date Handling
```go
// utils/date_utils.go
type DateRange struct {
    Start time.Time
    End   time.Time
    TZ    *time.Location
}

func ParseDateRange(startStr, endStr string) (*DateRange, error) {
    // Standard parsing dengan timezone support
}
```

#### 3. Fix Purchase Report SSOT
```sql
-- Ganti query untuk 100% dari journal
SELECT 
    ujl.source_id as purchase_id,
    p.code as purchase_code,
    SUM(CASE WHEN ujl.entry_type = 'DEBIT' THEN ujl.amount ELSE 0 END) as total_amount
FROM unified_journal_ledger ujl
INNER JOIN purchases p ON ujl.source_id = p.id
WHERE ujl.source_type = 'PURCHASE'
  AND ujl.status = 'POSTED'
  AND ujl.entry_date BETWEEN ? AND ?
GROUP BY ujl.source_id, p.code
```

#### 4. Add Missing Indexes
```sql
-- Untuk performa report
CREATE INDEX idx_journal_lines_account_date ON journal_lines(account_id, entry_date);
CREATE INDEX idx_purchases_date_status ON purchases(date, status);
CREATE INDEX idx_sales_date_status ON sales(date, status);
```

---

### 4.2 Mid-term Improvements (1-2 bulan)

#### 1. Implement Proper Caching Strategy

```go
type ReportCacheConfig struct {
    EnableCache     bool
    TTL             time.Duration
    MaxSize         int
    InvalidateOn    []string  // Event types that invalidate cache
}

// Example
config := ReportCacheConfig{
    EnableCache: true,
    TTL: 1 * time.Hour,
    InvalidateOn: []string{"JOURNAL_POSTED", "SALE_APPROVED"},
}
```

#### 2. Add Report Scheduling

```go
type ReportSchedule struct {
    ID          uint
    ReportType  string
    Frequency   string  // DAILY, WEEKLY, MONTHLY
    Parameters  JSONB
    Recipients  []string
    IsActive    bool
}

// Cron job
func ScheduledReportGenerator() {
    // Generate and send reports automatically
}
```

#### 3. Implement Audit Trail

```go
type ReportAuditLog struct {
    ID              uint
    ReportType      string
    GeneratedBy     uint
    Parameters      JSONB
    ExecutionTime   time.Duration
    RecordCount     int
    GeneratedAt     time.Time
}
```

#### 4. Create Report Dashboard

- Real-time financial metrics
- Quick access ke report yang sering digunakan
- Graphical representation (charts)

---

### 4.3 Long-term Architecture (3-6 bulan)

#### 1. Microservices untuk Report

```
┌─────────────────┐
│  API Gateway    │
└────────┬────────┘
         │
    ┌────┴─────┐
    │          │
┌───▼───┐  ┌──▼────┐
│ COA   │  │Report │
│Service│  │Service│
└───┬───┘  └──┬────┘
    │         │
    └────┬────┘
         │
┌────────▼────────┐
│ Journal Service │
│    (SSOT)       │
└─────────────────┘
```

#### 2. Event-Driven Architecture

```go
type DomainEvent struct {
    Type      string
    Aggregate string
    ID        uint
    Timestamp time.Time
    Payload   interface{}
}

// Example events
const (
    EVENT_JOURNAL_POSTED      = "journal.posted"
    EVENT_SALE_APPROVED       = "sale.approved"
    EVENT_PURCHASE_COMPLETED  = "purchase.completed"
)

// Event handlers
func OnJournalPosted(event DomainEvent) {
    // Invalidate related caches
    // Update materialized views
    // Trigger report regeneration if scheduled
}
```

#### 3. Materialized Views untuk Performa

```sql
-- Pre-calculated balances
CREATE MATERIALIZED VIEW mv_account_balances AS
SELECT 
    a.id,
    a.code,
    a.name,
    a.type,
    COALESCE(SUM(jl.debit_amount - jl.credit_amount), 0) as balance,
    MAX(je.entry_date) as last_transaction_date
FROM accounts a
LEFT JOIN journal_lines jl ON jl.account_id = a.id
LEFT JOIN journal_entries je ON je.id = jl.journal_entry_id
WHERE je.status = 'POSTED'
GROUP BY a.id, a.code, a.name, a.type;

-- Refresh setiap ada journal baru
CREATE OR REPLACE FUNCTION refresh_account_balances()
RETURNS TRIGGER AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY mv_account_balances;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_refresh_balances
AFTER INSERT OR UPDATE ON journal_entries
FOR EACH STATEMENT
EXECUTE FUNCTION refresh_account_balances();
```

#### 4. Advanced Analytics

```go
type AnalyticsDimension struct {
    Period      string  // Day, Week, Month, Quarter, Year
    Segment     string  // Product, Customer, Region
    Metric      string  // Revenue, Profit, Growth
    Value       float64
    PrevValue   float64
    Change      float64
    ChangeRate  float64
}

// ML-based forecasting
type ForecastModel struct {
    Type        string  // LINEAR, ARIMA, PROPHET
    Confidence  float64
    Predictions []ForecastPoint
}
```

---

## 📊 BAGIAN 5: METRICS & MONITORING

### 5.1 Report Performance Metrics

```go
type ReportMetrics struct {
    ReportType      string
    AverageTime     time.Duration
    P95Time         time.Duration
    P99Time         time.Duration
    ErrorRate       float64
    CacheHitRate    float64
    QueriesExecuted int
}
```

### 5.2 Data Quality Metrics

```go
type DataQualityMetrics struct {
    CompletenessScore  float64  // % of required fields filled
    AccuracyScore      float64  // % of validated records
    ConsistencyScore   float64  // % of consistent relationships
    TimelinessScore    float64  // % of recent data
    OverallScore       float64  // Weighted average
}
```

### 5.3 Monitoring Dashboard

**Metrics to Track:**
1. Report generation time
2. Database query performance
3. Cache hit ratio
4. Error rates
5. User engagement (most accessed reports)
6. Data quality trends

---

## 🚀 BAGIAN 6: MIGRATION PLAN

### Phase 1: Stabilization (Week 1-2)
- [ ] Fix critical COA sync issues
- [ ] Standardize date handling
- [ ] Add missing database indexes
- [ ] Fix SSOT purchase report queries

### Phase 2: Consolidation (Week 3-4)
- [ ] Merge COA services
- [ ] Remove unused code (enhanced report controller delegation)
- [ ] Standardize company info source
- [ ] Implement proper error handling

### Phase 3: Enhancement (Week 5-8)
- [ ] Implement comprehensive caching
- [ ] Add report scheduling
- [ ] Implement audit trail
- [ ] Create analytics dashboard

### Phase 4: Optimization (Week 9-12)
- [ ] Implement materialized views
- [ ] Optimize database queries
- [ ] Add batch processing for large reports
- [ ] Implement report archiving

---

## 📚 BAGIAN 7: DOCUMENTATION NEEDS

### 7.1 API Documentation
- Swagger/OpenAPI specs untuk semua endpoints
- Request/response examples
- Error codes documentation

### 7.2 Business Logic Documentation
- Accounting rules (debit/credit logic)
- Balance calculation methods
- Report generation algorithms
- Data validation rules

### 7.3 Operations Documentation
- Deployment procedures
- Database migration scripts
- Backup and recovery procedures
- Performance tuning guide

---

## ✅ KESIMPULAN

### Kondisi Saat Ini:

**✅ KELEBIHAN:**
1. Sistem COA yang flexible dengan support hierarchy
2. Report system yang comprehensive
3. SSOT integration untuk konsistensi data
4. Validation dan data quality checks
5. Multiple export formats (JSON, PDF, CSV)
6. Timezone-aware date handling (di beberapa bagian)

**⚠️ MASALAH:**
1. Balance sync issues (banyak script perbaikan)
2. Code duplication (multiple COA services)
3. Inconsistent SSOT implementation (purchase report)
4. Hard-coded values scattered
5. Incomplete transition dari Enhanced ke SSOT
6. Missing documentation

### Priority Fixes:

1. **CRITICAL:** Fix COA balance sync - migrate to calculated balances
2. **HIGH:** Complete SSOT migration untuk purchase report
3. **HIGH:** Consolidate COA services
4. **MEDIUM:** Standardize date/timezone handling
5. **MEDIUM:** Remove code duplication
6. **LOW:** Add advanced analytics

### Recommended Next Steps:

1. **Immediate (This Week):**
   - Document current architecture
   - Fix critical balance sync bugs
   - Add monitoring untuk track issues

2. **Short-term (This Month):**
   - Implement Phase 1 & 2 dari migration plan
   - Create comprehensive test suite
   - Update API documentation

3. **Long-term (This Quarter):**
   - Complete SSOT migration
   - Implement caching strategy
   - Add advanced analytics

---

**Dibuat:** 11 November 2025  
**Author:** AI Assistant  
**Version:** 1.0  
**Status:** Draft untuk Review
