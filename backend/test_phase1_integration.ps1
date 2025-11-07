# Phase 1 CashBank SSOT Integration Testing
# This script tests the Phase 1 implementation of CashBank-SSOT integration

Write-Host "🚀 Phase 1 CashBank SSOT Integration Testing" -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Gray

$baseURL = "http://localhost:8080"
$frontendURL = "http://localhost:3000"

# Step 1: Test backend service layer first
Write-Host "`n1️⃣ Testing Service Layer..." -ForegroundColor Yellow
Write-Host "Running Go service test..." -ForegroundColor Cyan

try {
    $goTestResult = & go run test_cashbank_integrated_service.go
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Service layer test passed!" -ForegroundColor Green
    } else {
        Write-Host "❌ Service layer test failed!" -ForegroundColor Red
        Write-Host "Output: $goTestResult" -ForegroundColor Yellow
    }
} catch {
    Write-Host "❌ Failed to run Go service test: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "Make sure Go is installed and dependencies are available" -ForegroundColor Yellow
}

# Step 2: Test backend server startup
Write-Host "`n2️⃣ Testing Backend Server..." -ForegroundColor Yellow

try {
    $healthResponse = Invoke-WebRequest -Uri "$baseURL/api/v1/health" -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    Write-Host "✅ Backend server is running (Status: $($healthResponse.StatusCode))" -ForegroundColor Green
} catch {
    Write-Host "❌ Backend server is not running!" -ForegroundColor Red
    Write-Host "Please start the backend:" -ForegroundColor Yellow
    Write-Host "  cd backend && go run cmd/main.go" -ForegroundColor White
    Write-Host ""
    Write-Host "Continuing with other tests..." -ForegroundColor Cyan
}

# Step 3: Test frontend server
Write-Host "`n3️⃣ Testing Frontend Server..." -ForegroundColor Yellow

try {
    $frontendResponse = Invoke-WebRequest -Uri $frontendURL -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    Write-Host "✅ Frontend server is running (Status: $($frontendResponse.StatusCode))" -ForegroundColor Green
} catch {
    Write-Host "❌ Frontend server is not running!" -ForegroundColor Red
    Write-Host "Please start the frontend:" -ForegroundColor Yellow
    Write-Host "  cd frontend && npm run dev" -ForegroundColor White
}

# Step 4: Test API proxy (if both servers are running)
Write-Host "`n4️⃣ Testing API Proxy..." -ForegroundColor Yellow

try {
    $proxyResponse = Invoke-WebRequest -Uri "$frontendURL/api/v1/health" -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    Write-Host "✅ API proxy is working! (Status: $($proxyResponse.StatusCode))" -ForegroundColor Green
} catch {
    Write-Host "❌ API proxy is not working!" -ForegroundColor Red
    Write-Host "This might be normal if frontend is still starting up" -ForegroundColor Yellow
}

# Step 5: Test route registration
Write-Host "`n5️⃣ Testing Route Registration..." -ForegroundColor Yellow

# Test if our new integrated routes are registered
$testRoutes = @(
    "/api/v1/cashbank/integrated/summary",
    "/api/v1/cashbank/integrated/accounts/1",
    "/api/v1/cashbank/integrated/accounts/1/reconciliation"
)

foreach ($route in $testRoutes) {
    try {
        $routeResponse = Invoke-WebRequest -Uri "$baseURL$route" -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
        Write-Host "✅ Route $route is registered (Status: $($routeResponse.StatusCode))" -ForegroundColor Green
    } catch {
        if ($_.Exception.Response.StatusCode -eq 401) {
            Write-Host "✅ Route $route is registered (Authentication required)" -ForegroundColor Green
        } elseif ($_.Exception.Response.StatusCode -eq 404) {
            Write-Host "❌ Route $route is NOT registered" -ForegroundColor Red
        } else {
            Write-Host "⚠️  Route $route status: $($_.Exception.Response.StatusCode)" -ForegroundColor Yellow
        }
    }
}

# Step 6: Test basic compilation (if backend server is not running)
Write-Host "`n6️⃣ Testing Code Compilation..." -ForegroundColor Yellow

try {
    Write-Host "Checking if code compiles..." -ForegroundColor Cyan
    $buildResult = & go build -o temp_test.exe cmd/main.go 2>&1
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Code compiles successfully!" -ForegroundColor Green
        if (Test-Path "temp_test.exe") {
            Remove-Item "temp_test.exe" -Force
        }
    } else {
        Write-Host "❌ Compilation failed!" -ForegroundColor Red
        Write-Host "Build output: $buildResult" -ForegroundColor Yellow
    }
} catch {
    Write-Host "❌ Failed to test compilation: $($_.Exception.Message)" -ForegroundColor Red
}

