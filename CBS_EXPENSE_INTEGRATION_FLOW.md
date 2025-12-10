# Analisis Integrasi CBS - Expenses - Budget vs Actual

## 📊 Gambaran Umum Sistem

Aplikasi Anda memiliki 3 modul utama yang saling terkait:

1. **CBS (Cost Breakdown Structure)** - `/cost-control/cbs`
2. **Expense Transactions** - `/cost-control/expenses`
3. **Budget vs Actual Report** - `/cost-control/budget-vs-actual`

## 🔗 Struktur Data & Relasi

### 1. CBS Node (`cbs_nodes`)
```
- ID, ProjectID, ParentID (tree structure)
- Code, Name, Description
- COAAccountID (link ke Chart of Account)
- BudgetAmount (budget yang dialokasikan)
- IsActive
```

### 2. COA Account (`coa_accounts`)
```
- ID, Code, Name
- Type (ASSET, LIABILITY, EXPENSE, etc)
- Category (Material, Labor, Equipment, etc)
- BudgetCategory (LABOUR_BUDGET, OPERASIONAL_BUDGET, OTHER)
- WorkPackage (Pekerjaan Persiapan, Pekerjaan Beton, etc)
- ParentID (tree structure)
- IsHeader (header tidak bisa digunakan untuk transaksi)
```

### 3. Project Budget (`project_budgets`)
```
- ProjectID, AccountID (COA)
- EstimatedAmount (budget estimasi per COA)
```

### 4. Expense Transaction (`expense_transactions`)
```
- ProjectID, TransactionDate
- COAAccountID (link ke COA)
- Description, Amount, Unit, Quantity
- TransactionType (LABOUR, MATERIAL, OPERATIONAL, OTHER)
- ReferenceType (PR, PO, MANUAL, CBS)
- ReferenceID, ReferenceNo
```

## 🔄 Flow Integrasi yang Direkomendasikan

### **FASE 1: Setup Budget (CBS → Project Budget)**

```
CBS Page
   ↓
1. User membuat CBS Tree Structure
   - Root: Project
   - Level 1: Major Work Packages (Pekerjaan Persiapan, Struktur, dll)
   - Level 2: Sub Work Packages
   - Level 3: Detail Activities
   
2. Setiap CBS Node di-link ke COA Account
   - CBS Node.COAAccountID → COA Account
   - CBS Node.BudgetAmount → Budget untuk node tersebut
   
3. Budget dari CBS Node otomatis create/update Project Budget
   - Saat CBS Node disimpan dengan BudgetAmount
   - Create/Update entry di project_budgets
   - project_budgets.AccountID = CBS.COAAccountID
   - project_budgets.EstimatedAmount = CBS.BudgetAmount
```

**Implementasi yang Diperlukan:**
- Tambahkan logic di `CBSService.Create/Update` untuk sync ke `project_budgets`
- Saat CBS node dengan COAAccountID & BudgetAmount disimpan → upsert project_budgets

---

### **FASE 2: Recording Expenses**

```
Expenses Page
   ↓
1. User pilih Project
2. User create Expense Transaction
   - Pilih COA Account (dari dropdown, filter: IsHeader=false)
   - Input Amount, Date, Description
   - ReferenceType bisa:
     * MANUAL (input manual)
     * CBS (dari CBS allocation)
     * PR (dari Purchase Request)
     * PO (dari Purchase Order)
   
3. Expense Transaction tersimpan dengan link ke:
   - ProjectID
   - COAAccountID
   - ReferenceType & ReferenceID (optional)
```

**Fitur Tambahan yang Bisa Ditambahkan:**
- **Quick Entry dari CBS**: Di CBS Tree, tambahkan button "Add Expense" per node
  - Auto-fill COAAccountID dari CBS Node
  - Auto-fill ReferenceType = "CBS"
  - Auto-fill ReferenceID = CBS Node ID
  
- **Batch Import**: Import expenses dari Excel/CSV
  - Mapping COA Code → COA Account
  - Bulk insert expense transactions

---

### **FASE 3: Budget vs Actual Report**

