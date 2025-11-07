# 🔧 PANDUAN PRAKTIS SWAGGER UI - SISTEM AKUNTANSI

## 🚀 QUICK START GUIDE

### Step 1: Buka Swagger UI
1. Pastikan backend server berjalan
2. Buka browser dan masuk ke: **http://localhost:8080/swagger**
3. Anda akan melihat interface Swagger UI dengan semua API endpoints

### Step 2: Login dan Authorize
1. **Cari section "Authentication"**
2. **Klik endpoint `POST /api/v1/auth/login`**
3. **Klik "Try it out"**
4. **Isi Request body**:
   ```json
   {
     "email": "admin@company.com",
     "password": "admin123"
   }
   ```
5. **Klik "Execute"**
6. **Copy token dari response** (tanpa tanda kutip)
7. **Klik tombol "Authorize" di pojok kanan atas**
8. **Paste token dengan format**: `Bearer <your_token_here>`
9. **Klik "Authorize"**

✅ **Sekarang Anda sudah authorized dan bisa menggunakan semua endpoints!**

---

## 📂 EKSPLORASI SECTIONS DI SWAGGER UI

### 1. 👥 **Users** - User Management
- **Purpose**: Mengelola user, role, dan permissions
- **Key endpoints**:
  - `GET /api/v1/users` - List semua users
  - `POST /api/v1/users` - Create user baru
  - `PUT /api/v1/users/{id}` - Update user

### 2. 📊 **Accounts** - Chart of Accounts  
- **Purpose**: Mengelola bagan akun (COA)
- **Key endpoints**:
  - `GET /api/v1/accounts` - List accounts
  - `POST /api/v1/accounts` - Create account baru
  - `GET /api/v1/accounts/{id}/balance-history` - History saldo

### 3. 👤 **Contacts** - Customer & Vendor Management
- **Purpose**: Mengelola data customer dan vendor
- **Key endpoints**:
  - `GET /api/v1/contacts` - List contacts
  - `POST /api/v1/contacts` - Create contact baru

### 4. 📦 **Products** - Product Management
- **Purpose**: Mengelola produk dan inventory
- **Key endpoints**:
  - `GET /api/v1/products` - List products
  - `POST /api/v1/products/{id}/stock-adjustment` - Stock adjustment

### 5. 🏢 **Assets** - Fixed Assets Management ⭐
- **Purpose**: Mengelola aset tetap dan penyusutan
- **Key endpoints**:
  - `GET /api/v1/assets` - List aset
  - `POST /api/v1/assets` - Create aset baru
  - `GET /api/v1/assets/{id}/depreciation-schedule` - Jadwal penyusutan
  - `GET /api/v1/assets/summary` - Summary aset

### 6. ⚙️ **Settings** - System Settings ⭐
- **Purpose**: Konfigurasi sistem dan company profile
- **Key endpoints**:
  - `GET /api/v1/settings` - Get settings
  - `PUT /api/v1/settings/company` - Update company info
  - `POST /api/v1/settings/company/logo` - Upload logo

### 7. 🏦 **Tax Accounts** - Tax Configuration ⭐
- **Purpose**: Konfigurasi akun pajak (PPN, PPh)
- **Key endpoints**:
  - `GET /api/v1/tax-accounts/current` - Current tax config
  - `GET /api/v1/tax-accounts/suggestions` - Saran akun pajak
  - `POST /api/v1/tax-accounts/validate` - Validasi konfigurasi

### 8. 💰 **Sales** - Sales Management
- **Purpose**: Mengelola transaksi penjualan
- **Key endpoints**:
  - `GET /api/v1/sales/summary` - Sales summary
  - `GET /api/v1/sales/{id}/invoice/pdf` - Invoice PDF

### 9. 🛒 **Purchases** - Purchase Management  
- **Purpose**: Mengelola transaksi pembelian
- **Key endpoints**:
  - `GET /api/v1/purchases/summary` - Purchase summary
  - `POST /api/v1/purchases/{id}/approve` - Approve purchase

### 10. 💳 **Payments** - Payment Processing
- **Purpose**: Mengelola pembayaran (receivable/payable)
- **Key endpoints**:
  - `POST /api/v1/payments/ssot/receivable` - Terima pembayaran
  - `POST /api/v1/payments/ssot/payable` - Bayar ke vendor

