# Flow Aplikasi untuk Budget Report

## Analisis Struktur Aplikasi Saat Ini

### 1. **Data Master yang Sudah Ada**
- ✅ **COA (Chart of Accounts)** - Sudah diupdate dengan `budget_category` dan `work_package`
- ✅ **Materials** - Master material dengan link ke COA
- ✅ **Vendors** - Master vendor
- ✅ **CBS (Cost Breakdown Structure)** - Struktur breakdown biaya per project dengan link ke COA

### 2. **Transaksi yang Sudah Ada**
- ✅ **Purchase Request (PR)** - Request pembelian dengan items
- ✅ **PR Items** - Detail item PR dengan link ke Material
- ✅ **PR CBS Mapping** - Mapping PR ke CBS Node
- ✅ **Project Budget** - Budget per project per COA account
- ✅ **Project Actual Cost** - View untuk actual cost dari berbagai sumber

### 3. **Yang Masih Kurang untuk Report Budget**
- ❌ **Transaction/Journal Entry** - Untuk mencatat transaksi harian seperti di referensi
- ❌ **Expense Transaction** - Untuk mencatat biaya operasional (bensin, tol, makan, dll)
- ❌ **Labour Transaction** - Untuk mencatat biaya tenaga kerja (mandor, tukang, dll)
- ❌ **Link PR ke COA** - PR items perlu link langsung ke COA untuk kategorisasi

---

## Flow untuk Membentuk Report Budget

### **FLOW 1: Setup Master Data**

```
1. Setup COA dengan Budget Category
   ├─ 5100 - LABOUR BUDGET (budget_category: LABOUR_BUDGET)
   │  ├─ 5101 - Mandor Civil & MEP
   │  ├─ 5102 - Tukang Bangunan
   │  └─ 5106 - Kompensasi & Kasbon
   │
   ├─ 5200 - OPERASIONAL BUDGET (budget_category: OPERASIONAL_BUDGET)
   │  ├─ 5201 - Pekerjaan Persiapan (work_package: PEKERJAAN PERSIAPAN)
   │  ├─ 5202 - Pekerjaan Tanah dan Pasir
   │  └─ 5213 - Pekerjaan Finishing
   │
   └─ 5300 - BIAYA OPERASIONAL LAINNYA (budget_category: OTHER)
      ├─ 5301 - Transportasi
      ├─ 5302 - Akomodasi
      └─ 5309 - Lain-lain

2. Setup Materials dengan link ke COA
   - Material "Semen" → COA 5203 (Pasangan dan Plesteran)
   - Material "Besi Beton" → COA 5204 (Pekerjaan Beton)

3. Setup CBS per Project dengan link ke COA
   - CBS Node "Pekerjaan Struktur" → COA 5204
   - CBS Node "Pekerjaan Finishing" → COA 5213
```

---

### **FLOW 2: Budget Planning (Estimation)**

```
1. Project Manager membuat Budget per COA
   ├─ Project: "Padel Court Bandung"
   ├─ COA 5101 (Mandor) → Budget: Rp 40,000,000
   ├─ COA 5201 (Prelim) → Budget: Rp 44,086,151
   └─ COA 5301 (Transportasi) → Budget: Rp 5,000,000

2. Budget disimpan di table: project_budgets
   - project_id
   - account_id (COA ID)
   - estimated_amount
```

---

### **FLOW 3: Transaksi Operasional (ACTUAL)**

#### **A. Labour Budget (Tenaga Kerja)**

```
1. User mencatat pembayaran tenaga kerja
   ├─ Date: 2025-08-16
   ├─ Description: "Pelunasan DP 30% (Mandor 2)"
   ├─ COA: 5101 (Mandor Civil & MEP)
   ├─ Amount: Rp 38,500,000
   └─ Unit: ls (lump sum)

2. Data disimpan di table: expense_transactions
   - project_id
   - transaction_date
   - coa_account_id
   - description
   - amount
   - unit
   - quantity
   - transaction_type: 'LABOUR'
   - reference_no
```

#### **B. Operasional Budget (Material & Pekerjaan)**

```
1. User membuat Purchase Request
   ├─ PR untuk Material Semen
   ├─ PR Item → Material ID → COA (via material.coa_account_id)
   └─ PR CBS Mapping → CBS Node → COA

2. Setelah PR Approved & PO Created
   ├─ Actual cost tercatat di project_actual_costs view
   └─ Linked ke COA via CBS Node atau Material

3. Untuk transaksi langsung (non-PR)
   ├─ User input manual expense
   ├─ Date: 2025-08-16
   ├─ Description: "Bensin"
   ├─ COA: 5301 (Transportasi)
   ├─ Amount: Rp 60,000
   └─ Saved to: expense_transactions
```

#### **C. Biaya Operasional Lainnya**

