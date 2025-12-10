# System Integration Plan - Unipro Project Management

## 📊 Analisis Fitur yang Ada

### **1. PROJECT MANAGEMENT**
- **Projects** - Master data proyek

### **2. COST CONTROL**
- **Cost Control Dashboard** - Overview biaya
- **Expense Transactions** - Transaksi biaya manual/otomatis ✨ NEW
- **Budget vs Actual Report** - Laporan budget ✨ NEW
- **Material Tracking** - Tracking material usage
- **Cost Breakdown Structure (CBS)** - Struktur breakdown biaya
- **Purchase Request Management** - Manajemen PR

### **3. MASTER DATA**
- **Chart of Accounts (COA)** - Akun biaya ✨ UPDATED
- **Material** - Master material
- **Vendor** - Master vendor

---

## 🔗 Integration Flow yang Harus Dibuat

### **FLOW 1: Purchase Request → Expense Transaction**

```
┌─────────────────────────────────────────────────────────────┐
│  Purchase Request (PR) Workflow                              │
└─────────────────────────────────────────────────────────────┘
                           ↓
    1. User creates PR with items
       - Item: Semen 100 sak
       - Material: Semen (linked to COA 5203)
       - Estimated Price: Rp 5,000,000
       - CBS Node: Pekerjaan Struktur
                           ↓
    2. PR goes through approval workflow
       - Cost Control reviews
       - Project Director approves
                           ↓
    3. PR Status = APPROVED
       ⚡ TRIGGER: Auto-create Expense Transaction
       - project_id: from PR
       - transaction_date: PR approved date
       - coa_account_id: from Material.coa_account_id
       - description: "PR-{code}: {item_name}"
       - amount: estimated_price
       - transaction_type: MATERIAL
       - reference_type: PR
       - reference_id: PR.id
       - reference_no: PR.code
                           ↓
    4. Expense Transaction created
       ✅ Budget tracking updated automatically
       ✅ Actual cost recorded
```

### **FLOW 2: Material Tracking → Expense Transaction**

```
┌─────────────────────────────────────────────────────────────┐
│  Material Usage Recording                                    │
└─────────────────────────────────────────────────────────────┘
                           ↓
    1. Field officer records material usage
       - Material: Besi Beton 500 kg
       - Quantity: 500
       - Unit Price: Rp 15,000/kg
       - Total: Rp 7,500,000
                           ↓
    2. Material usage saved
       ⚡ TRIGGER: Auto-create Expense Transaction
       - coa_account_id: from Material.coa_account_id
       - description: "Material Usage: {material_name}"
       - amount: quantity × unit_price
       - transaction_type: MATERIAL
       - reference_type: MATERIAL_USAGE
                           ↓
    3. Expense Transaction created
       ✅ Material cost tracked
       ✅ Budget updated
```

### **FLOW 3: CBS Budget → Project Budget → Expense Tracking**

```
┌─────────────────────────────────────────────────────────────┐
│  CBS Budget Allocation                                       │
└─────────────────────────────────────────────────────────────┘
                           ↓
    1. Project Manager creates CBS structure
       - Node: Pekerjaan Struktur
       - Budget: Rp 500,000,000
       - Linked to COA: 5204 (Pekerjaan Beton)
                           ↓
    2. CBS budget allocated
       ⚡ SYNC: Create/Update Project Budget
       - project_id
       - coa_account_id: from CBS.coa_account_id
       - estimated_amount: CBS.budget_amount
                           ↓
    3. Expenses recorded against CBS
       - All expense transactions with matching COA
       - Grouped by CBS node
       - Real-time variance calculation
```

### **FLOW 4: COA → Material → PR → Expense (Full Integration)**