```
Budget vs Actual Page
   ↓
1. User pilih Project & Date Range
2. System generate report dengan query:
   
   FOR EACH COA Account in Project:
     Budget = SUM(project_budgets.EstimatedAmount) 
              WHERE ProjectID = X AND AccountID = COA.ID
     
     Actual = SUM(expense_transactions.Amount)
              WHERE ProjectID = X 
              AND COAAccountID = COA.ID
              AND TransactionDate BETWEEN start_date AND end_date
     
     Variance = Budget - Actual
     
3. Group by BudgetCategory:
   - LABOUR_BUDGET
   - OPERASIONAL_BUDGET (group by WorkPackage)
   - OTHER
   
4. Show details:
   - Per COA Account
   - Per Work Package (for OPERASIONAL)
   - Transaction details
```

**Report sudah implemented dengan baik!** ✅

---

## 🎯 Rekomendasi Integrasi Flow

### **Option A: CBS-Centric (Recommended)**

CBS sebagai master budget allocation:

```
1. Setup CBS Tree
   ├─ Link setiap node ke COA Account
   ├─ Set BudgetAmount per node
   └─ Auto-sync ke project_budgets
   
2. Record Expenses
   ├─ Bisa dari CBS (quick entry)
   ├─ Bisa dari PR/PO (auto-create)
   └─ Bisa manual entry
   
3. View Report
   ├─ Budget dari project_budgets (sourced from CBS)
   ├─ Actual dari expense_transactions
   └─ Variance analysis
```

**Keuntungan:**
- CBS menjadi single source of truth untuk budget
- Traceability jelas: CBS → Budget → Expenses
- Mudah tracking per work package
- Support hierarchical budget rollup

---

### **Option B: Dual Entry (Flexible)**

CBS dan Project Budget terpisah:

```
1. Setup Budget
   ├─ Option 1: Via CBS (structured)
   └─ Option 2: Direct project_budgets entry (quick)
   
2. Record Expenses (sama seperti Option A)

3. View Report (sama seperti Option A)
```

**Keuntungan:**
- Lebih fleksibel
- Bisa budget tanpa CBS (untuk project sederhana)
- CBS optional untuk project kompleks

---

## 🛠️ Implementasi yang Diperlukan

### 1. **Backend: CBS → Project Budget Sync**

File: `backend/services/cbs_service.go`

```go
func (s *cbsService) Create(dto *CreateCBSNodeDTO) (*models.CBSNode, error) {
    // ... existing code ...
    
    // Sync to project_budgets if COAAccountID and BudgetAmount exist
    if node.COAAccountID != nil && node.BudgetAmount > 0 {
        err := s.syncToProjectBudget(node)
        if err != nil {
            log.Printf("Warning: Failed to sync CBS to project budget: %v", err)
        }
    }
    
    return node, nil
}

func (s *cbsService) syncToProjectBudget(node *models.CBSNode) error {
    // Upsert project_budgets
    budget := &models.ProjectBudget{
        ProjectID:       node.ProjectID,
        AccountID:       *node.COAAccountID,
        EstimatedAmount: float64(node.BudgetAmount),
    }
    
    return s.projectBudgetRepo.Upsert(budget)
}
```

### 2. **Frontend: Quick Expense Entry dari CBS**

File: `frontend/src/components/cost-control/CBSTreeView.tsx`

Tambahkan button "Add Expense" per CBS node:

```tsx
<IconButton
  icon={<FiDollarSign />}
  aria-label="Add Expense"
  size="sm"
  onClick={() => onAddExpense(node)}
  title="Quick add expense for this CBS item"
/>
```

Handler di `frontend/app/cost-control/cbs/page.tsx`:

```tsx
const handleAddExpense = (node: CBSNode) => {
  // Navigate to expenses page with pre-filled data
  router.push(`/cost-control/expenses?project=${node.project_id}&coa=${node.coa_account_id}&cbs=${node.id}`);
};
```

### 3. **Frontend: CBS Budget Summary Widget**

Di CBS page, tambahkan summary card:

```tsx
<Box>
  <Stat>
    <StatLabel>Total Budget (from CBS)</StatLabel>
    <StatNumber>{formatCurrency(totalBudget)}</StatNumber>
  </Stat>
  <Stat>
    <StatLabel>Total Actual</StatLabel>
    <StatNumber>{formatCurrency(totalActual)}</StatNumber>
  </Stat>
  <Stat>
    <StatLabel>Variance</StatLabel>
    <StatNumber color={variance >= 0 ? 'green' : 'red'}>
      {formatCurrency(variance)}
    </StatNumber>
  </Stat>
</Box>
```

### 4. **Navigation Flow Enhancement**

