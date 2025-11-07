# 📚 DOKUMENTASI API SISTEM AKUNTANSI

Selamat datang di dokumentasi lengkap API Sistem Akuntansi! Berikut adalah panduan untuk menggunakan dokumentasi ini.

## 🔗 Akses Cepat Swagger UI
- **URL**: http://localhost:8080/swagger
- **Alternatif**: http://localhost:8080/docs
- **JSON Spec**: http://localhost:8080/openapi/enhanced-doc.json

---

## 📋 DAFTAR DOKUMENTASI

### 1. 📖 [COMPLETE_API_DOCUMENTATION.md](./COMPLETE_API_DOCUMENTATION.md)
**Dokumentasi lengkap dan detail semua API endpoints**
- Penjelasan mendalam setiap endpoint
- Contoh request dan response lengkap
- Parameter dan query options
- Error handling
- Workflow examples
- Best practices

**Kapan digunakan**: Ketika Anda perlu pemahaman detail tentang API tertentu

### 2. 🔧 [SWAGGER_UI_PRACTICAL_GUIDE.md](./SWAGGER_UI_PRACTICAL_GUIDE.md)
**Panduan praktis menggunakan Swagger UI**
- Step-by-step tutorial menggunakan Swagger UI
- Testing scenarios praktis
- Troubleshooting tips
- Common use cases
- Setup workflows

**Kapan digunakan**: Untuk pemula yang baru menggunakan Swagger UI atau butuh panduan praktis testing

### 3. ⚡ [API_QUICK_REFERENCE.md](./API_QUICK_REFERENCE.md)
**Quick reference untuk semua endpoints**
- Tabel referensi cepat semua endpoints
- Common parameters
- Example curl commands
- Response codes
- Pro tips

**Kapan digunakan**: Ketika Anda sudah familiar dengan API dan butuh reference cepat

---

## 🎯 MULAI DARI MANA?

### 👶 **Pemula** (Belum pernah menggunakan API ini)
1. **Mulai dengan**: [SWAGGER_UI_PRACTICAL_GUIDE.md](./SWAGGER_UI_PRACTICAL_GUIDE.md)
2. **Ikuti**: Quick Start Guide untuk login dan authorization
3. **Coba**: Testing scenarios yang disediakan
4. **Reference**: [API_QUICK_REFERENCE.md](./API_QUICK_REFERENCE.md) untuk endpoints

### 🧑‍💻 **Developer** (Sudah familiar dengan REST API)
1. **Reference cepat**: [API_QUICK_REFERENCE.md](./API_QUICK_REFERENCE.md)
2. **Detail spesifik**: [COMPLETE_API_DOCUMENTATION.md](./COMPLETE_API_DOCUMENTATION.md)
3. **Langsung testing**: http://localhost:8080/swagger

### 🎓 **Advanced** (Butuh integrasi atau development)
1. **Full documentation**: [COMPLETE_API_DOCUMENTATION.md](./COMPLETE_API_DOCUMENTATION.md)
2. **Swagger JSON**: http://localhost:8080/openapi/enhanced-doc.json
3. **Custom testing**: Gunakan Postman, Insomnia, atau tools lain

---

## 🔑 LOGIN CREDENTIALS

### Default Admin Account
- **Email**: `admin@company.com`
- **Password**: `admin123`
- **Role**: `admin` (full access)

### Test Accounts (jika tersedia)
- **Finance**: `finance@company.com` / `finance123`
- **User**: `user@company.com` / `user123`

---

## 🗂️ STRUKTUR API

### 📂 **Core Modules** (Sudah ada di Swagger UI)
- 👥 **Users** - User management dan permissions
- 📊 **Accounts** - Chart of accounts (bagan akun)
- 👤 **Contacts** - Customer dan vendor management
- 📦 **Products** - Product dan inventory management
- 🏢 **Assets** - Fixed assets dan depreciation ⭐
- ⚙️ **Settings** - System settings dan configuration ⭐
- 🏦 **Tax Accounts** - Tax configuration (PPN, PPh) ⭐

