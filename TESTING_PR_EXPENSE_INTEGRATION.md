# Testing Guide: PR to Expense Integration

## Prerequisites

### 1. Database Setup
Ensure migrations are up to date:
```powershell
cd backend
$env:PGPASSWORD='postgres'; psql -U postgres -d db_unipro -f migrations/071_update_coa_for_budget_report.sql
$env:PGPASSWORD='postgres'; psql -U postgres -d db_unipro -f migrations/072_create_expense_transactions.sql
```

### 2. Master Data Setup
You need the following master data configured:

#### COA Accounts (from migration 071)
- 5203 - Pasangan dan Plesteran (OPERASIONAL_BUDGET)
- 5204 - Pekerjaan Beton (OPERASIONAL_BUDGET)
- 5205 - Pekerjaan Baja (OPERASIONAL_BUDGET)
- etc.

#### Materials
Create materials linked to COA accounts:
```sql
-- Example: Create material with COA mapping
INSERT INTO materials (name, code, unit, unit_price, coa_account_id, created_at, updated_at)
VALUES ('Semen Portland', 'MAT-001', 'sak', 50000, 
        (SELECT id FROM coa_accounts WHERE code = '5203'), 
        NOW(), NOW());
```

#### Vendors (Optional)
```sql
INSERT INTO vendors (name, code, contact_person, phone, email, address, created_at, updated_at)
VALUES ('PT Semen Indonesia', 'VEN-001', 'John Doe', '08123456789', 'contact@semen.com', 'Jakarta', NOW(), NOW());
```

---

## Test Scenarios

### Test 1: Basic PR to Expense Flow

#### Step 1: Start Backend
```powershell
cd backend
.\main.exe
```

#### Step 2: Start Frontend
```powershell
cd frontend
npm run dev
```

#### Step 3: Login
- Navigate to `http://localhost:3000`
- Login with admin credentials
- Username: `admin`
- Password: `admin123`

#### Step 4: Create Purchase Request
1. Go to **Cost Control → Purchase Request Management**
2. Click **Create New PR**
3. Fill in details:
   - **Project**: Select any project
   - **Request Date**: Today
   - **Required Date**: Next week
   - **Vendor**: Select vendor (optional)
   - **Notes**: "Test PR for expense integration"

4. Add Items:
   - **Item 1**:
     - Material: Semen Portland (or any material with COA mapping)
     - Quantity: 100
     - Unit: sak
     - Estimated Price: 50,000
     - Total: 5,000,000
   
   - **Item 2**:
     - Material: Besi Beton (if available)
     - Quantity: 500
     - Unit: kg
     - Estimated Price: 15,000
     - Total: 7,500,000

5. Click **Submit**

#### Step 5: Approve Purchase Request
1. Go to **Dashboard** (or **Employee → Approvals**)
2. Find the pending PR in approval list
3. Click **Review** or **Process**
4. As each approver role:
   - **Purchasing**: Approve
   - **Cost Control**: Approve
   - **GM**: Approve
   - **Project Director**: Approve
   - **Managing Director**: Approve (final approval)

#### Step 6: Verify Expense Creation
1. Go to **Cost Control → Expense Transactions**
2. Filter by the project used in PR
3. **Expected Results**:
   - ✅ 2 expense transactions created (one for each PR item)
   - ✅ Transaction Type: MATERIAL
   - ✅ Reference Type: PR
   - ✅ Reference No: PR-{PROJECT_ID}-{TIMESTAMP}
   - ✅ Description: "PR-{CODE}: {ITEM_NAME} (Material: {MATERIAL_NAME})"
   - ✅ Amount matches PR item total price
   - ✅ COA Account matches material's COA

#### Step 7: Check Budget Report
1. Go to **Cost Control → Budget vs Actual Report**
2. Select the same project
3. **Expected Results**:
   - ✅ Actual expenses show the PR amounts
   - ✅ Expenses grouped by budget category
   - ✅ Variance calculated correctly

#### Step 8: Check Backend Logs
Look for these log messages:
```
✅ Created approval request APP-PR-{TIMESTAMP} for PR PR-{PROJECT_ID}-{TIMESTAMP}
✅ PR {ID} approved - expense creation will be triggered
🔄 Processing approved PR {ID} - creating expense transactions...
✅ Created expense transaction for PR-{CODE} item: {ITEM_NAME} (Amount: {AMOUNT})
✅ Successfully created expense transactions for PR {ID}
✅ Post-approval processing completed for PR {ID} - expense transactions created
```

---

### Test 2: PR with Missing COA Mapping

#### Purpose
Verify graceful handling when material doesn't have COA mapping

#### Steps
1. Create a material WITHOUT COA mapping:
   ```sql
   INSERT INTO materials (name, code, unit, unit_price, created_at, updated_at)
   VALUES ('Test Material No COA', 'MAT-999', 'pcs', 10000, NOW(), NOW());
   ```

2. Create PR with this material
3. Approve PR through workflow
4. **Expected Results**:
   - ✅ PR approval succeeds
   - ✅ Warning logged: "Warning: No COA found for PR item {ID} ({NAME}), skipping expense creation"
   - ✅ No expense transaction created for this item
   - ✅ Other items with COA still create expenses

---

### Test 3: Multiple PRs Same Project

#### Purpose
Verify multiple PRs create separate expense transactions

#### Steps
1. Create PR #1 with 2 items → Approve
2. Create PR #2 with 3 items → Approve
3. Create PR #3 with 1 item → Approve

4. **Expected Results**:
   - ✅ 6 total expense transactions created
   - ✅ Each expense has correct reference to its PR
   - ✅ Budget report shows cumulative total
   - ✅ Can filter expenses by reference_no to see PR-specific expenses

---

### Test 4: PR Rejection (No Expense Creation)

