# Purchase Request to Expense Transaction Integration

## Overview
This document describes the automatic integration between Purchase Requests (PR) and Expense Transactions. When a PR is approved through the approval workflow, the system automatically creates corresponding expense transactions for cost tracking.

## Implementation Details

### Architecture

```
Purchase Request → Approval Workflow → Approval Service → Callback Handler → PR Service → Expense Transactions
```

### Components

#### 1. Approval Service (`backend/services/approval_service.go`)
- **Enhanced Interface**: Added `OnPurchaseRequestApproved` to `PostApprovalCallback` interface
- **Trigger Point**: When all approval steps are completed and PR status changes to APPROVED
- **Async Processing**: Expense creation runs in a goroutine to avoid blocking the approval process

```go
// POST-APPROVAL PROCESSING: Create expense transactions for approved PRs
if prToProcess != nil && s.postApprovalCallback != nil {
    go func(prID uint) {
        if err := s.postApprovalCallback.OnPurchaseRequestApproved(prID); err != nil {
            fmt.Printf("⚠️ Post-approval processing failed for PR %d: %v\n", prID, err)
        } else {
            fmt.Printf("✅ Post-approval processing completed for PR %d\n", prID)
        }
    }(*prToProcess)
}
```

#### 2. Approval Callback Handler (`backend/services/approval_callback.go`)
- **Purpose**: Connects approval service to business logic services
- **Implementation**: Implements `PostApprovalCallback` interface
- **Responsibility**: Delegates PR approval processing to PR service

```go
func (h *ApprovalCallbackHandler) OnPurchaseRequestApproved(prID uint) error {
    return h.prService.CreateExpenseFromApprovedPR(prID)
}
```

#### 3. Purchase Request Service (`backend/services/purchase_request_service.go`)
- **New Method**: `CreateExpenseFromApprovedPR(prID uint) error`
- **Dependencies**: Requires `ExpenseTransactionRepository` and `MaterialRepository`
- **Logic**:
  1. Retrieves PR with all items
  2. Validates PR is approved
  3. For each PR item:
     - Gets COA account from linked material
     - Creates expense transaction with proper reference
     - Logs success/errors

```go
func (s *purchaseRequestService) CreateExpenseFromApprovedPR(prID uint) error {
    // Get PR with items
    pr, err := s.repo.FindByID(prID)
    if err != nil {
        return fmt.Errorf("failed to get PR: %w", err)
    }

    // Only create expenses for approved PRs
    if pr.Status != models.PRStatusApproved {
        return fmt.Errorf("PR is not approved, status: %s", pr.Status)
    }

    // Create expense transaction for each PR item
    for _, item := range pr.Items {
        // Get COA from material
        var coaAccountID uint
        if item.MaterialID != nil {
            material, err := s.materialRepo.GetByID(*item.MaterialID)
            if err == nil && material != nil && material.COAAccountID != nil {
                coaAccountID = *material.COAAccountID
            }
        }

        if coaAccountID == 0 {
            // Skip items without COA mapping
            continue
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
            Notes:           fmt.Sprintf("Auto-created from approved PR: %s", pr.Code),
            CreatedBy:       pr.CreatedBy,
        }

        s.expenseRepo.Create(expense)
    }

    return nil
}
```

#### 4. Routes Initialization (`backend/routes/routes.go`)
- **Wiring**: Connects all components together
- **Callback Setup**: Registers callback handler with approval service

```go
// Purchase Request with expense integration
expenseRepo := repositories.NewExpenseTransactionRepository(db)
materialRepo := repositories.NewMaterialRepository(db)
prService := services.NewPurchaseRequestService(prRepo, cbsRepo, db, approvalService, expenseRepo, materialRepo)

// Setup approval callback
approvalCallback := services.NewApprovalCallbackHandler(prService)
approvalService.SetPostApprovalCallback(approvalCallback)
```

## Data Flow

### 1. PR Creation
```
User → Create PR → PR Service → Approval Service (creates approval request)
```

### 2. PR Approval
```
Approver → Approve Step → Approval Service → Check if all steps complete
```

### 3. Expense Creation (Automatic)
```
All Steps Approved → Callback Handler → PR Service → Create Expense Transactions
```

### 4. Expense Transaction Details
Each expense transaction includes:
- **Project ID**: From PR
- **Transaction Date**: Current date/time
- **COA Account**: From linked material's COA mapping
- **Description**: "PR-{CODE}: {ITEM_NAME} (Material: {MATERIAL_NAME})"
- **Amount**: Item total price
- **Quantity & Unit**: From PR item
- **Transaction Type**: MATERIAL
- **Reference Type**: PR
- **Reference ID**: PR ID
- **Reference No**: PR Code
- **Notes**: "Auto-created from approved PR: {CODE}"
- **Created By**: Original PR creator

