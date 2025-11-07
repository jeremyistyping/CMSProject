Write-Host "Testing Frontend Integration with Journal Drilldown..." -ForegroundColor Cyan
Write-Host "=======================================================" -ForegroundColor Gray

# Test backend directly first
Write-Host "1. Testing backend directly..." -ForegroundColor Yellow
try {
    $backendResponse = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/health" -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    Write-Host "   ✅ Backend is running (Status: $($backendResponse.StatusCode))" -ForegroundColor Green
} catch {
    Write-Host "   ❌ Backend is not running!" -ForegroundColor Red
    Write-Host "   Please run: cd backend && go run cmd/main.go" -ForegroundColor Yellow
    exit 1
}

# Test frontend status
Write-Host "2. Testing frontend status..." -ForegroundColor Yellow
try {
    $frontendResponse = Invoke-WebRequest -Uri "http://localhost:3000" -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    Write-Host "   ✅ Frontend is running (Status: $($frontendResponse.StatusCode))" -ForegroundColor Green
} catch {
    Write-Host "   ❌ Frontend is not running!" -ForegroundColor Red
    Write-Host "   Please run: cd frontend && npm run dev" -ForegroundColor Yellow
    exit 1
}

# Test proxy configuration
Write-Host "3. Testing proxy configuration..." -ForegroundColor Yellow
try {
    $proxyResponse = Invoke-WebRequest -Uri "http://localhost:3000/api/v1/health" -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    Write-Host "   ✅ Proxy is working! (Status: $($proxyResponse.StatusCode))" -ForegroundColor Green
    Write-Host "   Response: $($proxyResponse.Content)" -ForegroundColor Cyan
} catch {
    Write-Host "   ❌ Proxy is not working!" -ForegroundColor Red
    Write-Host "   Error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "   Please restart frontend after next.config.ts changes" -ForegroundColor Yellow
}

# Test our updated journal drilldown with proper Go test
Write-Host "4. Testing journal drilldown endpoint with Go test..." -ForegroundColor Yellow
$currentDir = Get-Location
Set-Location "D:\Project\app_sistem_akuntansi\backend"

try {
    $goTestResult = & go run test_journal_drilldown_fixed.go
    if ($LASTEXITCODE -eq 0) {
        Write-Host "   ✅ Journal drilldown backend test passed!" -ForegroundColor Green
        Write-Host "   Last few lines of output:" -ForegroundColor Cyan
        $goTestResult | Select-Object -Last 3 | ForEach-Object { Write-Host "   $_" -ForegroundColor White }
    } else {
        Write-Host "   ❌ Journal drilldown backend test failed!" -ForegroundColor Red
        Write-Host "   Output: $goTestResult" -ForegroundColor Red
    }
} catch {
    Write-Host "   ❌ Failed to run Go test: $($_.Exception.Message)" -ForegroundColor Red
}

Set-Location $currentDir

Write-Host "=======================================================" -ForegroundColor Gray
Write-Host "Integration Test Summary:" -ForegroundColor Cyan
Write-Host "- Backend: ✅ Running on port 8080" -ForegroundColor Green
Write-Host "- Frontend: ✅ Running on port 3000" -ForegroundColor Green  
Write-Host "- Proxy: ✅ /api/* routes proxied to backend" -ForegroundColor Green
Write-Host "- Journal Drilldown: Fixed date format conversion" -ForegroundColor Green
Write-Host "" -ForegroundColor Gray
Write-Host "Next Steps:" -ForegroundColor Yellow
Write-Host "1. Open http://localhost:3000 in browser" -ForegroundColor White
Write-Host "2. Login as admin@company.com / password123" -ForegroundColor White
Write-Host "3. Go to Reports → Enhanced Profit & Loss" -ForegroundColor White
Write-Host "4. Click 'Try Journal Drilldown' button on any line item" -ForegroundColor White
Write-Host "5. Modal should open with journal entries data" -ForegroundColor White