```
┌─────────────────────────────────────────────────────────────┐
│  Complete Integration Flow                                   │
└─────────────────────────────────────────────────────────────┘

1. SETUP MASTER DATA
   ├─ Create COA: 5203 - Pasangan dan Plesteran
   │  └─ budget_category: OPERASIONAL_BUDGET
   │  └─ work_package: PASANGAN DAN PLESTERAN
   │
   ├─ Create Material: Semen Portland
   │  └─ coa_account_id: 5203 (linked to COA)
   │  └─ unit_price: Rp 50,000
   │
   └─ Create Vendor: PT Semen Indonesia
      └─ supplies: Semen Portland

2. PROJECT EXECUTION
   ├─ Create CBS Node: Pekerjaan Dinding
   │  └─ coa_account_id: 5203
   │  └─ budget_amount: Rp 100,000,000
   │
   ├─ Create PR: Request Semen
   │  ├─ Item: Semen Portland 1000 sak
   │  ├─ material_id: Semen Portland
   │  ├─ cbs_node_id: Pekerjaan Dinding
   │  └─ estimated_price: Rp 50,000,000
   │
   ├─ PR Approved
   │  └─ ⚡ Auto-create Expense Transaction
   │     ├─ coa_account_id: 5203 (from Material)
   │     ├─ amount: Rp 50,000,000
   │     ├─ reference: PR-001
   │     └─ transaction_type: MATERIAL
   │
   └─ Material Delivered & Used
      └─ ⚡ Update Material Tracking
         └─ Stock updated
         └─ Usage recorded

3. REPORTING
   ├─ Budget vs Actual Report
   │  ├─ Budget: Rp 100,000,000 (from CBS)
   │  ├─ Actual: Rp 50,000,000 (from Expense)
   │  └─ Variance: Rp 50,000,000 (50%)
   │
   └─ Material Tracking Report
      ├─ Material: Semen Portland
      ├─ Purchased: 1000 sak
      ├─ Used: 800 sak
      └─ Remaining: 200 sak
```

---

## 🛠️ Implementation Tasks

### **Phase 1: PR → Expense Integration (HIGH PRIORITY)**

#### Backend Changes

**1. Update PR Service** (`backend/services/purchase_request_service.go`)
```go
// Add method to create expense transaction when PR is approved
func (s *purchaseRequestService) CreateExpenseFromPR(pr *models.PurchaseRequest) error {
    // For each PR item
    for _, item := range pr.Items {
        // Get COA from material
        var coaAccountID uint
        if item.MaterialID != nil {
            material, _ := s.materialRepo.GetByID(*item.MaterialID)
            if material != nil && material.COAAccountID != nil {
                coaAccountID = *material.COAAccountID
            }
        }
        
        // Create expense transaction
        expense := &models.ExpenseTransaction{
            ProjectID:       pr.ProjectID,
            TransactionDate: time.Now(),
            COAAccountID:    coaAccountID,
            Description:     fmt.Sprintf("PR-%s: %s", pr.Code, item.ItemName),
            Amount:          item.TotalPrice,
            Unit:            item.Unit,
            Quantity:        item.Quantity,
            TransactionType: models.ExpenseTypeMaterial,
            ReferenceType:   models.ExpenseRefTypePR,
            ReferenceID:     &pr.ID,
            ReferenceNo:     pr.Code,
            CreatedBy:       pr.CreatedBy,
        }
        
        s.expenseRepo.Create(expense)
    }
    return nil
}
```

**2. Update PR Approval Handler**
```go
// In approval service, after PR is approved
if pr.Status == models.PRStatusApproved {
    // Create expense transactions
    if err := prService.CreateExpenseFromPR(pr); err != nil {
        log.Printf("Failed to create expense from PR: %v", err)
    }
}
```

#### Database Changes
```sql
-- Add trigger to auto-create expense when PR is approved
-- Or handle in application layer (recommended)
```

---

### **Phase 2: Material Tracking → Expense Integration**

**1. Update Material Tracking Service**
```go
func (s *materialTrackingService) RecordUsage(usage *MaterialUsage) error {
    // Record material usage
    if err := s.repo.Create(usage); err != nil {
        return err
    }
    
    // Get material to get COA
    material, _ := s.materialRepo.GetByID(usage.MaterialID)
    
    // Create expense transaction
    expense := &models.ExpenseTransaction{
        ProjectID:       usage.ProjectID,
        TransactionDate: usage.UsageDate,
        COAAccountID:    material.COAAccountID,
        Description:     fmt.Sprintf("Material Usage: %s", material.Name),
        Amount:          usage.Quantity * material.UnitPrice,
        Unit:            material.Unit,
        Quantity:        usage.Quantity,
        TransactionType: models.ExpenseTypeMaterial,
        ReferenceType:   "MATERIAL_USAGE",
        ReferenceID:     &usage.ID,
    }
    
    s.expenseRepo.Create(expense)
    return nil
}
```

