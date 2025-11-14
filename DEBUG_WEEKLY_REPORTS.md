# 🔧 Debug Weekly Reports Error - Complete Guide

## 📋 Current Error

**Error Message:**
```
Error: Failed to create weekly report
TypeError: Cannot read properties of null (reading 'data')
404 Not Found: POST /projects/1/weekly-reports
```

## ✅ Fixes Applied

### **1. Fixed `weeklyReportService.ts`** ✅

All methods now have proper null checking:

```typescript
// ✅ GET - List reports (safe, returns empty array)
async getWeeklyReports(projectId: number): Promise<WeeklyReportDTO[]>

// ✅ GET - Single report (throws meaningful error)
async getWeeklyReport(projectId: number, reportId: number): Promise<WeeklyReportDTO>

// ✅ POST - Create report (throws meaningful error with details)
async createWeeklyReport(projectId: number, data: CreateWeeklyReportRequest): Promise<WeeklyReportDTO>

// ✅ PUT - Update report (throws meaningful error)
async updateWeeklyReport(projectId: number, reportId: number, data: UpdateWeeklyReportRequest): Promise<WeeklyReportDTO>

// ✅ DELETE - Delete report (throws meaningful error)
async deleteWeeklyReport(projectId: number, reportId: number): Promise<void>
```

### **2. Added NumberInput with Arrow Up/Down** ✅

Fields now have stepper arrows:
- Total Work Days
- Weather Delays
- Team Size

## 🔍 Debugging Steps

### Step 1: Verify Backend is Running

```powershell
# Check health endpoint
Invoke-WebRequest -Uri "http://localhost:8080/api/v1/health" -Method GET -UseBasicParsing
```

Expected: `{"status":"ok"}` with status code 200 ✅

### Step 2: Check Backend Logs

Open terminal where backend is running and look for:
```
✅ /api/v1/projects/:id/weekly-reports routes registered
```

If NOT found, backend needs restart:
```powershell
cd backend
go run cmd/main.go
```

### Step 3: Verify Route Registration

Check if routes are registered in `backend/routes/project_routes.go`:

```go
// Line 67-72: Weekly Reports routes
projects.GET("/:id/weekly-reports", weeklyReportController.GetWeeklyReports)
projects.GET("/:id/weekly-reports/:reportId", weeklyReportController.GetWeeklyReport)
projects.POST("/:id/weekly-reports", weeklyReportController.CreateWeeklyReport)
projects.PUT("/:id/weekly-reports/:reportId", weeklyReportController.UpdateWeeklyReport)
projects.DELETE("/:id/weekly-reports/:reportId", weeklyReportController.DeleteWeeklyReport)
projects.GET("/:id/weekly-reports/:reportId/pdf", weeklyReportController.GeneratePDF)
```

Check if called in `backend/routes/routes.go`:

```go
// Line 1168: Setup project routes
SetupProjectRoutes(protected, db)
```

### Step 4: Test with cURL (Manual Test)

#### A. Login First
```powershell
$body = @{
    email = "admin@company.com"
    password = "password123"
} | ConvertTo-Json

$response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" `
    -Method POST `
    -Body $body `
    -ContentType "application/json"

$token = $response.access_token
Write-Host "Token: $token"
```

#### B. Get Projects
```powershell
$headers = @{
    "Authorization" = "Bearer $token"
}

$projects = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/projects" `
    -Method GET `
    -Headers $headers

$projects.data
```

#### C. Test Create Weekly Report
```powershell
$projectId = 1  # Use actual project ID from step B

$reportData = @{
    project_id = $projectId
    week = 46
    year = 2025
    project_manager = "Test Manager"
    total_work_days = 5
    weather_delays = 0
    team_size = 10
    accomplishments = "Test accomplishments"
    challenges = "Test challenges"
    next_week_priorities = "Test priorities"
} | ConvertTo-Json

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

$result = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/projects/$projectId/weekly-reports" `
    -Method POST `
    -Body $reportData `
    -Headers $headers

