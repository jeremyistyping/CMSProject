Write-Host "Testing Frontend Proxy Configuration..." -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Gray

# Test if frontend is running
Write-Host "1. Checking frontend status..." -ForegroundColor Yellow
try {
    $frontendResponse = Invoke-WebRequest -Uri "http://localhost:3000" -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    Write-Host "   Frontend is running (Status: $($frontendResponse.StatusCode))" -ForegroundColor Green
} catch {
    Write-Host "   Frontend is not running or not ready yet" -ForegroundColor Red
    Write-Host "   Please start frontend manually: cd frontend && npm run dev" -ForegroundColor Yellow
    exit 1
}

# Test backend health directly
Write-Host "2. Checking backend status..." -ForegroundColor Yellow
try {
    $backendResponse = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/health" -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    Write-Host "   Backend is running (Status: $($backendResponse.StatusCode))" -ForegroundColor Green
} catch {
    Write-Host "   Backend is not running!" -ForegroundColor Red
    exit 1
}

# Test proxy by accessing API through frontend
Write-Host "3. Testing proxy configuration..." -ForegroundColor Yellow
try {
    # This should be proxied to backend
    $proxyResponse = Invoke-WebRequest -Uri "http://localhost:3000/api/v1/health" -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    Write-Host "   Proxy is working! (Status: $($proxyResponse.StatusCode))" -ForegroundColor Green
    Write-Host "   Response: $($proxyResponse.Content)" -ForegroundColor Cyan
} catch {
    Write-Host "   Proxy is not working yet. Response: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "   This might be normal if frontend is still starting up." -ForegroundColor Yellow
}

Write-Host "=========================================" -ForegroundColor Gray
Write-Host "Test completed!" -ForegroundColor Green
Write-Host "If proxy is working, journal drilldown should work in the frontend." -ForegroundColor Cyan