# Start Backend Server for Testing
# This script starts the backend server with test configuration

param(
    [string]$Port = "8080",
    [string]$Environment = "development"
)

Write-Host "🚀 Starting Backend Server for Testing..." -ForegroundColor Green
Write-Host "================================" -ForegroundColor Green
Write-Host "Port: $Port" -ForegroundColor Cyan
Write-Host "Environment: $Environment" -ForegroundColor Cyan
Write-Host ""

# Set environment variables
$env:PORT = $Port
$env:GO_ENV = $Environment
$env:ENABLE_DEBUG_ROUTES = "true"
$env:GIN_MODE = "debug"

# Check if main.exe exists, if not build it
if (-not (Test-Path "cmd/main.exe")) {
    Write-Host "🔨 Building server..." -ForegroundColor Yellow
    go build -o cmd/main.exe ./cmd/main.go
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "❌ Build failed!" -ForegroundColor Red
        exit 1
    }
    Write-Host "✅ Build successful!" -ForegroundColor Green
    Write-Host ""
}

Write-Host "📊 WebSocket endpoint will be available at:" -ForegroundColor Cyan
Write-Host "   ws://localhost:$Port/api/v1/journals/account-balances/ws" -ForegroundColor Cyan
Write-Host ""
Write-Host "🔗 Test WebSocket with:" -ForegroundColor Yellow
Write-Host "   node test_websocket_client.js ws://localhost:$Port" -ForegroundColor Yellow
Write-Host ""
Write-Host "🧪 Run integration test with:" -ForegroundColor Yellow
Write-Host "   .\test_ssot_high_priority.ps1 -BaseUrl 'http://localhost:$Port'" -ForegroundColor Yellow
Write-Host ""

Write-Host "Starting server... (Press Ctrl+C to stop)" -ForegroundColor Green
Write-Host ""

# Start the server
try {
    ./cmd/main.exe
} catch {
    Write-Host "❌ Server failed to start: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}