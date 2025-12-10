# Quick Start Guide: Phase 1 PR → Expense Integration

## 🚀 Quick Test (5 Minutes)

### 1. Start Services
```powershell
# Terminal 1: Backend
cd backend
.\main.exe

# Terminal 2: Frontend
cd frontend
npm run dev
```

### 2. Login
- URL: `http://localhost:3000`
- User: `admin`
- Pass: `admin123`

### 3. Create & Approve PR
1. **Cost Control → Purchase Request Management**
2. **Create New PR**:
   - Project: Any
   - Add Item: Semen Portland, Qty: 100, Price: 50,000
3. **Dashboard → Approvals**
4. **Approve** through all steps (Purchasing → Cost Control → GM → Project Director → Managing Director)

### 4. Verify Expenses
1. **Cost Control → Expense Transactions**
2. Look for:
   - ✅ New expense with description "PR-{CODE}: Semen Portland"
   - ✅ Amount: 5,000,000
   - ✅ Reference Type: PR

### 5. Check Logs
Backend console should show:
```
✅ Created expense transaction for PR-{CODE} item: Semen Portland (Amount: 5000000.00)
✅ Post-approval processing completed for PR {ID}
```

---

## 🔍 What to Look For

### Success Indicators:
- ✅ PR status changes to APPROVED
- ✅ Expense transaction appears automatically
- ✅ Expense has correct COA from material
- ✅ Reference fields link to PR
- ✅ Budget report shows the expense

### Common Issues:
- ❌ Material has no COA → Check logs for warning
- ❌ No expense created → Check PR status is APPROVED
- ❌ Wrong amount → Verify PR item total_price

---

## 📊 Quick Verification Queries

```sql
-- Check recent expenses from PRs
SELECT 
    et.description,
    et.amount,
    et.reference_no,
    coa.code,
    coa.name
FROM expense_transactions et
JOIN coa_accounts coa ON et.coa_account_id = coa.id
WHERE et.reference_type = 'PR'
ORDER BY et.created_at DESC
LIMIT 5;

-- Check PR approval status
SELECT code, status, total_amount, approved_at
FROM purchase_requests
ORDER BY created_at DESC
LIMIT 5;
```

---

## 🎯 Success Criteria

- [ ] PR approval completes successfully
- [ ] Expense transaction created automatically
- [ ] Correct COA mapping from material
- [ ] Reference fields properly set
- [ ] Budget report shows expense
- [ ] No errors in backend logs

---

## 📚 Full Documentation

- **Implementation**: `PR_EXPENSE_INTEGRATION.md`
- **Testing**: `TESTING_PR_EXPENSE_INTEGRATION.md`
- **Summary**: `PHASE1_COMPLETION_SUMMARY.md`
- **Integration Plan**: `SYSTEM_INTEGRATION_PLAN.md`

---

## 🆘 Quick Troubleshooting

### No Expense Created?
1. Check PR status: `SELECT status FROM purchase_requests WHERE id = {ID}`
2. Check material COA: `SELECT coa_account_id FROM materials WHERE id = {ID}`
3. Check backend logs for errors

### Wrong Amount?
1. Verify PR item: `SELECT quantity, estimated_price, total_price FROM purchase_request_items WHERE id = {ID}`
2. Check calculation: quantity × estimated_price = total_price

### Missing Reference?
1. Check expense: `SELECT reference_type, reference_id, reference_no FROM expense_transactions WHERE id = {ID}`
2. Should be: reference_type='PR', reference_id={PR_ID}, reference_no='PR-{CODE}'

---

## ✅ Next Steps After Testing

1. **If successful**: Mark as production-ready, train users
2. **If issues**: Document bugs, fix, re-test
3. **Future**: Implement Phase 2 (Material Tracking integration)