$result
```

## 🚨 Common Issues & Solutions

### Issue 1: 404 Not Found (Route Not Registered)

**Symptoms:**
- 404 error for `/projects/1/weekly-reports`
- Backend logs don't show route registration

**Solution:**
1. Restart backend:
   ```powershell
   cd backend
   go run cmd/main.go
   ```

2. Check logs for:
   ```
   ✅ /api/v1/projects routes registered
   ```

### Issue 2: 401 Unauthorized (Token Issue)

**Symptoms:**
- 401 error
- "Unauthorized" or "Invalid token" message

**Solution:**
1. Clear browser localStorage:
   ```javascript
   localStorage.clear()
   ```

2. Logout and login again

3. Check token in browser DevTools:
   ```javascript
   console.log(localStorage.getItem('token'))
   ```

### Issue 3: 400 Bad Request (Validation Error)

**Symptoms:**
- 400 error
- "Invalid request body" message

**Solution:**
Check required fields:
- ✅ `project_id` (number)
- ✅ `week` (number, 1-53)
- ✅ `year` (number)
- ✅ `total_work_days` (number, >= 0)
- ✅ `weather_delays` (number, >= 0)
- ✅ `team_size` (number, >= 1)

### Issue 4: 500 Internal Server Error (Backend Error)

**Symptoms:**
- 500 error
- Backend crashes or shows error

**Solution:**
1. Check backend logs for stack trace

2. Common causes:
   - Database connection issue
   - Missing database table `weekly_reports`
   - Validation error in backend

3. Check database:
   ```sql
   -- Connect to database
   psql -U admin -d sistem_akuntansi
   
   -- Check if table exists
   \dt weekly_reports
   
   -- Check table structure
   \d weekly_reports
   ```

## 🎯 Frontend Error Handling

The frontend now shows better error messages:

### Before:
```
Error: Failed to create weekly report
TypeError: Cannot read properties of null (reading 'data')
```

### After:
```
Error: Failed to create weekly report - Project not found
Error: Failed to create weekly report - Invalid week number
Error: Failed to create weekly report - empty response from server
```

## 📝 Testing Checklist

- [ ] Backend is running (`http://localhost:8080/api/v1/health` returns OK)
- [ ] Frontend is running (`http://localhost:3000`)
- [ ] User is logged in (check browser localStorage for token)
- [ ] Project exists (check `/api/v1/projects` returns data)
- [ ] Route is registered (backend logs show project routes)
- [ ] Database table exists (`weekly_reports` table in database)
- [ ] Form fields are valid (all required fields filled)
- [ ] Network requests visible in browser DevTools Network tab

## 🔍 Browser DevTools Debugging

### 1. Open DevTools (F12)

### 2. Network Tab
- Filter: `weekly-reports`
- Check:
  - Request URL: Should be `/api/v1/projects/1/weekly-reports`
  - Method: `POST`
  - Status: Should be `201` (Created) or `200` (OK)
  - Request Headers: Check `Authorization: Bearer <token>`
  - Request Payload: Check JSON data

### 3. Console Tab
Look for errors:
```
❌ Error creating weekly report: ...
❌ Failed to create report: ...
```

### 4. Application Tab
Check localStorage:
```javascript
localStorage.getItem('token')  // Should have JWT token
localStorage.getItem('user')   // Should have user data
```

## 🚀 Quick Fix Commands

### Restart Backend:
```powershell
cd "C:\Users\jeremia.kaligis\Desktop\CMS New\backend"
go run cmd/main.go
```

### Restart Frontend:
```powershell
cd "C:\Users\jeremia.kaligis\Desktop\CMS New\frontend"
npm run dev
```

### Clear Frontend Cache:
```powershell
# In browser console (F12)
localStorage.clear()
# Then refresh: CTRL + F5
```

### Test Backend Manually:
```powershell
cd "C:\Users\jeremia.kaligis\Desktop\CMS New"
.\test-weekly-reports.ps1
```

## 📊 Expected API Response

### Success Response (201 Created):
```json
{
  "status": "success",
  "message": "Weekly report created successfully",
  "data": {
    "id": 1,
    "project_id": 1,
    "project_name": "Project ABC",
    "week": 46,
    "year": 2025,
    "week_label": "2025-W46",
    "project_manager": "John Doe",
    "total_work_days": 5,
    "weather_delays": 0,
    "team_size": 10,
    "accomplishments": "...",
    "challenges": "...",
    "next_week_priorities": "...",
    "generated_date": "2025-11-13T09:00:00Z",
    "created_by": "admin@company.com",
    "created_at": "2025-11-13T09:00:00Z",
    "updated_at": "2025-11-13T09:00:00Z"
  }
}
```

### Error Response (400 Bad Request):
```json
{
  "error": "Invalid request body",
  "details": "week: must be between 1 and 53"
}
```

### Error Response (404 Not Found):
```json
{
  "error": "Project not found",
  "details": "Project with ID 999 does not exist"
}
```

## 💡 Tips

1. **Always check backend logs first** - Most issues show up there
2. **Test with small data** - Start with minimal required fields
3. **Check Network tab** - See actual request/response
4. **Clear cache** - When things don't make sense
5. **Restart both servers** - When in doubt

## ✅ Status

- [x] Frontend service fixed (null handling)
- [x] Frontend component fixed (error handling)
- [x] NumberInput added (arrow up/down)
- [x] Backend route exists (verified)
- [x] Backend controller exists (verified)
- [x] Backend service exists (verified)
- [ ] **Need to verify:** Backend is actually running
- [ ] **Need to verify:** Database table exists
- [ ] **Need to test:** Create weekly report end-to-end

## 🎯 Next Action

**Run this command to verify backend:**
```powershell
cd backend
go run cmd/main.go
```

Watch for:
```
✅ Database migrations complete
✅ Routes registered
🚀 Server starting on port 8080
```

Then test in browser:
1. Open `http://localhost:3000/projects`
2. Click on any project
3. Go to "Weekly Reports" tab
4. Fill form and click "Generate Report"
5. Should work! 🎉

