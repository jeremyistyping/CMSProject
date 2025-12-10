# Implementasi Integrasi CBS - Expenses - Budget vs Actual

## ✅ Fitur yang Telah Diimplementasikan

### 1. **Backend: CBS → Project Budget Auto-Sync**

#### File Baru:
- `backend/repositories/project_budget_repository.go` - Repository untuk mengelola project budgets

#### File yang Diupdate:
- `backend/services/cbs_service.go`
  - Menambahkan `ProjectBudgetRepository` dependency
  - Method `syncToProjectBudget()` - Auto-sync CBS budget ke project_budgets
  - Method `GetProjectBudgetSummary()` - Mendapatkan summary budget per project
  - Update `CreateCBSNode()` dan `UpdateCBSNode()` untuk trigger auto-sync

- `backend/controllers/cbs_controller.go`
  - Method `GetProjectBudgetSummary()` - Endpoint untuk mendapatkan budget summary

- `backend/routes/cbs_routes.go`
  - Route baru: `GET /api/v1/projects/:id/cbs/summary`

- `backend/routes/routes.go` & `backend/routes/project_routes.go`
  - Inisialisasi `ProjectBudgetRepository`
  - Inject ke `CBSService`

#### Cara Kerja:
```
1. User membuat/update CBS Node dengan:
   - COAAccountID (link ke Chart of Account)
   - BudgetAmount (budget allocation)

2. System otomatis:
   - Upsert entry di table project_budgets
   - project_budgets.ProjectID = CBS.ProjectID
   - project_budgets.AccountID = CBS.COAAccountID
   - project_budgets.EstimatedAmount = CBS.BudgetAmount

3. Budget dari CBS menjadi source untuk Budget vs Actual Report
```

---

### 2. **Frontend: Budget Summary Widget di CBS Page**

#### File Baru:
- `frontend/src/components/cost-control/CBSBudgetSummary.tsx`
  - Widget untuk menampilkan summary budget
  - Menampilkan: Total Budget, Total Actual, Variance, Utilization %
  - Color coding: green (under budget), red (over budget)
  - Real-time data dari API

#### File yang Diupdate:
- `frontend/src/services/cbsService.ts`
  - Method `getProjectBudgetSummary()` - Fetch budget summary dari API

- `frontend/app/cost-control/cbs/page.tsx`
  - Import `CBSBudgetSummary` component
  - Menampilkan budget summary di atas CBS tree
  - Tambahkan navigation buttons ke Expenses dan Budget Report

#### Tampilan:
```
┌─────────────────────────────────────────────────────────┐
│ Budget Summary                                          │
├──────────────┬──────────────┬──────────────┬───────────┤
│ Total Budget │ Total Actual │ Variance     │ Utilization│
│ Rp 500.000K  │ Rp 350.000K  │ +Rp 150.000K │ 70%       │
│ 15 CBS nodes │ From PR      │ ↑ 30%        │ Under     │
└──────────────┴──────────────┴──────────────┴───────────┘
```

---

### 3. **Frontend: Enhanced Navigation Antar Modul**

#### CBS Page (`/cost-control/cbs`):
- ✅ Button "View Expenses" → Navigate ke `/cost-control/expenses`
- ✅ Button "Budget Report" → Navigate ke `/cost-control/budget-vs-actual`
- ✅ Budget Summary Widget (real-time)

#### Expenses Page (`/cost-control/expenses`):
- ✅ Button "View CBS" → Navigate ke `/cost-control/cbs`
- ✅ Button "Budget Report" → Navigate ke `/cost-control/budget-vs-actual`

#### Budget vs Actual Page (`/cost-control/budget-vs-actual`):
- ✅ Button "View CBS" → Navigate ke `/cost-control/cbs`
- ✅ Button "View Expenses" → Navigate ke `/cost-control/expenses`

---

## 🔄 Data Flow yang Telah Diimplementasikan