Tambahkan breadcrumb/link antar halaman:

```
CBS Page:
  → "View Budget Report" button → Budget vs Actual
  → "View Expenses" button → Expenses (filtered by project)
  → Per node: "Add Expense" → Expenses (pre-filled)

Expenses Page:
  → "View Budget Report" button → Budget vs Actual
  → "View CBS Structure" button → CBS

Budget vs Actual Page:
  → "View CBS Structure" button → CBS
  → "View Expenses" button → Expenses
  → Per COA line: Click → Expenses (filtered by COA)
```

---

## 📈 Data Flow Diagram

```
┌─────────────────┐
│   CBS TREE      │
│  (Structure)    │
│                 │
│ - Work Packages │
│ - Activities    │
│ - Link to COA   │
│ - Budget Amount │
└────────┬────────┘
         │
         │ Auto-sync
         ↓
┌─────────────────┐
│ PROJECT BUDGETS │
│   (Estimates)   │
│                 │
│ - Per COA       │
│ - Amount        │
└────────┬────────┘
         │
         │ Compare with
         ↓
┌─────────────────┐      ┌──────────────┐
│    EXPENSES     │←─────│  PR/PO       │
│ (Transactions)  │      │  (Auto)      │
│                 │      └──────────────┘
│ - Per COA       │
│ - Amount        │      ┌──────────────┐
│ - Date          │←─────│  CBS         │
│ - Reference     │      │  (Quick)     │
└────────┬────────┘      └──────────────┘
         │
         │               ┌──────────────┐
         └───────────────│  Manual      │
                         │  (Direct)    │
                         └──────────────┘
         │
         ↓
┌─────────────────┐
│ BUDGET REPORT   │
│ (Analysis)      │
│                 │
│ - Budget        │
│ - Actual        │
│ - Variance      │
│ - By Category   │
│ - By Work Pkg   │
└─────────────────┘
```

---

## 🎨 UI/UX Recommendations

### 1. **CBS Page Enhancements**
- Show budget vs actual per node (inline)
- Color coding: green (under budget), red (over budget)
- Progress bar per node
- Quick expense entry button per node

### 2. **Expenses Page Enhancements**
- Filter by CBS Node (dropdown)
- Filter by Work Package
- Bulk import from Excel
- Template download
- Quick entry form (minimal fields)

### 3. **Budget vs Actual Page Enhancements**
- Export to Excel/PDF
- Drill-down: Click COA → Show transactions
- Drill-down: Click Work Package → Show CBS nodes
- Trend chart (budget vs actual over time)
- Alert indicators (over budget items)

### 4. **Cross-Module Navigation**
- Floating action button: "Quick Actions"
  - Add Expense
  - View Report
  - View CBS
- Contextual links (e.g., from report → expenses → CBS)

---

## 🚀 Implementation Priority

### **Phase 1: Core Integration** (High Priority)
1. ✅ Expense Transaction model (DONE)
2. ✅ Budget vs Actual report (DONE)
3. 🔲 CBS → Project Budget sync
4. 🔲 Quick expense entry from CBS

### **Phase 2: Enhanced Features** (Medium Priority)
5. 🔲 Expense filtering by CBS node
6. 🔲 Budget summary widget in CBS page
7. 🔲 Cross-module navigation links
8. 🔲 Drill-down in reports

### **Phase 3: Advanced Features** (Low Priority)
9. 🔲 Bulk expense import
10. 🔲 Trend analysis charts
11. 🔲 Budget alerts & notifications
12. 🔲 Export reports (Excel/PDF)

---

## 💡 Key Insights

1. **COA adalah penghubung utama** antara CBS, Budget, dan Expenses
2. **CBS bisa menjadi master budget** dengan auto-sync ke project_budgets
3. **Expenses bisa berasal dari multiple sources**: CBS, PR/PO, Manual
4. **Budget vs Actual report sudah bagus**, tinggal enhance dengan drill-down
5. **Navigation flow** perlu diperbaiki untuk seamless user experience

---

## 📝 Next Steps

Apakah Anda ingin saya implementasikan:
1. **CBS → Project Budget sync** (backend logic)?
2. **Quick expense entry dari CBS** (frontend integration)?
3. **Enhanced navigation** antar modul?
4. **Budget summary widget** di CBS page?

Atau ada prioritas lain yang ingin Anda fokuskan?
