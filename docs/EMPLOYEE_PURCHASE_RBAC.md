# Employee Purchase RBAC Implementation

## Overview
Implementasi Role-Based Access Control (RBAC) untuk membatasi akses Employee pada modul Purchase agar hanya dapat melihat dan mengelola purchase request yang mereka buat sendiri.

## Perubahan yang Diimplementasikan

### 1. Model Layer (`models/purchase.go`)
**File:** `backend/models/purchase.go`

**Perubahan:**
- Menambahkan field `UserID` pada struct `PurchaseFilter` untuk mendukung filtering berdasarkan user yang membuat purchase request.

```go
type PurchaseFilter struct {
    Status           string `json:"status"`
    VendorID         string `json:"vendor_id"`
    UserID           uint   `json:"user_id"` // NEW: Filter by requester user ID
    StartDate        string `json:"start_date"`
    EndDate          string `json:"end_date"`
    // ... fields lainnya
}
```

### 2. Repository Layer (`repositories/purchase_repository.go`)
**File:** `backend/repositories/purchase_repository.go`

**Perubahan:**
- Menambahkan kondisi WHERE di `FindWithFilter` untuk memfilter berdasarkan `user_id` ketika `filter.UserID` tidak kosong.

```go
// Filter by user_id (for employee role - restrict to their own purchases)
if filter.UserID != 0 {
    query = query.Where("user_id = ?", filter.UserID)
}
```

### 3. Controller Layer (`controllers/purchase_controller.go`)
**File:** `backend/controllers/purchase_controller.go`

**Perubahan pada 4 endpoints:**

#### a. `GetPurchases` (List semua purchases)
```go
// RBAC: Employee role can only see their own purchases
if userRole != nil && userRole.(string) == models.RoleEmployee {
    if userID != nil {
        filter.UserID = userID.(uint)
    }
}
```

#### b. `GetPurchase` (Detail purchase by ID)
```go
// RBAC: Employee role can only view their own purchases
userRole, _ := c.Get("role")
userID, _ := c.Get("user_id")
if userRole != nil && userRole.(string) == models.RoleEmployee {
    if userID != nil && purchase.UserID != userID.(uint) {
        c.JSON(http.StatusForbidden, gin.H{
            "error": "Access denied: You can only view your own purchase requests",
            "code":  "INSUFFICIENT_PERMISSION",
        })
        return
    }
}
```

#### c. `UpdatePurchase` (Update purchase)
```go
// RBAC: Employee role can only update their own purchases
userRole, _ := c.Get("role")
userID := c.MustGet("user_id").(uint)
if userRole != nil && userRole.(string) == models.RoleEmployee {
    purchase, err := pc.purchaseService.GetPurchaseByID(uint(id))
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{"error": "Purchase not found"})
        return
    }
    if purchase.UserID != userID {
        c.JSON(http.StatusForbidden, gin.H{
            "error": "Access denied: You can only update your own purchase requests",
            "code":  "INSUFFICIENT_PERMISSION",
        })
        return
    }
}
```

#### d. `DeletePurchase` (Delete purchase)
```go
// RBAC: Employee role can only delete their own purchases
if userRole == models.RoleEmployee && purchase.UserID != userID {
    c.JSON(http.StatusForbidden, gin.H{
        "error": "Access denied: You can only delete your own purchase requests",
        "code":  "INSUFFICIENT_PERMISSION",
    })
    return
}
```

## Behavior Summary

| Role | Akses Purchases |
|------|-----------------|
| **Employee** | Hanya dapat melihat, mengedit, dan menghapus purchase request yang mereka buat sendiri |
| **Finance** | Dapat melihat semua purchases (tidak ada pembatasan) |
| **Director** | Dapat melihat semua purchases (tidak ada pembatasan) |
| **Admin** | Dapat melihat semua purchases (tidak ada pembatasan) |
| **Inventory Manager** | Dapat melihat semua purchases (tidak ada pembatasan) |
| **Auditor** | Dapat melihat semua purchases (tidak ada pembatasan) |

## Testing Scenarios

### Test Case 1: Employee melihat list purchases
**Expected:** Hanya menampilkan purchases yang dibuat oleh employee tersebut.

**Request:**
```http
GET /api/v1/purchases?page=1&limit=10
Authorization: Bearer <employee_token>
```

**Response:**
```json
{
  "data": [
    // Hanya purchases yang user_id = employee's ID
  ],
  "total": 3,
  "page": 1,
  "limit": 10
}
```

### Test Case 2: Employee mengakses purchase milik employee lain
**Expected:** HTTP 403 Forbidden

**Request:**
```http
GET /api/v1/purchases/123
Authorization: Bearer <employee_token>
```

**Response:**
```json
{
  "error": "Access denied: You can only view your own purchase requests",
  "code": "INSUFFICIENT_PERMISSION"
}
```

### Test Case 3: Finance/Admin/Director melihat semua purchases
**Expected:** Menampilkan semua purchases tanpa filter

**Request:**
```http
GET /api/v1/purchases?page=1&limit=10
Authorization: Bearer <finance_token>
```

**Response:**
```json
{
  "data": [
    // Semua purchases dari semua users
  ],
  "total": 150,
  "page": 1,
  "limit": 10
}
```

## Security Considerations

1. **Authorization Check**: Dilakukan di controller layer sebelum data dikirim ke client
2. **Database Level Filter**: Diterapkan di repository layer untuk memastikan query hanya mengambil data yang sesuai
3. **Ownership Validation**: Setiap operasi CRUD pada purchase di-validasi berdasarkan ownership (user_id)
4. **Error Messages**: Memberikan pesan error yang jelas namun tidak mengekspos informasi sensitif

## Future Enhancements

1. Menambahkan audit log untuk setiap attempt akses yang ditolak
2. Implementasi "shared purchases" dimana employee bisa berbagi akses dengan employee lain
3. Menambahkan permission khusus untuk "view all purchases in same department"

## Related Files

- `backend/models/purchase.go` - Model definition dan filter struct
- `backend/repositories/purchase_repository.go` - Database query layer
- `backend/controllers/purchase_controller.go` - API endpoint handlers
- `backend/middleware/rbac.go` - RBAC middleware (existing)
- `backend/models/auth.go` - Role constants definition

## API Endpoints Affected

- `GET /api/v1/purchases` - List purchases (dengan auto-filter untuk employee)
- `GET /api/v1/purchases/:id` - Get purchase detail (dengan ownership check)
- `PUT /api/v1/purchases/:id` - Update purchase (dengan ownership check)
- `DELETE /api/v1/purchases/:id` - Delete purchase (dengan ownership check)

## Notes

- Permission checks menggunakan existing RBAC middleware yang sudah ada
- Tidak ada perubahan pada database schema
- Tidak ada breaking changes untuk role lain (Finance, Director, Admin, dll)
- Employee tetap dapat membuat purchase baru dengan endpoint `POST /api/v1/purchases`
