# Weekly Reports PDF API Documentation

## Overview
API endpoints untuk mengunduh Weekly Reports dalam format PDF, baik individual maupun batch (semua reports dalam ZIP).

## Base URL
```
http://localhost:8080/api/v1/projects/{projectId}/weekly-reports
```

---

## Endpoints

### 1. Download Individual PDF
Download single weekly report sebagai PDF.

**Endpoint:** `GET /api/v1/projects/{projectId}/weekly-reports/{reportId}/pdf`

**Headers:**
```
Authorization: Bearer <your-jwt-token>
```

**Path Parameters:**
- `projectId` (int, required) - ID dari project
- `reportId` (int, required) - ID dari weekly report

**Response:**
- Content-Type: `application/pdf`
- File akan otomatis di-download dengan nama: `weekly_report_{ProjectName}_week{Week}_{Year}.pdf`

**Example:**
```bash
curl -X GET "http://localhost:8080/api/v1/projects/1/weekly-reports/1/pdf" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  --output report.pdf
```

**JavaScript/Frontend Example:**
```javascript
// Using fetch with authorization
async function downloadWeeklyReportPDF(projectId, reportId) {
  const token = localStorage.getItem('token'); // atau dari state management
  
  try {
    const response = await fetch(
      `http://localhost:8080/api/v1/projects/${projectId}/weekly-reports/${reportId}/pdf`,
      {
        method: 'GET',
        headers: {
          'Authorization': `Bearer ${token}`
        }
      }
    );

    if (!response.ok) {
      throw new Error('Failed to download PDF');
    }

    // Convert response to blob
    const blob = await response.blob();
    
    // Create download link
    const url = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `weekly_report_${reportId}.pdf`;
    document.body.appendChild(a);
    a.click();
    a.remove();
    window.URL.revokeObjectURL(url);
    
    console.log('PDF downloaded successfully');
  } catch (error) {
    console.error('Error downloading PDF:', error);
    throw error;
  }
}

// Usage
downloadWeeklyReportPDF(1, 5);
```

---

### 2. Export All PDFs (ZIP)
Download semua weekly reports untuk project sebagai file ZIP yang berisi multiple PDFs.

**Endpoint:** `GET /api/v1/projects/{projectId}/weekly-reports/export-all`

**Headers:**
```
Authorization: Bearer <your-jwt-token>
```

**Path Parameters:**
- `projectId` (int, required) - ID dari project

**Query Parameters (optional):**
- `year` (int, optional) - Filter by specific year (e.g., 2025)

**Response:**
- Content-Type: `application/zip`
- File akan otomatis di-download dengan nama: `weekly_reports_{ProjectName}_{Year|all}.zip`

**Example:**
```bash
# Download all reports
curl -X GET "http://localhost:8080/api/v1/projects/1/weekly-reports/export-all" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  --output all_reports.zip

# Download reports for specific year
curl -X GET "http://localhost:8080/api/v1/projects/1/weekly-reports/export-all?year=2025" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  --output reports_2025.zip
```

**JavaScript/Frontend Example:**
```javascript
// Using fetch with authorization
async function exportAllWeeklyReportsPDF(projectId, year = null) {
  const token = localStorage.getItem('token'); // atau dari state management
  
  // Build URL dengan optional year parameter
  let url = `http://localhost:8080/api/v1/projects/${projectId}/weekly-reports/export-all`;
  if (year) {
    url += `?year=${year}`;
  }
  
  try {
    const response = await fetch(url, {
      method: 'GET',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });

    if (!response.ok) {
      const errorData = await response.json();
      throw new Error(errorData.error || 'Failed to export PDFs');
    }

    // Convert response to blob
    const blob = await response.blob();
    
    // Create download link
    const downloadUrl = window.URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = downloadUrl;
    a.download = year 
      ? `weekly_reports_${year}.zip` 
      : `weekly_reports_all.zip`;
    document.body.appendChild(a);
    a.click();
    a.remove();
    window.URL.revokeObjectURL(downloadUrl);
    
    console.log('ZIP file downloaded successfully');
    return true;
  } catch (error) {
    console.error('Error exporting PDFs:', error);
    throw error;
  }
}

// Usage Examples
exportAllWeeklyReportsPDF(1);        // Export all reports
exportAllWeeklyReportsPDF(1, 2025);  // Export only 2025 reports
```

---

## React/TypeScript Example

```typescript
import { useState } from 'react';

interface DownloadButtonProps {
  projectId: number;
  reportId?: number;
  year?: number;
}

