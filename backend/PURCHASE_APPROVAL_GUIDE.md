# 🔐 Purchase Approval Login Guide

## 📋 Approval Flow
```
Purchasing (Employee) → Cost Control (Patrick) → GM (Director) → APPROVED
```

---

## 👥 Login Credentials

### 1️⃣ **PURCHASING (Andi/Employee)** - Pembuat PR
**Role:** `employee`

```
Email: employee@company.com
Password: password123
```

**Akses:**
- ✅ Create Purchase Request
- ✅ View own purchases
- ❌ Cannot approve

**URL:** http://localhost:3000/purchases

---

### 2️⃣ **COST CONTROL (Patrick)** - Approval Step 1
**Role:** `cost_control`

```
Email: patrick@company.com
Password: password123
```

**Akses:**
- ✅ View all purchases
- ✅ **Approve purchases (Step 1)**
- ✅ View cost tracking
- ✅ Access cost control dashboard
- ❌ Cannot create purchases

**URL:** 
- Purchase List: http://localhost:3000/purchases
- Cost Control Dashboard: http://localhost:3000/cost-control

---

### 3️⃣ **GM/DIRECTOR (Pak Marlin)** - Final Approval
**Role:** `director`

```
Email: director@company.com
Password: password123
```

**Akses:**
- ✅ View all purchases
- ✅ **Final approve purchases (Step 2)**
- ✅ View all reports
- ✅ Full oversight

**URL:** http://localhost:3000/purchases

---

### 4️⃣ **ADMIN** - Super User (For Testing)
**Role:** `admin`

```
Email: admin@company.com
Password: password123
```

**Akses:**
- ✅ Full access to everything
- ✅ Create, approve, delete anything
- ✅ User management

**URL:** http://localhost:3000

---

## 🔄 Complete Approval Workflow Test

### **Step 1: Create Purchase Request**
1. Login sebagai **EMPLOYEE**
   ```
   Email: employee@company.com
   Password: password123
   ```
2. Buka http://localhost:3000/purchases
3. Click "Create Purchase" / "Tambah Pembelian"
4. Isi form:
   - **Project**: Pilih project (optional)
   - **Vendor**: Pilih vendor
   - **Items**: Tambah item material
   - **Payment Method**: Pilih metode pembayaran
5. **Save** → Status: `DRAFT`
6. **Submit for Approval** → Status: `PENDING_COST_CONTROL`

---

### **Step 2: Cost Control Approval**
1. **Logout** dari Employee
2. Login sebagai **COST CONTROL**
   ```
   Email: patrick@company.com
   Password: password123
   ```
3. Buka http://localhost:3000/purchases
4. Filter: Status = "Pending Cost Control"
5. Click PR yang baru dibuat
6. Review:
   - Check item details
   - Check price
   - Check budget (if linked to project)
7. **Action:**
   - ✅ **Approve** → Status: `APPROVED_COST_CONTROL` → Forward ke GM
   - ❌ **Reject** → Status: `REJECTED_BY_CC` → Back to purchasing
8. Tambahkan **Comments** (optional)

---

### **Step 3: GM Final Approval**
1. **Logout** dari Cost Control
2. Login sebagai **DIRECTOR**
   ```
   Email: director@company.com
   Password: password123
   ```
3. Buka http://localhost:3000/purchases
4. Filter: Status = "Pending GM"
5. Click PR yang sudah di-approve Cost Control
6. Final Review
7. **Action:**
   - ✅ **Approve** → Status: `APPROVED` ✅ (FINAL)
   - ❌ **Reject** → Status: `REJECTED_BY_GM`
8. Tambahkan **Comments** (optional)

---

### **Step 4: Verify Approval**
1. Login kembali sebagai **EMPLOYEE** atau **ADMIN**
2. Buka Purchase detail
3. Check:
   - ✅ Status: `APPROVED`
   - ✅ Cost Control Approved By: Patrick
   - ✅ Cost Control Approved At: [timestamp]
   - ✅ GM Approved By: Director
   - ✅ GM Approved At: [timestamp]
   - ✅ Approval Comments

---

## 📊 Status Purchase

| Status | Deskripsi | Next Action |
|--------|-----------|-------------|
| `DRAFT` | Belum disubmit | Submit for approval |
| `PENDING_COST_CONTROL` | Menunggu Cost Control | Cost Control approve |
| `APPROVED_COST_CONTROL` | CC approved, pending GM | GM final approve |
| `REJECTED_BY_CC` | Ditolak Cost Control | Revisi & resubmit |
| `PENDING_GM` | Menunggu GM approval | GM approve |
| `APPROVED` | Fully approved ✅ | Proceed to payment |
| `REJECTED_BY_GM` | Ditolak GM | Revisi major |

---

## 🎯 Testing Scenarios

### Scenario 1: Normal Approval Flow
```
Employee (create) → Cost Control (approve) → GM (approve) → APPROVED ✅
```

### Scenario 2: Rejected by Cost Control
```
Employee (create) → Cost Control (reject) → REJECTED_BY_CC
→ Employee (revise) → Cost Control (approve) → GM (approve) → APPROVED ✅
```

### Scenario 3: Rejected by GM
```
Employee (create) → Cost Control (approve) → GM (reject) → REJECTED_BY_GM
→ Employee (major revise) → Start over
```

---

## 🔍 How to Check Approval Status

### Via API (for testing):
```bash
# Get purchase detail with approval info
curl -X GET "http://localhost:8080/api/v1/purchases/1" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

### Response will include:
```json
{
  "id": 1,
  "code": "PR-2025-001",
  "status": "APPROVED",
  "approval_status": "APPROVED",
  "current_approval_step": "COMPLETED",
  "cost_control_approved_by": 6,
  "cost_control_approved_at": "2025-11-11T13:00:00Z",
  "cost_control_comments": "Budget OK, material sesuai",
  "gm_approved_by": 4,
  "gm_approved_at": "2025-11-11T14:00:00Z",
  "gm_comments": "Approved for procurement",
  "project_id": 2,
  "project": {
    "project_name": "Pabrik Gresik"
  }
}
```

---

## 💡 Tips

1. **Gunakan Admin** untuk quick testing tanpa approval flow
2. **Cost Control** bisa melihat budget impact jika PR linked to project
3. **GM** bisa melihat summary all pending approvals
4. Setiap approval **auto-send notification** (jika notifikasi enabled)
5. **Approval history** tersimpan untuk audit trail

---

## 🚨 Troubleshooting

### "Cannot approve" atau button disabled?
- ✅ Check: Anda login dengan role yang tepat?
- ✅ Check: Approval step sudah sesuai? (CC dulu, baru GM)
- ✅ Check: Purchase status sudah `PENDING_COST_CONTROL` atau `PENDING_GM`?

### "Forbidden" error saat approve?
- ✅ Check: Token masih valid?
- ✅ Check: Permission `CanApprove` untuk module `purchases` = true?
- ✅ Check: Role benar? (`cost_control` atau `director`)

### Purchase tidak muncul di list?
- ✅ Check filter status
- ✅ Check: Employee hanya bisa lihat purchase sendiri
- ✅ Cost Control & GM bisa lihat semua

---

## 📞 Support

Jika ada issue:
1. Check backend logs: `D:\Project\CMSProject\backend`
2. Check user permissions: `go run scripts/list_user_credentials.go`
3. Verify approval workflow exists di database

---

**Happy Testing! 🚀**
