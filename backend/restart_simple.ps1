Write-Host "Restarting Backend and Frontend Services..." -ForegroundColor Cyan
Write-Host "==============================================" -ForegroundColor Gray

# Stop all existing processes
Write-Host "Stopping existing processes..." -ForegroundColor Yellow
Get-Process -Name "go" -ErrorAction SilentlyContinue | Stop-Process -Force
Get-Process -Name "node" -ErrorAction SilentlyContinue | Stop-Process -Force

# Wait a moment for processes to fully terminate
Start-Sleep -Seconds 2

# Start Backend
Write-Host "Starting Backend (port 8080)..." -ForegroundColor Green
$backendPath = "D:\Project\app_sistem_akuntansi\backend"
Set-Location $backendPath
$backendProcess = Start-Process -FilePath "go" -ArgumentList "run", "cmd/main.go" -WindowStyle Hidden -PassThru
Write-Host "Backend started with PID: $($backendProcess.Id)" -ForegroundColor Green

# Wait for backend to initialize
Write-Host "Waiting for backend to initialize..." -ForegroundColor Yellow
Start-Sleep -Seconds 5

# Test backend endpoint
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/health" -UseBasicParsing -ErrorAction Stop
    Write-Host "Backend health check passed (Status: $($response.StatusCode))" -ForegroundColor Green
} catch {
    Write-Host "Backend health check failed, but continuing..." -ForegroundColor Yellow
}

Write-Host "==============================================" -ForegroundColor Gray
Write-Host "Services restarted successfully!" -ForegroundColor Green
Write-Host "Frontend: http://localhost:3000" -ForegroundColor Cyan
Write-Host "Backend: http://localhost:8080" -ForegroundColor Cyan
Write-Host "Journal Drilldown: /api/v1/journal-drilldown (proxied to backend)" -ForegroundColor Cyan

Write-Host "Current directory: $(Get-Location)" -ForegroundColor Gray