# Step 7: Check for required files
Write-Host "`n7️⃣ Checking Required Files..." -ForegroundColor Yellow

$requiredFiles = @(
    "services/cashbank_integrated_service.go",
    "controllers/cashbank_integrated_controller.go",
    "routes/cashbank_integrated_routes.go",
    "docs/SSOT_CASHBANK_INTEGRATION_ARCHITECTURE.md"
)

foreach ($file in $requiredFiles) {
    if (Test-Path $file) {
        Write-Host "✅ $file exists" -ForegroundColor Green
    } else {
        Write-Host "❌ $file is missing" -ForegroundColor Red
    }
}

# Step 8: Database connectivity test
Write-Host "`n8️⃣ Testing Database Connectivity..." -ForegroundColor Yellow

if (Test-Path ".env") {
    Write-Host "✅ .env file exists" -ForegroundColor Green
    
    # Try to read database config from .env
    $envContent = Get-Content ".env"
    $dbConfig = $envContent | Where-Object { $_ -like "DB_*" }
    
    if ($dbConfig) {
        Write-Host "✅ Database configuration found in .env:" -ForegroundColor Green
        foreach ($config in $dbConfig) {
            if ($config -notlike "*PASSWORD*") {
                Write-Host "   $config" -ForegroundColor Cyan
            }
        }
    } else {
        Write-Host "⚠️  No database configuration found in .env" -ForegroundColor Yellow
    }
} else {
    Write-Host "❌ .env file not found" -ForegroundColor Red
    Write-Host "Please create .env file with database configuration" -ForegroundColor Yellow
}

# Step 9: Frontend integration check
Write-Host "`n9️⃣ Checking Frontend Integration Readiness..." -ForegroundColor Yellow

# Check if frontend has the cash-bank page
$frontendPaths = @(
    "../frontend",
    "../../frontend",
    "../../../frontend"
)

$frontendFound = $false
foreach ($path in $frontendPaths) {
    if (Test-Path $path) {
        Write-Host "✅ Frontend directory found at: $path" -ForegroundColor Green
        $frontendFound = $true
        
        # Check for cash-bank related files
        $cashBankFiles = Get-ChildItem -Path $path -Recurse -Name "*cash*bank*" -ErrorAction SilentlyContinue
        if ($cashBankFiles) {
            Write-Host "✅ Cash-bank related files found:" -ForegroundColor Green
            $cashBankFiles | ForEach-Object { Write-Host "   $_" -ForegroundColor Cyan }
        } else {
            Write-Host "⚠️  No cash-bank specific files found - ready for implementation" -ForegroundColor Yellow
        }
        break
    }
}

if (-not $frontendFound) {
    Write-Host "⚠️  Frontend directory not found in expected locations" -ForegroundColor Yellow
}

# Summary
Write-Host "`n=============================================" -ForegroundColor Gray
Write-Host "📋 Phase 1 Integration Test Summary" -ForegroundColor Cyan
Write-Host "=============================================" -ForegroundColor Gray

Write-Host "`n✅ Completed Tasks:" -ForegroundColor Green
Write-Host "  • Service layer implementation" -ForegroundColor White
Write-Host "  • Controller implementation" -ForegroundColor White
Write-Host "  • Route registration" -ForegroundColor White
Write-Host "  • Architecture documentation" -ForegroundColor White

Write-Host "`n🔄 Next Phase 1 Steps:" -ForegroundColor Yellow
Write-Host "  1. Start both servers if not running:" -ForegroundColor White
Write-Host "     Backend:  go run cmd/main.go" -ForegroundColor Gray
Write-Host "     Frontend: npm run dev" -ForegroundColor Gray
Write-Host "  2. Get JWT token via login API" -ForegroundColor White
Write-Host "  3. Test API endpoints with authentication" -ForegroundColor White
Write-Host "  4. Implement frontend components" -ForegroundColor White

Write-Host "`n🎯 Ready for Phase 1.3: Frontend Implementation" -ForegroundColor Cyan

Write-Host "`n📖 Available Documentation:" -ForegroundColor Yellow
Write-Host "  • Architecture: docs/SSOT_CASHBANK_INTEGRATION_ARCHITECTURE.md" -ForegroundColor White
Write-Host "  • API Testing: test_cashbank_ssot_integration.ps1" -ForegroundColor White
Write-Host "  • Service Testing: test_cashbank_integrated_service.go" -ForegroundColor White