```
1. User mencatat biaya harian
   ├─ Date: 2025-08-16
   ├─ Description: "Top Up Tol"
   ├─ COA: 5301 (Transportasi)
   ├─ Amount: Rp 27,500
   └─ Unit: ls

2. Semua transaksi masuk ke: expense_transactions
```

---

### **FLOW 4: Generate Report Budget vs Actual**

```sql
-- Query untuk Labour Budget Report
SELECT 
    et.transaction_date as date,
    et.description,
    et.unit,
    et.amount as total_price,
    coa.code,
    coa.name as coa_name,
    coa.work_package,
    pb.estimated_amount as budget_estimation,
    SUM(et.amount) OVER (PARTITION BY et.coa_account_id) as actual_total
FROM expense_transactions et
JOIN coa_accounts coa ON et.coa_account_id = coa.id
LEFT JOIN project_budgets pb ON pb.project_id = et.project_id 
    AND pb.account_id = et.coa_account_id
WHERE et.project_id = ?
    AND coa.budget_category = 'LABOUR_BUDGET'
    AND et.transaction_date BETWEEN ? AND ?
ORDER BY et.transaction_date, coa.code;

-- Query untuk Operasional Budget Report
SELECT 
    et.transaction_date as date,
    et.description,
    et.unit,
    et.amount as total_price,
    coa.code,
    coa.name as coa_name,
    coa.work_package,
    pb.estimated_amount as budget_estimation,
    SUM(et.amount) OVER (PARTITION BY coa.work_package) as actual_by_work_package
FROM expense_transactions et
JOIN coa_accounts coa ON et.coa_account_id = coa.id
LEFT JOIN project_budgets pb ON pb.project_id = et.project_id 
    AND pb.account_id = et.coa_account_id
WHERE et.project_id = ?
    AND coa.budget_category = 'OPERASIONAL_BUDGET'
    AND et.transaction_date BETWEEN ? AND ?
ORDER BY et.transaction_date, coa.work_package, coa.code;
```

---

## Struktur Report Output

### **Format Report (Sesuai Referensi)**

```
┌─────────────────────────────────────────────────────────────────────────┐
│                         LABOUR BUDGET                                    │
├──────────┬─────────────────────┬──────┬──────────────┬──────────────────┤
│   DATE   │    DESCRIPTION      │ UNIT │ TOTAL PRICE  │ BUDGET ESTIMATION│
├──────────┼─────────────────────┼──────┼──────────────┼──────────────────┤
│ 8/16/2025│ Mandor 01 Civil&MEP │  ls  │ 20,000,000   │                  │
│ 8/16/2025│ Mandor 02 Civil&MEP │  ls  │ 20,000,000   │                  │
│ 8/16/2025│ Pelunasan DP 30%    │  ls  │ 38,500,000   │                  │
├──────────┴─────────────────────┴──────┴──────────────┼──────────────────┤
│                                    TOTAL ACTUAL       │   466,703,196    │
│                                    BUDGET ESTIMATION  │   492,016,432    │
│                                    BALANCING BUDGET   │    25,313,236    │
└───────────────────────────────────────────────────────┴──────────────────┘

┌─────────────────────────────────────────────────────────────────────────┐
│                      OPERASIONAL BUDGET                                  │
├──────────┬─────────────────────┬──────┬──────────────┬──────────────────┤
│   DATE   │    DESCRIPTION      │ UNIT │ TOTAL PRICE  │ BUDGET ESTIMATION│
├──────────┼─────────────────────┼──────┼──────────────┼──────────────────┤
│          │ PEKERJAAN PERSIAPAN │      │              │   44,086,151     │
│ 8/16/2025│ Prelim Air Kerja    │  ls  │  37,790,994  │                  │
│ 8/16/2025│ Prelim              │  ls  │   7,476,500  │                  │
├──────────┼─────────────────────┼──────┼──────────────┼──────────────────┤
│          │ PEKERJAAN BETON     │      │              │   23,155,915     │
│ 8/27/2025│ Pekerjaan Lapangan  │  ls  │  29,306,160  │                  │
└──────────┴─────────────────────┴──────┴──────────────┴──────────────────┘
```

---

## Database Schema yang Dibutuhkan

### **Table: expense_transactions** (BARU - Perlu dibuat)

```sql
CREATE TABLE expense_transactions (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id),
    transaction_date DATE NOT NULL,
    coa_account_id INTEGER NOT NULL REFERENCES coa_accounts(id),
    description VARCHAR(500) NOT NULL,
    amount DECIMAL(15,2) NOT NULL,
    unit VARCHAR(20) DEFAULT 'ls',
    quantity DECIMAL(10,2) DEFAULT 1,
    transaction_type VARCHAR(30), -- LABOUR, MATERIAL, OPERATIONAL, OTHER
    reference_no VARCHAR(50), -- PR number, PO number, or manual ref
    notes TEXT,
    created_by INTEGER NOT NULL REFERENCES users(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    INDEX idx_expense_project (project_id),
    INDEX idx_expense_coa (coa_account_id),
    INDEX idx_expense_date (transaction_date),
    INDEX idx_expense_type (transaction_type)
);
```

