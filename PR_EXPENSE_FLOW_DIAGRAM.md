# PR to Expense Integration Flow Diagram

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         USER INTERFACE                               │
│  (Frontend: Purchase Request Form, Approval Panel, Expense List)    │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         API LAYER (Routes)                           │
│  - POST /api/v1/purchase-requests                                   │
│  - POST /api/v1/employee/approvals/:id/process                      │
│  - GET  /api/v1/projects/:id/expenses                               │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         SERVICE LAYER                                │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐ │
│  │ PR Service       │  │ Approval Service │  │ Expense Service  │ │
│  │ - CreatePR       │  │ - ProcessApproval│  │ - Create         │ │
│  │ - CreateExpense  │◄─┤ - Callback       │  │ - GetByProject   │ │
│  └──────────────────┘  └──────────────────┘  └──────────────────┘ │
│           │                      │                      ▲            │
│           │                      │                      │            │
│           │         ┌────────────┴──────────┐          │            │
│           │         │ Approval Callback     │          │            │
│           │         │ Handler               │──────────┘            │
│           │         └───────────────────────┘                       │
└───────────┼──────────────────────────────────────────────────────────┘
            │
            ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      REPOSITORY LAYER                                │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐             │
│  │ PR Repo      │  │ Expense Repo │  │ Material Repo│             │
│  │ - Create     │  │ - Create     │  │ - GetByID    │             │
│  │ - FindByID   │  │ - FindAll    │  │              │             │
│  └──────────────┘  └──────────────┘  └──────────────┘             │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         DATABASE                                     │
│  - purchase_requests                                                 │
│  - purchase_request_items                                            │
│  - expense_transactions                                              │
│  - materials                                                         │
│  - coa_accounts                                                      │
│  - approval_requests                                                 │
│  - approval_actions                                                  │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Detailed Flow: PR Creation to Expense

```
┌─────────────────────────────────────────────────────────────────────┐
│ STEP 1: User Creates Purchase Request                               │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────┐
                    │ PR Service: CreatePR()    │
                    │ - Generate PR code        │
                    │ - Calculate total amount  │
                    │ - Save PR to database     │
                    └───────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────┐
                    │ Approval Service:         │
                    │ CreateApprovalRequest()   │
                    │ - Find workflow           │
                    │ - Create approval request │
                    │ - Create approval actions │
                    │ - Notify first approver   │
                    └───────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│ STEP 2: Approval Workflow (Multiple Steps)                          │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                    ┌───────────────┴───────────────┐
                    │                               │
                    ▼                               ▼
        ┌─────────────────────┐       ┌─────────────────────┐
        │ Purchasing Approves │       │ Cost Control        │
        │ (Step 1)            │──────►│ Approves (Step 2)   │
        └─────────────────────┘       └─────────────────────┘
                                                │
                    ┌───────────────────────────┘
                    │
                    ▼
        ┌─────────────────────┐
        │ GM Approves         │
        │ (Step 3)            │
        └─────────────────────┘
                    │
                    ▼
        ┌─────────────────────┐
        │ Project Director    │
        │ Approves (Step 4)   │
        └─────────────────────┘
                    │
                    ▼
        ┌─────────────────────┐
        │ Managing Director   │
        │ Approves (Step 5)   │◄─── FINAL APPROVAL
        └─────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│ STEP 3: All Steps Approved - Trigger Expense Creation               │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────┐
                    │ Approval Service:         │
                    │ ProcessApprovalAction()   │
                    │ - Check all steps done    │
                    │ - Update PR status        │
                    │ - Detect PR approval      │
                    └───────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────┐
                    │ Trigger Callback          │
                    │ (Async - Goroutine)       │
                    └───────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────┐
                    │ Approval Callback Handler │
                    │ OnPurchaseRequestApproved │
                    └───────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│ STEP 4: Create Expense Transactions                                 │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────┐
                    │ PR Service:               │
                    │ CreateExpenseFromApproved │
                    │ - Get PR with items       │
                    │ - Validate PR approved    │
                    └───────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────┐
                    │ For Each PR Item:         │
                    └───────────────────────────┘
                                    │
                    ┌───────────────┴───────────────┐
                    │                               │
                    ▼                               ▼
        ┌─────────────────────┐       ┌─────────────────────┐
        │ Get Material        │       │ Get COA Account     │
        │ by MaterialID       │──────►│ from Material       │
        └─────────────────────┘       └─────────────────────┘
                                                │
                                                ▼
                                    ┌───────────────────────────┐
                                    │ Create Expense:           │
                                    │ - project_id              │
                                    │ - coa_account_id          │
                                    │ - description             │
                                    │ - amount                  │
                                    │ - quantity, unit          │
                                    │ - transaction_type: MAT   │
                                    │ - reference_type: PR      │
                                    │ - reference_id: PR.id     │
                                    │ - reference_no: PR.code   │
                                    └───────────────────────────┘
                                                │
                                                ▼
                                    ┌───────────────────────────┐
                                    │ Save to Database          │
                                    │ expense_transactions      │
                                    └───────────────────────────┘
                                                │
                                                ▼
                                    ┌───────────────────────────┐
                                    │ Log Success               │
                                    │ ✅ Created expense for    │
                                    │    PR-{CODE} item         │
                                    └───────────────────────────┘
                                                │
                    ┌───────────────────────────┘
                    │
                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│ STEP 5: Expenses Available for Reporting                            │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                    ┌───────────────┴───────────────┐
                    │                               │
                    ▼                               ▼
        ┌─────────────────────┐       ┌─────────────────────┐
        │ Expense Transaction │       │ Budget vs Actual    │
        │ List                │       │ Report              │
        │ - Shows all expenses│       │ - Budget from CBS   │
        │ - Filter by PR      │       │ - Actual from       │
        │ - Trace to source   │       │   Expenses          │
        └─────────────────────┘       └─────────────────────┘
```

