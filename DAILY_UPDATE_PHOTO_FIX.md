# Daily Update Photo Upload & Display Fix

## Problem
Photo yang diupload ketika create daily update tidak muncul (white blank) di preview dan ketika di download hanya membuka tab baru dengan URL `about:blank#blocked`.

## Root Cause
1. **Windows path separators** - Backend menyimpan `\` instead of `/` dalam URLs
2. **Missing leading slash** - URLs disimpan tanpa `/` di depan (e.g., `uploads\...` instead of `/uploads/...`)
3. **Static file serving** tidak memiliki CORS headers yang tepat
4. **Tidak ada logging** yang cukup untuk debugging photo upload
5. **Error handling** di frontend kurang informatif

## Solutions Implemented

### 1. Backend: Fixed Static File Serving with CORS Headers
**File:** `backend/main.go`

**Changes:**
- Replaced simple `r.Static()` call with custom handler
- Added proper CORS headers for image serving:
  - `Access-Control-Allow-Origin: *`
  - `Access-Control-Allow-Methods: GET, OPTIONS`
  - `Cache-Control: public, max-age=31536000`
- Added file existence check with detailed error logging
- Added OPTIONS method support for CORS preflight

**Why this fixes the issue:**
- Browser was blocking cross-origin image requests
- Now images can be loaded from backend (port 8080) to frontend (port 3000)
- Better error messages when files are not found

### 2. Backend: Enhanced Logging for Photo Uploads
**File:** `backend/controllers/daily_update_controller.go`

**Changes:**
- Added detailed logging at each step of photo upload:
  - Number of files received
  - Save success/failure
  - Path conversion (file path → public URL)
  - URLs stored in database

**Logging output examples:**
```
📷 Received 3 photo files for upload
✅ Saved 3 photos to disk
🔗 Photo 1: ./uploads/daily-updates/20241112-abc123.jpg -> /uploads/daily-updates/20241112-abc123.jpg
🔗 Photo 2: ./uploads/daily-updates/20241112-def456.jpg -> /uploads/daily-updates/20241112-def456.jpg
🔗 Photo 3: ./uploads/daily-updates/20241112-ghi789.jpg -> /uploads/daily-updates/20241112-ghi789.jpg
💾 Storing 3 photo URLs in database: [/uploads/daily-updates/20241112-abc123.jpg /uploads/daily-updates/20241112-def456.jpg /uploads/daily-updates/20241112-ghi789.jpg]
```

### 3. Backend: Fixed Windows Path Separators (CRITICAL FIX)
**File:** `backend/utils/file_upload.go`

**The Problem:**
- Windows uses backslash `\` for file paths
- URLs must use forward slash `/`
- Original code: `http://localhost:8080uploads\daily-updates\...` ❌
- Fixed code: `http://localhost:8080/uploads/daily-updates/...` ✅

**Changes:**
```go
func GetPublicURL(filePath string) string {
    // 1. Convert Windows backslashes to forward slashes
    publicPath := strings.ReplaceAll(filePath, "\\", "/")
    
    // 2. Replace ./uploads with /uploads
    publicPath = strings.Replace(publicPath, "./uploads", "/uploads", 1)
    
    // 3. Ensure leading /
    if !strings.HasPrefix(publicPath, "/") {
        if strings.HasPrefix(publicPath, "uploads") {
            publicPath = "/" + publicPath
        }
    }
    
    return publicPath
}
```

**Why this fixes the `about:blank#blocked` issue:**
- Browser was trying to navigate to malformed URL
- Now URLs are properly formatted for web access

### 4. Frontend: Improved Error Handling & Debugging
**File:** `frontend/src/components/projects/PhotoGallery.tsx`

**Changes:**
- Added console logging for URL conversion
- Added checks for empty URLs
- Added better error messages
- Existing fallback images still work

**Console output examples:**
```
Converting relative URL to absolute: /uploads/daily-updates/20241112-abc123.jpg -> http://localhost:8080/uploads/daily-updates/20241112-abc123.jpg
```

## IMPORTANT: Old Data in Database

⚠️ **If you already uploaded photos before this fix**, those old URLs in database still have the wrong format!

**Old format** (stored in database before fix):
- `uploads\daily-updates\20251112-102309-841fd66c.png` ❌
- Missing leading `/` and has backslashes

