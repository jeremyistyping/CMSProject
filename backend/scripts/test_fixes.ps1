# CashBank Integration Fixes Validation Script
# Tests backend compilation, service layer, and frontend compatibility

Write-Host "🔧 CashBank Integration - Fixes Validation" -ForegroundColor Blue
Write-Host "=========================================" -ForegroundColor Blue

$ErrorCount = 0

# Function to log errors
function Log-Error($message) {
    Write-Host "❌ ERROR: $message" -ForegroundColor Red
    $global:ErrorCount++
}

# Function to log success
function Log-Success($message) {
    Write-Host "✅ SUCCESS: $message" -ForegroundColor Green
}

# Function to log info
function Log-Info($message) {
    Write-Host "ℹ️  INFO: $message" -ForegroundColor Yellow
}

# 1. Test Go compilation
Write-Host "`n🔨 1. Testing Go Backend Compilation" -ForegroundColor Cyan
try {
    $env:CGO_ENABLED = "0"
    $buildResult = go build -o temp_build.exe . 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Log-Success "Backend compiles successfully"
        if (Test-Path "temp_build.exe") {
            Remove-Item "temp_build.exe" -Force
        }
    } else {
        Log-Error "Backend compilation failed: $buildResult"
    }
} catch {
    Log-Error "Failed to compile backend: $($_.Exception.Message)"
}

# 2. Test service types and interfaces
Write-Host "`n🔍 2. Validating Service Types" -ForegroundColor Cyan

$serviceFile = ".\services\cashbank_integrated_service.go"
if (Test-Path $serviceFile) {
    Log-Success "CashBank integrated service file exists"
    
    # Check for required types
    $serviceContent = Get-Content $serviceFile -Raw
    
    $requiredTypes = @(
        "IntegratedAccountDetail",
        "ReconciliationData", 
        "JournalEntriesResponse",
        "TransactionHistoryResponse",
        "PaginationInfo"
    )
    
    foreach ($type in $requiredTypes) {
        if ($serviceContent -match "type $type struct") {
            Log-Success "Found required type: $type"
        } else {
            Log-Error "Missing required type: $type"
        }
    }
    
    # Check for required methods
    $requiredMethods = @(
        "GetAccountReconciliation",
        "GetJournalEntriesPaginated", 
        "GetTransactionHistoryPaginated"
    )
    
    foreach ($method in $requiredMethods) {
        if ($serviceContent -match "func.*$method") {
            Log-Success "Found required method: $method"
        } else {
            Log-Error "Missing required method: $method"
        }
    }
    
} else {
    Log-Error "CashBank integrated service file not found"
}

# 3. Test controller fixes
Write-Host "`n🎮 3. Validating Controller Updates" -ForegroundColor Cyan

$controllerFile = ".\controllers\cashbank_integrated_controller.go"
if (Test-Path $controllerFile) {
    Log-Success "CashBank integrated controller file exists"
    
    $controllerContent = Get-Content $controllerFile -Raw
    
    # Check for updated controller methods
    if ($controllerContent -match "GetAccountReconciliation.*integratedService\.GetAccountReconciliation") {
        Log-Success "Controller uses new reconciliation service method"
    } else {
        Log-Error "Controller not updated to use new reconciliation service method"
    }
    
    if ($controllerContent -match "GetJournalEntriesPaginated") {
        Log-Success "Controller uses paginated journal entries method"
    } else {
        Log-Error "Controller not updated to use paginated journal entries"
    }
    
    if ($controllerContent -match "GetTransactionHistoryPaginated") {
        Log-Success "Controller uses paginated transaction history method"
    } else {
        Log-Error "Controller not updated to use paginated transaction history"
    }
    
} else {
    Log-Error "CashBank integrated controller file not found"
}

# 4. Test route registration
Write-Host "`n🛣️  4. Validating Route Registration" -ForegroundColor Cyan

$routesFile = ".\routes\routes.go"
if (Test-Path $routesFile) {
    $routesContent = Get-Content $routesFile -Raw
    
    if ($routesContent -match "SetupCashBankIntegratedRoutes") {
        Log-Success "CashBank integrated routes are registered"
    } else {
        Log-Error "CashBank integrated routes not found in main routes"
    }
} else {
    Log-Error "Main routes file not found"
}

$integratedRoutesFile = ".\routes\cashbank_integrated_routes.go"
if (Test-Path $integratedRoutesFile) {
    Log-Success "CashBank integrated routes file exists"
    
    $integratedRoutesContent = Get-Content $integratedRoutesFile -Raw
    
    $expectedRoutes = @(
        "/accounts/:id.*GetIntegratedAccountDetails",
        "/summary.*GetIntegratedSummary", 
        "/accounts/:id/reconciliation.*GetAccountReconciliation",
        "/accounts/:id/journal-entries.*GetAccountJournalEntries",
        "/accounts/:id/transactions.*GetAccountTransactionHistory"
    )
    
    foreach ($route in $expectedRoutes) {
        if ($integratedRoutesContent -match $route) {
            Log-Success "Found expected route pattern: $route"
        } else {
            Log-Error "Missing expected route pattern: $route"
        }
    }
} else {
    Log-Error "Integrated routes file not found"
}

# 5. Test frontend types compatibility
Write-Host "`n🎨 5. Validating Frontend Types" -ForegroundColor Cyan

