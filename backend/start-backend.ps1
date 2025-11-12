# Start backend with logging
Write-Host "🔨 Building backend..." -ForegroundColor Cyan
go build -o main.exe .

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Build failed!" -ForegroundColor Red
    exit 1
}

Write-Host "✅ Build successful!" -ForegroundColor Green
Write-Host "🚀 Starting backend..." -ForegroundColor Cyan
Write-Host ""
Write-Host "📝 Watch for these logs:" -ForegroundColor Yellow
Write-Host "   - 📁 Serving static files from: ..." -ForegroundColor Gray
Write-Host "   - 📝 Serving file: ... (when accessing photos)" -ForegroundColor Gray
Write-Host ""

.\main.exe