#### Purpose
Verify rejected PRs don't create expenses

#### Steps
1. Create PR
2. Start approval process
3. Reject at any approval step
4. **Expected Results**:
   - ✅ PR status = REJECTED
   - ✅ No expense transactions created
   - ✅ No error logs

---

### Test 5: Concurrent PR Approvals

#### Purpose
Test async processing doesn't cause race conditions

#### Steps
1. Create 5 PRs quickly
2. Approve all 5 PRs in rapid succession
3. **Expected Results**:
   - ✅ All expense transactions created correctly
   - ✅ No duplicate expenses
   - ✅ No missing expenses
   - ✅ All references correct

---

## Verification Checklist

### Database Verification
```sql
-- Check expense transactions were created
SELECT 
    et.id,
    et.project_id,
    et.description,
    et.amount,
    et.reference_type,
    et.reference_no,
    coa.code as coa_code,
    coa.name as coa_name
FROM expense_transactions et
JOIN coa_accounts coa ON et.coa_account_id = coa.id
WHERE et.reference_type = 'PR'
ORDER BY et.created_at DESC;

-- Check PR to Expense linkage
SELECT 
    pr.code as pr_code,
    pr.status,
    pr.total_amount as pr_total,
    COUNT(et.id) as expense_count,
    SUM(et.amount) as expense_total
FROM purchase_requests pr
LEFT JOIN expense_transactions et ON et.reference_id = pr.id AND et.reference_type = 'PR'
WHERE pr.status = 'APPROVED'
GROUP BY pr.id, pr.code, pr.status, pr.total_amount;

-- Check materials with COA mapping
SELECT 
    m.id,
    m.name,
    m.code,
    m.coa_account_id,
    coa.code as coa_code,
    coa.name as coa_name
FROM materials m
LEFT JOIN coa_accounts coa ON m.coa_account_id = coa.id;
```

### API Verification
```bash
# Get expense transactions for a project
curl -X GET "http://localhost:8080/api/v1/projects/{PROJECT_ID}/expenses" \
  -H "Authorization: Bearer {TOKEN}"

# Get budget vs actual report
curl -X GET "http://localhost:8080/api/v1/projects/{PROJECT_ID}/budget-report" \
  -H "Authorization: Bearer {TOKEN}"

# Get PR details
curl -X GET "http://localhost:8080/api/v1/purchase-requests/{PR_ID}" \
  -H "Authorization: Bearer {TOKEN}"
```

---

## Troubleshooting

### Issue: Expenses Not Created

**Possible Causes**:
1. Material doesn't have COA mapping
2. PR status is not APPROVED
3. Approval callback not registered
4. Database connection issues

**Debug Steps**:
```sql
-- Check PR status
SELECT id, code, status, approved_at FROM purchase_requests WHERE id = {PR_ID};

-- Check material COA mapping
SELECT m.id, m.name, m.coa_account_id, coa.code 
FROM materials m 
LEFT JOIN coa_accounts coa ON m.coa_account_id = coa.id
WHERE m.id IN (SELECT material_id FROM purchase_request_items WHERE purchase_request_id = {PR_ID});

-- Check if expense table exists
SELECT COUNT(*) FROM expense_transactions;
```

**Backend Logs to Check**:
```
⚠️ Warning: No COA found for PR item {ID} ({NAME}), skipping expense creation
⚠️ Post-approval processing failed for PR {ID}: {ERROR}
```

### Issue: Wrong Amounts

**Check**:
1. PR item total_price calculation
2. Material unit_price
3. Quantity and unit values

```sql
-- Verify PR item calculations
SELECT 
    pri.id,
    pri.item_name,
    pri.quantity,
    pri.estimated_price,
    pri.total_price,
    (pri.quantity * pri.estimated_price) as calculated_total
FROM purchase_request_items pri
WHERE pri.purchase_request_id = {PR_ID};
```

### Issue: Missing References

**Check**:
```sql
-- Verify reference fields
SELECT 
    id,
    description,
    reference_type,
    reference_id,
    reference_no
FROM expense_transactions
WHERE reference_type = 'PR'
ORDER BY created_at DESC
LIMIT 10;
```

---

## Performance Testing

### Load Test: 100 PRs
```bash
# Create script to generate 100 PRs
# Approve all
# Measure:
# - Time to create all expenses
# - Database query performance
# - Memory usage
# - No deadlocks or race conditions
```

### Expected Performance:
- Expense creation: < 100ms per PR
- No blocking of approval process
- Async processing completes within 1 second
- No database locks

---

## Success Criteria

✅ **Functional Requirements**:
- [ ] PR approval triggers expense creation
- [ ] Correct COA mapping from materials
- [ ] Proper reference tracking
- [ ] Graceful handling of missing COA
- [ ] No duplicate expenses
- [ ] Budget report shows correct data

✅ **Non-Functional Requirements**:
- [ ] Approval process not blocked by expense creation
- [ ] Error handling doesn't fail PR approval
- [ ] Logging provides clear audit trail
- [ ] Performance acceptable under load
- [ ] No data inconsistencies

✅ **User Experience**:
- [ ] No manual expense entry needed
- [ ] Budget reports update automatically
- [ ] Clear indication of PR-sourced expenses
- [ ] Easy to trace expense back to PR

---

## Next Steps After Testing

1. **If tests pass**:
   - Mark Phase 1 as production-ready
   - Document any edge cases found
   - Train users on new workflow
   - Monitor production usage

2. **If issues found**:
   - Document issues in detail
   - Create bug fixes
   - Re-test after fixes
   - Update documentation

3. **Future enhancements**:
   - Implement Phase 2 (Material Tracking)
   - Add expense reversal for cancelled PRs
   - Enhance reporting with PR analytics
   - Add notifications for expense creation