---

### **Phase 3: CBS → Project Budget Sync**

**1. Update CBS Service**
```go
func (s *cbsService) Create(cbs *models.CBSNode) error {
    // Create CBS node
    if err := s.repo.Create(cbs); err != nil {
        return err
    }
    
    // Sync with project budget
    if cbs.COAAccountID != nil {
        projectBudget := &models.ProjectBudget{
            ProjectID:       cbs.ProjectID,
            AccountID:       *cbs.COAAccountID,
            EstimatedAmount: float64(cbs.BudgetAmount),
        }
        s.projectBudgetRepo.Upsert(projectBudget)
    }
    
    return nil
}
```

---

### **Phase 4: Dashboard Integration**

**1. Cost Control Dashboard Enhancement**
```typescript
// Show real-time data from all sources
interface DashboardData {
    // From Expense Transactions
    totalExpenses: number;
    expensesByCategory: CategoryBreakdown[];
    
    // From Purchase Requests
    pendingPRs: number;
    approvedPRAmount: number;
    
    // From Material Tracking
    materialCost: number;
    lowStockItems: number;
    
    // From CBS
    budgetUtilization: number;
    overBudgetNodes: CBSNode[];
}
```

---

## 📋 Integration Checklist

### **Immediate (Week 1)**
- [ ] PR Approval → Auto-create Expense Transaction
- [ ] Link Material to COA in PR items
- [ ] Update PR service to create expenses
- [ ] Test PR → Expense flow

### **Short Term (Week 2-3)**
- [ ] Material Usage → Auto-create Expense Transaction
- [ ] CBS Budget → Project Budget sync
- [ ] Update Budget vs Actual to use real data
- [ ] Dashboard integration

### **Medium Term (Month 1)**
- [ ] Expense reversal when PR is rejected
- [ ] Batch expense creation for multiple PRs
- [ ] Material cost variance tracking
- [ ] Budget alert notifications

### **Long Term (Month 2+)**
- [ ] Purchase Order integration
- [ ] Vendor payment tracking
- [ ] Cash flow projection
- [ ] Advanced analytics

---

## 🎯 Benefits of Integration

### **1. Automatic Data Flow**
- ✅ No manual entry of expenses from PR
- ✅ Real-time budget tracking
- ✅ Accurate cost reporting

### **2. Data Consistency**
- ✅ Single source of truth
- ✅ No data duplication
- ✅ Synchronized across modules

### **3. Better Decision Making**
- ✅ Real-time budget status
- ✅ Early warning for over-budget
- ✅ Material cost optimization

### **4. Audit Trail**
- ✅ Complete transaction history
- ✅ Traceable from PR to expense
- ✅ Compliance ready

---

## 🚀 Quick Start Implementation

### **Step 1: Update PR Service**
File: `backend/services/purchase_request_service.go`

Add dependency:
```go
type PurchaseRequestService struct {
    repo        repositories.PurchaseRequestRepository
    cbsRepo     repositories.CBSRepository
    db          *gorm.DB
    approval    services.ApprovalService
    expenseRepo repositories.ExpenseTransactionRepository  // ADD THIS
    materialRepo repositories.MaterialRepository            // ADD THIS
}
```

### **Step 2: Add Auto-Expense Method**
```go
func (s *purchaseRequestService) CreateExpenseFromApprovedPR(prID uint) error {
    // Implementation as shown above
}
```

### **Step 3: Hook into Approval Flow**
```go
// In approval handler, after status update
if newStatus == "APPROVED" {
    go prService.CreateExpenseFromApprovedPR(pr.ID)
}
```

---

## 📊 Data Flow Diagram