$typesFile = ".\frontend_components\types\cashBankIntegration.types.ts"
if (Test-Path $typesFile) {
    Log-Success "Frontend types file exists"
    
    $typesContent = Get-Content $typesFile -Raw
    
    # Check for updated interfaces
    $requiredInterfaces = @(
        "IntegratedAccountDetail",
        "ReconciliationData",
        "JournalEntriesResponse", 
        "TransactionHistoryResponse",
        "JournalEntryDetail",
        "TransactionEntry"
    )
    
    foreach ($interface in $requiredInterfaces) {
        if ($typesContent -match "interface $interface") {
            Log-Success "Found required frontend interface: $interface"
        } else {
            Log-Error "Missing required frontend interface: $interface"
        }
    }
    
    # Check for correct field mappings
    if ($typesContent -match "recent_transactions.*IntegratedTransaction") {
        Log-Success "Frontend response structure matches backend"
    } else {
        Log-Error "Frontend response structure mismatch with backend"
    }
    
} else {
    Log-Error "Frontend types file not found"
}

# 6. Test frontend service compatibility
Write-Host "`n🔗 6. Validating Frontend Service" -ForegroundColor Cyan

$serviceFile = ".\frontend_components\services\cashBankIntegrationService.ts"
if (Test-Path $serviceFile) {
    Log-Success "Frontend service file exists"
    
    $serviceContent = Get-Content $serviceFile -Raw
    
    # Check for required methods
    $requiredServiceMethods = @(
        "getIntegratedSummary",
        "getIntegratedAccountDetails",
        "getAccountReconciliation", 
        "getAccountJournalEntries",
        "getAccountTransactionHistory"
    )
    
    foreach ($method in $requiredServiceMethods) {
        if ($serviceContent -match "async $method") {
            Log-Success "Found required frontend service method: $method"
        } else {
            Log-Error "Missing required frontend service method: $method"
        }
    }
    
    # Check for method aliases
    if ($serviceContent -match "getIntegratedAccount.*getIntegratedAccountDetails") {
        Log-Success "Frontend service has proper method aliases"
    } else {
        Log-Error "Frontend service missing method aliases"
    }
    
} else {
    Log-Error "Frontend service file not found"
}

# 7. Test database models compatibility
Write-Host "`n🗃️  7. Validating Database Models" -ForegroundColor Cyan

# Check if models are compatible
$modelsDir = ".\models"
if (Test-Path $modelsDir) {
    Log-Success "Models directory exists"
    
    $cashbankModel = "$modelsDir\cashbank.go"
    $journalModel = "$modelsDir\ssot_journal.go"
    
    if (Test-Path $cashbankModel) {
        Log-Success "CashBank model exists"
    } else {
        Log-Error "CashBank model not found"
    }
    
    if (Test-Path $journalModel) {
        Log-Success "SSOT Journal model exists"
    } else {
        Log-Error "SSOT Journal model not found"
    }
} else {
    Log-Error "Models directory not found"
}

# 8. Test for potential runtime issues
Write-Host "`n⚠️  8. Checking for Potential Issues" -ForegroundColor Cyan

# Check for Go module issues
if (Test-Path "go.mod") {
    try {
        $modTidyResult = go mod tidy 2>&1
        if ($LASTEXITCODE -eq 0) {
            Log-Success "Go modules are clean"
        } else {
            Log-Error "Go mod tidy failed: $modTidyResult"
        }
    } catch {
        Log-Error "Failed to run go mod tidy"
    }
} else {
    Log-Error "go.mod file not found"
}

# Check for import issues
try {
    $fmtResult = go fmt ./... 2>&1
    if ($LASTEXITCODE -eq 0) {
        Log-Success "Go formatting is correct"
    } else {
        Log-Error "Go format issues found: $fmtResult"
    }
} catch {
    Log-Error "Failed to run go fmt"
}

# 9. Summary
Write-Host "`n📊 VALIDATION SUMMARY" -ForegroundColor Magenta
Write-Host "=====================" -ForegroundColor Magenta

if ($ErrorCount -eq 0) {
    Write-Host "🎉 ALL CHECKS PASSED! Integration fixes are ready for testing." -ForegroundColor Green
    Write-Host "`n✨ Ready for Phase 1.4 - End-to-end Testing" -ForegroundColor Green
} else {
    Write-Host "⚠️  FOUND $ErrorCount ISSUES that need to be fixed before proceeding." -ForegroundColor Red
    Write-Host "`n🔧 Please fix the reported issues before testing." -ForegroundColor Yellow
}

Write-Host "`n🚀 Next Steps:" -ForegroundColor Blue
Write-Host "1. If all checks passed: Run the backend server" -ForegroundColor White
Write-Host "2. Test the integration endpoints manually or via Postman" -ForegroundColor White  
Write-Host "3. Test frontend components with real data" -ForegroundColor White
Write-Host "4. Proceed to Phase 2 enhancements" -ForegroundColor White

Write-Host "`n📝 Test Commands:" -ForegroundColor Blue
Write-Host "   Backend: go run ." -ForegroundColor White
Write-Host "   Test API: curl -H 'Authorization: Bearer <token>' http://localhost:8080/api/v1/cashbank/integrated/summary" -ForegroundColor White

Exit $ErrorCount