---

## 💡 TESTING SCENARIOS PRAKTIS

### Scenario 1: Setup Awal Perusahaan

#### 1.1 Update Company Profile
```
1. Buka: PUT /api/v1/settings/company
2. Try it out
3. Request body:
{
  "company_name": "PT. Contoh Perusahaan",
  "company_address": "Jl. Sudirman No. 123, Jakarta",
  "company_phone": "+6221-5555-1234",
  "company_email": "admin@contohperusahaan.com",
  "tax_number": "01.234.567.8-901.000"
}
4. Execute
```

#### 1.2 Setup Tax Accounts Configuration
```
1. Buka: GET /api/v1/tax-accounts/suggestions
2. Execute untuk melihat saran akun pajak
3. Buka: POST /api/v1/tax-accounts
4. Request body:
{
  "name": "Tax Config 2025",
  "input_tax_account_id": 15,
  "output_tax_account_id": 25,
  "description": "Standard tax configuration"
}
5. Execute
```

### Scenario 2: Mengelola Master Data

#### 2.1 Create Customer Baru
```
1. Buka: POST /api/v1/contacts
2. Request body:
{
  "name": "PT. Pelanggan Utama",
  "email": "contact@pelangganutama.com",
  "phone": "+6221-7777-8888",
  "type": "customer",
  "address": "Jakarta Selatan",
  "tax_number": "02.345.678.9-012.000",
  "payment_terms": 30,
  "credit_limit": 50000000.00
}
3. Execute
```

#### 2.2 Create Product Baru
```
1. Buka: POST /api/v1/products
2. Request body:
{
  "sku": "LAPTOP-001",
  "name": "Laptop Gaming",
  "description": "High performance gaming laptop",
  "category": "Electronics",
  "unit": "unit",
  "purchase_price": 15000000.00,
  "selling_price": 18000000.00,
  "stock_quantity": 10,
  "min_stock": 2,
  "is_active": true,
  "tax_rate": 0.11
}
3. Execute
```

### Scenario 3: Asset Management

#### 3.1 Create Fixed Asset Baru
```
1. Buka: POST /api/v1/assets
2. Request body:
{
  "name": "Server Komputer",
  "asset_code": "IT-SRV-001",
  "category": "IT Equipment",
  "purchase_date": "2025-01-01",
  "purchase_price": 25000000.00,
  "useful_life": 4,
  "depreciation_method": "straight_line",
  "salvage_value": 2500000.00,
  "location": "Server Room",
  "condition": "new",
  "status": "active"
}
3. Execute
```

#### 3.2 Lihat Depreciation Schedule
```
1. Dari response di atas, catat asset ID
2. Buka: GET /api/v1/assets/{id}/depreciation-schedule
3. Masukkan ID di parameter path
4. Execute untuk melihat jadwal penyusutan
```

#### 3.3 Assets Summary
```
1. Buka: GET /api/v1/assets/summary
2. Execute untuk melihat ringkasan semua aset
```

### Scenario 4: Transaksi Bisnis

#### 4.1 Create Sale Transaction
```
1. Buka: POST /api/v1/sales
2. Request body:
{
  "customer_id": 1,
  "sale_date": "2025-01-15",
  "due_date": "2025-02-14",
  "items": [
    {
      "product_id": 1,
      "quantity": 2,
      "unit_price": 18000000.00,
      "discount_percent": 5.0,
      "tax_rate": 0.11
    }
  ],
  "notes": "Sale for 2 gaming laptops",
  "payment_terms": 30
}
3. Execute
```

#### 4.2 Record Payment dari Customer
```
1. Buka: POST /api/v1/payments/ssot/receivable
2. Request body:
{
  "customer_id": 1,
  "amount": 38340000.00,
  "payment_date": "2025-01-20",
  "payment_method": "bank_transfer",
  "bank_account_id": 1,
  "reference": "TRF-20250120-001",
  "notes": "Payment received for invoice",
  "invoices": [
    {
      "invoice_id": 1,
      "amount": 38340000.00
    }
  ]
}
3. Execute
```

### Scenario 5: Reporting & Analysis

