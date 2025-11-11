# Setup Instructions: Project-Based Reports

## ✅ Files Yang Sudah Dibuat

### Backend (Go):
1. **`models/project_report.go`** - Data structures untuk 4 reports
2. **`services/project_report_service.go`** - Business logic & database queries
3. **`controllers/project_report_controller.go`** - HTTP handlers
4. **`routes/project_report_routes.go`** - Route registration

### Frontend (Next.js/React):
5. **`app/reports/page.tsx`** - UI untuk display 4 reports (REPLACED)
6. **`app/reports/page.tsx.backup`** - Backup old reports page

---

## 📝 Yang Perlu Ditambahkan

### Step 1: Register Routes di `main.go` atau `routes/routes.go`

Tambahkan ini di file tempat Anda setup routes (biasanya `main.go` atau `routes/routes.go`):

```go
import (
    "app-sistem-akuntansi/routes"
    // ... imports lain
)

func setupRoutes() {
    // ... existing route setup ...
    
    // Register project reports routes
    routes.SetupProjectReportRoutes(v1, db, jwtManager)
    
    // ... existing route setup ...
}
```

**Contoh lengkap di main.go:**

```go
func main() {
    // ... database init, JWT init ...
    
    router := gin.Default()
    v1 := router.Group("/api/v1")
    
    // Existing routes
    routes.SetupAuthRoutes(v1, db, jwtManager)
    routes.SetupProjectRoutes(v1, db, jwtManager)
    // ... other routes ...
    
    // ADD THIS LINE - Project Reports
    routes.SetupProjectReportRoutes(v1, db, jwtManager)
    
    router.Run(":8080")
}
```

---

## 🔗 API Endpoints Yang Akan Tersedia

Setelah register routes, endpoints berikut akan aktif:

### 1. List Available Reports
```
GET /api/v1/project-reports/available
```

**Response:**
```json
{
  "status": "success",
  "data": [
    {
      "id": "budget-vs-actual",
      "name": "Budget vs Actual by COA Group",
      "description": "Menampilkan total estimasi vs realisasi per akun",
      "endpoint": "/api/v1/project-reports/budget-vs-actual",
      "type": "PROJECT"
    },
    // ... 3 reports lainnya
  ]
}
```

### 2. Budget vs Actual Report
```
GET /api/v1/project-reports/budget-vs-actual?start_date=2025-01-01&end_date=2025-12-31&project_id=1
```

**Parameters:**
- `start_date` (required): YYYY-MM-DD
- `end_date` (required): YYYY-MM-DD
- `project_id` (optional): Filter by specific project

**Response:**
```json
{
  "status": "success",
  "data": {
    "report_date": "2025-11-11T...",
    "start_date": "2025-01-01T...",
    "end_date": "2025-12-31T...",
    "project_name": "Project ABC",
    "total_budget": 1000000,
    "total_actual": 950000,
    "total_variance": -50000,
    "variance_rate": -5.0,
    "coa_groups": [
      {
        "coa_code": "5101",
        "coa_name": "Material Cost",
        "coa_type": "EXPENSE",
        "budget": 500000,
        "actual": 480000,
        "variance": -20000,
        "variance_rate": -4.0,
        "status": "UNDER_BUDGET"
      }
    ]
  }
}
```

### 3. Profitability Report
```
GET /api/v1/project-reports/profitability?start_date=2025-01-01&end_date=2025-12-31
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "report_date": "2025-11-11T...",
    "start_date": "2025-01-01T...",
    "end_date": "2025-12-31T...",
    "total_revenue": 5000000,
    "total_direct_cost": 3000000,
    "total_operational": 500000,
    "total_profit": 1500000,
    "overall_margin": 30.0,
    "projects": [
      {
        "project_id": 1,
        "project_code": "PRJ001",
        "project_name": "Project ABC",
        "project_status": "IN_PROGRESS",
        "revenue": 2000000,
        "direct_cost": 1200000,
        "operational_cost": 200000,
        "total_cost": 1400000,
        "gross_profit": 800000,
        "net_profit": 600000,
        "gross_profit_margin": 40.0,
        "net_profit_margin": 30.0
      }
    ]
  }
}
```

