# Stop existing backend process
Write-Host "Stopping existing backend..." -ForegroundColor Yellow
Stop-Process -Name "backend" -Force -ErrorAction SilentlyContinue
Stop-Process -Name "main" -Force -ErrorAction SilentlyContinue
Start-Sleep -Seconds 2

# Clean old executable if exists
if (Test-Path ".\backend.exe") {
    Remove-Item ".\backend.exe" -Force -ErrorAction SilentlyContinue
}

Write-Host "`nBuilding backend..." -ForegroundColor Cyan
go build -o backend.exe cmd/main.go

if ($LASTEXITCODE -eq 0) {
    Write-Host "`nStarting backend..." -ForegroundColor Green
    Write-Host "Watch for migration logs below:" -ForegroundColor Yellow
    Write-Host "======================================`n" -ForegroundColor Gray
    
    # Run backend and show output
    .\backend.exe
} else {
    Write-Host "`nBuild failed!" -ForegroundColor Red
    exit 1
}