```
┌──────────────┐
│   Master     │
│    Data      │
│  (COA, Mat)  │
└──────┬───────┘
       │
       ├─────────────────────────────────┐
       ↓                                 ↓
┌──────────────┐                  ┌──────────────┐
│   Purchase   │                  │   Material   │
│   Request    │                  │   Tracking   │
└──────┬───────┘                  └──────┬───────┘
       │                                 │
       │ (Approved)                      │ (Usage)
       │                                 │
       └─────────────┬───────────────────┘
                     ↓
              ┌──────────────┐
              │   Expense    │
              │ Transactions │
              └──────┬───────┘
                     │
                     ├──────────────┬──────────────┐
                     ↓              ↓              ↓
              ┌──────────┐   ┌──────────┐  ┌──────────┐
              │  Budget  │   │   CBS    │  │Dashboard │
              │  Report  │   │ Tracking │  │ Analytics│
              └──────────┘   └──────────┘  └──────────┘
```

---

## 💡 Recommendation

**Start with Phase 1** (PR → Expense Integration) karena:
1. Paling banyak digunakan
2. Paling besar impact-nya
3. Relatif mudah diimplementasikan
4. Foundation untuk integrasi lainnya

Setelah Phase 1 stabil, lanjut ke Phase 2 dan 3 secara parallel.


---

## 📋 Implementation Status

### Phase 1: PR → Expense Integration ✅ COMPLETED
**Completion Date**: December 7, 2025

**Implemented Components**:
- ✅ Approval callback mechanism (`approval_callback.go`)
- ✅ PR service method `CreateExpenseFromApprovedPR`
- ✅ Automatic expense creation on PR approval
- ✅ COA mapping through materials
- ✅ Reference tracking (PR → Expense)
- ✅ Error handling and logging
- ✅ Service wiring in routes

**Documentation**:
- See `PR_EXPENSE_INTEGRATION.md` for detailed implementation guide

**Testing**:
- ✅ Backend compilation successful
- ⏳ Integration testing pending
- ⏳ End-to-end workflow testing pending

**Next Steps**:
1. Test complete PR approval workflow
2. Verify expense transactions are created
3. Check Budget vs Actual report shows correct data
4. Test error scenarios (missing COA, etc.)

---

### Phase 2: Material Tracking → Expense Integration 🔴 NOT STARTED
**Priority**: Medium
**Estimated Effort**: 2-3 hours

**Requirements**:
- Update material tracking service to create expenses on usage
- Link material movements to expense transactions
- Track actual vs estimated material costs

---

### Phase 3: CBS → Project Budget Sync 🔴 NOT STARTED
**Priority**: Medium
**Estimated Effort**: 3-4 hours

**Requirements**:
- Sync CBS budget allocations with project budgets
- Validate expenses against CBS budget limits
- Alert when budget thresholds are exceeded

---

### Phase 4: Dashboard Integration 🔴 NOT STARTED
**Priority**: Low
**Estimated Effort**: 4-5 hours

**Requirements**:
- Show PR-to-Expense conversion metrics
- Display pending PRs vs approved expenses
- Cost control dashboard with real-time data
- Material tracking dashboard enhancements

---

## 🎯 Success Metrics

### Phase 1 Success Criteria
- [x] PR approval triggers expense creation automatically
- [x] Expense transactions have correct COA from materials
- [x] Reference fields properly link expenses to PRs
- [ ] Budget vs Actual report shows PR expenses
- [ ] No manual expense entry needed for approved PRs

### Overall Integration Goals
- Reduce manual data entry by 80%
- Improve cost tracking accuracy to 95%+
- Enable real-time budget monitoring
- Provide complete audit trail from PR to expense
- Support project cost forecasting

---

## 📝 Notes

### Design Decisions
1. **Async Processing**: Expense creation runs in goroutine to avoid blocking approval
2. **Graceful Degradation**: Missing COA mappings log warnings but don't fail approval
3. **Callback Pattern**: Clean separation between approval and business logic
4. **Reference Tracking**: Every expense links back to source (PR, Material Usage, etc.)

### Known Limitations
1. Materials must have COA mappings for expense creation
2. Expense creation is one-way (PR → Expense, not bidirectional)
3. No automatic expense deletion if PR is cancelled after approval
4. Material price changes don't update existing expenses

### Future Enhancements
1. Support for expense reversal when PR is cancelled
2. Automatic price adjustment based on actual PO prices
3. Multi-currency support for international purchases
4. Integration with accounting systems (export to ERP)
5. Predictive analytics for budget forecasting
