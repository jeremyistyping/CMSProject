# Test Photo Fix Script
# This script tests the GetPublicURL fix and restarts the backend

Write-Host "==============================================" -ForegroundColor Cyan
Write-Host "  Daily Update Photo Fix - Test & Restart" -ForegroundColor Cyan
Write-Host "==============================================" -ForegroundColor Cyan
Write-Host ""

# Navigate to backend directory
$backendDir = Join-Path $PSScriptRoot "backend"
Set-Location $backendDir

Write-Host "📍 Current Directory: $backendDir" -ForegroundColor Yellow
Write-Host ""

# Step 1: Run unit tests
Write-Host "🧪 Step 1: Running unit tests..." -ForegroundColor Green
Write-Host "Testing GetPublicURL function with Windows paths..." -ForegroundColor Gray
go test -v ./utils -run TestGetPublicURL

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ All tests passed!" -ForegroundColor Green
} else {
    Write-Host "❌ Tests failed! Please check the error above." -ForegroundColor Red
    Write-Host "Press any key to exit..."
    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
    exit 1
}

Write-Host ""

# Step 2: Check if uploads directory exists
Write-Host "📁 Step 2: Checking uploads directory..." -ForegroundColor Green
$uploadsDir = Join-Path $backendDir "uploads\daily-updates"
if (-not (Test-Path $uploadsDir)) {
    Write-Host "Creating uploads directory: $uploadsDir" -ForegroundColor Yellow
    New-Item -ItemType Directory -Path $uploadsDir -Force | Out-Null
    Write-Host "✅ Directory created!" -ForegroundColor Green
} else {
    Write-Host "✅ Directory exists: $uploadsDir" -ForegroundColor Green
}

Write-Host ""

# Step 3: Show existing photos (if any)
Write-Host "📷 Step 3: Checking existing photos..." -ForegroundColor Green
$photos = Get-ChildItem -Path $uploadsDir -File -ErrorAction SilentlyContinue
if ($photos) {
    Write-Host "Found $($photos.Count) existing photo(s):" -ForegroundColor Yellow
    $photos | ForEach-Object {
        Write-Host "  - $($_.Name)" -ForegroundColor Gray
    }
} else {
    Write-Host "No existing photos found" -ForegroundColor Gray
}

Write-Host ""

# Step 4: Restart backend
Write-Host "🔄 Step 4: Restarting backend server..." -ForegroundColor Green
Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Backend Server Starting..." -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "📌 Watch for these log messages:" -ForegroundColor Yellow
Write-Host "   📁 Serving static files from: ..." -ForegroundColor Gray
Write-Host "   📷 Received X photo files for upload" -ForegroundColor Gray
Write-Host "   🔗 Photo 1: ... -> ..." -ForegroundColor Gray
Write-Host "   ✅ Saved X photos to disk" -ForegroundColor Gray
Write-Host ""
Write-Host "Press Ctrl+C to stop the server" -ForegroundColor Yellow
Write-Host ""

# Run the backend
go run main.go