**New format** (after fix):
- `/uploads/daily-updates/20251112-114712-b1a4464f.png` ✅
- Has leading `/` and uses forward slashes

### How to Fix Old Data

**Option 1: Use Complete Fix Script (RECOMMENDED)**
```powershell
cd "C:\Users\jeremia.kaligis\Desktop\CMS New"
.\complete-photo-fix.ps1
```

This script will:
1. ✅ Fix ALL old photo URLs in database
2. ✅ Check if files exist on disk
3. ✅ Restart backend server

**Option 2: Manual Fix**
```powershell
cd "C:\Users\jeremia.kaligis\Desktop\CMS New\backend"
go run ./cmd/scripts/fix_old_photo_urls.go
```

Then restart backend manually.

## How to Test

### Quick Test with Script (RECOMMENDED)

**Run the automated test script:**
```powershell
# Navigate to project root
cd "C:\Users\jeremia.kaligis\Desktop\CMS New"

# Run the test script
.\test-photo-fix.ps1
```

**The script will:**
1. ✅ Run unit tests for GetPublicURL function
2. ✅ Create uploads directory if needed
3. ✅ Show existing photos
4. ✅ Start the backend server with detailed logging

### Manual Test Steps

#### Step 1: Run Unit Tests
```powershell
cd "C:\Users\jeremia.kaligis\Desktop\CMS New\backend"
go test -v ./utils -run TestGetPublicURL
```

**Expected output:**
```
=== RUN   TestGetPublicURL
=== RUN   TestGetPublicURL/Windows_path_with_backslashes
=== RUN   TestGetPublicURL/Unix_path_with_forward_slashes
...
--- PASS: TestGetPublicURL (0.00s)
PASS
```

#### Step 2: Restart Backend
```powershell
cd "C:\Users\jeremia.kaligis\Desktop\CMS New\backend"
go run main.go
```

**Expected output:**
```
📁 Serving static files from: C:\Users\jeremia.kaligis\Desktop\CMS New\backend\uploads
Server starting on port 8080
```

### Step 2: Test Photo Upload

1. **Open the application** in your browser
2. **Navigate to Projects** page
3. **Open a project** details
4. **Click "Add Daily Update"** button
5. **Fill in the form:**
   - Date: Select today
   - Weather: Select any
   - Workers Present: Enter a number
   - Work Description: Enter some text
   - **Upload Photos:** Select 1-3 photos

6. **Click "Create Daily Update"**

**Expected backend logs:**
```
📝 CreateDailyUpdate - Project ID: 1, Content-Type: multipart/form-data; boundary=...
📋 Parsed Data - Date: 2025-11-12...
📷 Received 2 photo files for upload
✅ Saved 2 photos to disk
🔗 Photo 1: ./uploads/daily-updates/20241112-abc123.jpg -> /uploads/daily-updates/20241112-abc123.jpg
🔗 Photo 2: ./uploads/daily-updates/20241112-def456.jpg -> /uploads/daily-updates/20241112-def456.jpg
💾 Storing 2 photo URLs in database: [...]
💾 Calling service.CreateDailyUpdate...
✅ Daily update created successfully - ID: 1
```

### Step 3: Verify Photo Display

1. **View the daily update** you just created
2. **Click on the photo badge** (e.g., "2 photos")
3. **Photo gallery should open** showing thumbnail grid
4. **Photos should be visible** (not blank/white)

**Expected browser console logs:**
```
Converting relative URL to absolute: /uploads/daily-updates/20241112-abc123.jpg -> http://localhost:8080/uploads/daily-updates/20241112-abc123.jpg
Converting relative URL to absolute: /uploads/daily-updates/20241112-def456.jpg -> http://localhost:8080/uploads/daily-updates/20241112-def456.jpg
```

### Step 4: Test Photo Download

1. **Hover over a photo** in the gallery
2. **Click the download button** (blue icon in top-right corner)
3. **Photo should download** OR **open in a new tab** (not `about:blank#blocked`)

**Expected browser console log:**
```
Downloading photo: http://localhost:8080/uploads/daily-updates/20241112-abc123.jpg
```

### Step 5: Test Photo Viewing (Lightbox)