```
┌─────────────────┐
│   CBS TREE      │
│                 │
│ User creates/   │
│ updates node    │
│ with:           │
│ - COAAccountID  │
│ - BudgetAmount  │
└────────┬────────┘
         │
         │ AUTO-SYNC (Backend)
         ↓
┌─────────────────┐
│ PROJECT BUDGETS │
│                 │
│ Upsert:         │
│ - ProjectID     │
│ - AccountID     │
│ - EstimatedAmt  │
└────────┬────────┘
         │
         │ Used by
         ↓
┌─────────────────┐
│ BUDGET REPORT   │
│                 │
│ Compare:        │
│ - Budget (from  │
│   project_budgets)│
│ - Actual (from  │
│   expense_trans)│
└─────────────────┘
```

---

## 📊 API Endpoints Baru

### 1. Get Project Budget Summary
```
GET /api/v1/projects/:id/cbs/summary
```

**Response:**
```json
{
  "project_id": 1,
  "total_budget": 500000000,
  "total_actual": 350000000,
  "total_variance": 150000000,
  "node_count": 15
}
```

**Deskripsi:**
- `total_budget`: Total budget dari semua CBS nodes
- `total_actual`: Total actual cost dari PR allocations
- `total_variance`: Budget - Actual (positive = under budget)
- `node_count`: Jumlah CBS nodes dalam project

---

## 🎯 Cara Menggunakan Fitur Baru

### Scenario 1: Setup Budget via CBS

1. **Buka CBS Page** (`/cost-control/cbs`)
2. **Pilih Project** dari dropdown
3. **Create CBS Node**:
   - Input Code, Name, Description
   - **Pilih COA Account** (link ke Chart of Account)
   - **Input Budget Amount** (e.g., 50.000.000)
   - Save
4. **System otomatis**:
   - Sync budget ke `project_budgets` table
   - Update budget summary widget
5. **Lihat Budget Summary** di atas CBS tree:
   - Total Budget, Actual, Variance
   - Budget Utilization %

### Scenario 2: View Budget Summary

1. **Buka CBS Page**
2. **Pilih Project**
3. **Lihat Budget Summary Widget**:
   - Total Budget (from CBS)
   - Total Actual (from PR allocations)
   - Variance (green = under, red = over)
   - Utilization percentage

### Scenario 3: Navigate Between Modules

**Dari CBS Page:**
- Click "View Expenses" → Langsung ke Expenses page
- Click "Budget Report" → Langsung ke Budget vs Actual page

**Dari Expenses Page:**
- Click "View CBS" → Langsung ke CBS page
- Click "Budget Report" → Langsung ke Budget vs Actual page

**Dari Budget vs Actual Page:**
- Click "View CBS" → Langsung ke CBS page
- Click "View Expenses" → Langsung ke Expenses page

---

## 🔧 Technical Details

### Database Schema

#### Table: `project_budgets`
```sql
CREATE TABLE project_budgets (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id),
    account_id INTEGER NOT NULL REFERENCES coa_accounts(id),
    estimated_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    UNIQUE(project_id, account_id)
);
```

### Auto-Sync Logic

**Trigger:** Saat CBS Node dibuat/diupdate dengan `COAAccountID` dan `BudgetAmount`

**Process:**
```go
func (s *cbsService) syncToProjectBudget(node *models.CBSNode) error {
    if node.COAAccountID == nil {
        return nil
    }

    budget := &models.ProjectBudget{
        ProjectID:       node.ProjectID,
        AccountID:       *node.COAAccountID,
        EstimatedAmount: float64(node.BudgetAmount),
    }

    return s.projectBudgetRepo.Upsert(budget)
}
```

**Upsert Logic:**
- Jika entry sudah ada (same ProjectID + AccountID) → Update amount
- Jika entry belum ada → Create new entry

---

## 🚀 Next Steps (Belum Diimplementasikan)

### Phase 2: Enhanced Features