#### 5.1 Sales Summary Report
```
1. Buka: GET /api/v1/sales/summary
2. Parameter:
   - period: monthly
   - start_date: 2025-01-01
   - end_date: 2025-01-31
3. Execute untuk melihat summary penjualan
```

#### 5.2 Balance Sheet Report
```
1. Buka: GET /api/v1/reports/balance-sheet
2. Parameter:
   - as_of_date: 2025-01-31
   - format: json
3. Execute untuk melihat neraca
```

---

## 🛠️ TROUBLESHOOTING TIPS

### Issue 1: "Authorization Required" Error
**Solution**:
1. Pastikan sudah login via `POST /api/v1/auth/login`
2. Copy token dari response
3. Klik "Authorize" button
4. Paste dengan format: `Bearer <token>`

### Issue 2: "404 Not Found" Error  
**Possible Causes**:
1. **Endpoint path salah** - Cek spelling dan case sensitivity
2. **Method salah** - Pastikan menggunakan GET/POST/PUT/DELETE yang benar
3. **ID tidak ada** - Untuk endpoint dengan `{id}`, pastikan ID exists di database

### Issue 3: "400 Bad Request" Error
**Possible Causes**:
1. **Request body format salah** - Cek JSON syntax
2. **Required field missing** - Pastikan semua required fields terisi
3. **Data type salah** - Cek apakah number/string/boolean sesuai

### Issue 4: "403 Forbidden" Error
**Causes**:
1. **Insufficient permissions** - User tidak memiliki role/permission yang diperlukan
2. **Admin-only endpoint** - Beberapa endpoint hanya bisa diakses admin

---

## 📊 UNDERSTANDING RESPONSE FORMATS

### Success Response Format
```json
{
  "status": "success",
  "message": "Operation completed successfully",
  "data": {
    // Actual data here
  },
  "pagination": {
    // Only for list endpoints
    "page": 1,
    "limit": 20,
    "total": 100
  }
}
```

### Error Response Format
```json
{
  "status": "error",
  "message": "Error description",
  "errors": {
    // Detailed validation errors
    "field_name": ["error message"]
  },
  "code": "ERROR_CODE"
}
```

---

## 🎯 BEST PRACTICES

### 1. **Testing Strategy**
1. **Start with GET endpoints** - Tidak mengubah data
2. **Test with simple data first** - Minimal required fields saja
3. **Save successful request examples** - Untuk reference nanti
4. **Test error scenarios** - Coba dengan data invalid

### 2. **Data Management**
1. **Create test data in logical order**:
   - Users & Settings dulu
   - Master data (Accounts, Contacts, Products)
   - Transactions (Sales, Purchases)
   - Reports terakhir

### 3. **Security**
1. **Jangan share JWT token** - Token memberikan akses penuh
2. **Logout after testing** - Token akan expire otomatis
3. **Use proper permissions** - Test dengan different user roles

---

## 🔗 USEFUL ENDPOINTS FOR DEVELOPMENT

### Quick Health Check
```
GET /api/v1/health
```

### Check Current User Profile  
```
GET /api/v1/profile
```

### System Status
```
GET /api/v1/monitoring/status
```

---

## 📋 COMMON USE CASES CHECKLIST

### ✅ Setup & Configuration
- [ ] Login successfully
- [ ] Update company profile
- [ ] Configure tax accounts
- [ ] Create chart of accounts
- [ ] Setup bank accounts

### ✅ Master Data Management
- [ ] Create customers
- [ ] Create vendors  
- [ ] Create products
- [ ] Create fixed assets
- [ ] Setup users & permissions

### ✅ Transaction Processing
- [ ] Create sales transactions
- [ ] Generate invoices
- [ ] Record payments (receivable)
- [ ] Create purchases
- [ ] Record vendor payments (payable)

### ✅ Asset Management
- [ ] Register new assets
- [ ] Calculate depreciation
- [ ] Track asset locations
- [ ] Generate asset reports

### ✅ Reporting & Analytics
- [ ] Generate balance sheet
- [ ] Create profit & loss report
- [ ] View cash flow
- [ ] Export reports to PDF/Excel

---

**💡 Pro Tip**: Bookmark halaman Swagger UI dan jadikan sebagai reference utama untuk API development dan testing!

**🎉 Happy API Testing!**