---

## Data Flow: PR Item to Expense Transaction

```
┌─────────────────────────────────────────────────────────────────────┐
│                    PURCHASE REQUEST ITEM                             │
├─────────────────────────────────────────────────────────────────────┤
│ id: 123                                                              │
│ purchase_request_id: 456                                             │
│ item_name: "Semen Portland"                                          │
│ material_id: 789                                                     │
│ quantity: 100                                                        │
│ unit: "sak"                                                          │
│ estimated_price: 50,000                                              │
│ total_price: 5,000,000                                               │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────┐
                    │ Get Material (ID: 789)    │
                    └───────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         MATERIAL                                     │
├─────────────────────────────────────────────────────────────────────┤
│ id: 789                                                              │
│ name: "Semen Portland"                                               │
│ code: "MAT-001"                                                      │
│ unit: "sak"                                                          │
│ unit_price: 50,000                                                   │
│ coa_account_id: 5203  ◄────────────────────────────────────────────┐│
└─────────────────────────────────────────────────────────────────────┘│
                                    │                                   │
                                    ▼                                   │
                    ┌───────────────────────────┐                      │
                    │ Get COA (ID: 5203)        │                      │
                    └───────────────────────────┘                      │
                                    │                                   │
                                    ▼                                   │
┌─────────────────────────────────────────────────────────────────────┐│
│                      COA ACCOUNT                                     ││
├─────────────────────────────────────────────────────────────────────┤│
│ id: 5203                                                             ││
│ code: "5203"                                                         ││
│ name: "Pasangan dan Plesteran"                                      ││
│ budget_category: "OPERASIONAL_BUDGET"                                ││
│ work_package: "PASANGAN DAN PLESTERAN"                               ││
└─────────────────────────────────────────────────────────────────────┘│
                                    │                                   │
                                    ▼                                   │
                    ┌───────────────────────────┐                      │
                    │ Create Expense            │                      │
                    │ Transaction               │                      │
                    └───────────────────────────┘                      │
                                    │                                   │
                                    ▼                                   │
┌─────────────────────────────────────────────────────────────────────┐│
│                   EXPENSE TRANSACTION                                ││
├─────────────────────────────────────────────────────────────────────┤│
│ id: 999                                                              ││
│ project_id: 1                                                        ││
│ transaction_date: 2025-12-07                                         ││
│ coa_account_id: 5203  ◄──────────────────────────────────────────────┘
│ description: "PR-PR-1-20251207: Semen Portland (Material: Semen...)"│
│ amount: 5,000,000                                                    │
│ unit: "sak"                                                          │
│ quantity: 100                                                        │
│ transaction_type: "MATERIAL"                                         │
│ reference_type: "PR"                                                 │
│ reference_id: 456                                                    │
│ reference_no: "PR-1-20251207"                                        │
│ notes: "Auto-created from approved PR: PR-1-20251207"               │
│ created_by: 1                                                        │
└─────────────────────────────────────────────────────────────────────┘
```

---