### **Update: purchase_request_items** (Tambah COA link)

```sql
ALTER TABLE purchase_request_items 
ADD COLUMN coa_account_id INTEGER REFERENCES coa_accounts(id);

CREATE INDEX idx_pr_items_coa ON purchase_request_items(coa_account_id);
```

---

## API Endpoints yang Dibutuhkan

### **1. Expense Transaction Management**

```
POST   /api/v1/projects/:projectId/expenses
GET    /api/v1/projects/:projectId/expenses
GET    /api/v1/projects/:projectId/expenses/:id
PUT    /api/v1/projects/:projectId/expenses/:id
DELETE /api/v1/projects/:projectId/expenses/:id
```

### **2. Budget Report**

```
GET /api/v1/projects/:projectId/reports/budget-vs-actual
    Query params:
    - start_date
    - end_date
    - budget_category (LABOUR_BUDGET, OPERASIONAL_BUDGET, OTHER)
    - work_package (optional)
    - coa_code (optional)

Response:
{
  "project_id": 1,
  "project_name": "Padel Court Bandung",
  "report_date": "2025-12-07",
  "start_date": "2025-08-01",
  "end_date": "2025-10-31",
  "labour_budget": {
    "budget_estimation": 492016432,
    "actual": 466703196,
    "variance": 25313236,
    "transactions": [
      {
        "date": "2025-08-16",
        "description": "Mandor 01 Civil & MEP",
        "unit": "ls",
        "total_price": 20000000,
        "coa_code": "5101",
        "coa_name": "Mandor Civil & MEP"
      }
    ]
  },
  "operasional_budget": {
    "budget_estimation": 732959158,
    "actual": 0,
    "variance": 732959158,
    "by_work_package": [
      {
        "work_package": "PEKERJAAN PERSIAPAN",
        "budget_estimation": 44086151,
        "actual": 45267494,
        "variance": -1181343,
        "transactions": [...]
      }
    ]
  }
}
```

---

## Frontend Components yang Dibutuhkan

### **1. Expense Transaction Form**

```typescript
// frontend/src/components/cost-control/ExpenseTransactionForm.tsx
- Input: Date, Description, COA (dropdown), Amount, Unit, Quantity
- Auto-calculate total
- Link to PR/PO (optional)
- Transaction type selection
```

### **2. Budget Report Viewer**

```typescript
// frontend/app/cost-control/budget-report/page.tsx
- Filter: Project, Date Range, Budget Category
- Display: Labour Budget table, Operasional Budget table
- Summary: Total Budget, Total Actual, Variance
- Export: Excel, PDF
```

### **3. Daily Expense Entry**

```typescript
// frontend/src/components/cost-control/DailyExpenseEntry.tsx
- Quick entry form for daily expenses
- Batch entry support
- Mobile-friendly
```

---

## Implementation Priority

### **Phase 1: Database & Backend** (High Priority)
1. ✅ Update COA with budget_category & work_package (DONE)
2. ⏳ Create expense_transactions table
3. ⏳ Create ExpenseTransaction model
4. ⏳ Create ExpenseTransaction repository, service, controller
5. ⏳ Create Budget Report service & endpoint
6. ⏳ Update PR items to link to COA

### **Phase 2: Frontend** (Medium Priority)
1. ⏳ Create Expense Transaction Form
2. ⏳ Create Expense Transaction List
3. ⏳ Create Budget Report Page
4. ⏳ Update PR form to select COA per item

### **Phase 3: Integration** (Medium Priority)
1. ⏳ Auto-create expense transaction from approved PR
2. ⏳ Link CBS to expense transactions
3. ⏳ Real-time budget tracking

### **Phase 4: Enhancement** (Low Priority)
1. ⏳ Budget alert & notification
2. ⏳ Budget approval workflow
3. ⏳ Multi-currency support
4. ⏳ Budget forecasting

---

## Summary

Untuk membentuk report budget seperti referensi, flow-nya adalah:

1. **Setup COA** dengan `budget_category` (LABOUR_BUDGET, OPERASIONAL_BUDGET, OTHER) dan `work_package`
2. **Buat table `expense_transactions`** untuk mencatat semua transaksi harian
3. **Link semua transaksi ke COA** (PR items, expense transactions, dll)
4. **Query data** berdasarkan project, date range, dan budget_category
5. **Group by work_package** untuk operasional budget
6. **Calculate variance** antara budget estimation dan actual

Struktur COA yang sudah diupdate akan menjadi backbone untuk kategorisasi semua transaksi, sehingga report bisa di-generate secara otomatis berdasarkan data transaksi yang tercatat.