1. **Quick Expense Entry dari CBS**
   - Tambahkan button "Add Expense" per CBS node
   - Auto-fill COAAccountID, ReferenceType=CBS, ReferenceID=CBS Node ID
   - Navigate ke Expenses page dengan pre-filled form

2. **Expense Filtering by CBS Node**
   - Di Expenses page, tambahkan filter by CBS Node
   - Show expenses yang terkait dengan CBS node tertentu

3. **Drill-down in Budget Report**
   - Click COA line → Show related expenses
   - Click Work Package → Show related CBS nodes

### Phase 3: Advanced Features

4. **Bulk Expense Import**
   - Import expenses dari Excel/CSV
   - Mapping COA Code → COA Account
   - Bulk insert expense transactions

5. **Trend Analysis Charts**
   - Budget vs Actual over time (line chart)
   - Budget utilization by category (pie chart)
   - Variance trend (bar chart)

6. **Budget Alerts & Notifications**
   - Alert when budget utilization > 80%
   - Alert when over budget
   - Email notifications to stakeholders

7. **Export Reports**
   - Export Budget vs Actual to Excel
   - Export Budget vs Actual to PDF
   - Include charts and summary

---

## 📝 Testing Checklist

### Backend Testing

- [x] Build berhasil tanpa error
- [ ] Test CBS Node creation dengan COAAccountID dan BudgetAmount
- [ ] Verify entry dibuat di `project_budgets` table
- [ ] Test CBS Node update dengan perubahan BudgetAmount
- [ ] Verify entry di `project_budgets` terupdate
- [ ] Test API endpoint `/api/v1/projects/:id/cbs/summary`
- [ ] Verify response sesuai dengan data CBS

### Frontend Testing

- [ ] Budget Summary Widget muncul di CBS page
- [ ] Data budget summary sesuai dengan CBS nodes
- [ ] Color coding variance bekerja (green/red)
- [ ] Navigation buttons berfungsi
- [ ] Navigate dari CBS → Expenses
- [ ] Navigate dari CBS → Budget Report
- [ ] Navigate dari Expenses → CBS
- [ ] Navigate dari Expenses → Budget Report
- [ ] Navigate dari Budget Report → CBS
- [ ] Navigate dari Budget Report → Expenses

### Integration Testing

- [ ] Create CBS node → Verify budget summary update
- [ ] Update CBS node budget → Verify budget summary update
- [ ] Create expense transaction → Verify actual cost update
- [ ] Generate budget report → Verify budget from CBS

---

## 🐛 Known Issues & Limitations

1. **Budget Summary Actual Cost**
   - Saat ini actual cost dihitung dari PR CBS mappings
   - Belum include expense transactions yang dibuat manual
   - **Solution:** Update `GetProjectBudgetSummary` untuk include expense transactions

2. **Real-time Update**
   - Budget summary tidak auto-refresh saat CBS node diupdate
   - User perlu refresh page atau click Refresh button
   - **Solution:** Implement WebSocket atau polling untuk real-time update

3. **Multiple COA per CBS Node**
   - Saat ini 1 CBS node hanya bisa link ke 1 COA account
   - Untuk complex projects, mungkin perlu multiple COA per node
   - **Solution:** Buat junction table `cbs_coa_mappings`

---

## 📚 Documentation Updates

File dokumentasi yang telah dibuat:
1. `CBS_EXPENSE_INTEGRATION_FLOW.md` - Analisis dan design flow
2. `CBS_EXPENSE_INTEGRATION_IMPLEMENTATION.md` - Dokumentasi implementasi (file ini)

---

## 🎉 Summary

Implementasi Phase 1 telah selesai dengan fitur:
- ✅ CBS → Project Budget auto-sync (Backend)
- ✅ Budget Summary Widget (Frontend)
- ✅ Enhanced Navigation antar modul (Frontend)
- ✅ API endpoint untuk budget summary

**Total Files Created:** 2
**Total Files Modified:** 9
**Build Status:** ✅ Success

Aplikasi siap untuk testing dan deployment!