export const WeeklyReportDownloadButton: React.FC<DownloadButtonProps> = ({ 
  projectId, 
  reportId, 
  year 
}) => {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const getAuthToken = () => {
    // Sesuaikan dengan state management Anda
    return localStorage.getItem('token') || '';
  };

  const downloadPDF = async () => {
    if (!reportId) return;
    
    setLoading(true);
    setError(null);

    try {
      const response = await fetch(
        `http://localhost:8080/api/v1/projects/${projectId}/weekly-reports/${reportId}/pdf`,
        {
          headers: {
            'Authorization': `Bearer ${getAuthToken()}`
          }
        }
      );

      if (!response.ok) throw new Error('Download failed');

      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `weekly_report_${reportId}.pdf`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      window.URL.revokeObjectURL(url);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Download failed');
    } finally {
      setLoading(false);
    }
  };

  const exportAll = async () => {
    setLoading(true);
    setError(null);

    try {
      let url = `http://localhost:8080/api/v1/projects/${projectId}/weekly-reports/export-all`;
      if (year) url += `?year=${year}`;

      const response = await fetch(url, {
        headers: {
          'Authorization': `Bearer ${getAuthToken()}`
        }
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.error || 'Export failed');
      }

      const blob = await response.blob();
      const downloadUrl = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = downloadUrl;
      a.download = year 
        ? `weekly_reports_${year}.zip` 
        : `weekly_reports_all.zip`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      window.URL.revokeObjectURL(downloadUrl);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Export failed');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="download-buttons">
      {reportId && (
        <button 
          onClick={downloadPDF} 
          disabled={loading}
          className="btn-download"
        >
          {loading ? 'Downloading...' : 'Download PDF'}
        </button>
      )}
      
      <button 
        onClick={exportAll} 
        disabled={loading}
        className="btn-export-all"
      >
        {loading ? 'Exporting...' : 'Export All PDF'}
      </button>
      
      {error && <div className="error-message">{error}</div>}
    </div>
  );
};
```

---

## Response Codes

### Success
- `200 OK` - File successfully generated and downloaded

### Error Responses

#### 400 Bad Request
```json
{
  "error": "Invalid project ID",
  "details": "strconv.ParseUint: parsing \"abc\": invalid syntax"
}
```

#### 401 Unauthorized
```json
{
  "code": "AUTH_HEADER_MISSING",
  "message": "Please include \"Authorization: Bearer \\u003ctoken\\u003e\" header in your request"
}
```

#### 404 Not Found
```json
{
  "error": "No weekly reports found",
  "message": "There are no weekly reports available for this project"
}
```

#### 500 Internal Server Error
```json
{
  "error": "Failed to generate PDF",
  "details": "pdf generation error details"
}
```

---

## Important Notes

### Authorization
**SEMUA endpoint memerlukan JWT token** dalam Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

Jika token tidak disertakan atau invalid, akan menerima `401 Unauthorized` response.

### ZIP File Structure
Ketika menggunakan export-all endpoint, ZIP file akan berisi:
```
weekly_reports_ProjectName_2025.zip
├── weekly_report_ProjectName_week1_2025.pdf
├── weekly_report_ProjectName_week2_2025.pdf
├── weekly_report_ProjectName_week3_2025.pdf
└── ...
```

### Performance
- Individual PDF: ~500ms - 1s
- Export All (10 reports): ~2s - 5s
- Export All (50+ reports): Bisa memakan waktu lebih lama, pertimbangkan untuk menambahkan loading indicator

### Browser Compatibility
- Modern browsers (Chrome, Firefox, Safari, Edge) fully supported
- IE11 mungkin memerlukan polyfill untuk `fetch` dan `URL.createObjectURL`

---

## Testing

### Manual Testing dengan cURL
```bash
# 1. Login terlebih dahulu untuk mendapatkan token
curl -X POST "http://localhost:8080/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "your-username",
    "password": "your-password"
  }'

# Response akan berisi token
# {"token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}

# 2. Gunakan token untuk download PDF
TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."

curl -X GET "http://localhost:8080/api/v1/projects/1/weekly-reports/1/pdf" \
  -H "Authorization: Bearer $TOKEN" \
  --output test_report.pdf

# 3. Export all PDFs
curl -X GET "http://localhost:8080/api/v1/projects/1/weekly-reports/export-all" \
  -H "Authorization: Bearer $TOKEN" \
  --output all_reports.zip
```

---

## Troubleshooting

### Error: "AUTH_HEADER_MISSING"
**Problem:** Token tidak disertakan dalam request
**Solution:** Pastikan mengirim header `Authorization: Bearer <token>`

### Error: "Invalid project ID"
**Problem:** projectId bukan angka yang valid
**Solution:** Pastikan projectId adalah integer yang valid

### Error: "No weekly reports found"
**Problem:** Tidak ada reports untuk project tersebut
**Solution:** Buat weekly reports terlebih dahulu atau cek projectId

### File tidak ter-download
**Problem:** Browser blocking download atau CORS issue
**Solution:** 
1. Cek browser console untuk error
2. Pastikan CORS headers sudah di-set dengan benar di backend
3. Coba gunakan `a.click()` method seperti pada contoh

---

## Support
Jika ada pertanyaan atau issue, silakan hubungi tim development atau buat issue ticket.

