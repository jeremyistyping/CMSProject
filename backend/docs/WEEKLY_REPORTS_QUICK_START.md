# Quick Start - Weekly Reports PDF Download

## ✅ Backend Implementation Complete

Fitur download PDF dan export all PDF untuk Weekly Reports sudah selesai diimplementasikan!

## 📦 Yang Sudah Dibuat

### 1. Backend Endpoints
✅ **Individual PDF Download**
```
GET /api/v1/projects/{projectId}/weekly-reports/{reportId}/pdf
```

✅ **Export All PDFs (ZIP)**
```
GET /api/v1/projects/{projectId}/weekly-reports/export-all
GET /api/v1/projects/{projectId}/weekly-reports/export-all?year=2025
```

### 2. Files Modified
- ✅ `controllers/weekly_report_controller.go` - Added `ExportAllPDF()` method
- ✅ `routes/project_routes.go` - Added export-all route

### 3. Documentation Created
- ✅ `docs/WEEKLY_REPORTS_PDF_API.md` - Comprehensive API documentation
- ✅ `docs/weekly-reports-helpers.js` - Ready-to-use frontend helper functions

---

## 🚀 Testing the Backend

### 1. Start the Backend Server
```bash
go run cmd/main.go
```

### 2. Login to Get Token
```bash
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "admin",
    "password": "your-password"
  }'
```

Response:
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {...}
}
```

### 3. Test Individual PDF Download
```bash
# Replace {TOKEN} with your actual token
curl -X GET "http://localhost:8080/api/v1/projects/1/weekly-reports/1/pdf" \
  -H "Authorization: Bearer {TOKEN}" \
  --output test_report.pdf
```

### 4. Test Export All PDFs
```bash
# Export all reports
curl -X GET "http://localhost:8080/api/v1/projects/1/weekly-reports/export-all" \
  -H "Authorization: Bearer {TOKEN}" \
  --output all_reports.zip

# Export reports for specific year
curl -X GET "http://localhost:8080/api/v1/projects/1/weekly-reports/export-all?year=2025" \
  -H "Authorization: Bearer {TOKEN}" \
  --output reports_2025.zip
```

---

## 💻 Frontend Integration

### Option 1: Copy Helper File
1. Copy `docs/weekly-reports-helpers.js` ke frontend project Anda
2. Sesuaikan `API_BASE_URL` dan `getAuthToken()` function
3. Import dan gunakan functions

### Option 2: Direct Implementation
Lihat contoh lengkap di `docs/WEEKLY_REPORTS_PDF_API.md`

### Simple Example
```javascript
// Import helper functions
import { downloadWeeklyReportPDF, exportAllWeeklyReportsPDF } from './weekly-reports-helpers';

// Download single PDF
async function handleDownload() {
  try {
    await downloadWeeklyReportPDF(1, 5); // projectId=1, reportId=5
    alert('PDF downloaded!');
  } catch (error) {
    alert('Error: ' + error.message);
  }
}

// Export all PDFs
async function handleExportAll() {
  try {
    await exportAllWeeklyReportsPDF(1); // projectId=1
    alert('All PDFs exported!');
  } catch (error) {
    alert('Error: ' + error.message);
  }
}
```

---

## 🔧 UI Integration Examples

### Button untuk Download Individual PDF
```html
<button 
  onclick="downloadWeeklyReportPDF(1, 5)"
  class="btn-download"
>
  <i class="icon-download"></i>
  Download PDF
</button>
```

### Button untuk Export All
```html
<button 
  onclick="exportAllWeeklyReportsPDF(1)"
  class="btn-export-all"
>
  <i class="icon-export"></i>
  Export All PDF
</button>
```

---

## 📝 Important Notes

### Authorization Required
**SEMUA endpoint memerlukan JWT token!**

Error yang Anda lihat di screenshot:
```json
{
  "code": "AUTH_HEADER_MISSING",
  "message": "Please include \"Authorization: Bearer \\u003ctoken\\u003e\" header"
}
```

**Solution:** Pastikan selalu mengirim header Authorization:
```javascript
headers: {
  'Authorization': `Bearer ${token}`
}
```

### CORS Configuration
Jika ada CORS error, pastikan backend sudah configure CORS headers dengan benar:
```go
// Biasanya sudah ada di main.go atau middleware
router.Use(cors.Default())
```

### File Size Consideration
- Individual PDF: ~100-500KB
- ZIP with 10 reports: ~1-5MB
- ZIP with 50+ reports: Bisa mencapai 10-20MB

Pertimbangkan untuk:
- Menambahkan progress indicator untuk export all
- Menambahkan pagination atau filtering untuk project dengan banyak reports

---

## 📊 Monitoring & Logging

Backend akan log setiap request:
```
[INFO] Generating PDF for report ID: 5
[INFO] PDF generated successfully: weekly_report_ProjectName_week5_2025.pdf
[INFO] Exporting all PDFs for project ID: 1 (10 reports)
[INFO] ZIP file created successfully: weekly_reports_ProjectName_all.zip
```

---

## 🐛 Troubleshooting

### Problem: "AUTH_HEADER_MISSING"
**Cause:** Token tidak dikirim  
**Fix:** Pastikan header Authorization ada dalam setiap request

### Problem: "No weekly reports found"
**Cause:** Project tidak punya weekly reports  
**Fix:** Buat weekly reports terlebih dahulu

### Problem: PDF tidak ter-download
**Cause:** Browser blocking atau CORS  
**Fix:** 
1. Check browser console
2. Verify CORS configuration
3. Test dengan cURL dulu

### Problem: ZIP file corrupt
**Cause:** Incomplete download atau error saat generate  
**Fix:** 
1. Check backend logs
2. Verify semua PDFs ter-generate dengan benar
3. Try download ulang

---

## 📚 Additional Resources

### Full Documentation
- **API Docs:** `docs/WEEKLY_REPORTS_PDF_API.md`
- **Frontend Helpers:** `docs/weekly-reports-helpers.js`

### Related Files
- **Controller:** `controllers/weekly_report_controller.go`
- **Service:** `services/weekly_report_service.go`
- **Model:** `models/weekly_report.go`
- **Routes:** `routes/project_routes.go`

---

## ✨ Next Steps

1. ✅ Backend sudah selesai dan tested
2. 📱 **TODO:** Integrate dengan frontend UI
3. 🎨 **TODO:** Add loading indicators
4. 📊 **TODO:** Add download progress tracking (optional)
5. 🔔 **TODO:** Add success/error notifications

---

## 🆘 Need Help?

Jika ada error atau pertanyaan:
1. Check backend logs
2. Review documentation di `docs/WEEKLY_REPORTS_PDF_API.md`
3. Test dengan cURL terlebih dahulu
4. Check browser console untuk frontend errors

---

## 🎉 Summary

Backend untuk fitur download PDF dan export all PDF sudah **READY TO USE**!

**Endpoints:**
- ✅ Individual PDF: `/api/v1/projects/{id}/weekly-reports/{reportId}/pdf`
- ✅ Export All: `/api/v1/projects/{id}/weekly-reports/export-all`

**Next:** Copy helper functions ke frontend dan integrate dengan UI! 🚀