## Error Handling Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│ PR Item Processing                                                   │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────┐
                    │ Get Material by ID        │
                    └───────────────────────────┘
                                    │
                    ┌───────────────┴───────────────┐
                    │                               │
                    ▼                               ▼
        ┌─────────────────────┐       ┌─────────────────────┐
        │ Material Found      │       │ Material Not Found  │
        │                     │       │ or No COA           │
        └─────────────────────┘       └─────────────────────┘
                    │                               │
                    ▼                               ▼
        ┌─────────────────────┐       ┌─────────────────────┐
        │ COA Account ID      │       │ Log Warning:        │
        │ Available           │       │ "No COA found for   │
        │                     │       │  PR item {ID}"      │
        └─────────────────────┘       └─────────────────────┘
                    │                               │
                    ▼                               ▼
        ┌─────────────────────┐       ┌─────────────────────┐
        │ Create Expense      │       │ Skip This Item      │
        │ Transaction         │       │ Continue with Next  │
        └─────────────────────┘       └─────────────────────┘
                    │                               │
                    ▼                               │
        ┌─────────────────────┐                     │
        │ Save Success?       │                     │
        └─────────────────────┘                     │
                    │                               │
        ┌───────────┴───────────┐                   │
        │                       │                   │
        ▼                       ▼                   │
┌─────────────┐       ┌─────────────┐              │
│ Success     │       │ Error       │              │
│ Log: ✅     │       │ Log: ⚠️     │              │
│ Continue    │       │ Continue    │              │
└─────────────┘       └─────────────┘              │
        │                       │                   │
        └───────────┬───────────┘                   │
                    │                               │
                    └───────────────┬───────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────┐
                    │ Process Next Item         │
                    │ (Graceful Degradation)    │
                    └───────────────────────────┘
```

---

## Async Processing Flow

```
┌─────────────────────────────────────────────────────────────────────┐
│ Main Thread: Approval Processing                                    │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────┐
                    │ All Approval Steps Done   │
                    │ Update PR Status          │
                    │ Commit Transaction        │
                    └───────────────────────────┘
                                    │
                                    ▼
                    ┌───────────────────────────┐
                    │ Launch Goroutine          │
                    │ (Async Processing)        │
                    └───────────────────────────┘
                                    │
                    ┌───────────────┴───────────────┐
                    │                               │
                    ▼                               ▼
        ┌─────────────────────┐       ┌─────────────────────┐
        │ Main Thread         │       │ Background Thread   │
        │ - Return success    │       │ - Create expenses   │
        │ - Send response     │       │ - Log results       │
        │ - Continue          │       │ - Handle errors     │
        └─────────────────────┘       └─────────────────────┘
                    │                               │
                    ▼                               ▼
        ┌─────────────────────┐       ┌─────────────────────┐
        │ User sees approval  │       │ Expenses created    │
        │ success immediately │       │ in background       │
        └─────────────────────┘       └─────────────────────┘
```

---

## Component Interaction Diagram

```
┌──────────────────────────────────────────────────────────────────────┐
│                                                                       │
│  ┌─────────────┐         ┌─────────────┐         ┌─────────────┐   │
│  │   Routes    │────────►│  PR Service │────────►│  PR Repo    │   │
│  │             │         │             │         │             │   │
│  └─────────────┘         └─────────────┘         └─────────────┘   │
│         │                        │                                   │
│         │                        │                                   │
│         ▼                        ▼                                   │
│  ┌─────────────┐         ┌─────────────┐                            │
│  │  Approval   │◄────────│  Callback   │                            │
│  │  Service    │         │  Handler    │                            │
│  └─────────────┘         └─────────────┘                            │
│         │                        │                                   │
│         │                        │                                   │
│         ▼                        ▼                                   │
│  ┌─────────────┐         ┌─────────────┐         ┌─────────────┐   │
│  │  Approval   │         │  Expense    │────────►│  Expense    │   │
│  │  Repo       │         │  Service    │         │  Repo       │   │
│  └─────────────┘         └─────────────┘         └─────────────┘   │
│                                  │                                   │
│                                  │                                   │
│                                  ▼                                   │
│                          ┌─────────────┐         ┌─────────────┐   │
│                          │  Material   │────────►│  Material   │   │
│                          │  Service    │         │  Repo       │   │
│                          └─────────────┘         └─────────────┘   │
│                                                                       │
└──────────────────────────────────────────────────────────────────────┘

Legend:
────► : Direct call
◄──── : Callback
```

---

## Summary

This integration provides:
- ✅ **Automatic**: No manual expense entry needed
- ✅ **Traceable**: Complete audit trail from PR to expense
- ✅ **Accurate**: COA mapping ensures correct accounting
- ✅ **Non-blocking**: Async processing doesn't delay approvals
- ✅ **Resilient**: Graceful error handling prevents failures
- ✅ **Scalable**: Can handle multiple concurrent PRs