### 4. Cash Flow Report
```
GET /api/v1/project-reports/cash-flow?start_date=2025-01-01&end_date=2025-12-31
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "report_date": "2025-11-11T...",
    "beginning_balance": 1000000,
    "total_cash_in": 5000000,
    "total_cash_out": 3000000,
    "net_cash_flow": 2000000,
    "ending_balance": 3000000,
    "projects": [
      {
        "project_id": 1,
        "project_code": "PRJ001",
        "project_name": "Project ABC",
        "cash_in": 2000000,
        "cash_out": 1500000,
        "net_cash_flow": 500000,
        "cash_in_details": [...],
        "cash_out_details": [...]
      }
    ]
  }
}
```

### 5. Cost Summary Report
```
GET /api/v1/project-reports/cost-summary?start_date=2025-01-01&end_date=2025-12-31&project_id=1
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "report_date": "2025-11-11T...",
    "project_name": "Project ABC",
    "total_cost": 3500000,
    "largest_category": "Material",
    "largest_amount": 1500000,
    "categories": [
      {
        "category_code": "MATERIAL",
        "category_name": "Material",
        "total_amount": 1500000,
        "percentage": 42.86,
        "item_count": 25,
        "items": [...]
      },
      {
        "category_code": "LABOUR",
        "category_name": "Labour",
        "total_amount": 1000000,
        "percentage": 28.57,
        "item_count": 15,
        "items": [...]
      }
    ]
  }
}
```

---

## 🗄️ Database Requirements

### Tables yang Digunakan:

1. **`accounts`** - Chart of Accounts
2. **`unified_journal_ledger`** - Journal entries (SSOT)
3. **`projects`** - Project master data
4. **`project_budgets`** - Budget per project per account

### Pastikan table `project_budgets` ada:

```sql
CREATE TABLE IF NOT EXISTS project_budgets (
    id SERIAL PRIMARY KEY,
    project_id INTEGER NOT NULL REFERENCES projects(id),
    account_id INTEGER NOT NULL REFERENCES accounts(id),
    estimated_amount DECIMAL(15,2) NOT NULL DEFAULT 0,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    UNIQUE(project_id, account_id, deleted_at)
);
```

---

## 🧪 Testing

### 1. Test Backend API

```bash
# Test available reports
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/v1/project-reports/available

# Test budget vs actual
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "http://localhost:8080/api/v1/project-reports/budget-vs-actual?start_date=2025-01-01&end_date=2025-12-31"

# Test profitability
curl -H "Authorization: Bearer YOUR_TOKEN" \
  "http://localhost:8080/api/v1/project-reports/profitability?start_date=2025-01-01&end_date=2025-12-31"
```

### 2. Test Frontend

1. Start backend: `go run main.go`
2. Start frontend: `npm run dev`
3. Open browser: `http://localhost:3000/reports`
4. Login dengan user yang memiliki role `admin`, `finance`, atau `director`
5. Click "View Report" pada salah satu report card
6. Input date range dan optional project
7. Click "Generate"

---

## 🎯 Features Summary

### 4 Project-Based Reports:

1. **Budget vs Actual** → Budget planning vs actual spending
2. **Profitability** → Revenue - Costs per project
3. **Cash Flow** → Cash in/out tracking per project
4. **Cost Summary** → Cost breakdown by category

### Key Benefits:

✅ **Project-centric** - Semua report fokus ke project
✅ **COA Integration** - Menggunakan Chart of Accounts
✅ **SSOT** - Data dari unified_journal_ledger
✅ **Flexible Filtering** - By date range & project
✅ **Clean UI** - Simple grid layout dengan modals

---

## 🚀 Next Steps

1. ✅ Add route registration ke main.go
2. ✅ Test all 4 endpoints
3. ✅ Verify database tables exist
4. ✅ Test frontend UI
5. ✅ Create sample project budgets (optional)
6. ✅ Add PDF/CSV export (optional)

---

## 📞 Troubleshooting

### Problem: 404 Not Found
**Solution:** Pastikan routes sudah di-register di main.go

### Problem: 403 Forbidden
**Solution:** User harus login dengan role yang tepat (admin/finance/director)

### Problem: Empty Data
**Solution:** 
- Pastikan ada data di `unified_journal_ledger`
- Pastikan ada data di `projects`
- Untuk Budget vs Actual, pastikan ada data di `project_budgets`

### Problem: Frontend tidak bisa connect
**Solution:**
- Pastikan backend running di port 8080
- Pastikan frontend running di port 3000
- Check CORS settings di backend

---

**Created:** 2025-11-11
**Version:** 1.0
**Status:** Ready for Implementation