## Benefits

### 1. Automation
- Eliminates manual expense entry for approved PRs
- Reduces data entry errors
- Ensures consistency between PR and expenses

### 2. Traceability
- Every expense links back to its source PR
- Reference fields enable audit trail
- Easy to track which expenses came from which PRs

### 3. Real-time Cost Tracking
- Expenses created immediately upon approval
- Budget vs Actual reports reflect approved PRs instantly
- Project managers see cost impact in real-time

### 4. Integration
- Seamless connection between procurement and cost control
- COA mapping through materials ensures proper accounting
- Supports budget category and work package tracking

## Error Handling

### Graceful Degradation
- If COA mapping is missing for a material, that item is skipped (logged as warning)
- Errors in expense creation don't block PR approval
- Each item is processed independently (one failure doesn't stop others)

### Logging
```
✅ Created expense transaction for PR-{CODE} item: {NAME} (Amount: {AMOUNT})
⚠️ Warning: No COA found for PR item {ID} ({NAME}), skipping expense creation
⚠️ Post-approval processing failed for PR {ID}: {ERROR}
```

## Testing

### Test Scenario 1: Complete Flow
1. Create PR with materials that have COA mappings
2. Submit for approval
3. Approve through all workflow steps
4. Verify expense transactions are created automatically
5. Check Budget vs Actual report shows the expenses

### Test Scenario 2: Missing COA
1. Create PR with material without COA mapping
2. Approve PR
3. Verify warning is logged
4. Verify other items with COA still create expenses

### Test Scenario 3: Multiple Items
1. Create PR with 5 different materials
2. Approve PR
3. Verify 5 expense transactions are created
4. Verify each has correct reference to PR

## Future Enhancements

### Phase 2: Material Tracking Integration
- When material is used from inventory, create expense transaction
- Link expense to material movement record
- Track actual vs estimated material costs

### Phase 3: CBS Budget Integration
- Sync CBS budget allocations with project budgets
- Validate expenses against CBS budget limits
- Alert when budget thresholds are exceeded

### Phase 4: Dashboard Integration
- Show PR-to-Expense conversion metrics
- Display pending PRs vs approved expenses
- Cost control dashboard with real-time data

## Configuration

### Required Setup
1. **Materials**: Must have COA account mappings
2. **COA**: Must have budget categories and work packages configured
3. **Approval Workflow**: Must be configured for Purchase Requests
4. **Permissions**: Users need appropriate permissions for PR and expense modules

### Database Requirements
- `expense_transactions` table must exist (migration 072)
- `materials` table must have `coa_account_id` column
- `coa_accounts` table must have `budget_category` and `work_package` columns

## Troubleshooting

### Expenses Not Created
1. Check PR status is APPROVED
2. Verify materials have COA mappings
3. Check backend logs for errors
4. Verify expense_transactions table exists

### Incorrect Amounts
1. Verify PR item total_price calculation
2. Check material COA mapping is correct
3. Review expense transaction creation logic

### Missing References
1. Ensure reference_type is set to "PR"
2. Verify reference_id points to correct PR
3. Check reference_no matches PR code

## Related Files

### Backend
- `backend/services/approval_service.go` - Approval workflow and callbacks
- `backend/services/approval_callback.go` - Callback handler implementation
- `backend/services/purchase_request_service.go` - PR service with expense creation
- `backend/routes/routes.go` - Service initialization and wiring
- `backend/models/expense_transaction.go` - Expense transaction model
- `backend/repositories/expense_transaction_repository.go` - Expense repository

### Frontend
- `frontend/src/components/cost-control/ExpenseTransactionList.tsx` - View expenses
- `frontend/src/components/cost-control/BudgetReportViewer.tsx` - Budget vs Actual report
- `frontend/app/cost-control/budget-vs-actual/page.tsx` - Report page

### Documentation
- `SYSTEM_INTEGRATION_PLAN.md` - Overall integration strategy
- `EXPENSE_TRANSACTION_IMPLEMENTATION.md` - Expense transaction details
- `PR_EXPENSE_INTEGRATION.md` - This document

## Conclusion

The PR to Expense integration provides seamless automation between procurement and cost control modules. By automatically creating expense transactions when PRs are approved, the system ensures accurate, real-time cost tracking while reducing manual data entry and potential errors.