### 💼 **Business Modules**
- 💰 **Sales** - Sales transactions dan invoicing
- 🛒 **Purchases** - Purchase orders dan approval
- 💳 **Payments** - Payment processing (SSOT)
- 🏦 **Cash & Bank** - Cash and bank management

### 📊 **Reporting & Analytics**
- 📈 **Reports** - Financial reports (Balance Sheet, P&L, etc.)
- 📊 **Analytics** - Business intelligence dan insights
- 🔍 **Monitoring** - System monitoring dan health checks

---

## 🚀 QUICK START

### 1. Pastikan Server Running
```bash
# Cek apakah server berjalan
curl http://localhost:8080/api/v1/health
```

### 2. Login dan Get Token
```bash
# Login
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email": "admin@company.com", "password": "admin123"}'

# Copy token dari response
```

### 3. Test API Call
```bash
# Test dengan token
curl -H "Authorization: Bearer <your_token>" \
  http://localhost:8080/api/v1/users
```

### 4. Buka Swagger UI
- URL: http://localhost:8080/swagger
- Klik "Authorize" dan masukkan token
- Explore dan test semua endpoints

---

## ✨ FITUR TERBARU

### 🎉 Assets Management
- ✅ Create dan manage fixed assets
- ✅ Automatic depreciation calculation
- ✅ Depreciation schedules
- ✅ Asset categories
- ✅ Asset summary reports

### ⚙️ Settings Management  
- ✅ Company profile management
- ✅ System configuration
- ✅ Logo upload
- ✅ Settings history tracking

### 🏦 Tax Accounts Configuration
- ✅ Tax account setup
- ✅ Account suggestions
- ✅ Validation tools
- ✅ Cache management

---

## 🛠️ TROUBLESHOOTING

### Common Issues

#### 1. "Authorization Required" Error
**Solution**: Pastikan sudah login dan menggunakan Bearer token

#### 2. "404 Not Found" untuk Assets/Settings
**Solution**: Sudah diperbaiki! Assets dan Settings sekarang tersedia di Swagger UI

#### 3. Server tidak response
**Solution**: 
```bash
# Restart server
go run main.go

# Atau cek status
curl http://localhost:8080/api/v1/health
```

#### 4. Swagger UI tidak load
**Solution**:
1. Clear browser cache (Ctrl+Shift+Delete)
2. Refresh page (Ctrl+F5)
3. Check browser console untuk errors

---

## 📞 SUPPORT

### 📚 Documentation
- **Complete Guide**: [COMPLETE_API_DOCUMENTATION.md](./COMPLETE_API_DOCUMENTATION.md)
- **Practical Guide**: [SWAGGER_UI_PRACTICAL_GUIDE.md](./SWAGGER_UI_PRACTICAL_GUIDE.md)  
- **Quick Reference**: [API_QUICK_REFERENCE.md](./API_QUICK_REFERENCE.md)

### 🔗 Links
- **Swagger UI**: http://localhost:8080/swagger
- **Health Check**: http://localhost:8080/api/v1/health
- **System Status**: http://localhost:8080/api/v1/monitoring/status

### 🎯 Best Practices
1. **Always use Authorization header** untuk protected endpoints
2. **Check response codes** untuk error handling
3. **Use pagination** untuk large datasets
4. **Test in development** sebelum production
5. **Backup data** sebelum testing destructive operations

---

## 📅 VERSION HISTORY

### v2.0.0 (Current)
- ✅ Added Assets Management API
- ✅ Added Settings Management API  
- ✅ Added Tax Accounts Configuration API
- ✅ Fixed Swagger UI display issues
- ✅ Enhanced documentation
- ✅ Improved error handling

### v1.0.0
- ✅ Core accounting modules
- ✅ Authentication & authorization
- ✅ Basic CRUD operations
- ✅ Financial reporting

---

**🎉 Happy API Development!**

*Untuk pertanyaan atau masalah, silakan refer ke dokumentasi di atas atau test langsung di Swagger UI.*