1. **Click on a photo** in the gallery grid
2. **Full-screen lightbox should open**
3. **Photo should be visible** in large size
4. **Navigation arrows** should work (if multiple photos)
5. **Download button** in bottom bar should work

## Troubleshooting

### Issue: Photos still not showing (blank/white)

**Check 1: Verify files are saved**
```powershell
# Check if uploads directory exists
Get-ChildItem "C:\Users\jeremia.kaligis\Desktop\CMS New\backend\uploads\daily-updates"
```

**Expected:** You should see files like `20241112-abc123.jpg`

**Check 2: Test direct URL access**
Open browser and go to: `http://localhost:8080/uploads/daily-updates/[filename].jpg`

**Expected:** Photo should display directly in browser

**Check 3: Check browser console for errors**
Open browser DevTools (F12) → Console tab
Look for:
- Network errors (404, CORS errors)
- Console warnings/errors

### Issue: 404 Error when accessing photos

**Possible causes:**
1. Backend server not running
2. Uploads directory doesn't exist
3. File path stored in database is incorrect

**Solution:**
```powershell
# Create uploads directory if it doesn't exist
New-Item -ItemType Directory -Path "C:\Users\jeremia.kaligis\Desktop\CMS New\backend\uploads\daily-updates" -Force
```

### Issue: CORS Error in browser console

**Error message:**
```
Access to fetch at 'http://localhost:8080/uploads/...' from origin 'http://localhost:3000' has been blocked by CORS policy
```

**Solution:**
This should be fixed by the changes in main.go. Verify:
1. Backend server was restarted after changes
2. Check backend logs for file serving requests

### Issue: `about:blank#blocked` when downloading

**This was the original issue - should be fixed now.**

**If it still happens:**
1. Check browser console for error messages
2. Verify the download URL is correct
3. Try right-click → "Save link as..." instead

## Technical Details

### Photo Upload Flow

```
1. Frontend (DailyUpdateModal)
   ↓ FormData with photos
2. Backend (DailyUpdateController.CreateDailyUpdate)
   ↓ Saves to ./uploads/daily-updates/[timestamp]-[uuid].jpg
3. utils.SaveMultipleFiles
   ↓ Returns file paths: ["./uploads/daily-updates/..."]
4. utils.GetPublicURL
   ↓ Converts to: ["/uploads/daily-updates/..."]
5. Database (PostgreSQL)
   ↓ Stores URLs in photos[] array field
6. Response sent to frontend
```

### Photo Display Flow

```
1. Frontend fetches daily update data
   ↓ Receives photos: ["/uploads/daily-updates/..."]
2. PhotoGallery.getFullUrl()
   ↓ Converts to: "http://localhost:8080/uploads/daily-updates/..."
3. Browser requests image
   ↓ GET http://localhost:8080/uploads/daily-updates/...
4. Backend static file handler (main.go)
   ↓ Serves file with CORS headers
5. Image displays in browser
```

## Environment Variables

Make sure your `.env` file has:

```env
# Backend URL for file serving
NEXT_PUBLIC_API_URL=http://localhost:8080
```

Or set it in your shell:
```powershell
$env:NEXT_PUBLIC_API_URL = "http://localhost:8080"
```

## File Permissions

Ensure the `uploads` directory has write permissions:
```powershell
# Check permissions
Get-Acl "C:\Users\jeremia.kaligis\Desktop\CMS New\backend\uploads"

# If needed, grant full control (be careful with this in production)
$acl = Get-Acl "C:\Users\jeremia.kaligis\Desktop\CMS New\backend\uploads"
$rule = New-Object System.Security.AccessControl.FileSystemAccessRule("Everyone","FullControl","Allow")
$acl.SetAccessRule($rule)
Set-Acl "C:\Users\jeremia.kaligis\Desktop\CMS New\backend\uploads" $acl
```

## Summary

The fix addresses three main issues:

1. ✅ **CORS Headers** - Photos can now be loaded cross-origin
2. ✅ **Better Logging** - Easy to debug photo upload issues
3. ✅ **Error Handling** - Clear error messages when things go wrong

The photos should now:
- ✅ Display in preview (not blank)
- ✅ Download properly (not `about:blank#blocked`)
- ✅ Work in lightbox view
- ✅ Have clear error messages if something fails

---

**Created:** November 12, 2025
**Author:** AI Assistant
**Status:** Ready for Testing

