# CashBank Integration Fixes Validation Script
Write-Host "CashBank Integration - Fixes Validation" -ForegroundColor Blue
Write-Host "=======================================" -ForegroundColor Blue

$ErrorCount = 0

Write-Host "`n1. Testing Go Backend Compilation" -ForegroundColor Cyan
$buildResult = go build -o temp_build.exe . 2>&1
if ($LASTEXITCODE -eq 0) {
    Write-Host "SUCCESS: Backend compiles successfully" -ForegroundColor Green
    if (Test-Path "temp_build.exe") {
        Remove-Item "temp_build.exe" -Force
    }
} else {
    Write-Host "ERROR: Backend compilation failed" -ForegroundColor Red
    Write-Host $buildResult -ForegroundColor Red
    $ErrorCount++
}

Write-Host "`n2. Validating Service Types" -ForegroundColor Cyan
$serviceFile = ".\services\cashbank_integrated_service.go"
if (Test-Path $serviceFile) {
    Write-Host "SUCCESS: CashBank integrated service file exists" -ForegroundColor Green
    $serviceContent = Get-Content $serviceFile -Raw
    
    if ($serviceContent -match "type IntegratedAccountDetail struct") {
        Write-Host "SUCCESS: Found IntegratedAccountDetail type" -ForegroundColor Green
    } else {
        Write-Host "ERROR: Missing IntegratedAccountDetail type" -ForegroundColor Red
        $ErrorCount++
    }
    
    if ($serviceContent -match "func.*GetAccountReconciliation") {
        Write-Host "SUCCESS: Found GetAccountReconciliation method" -ForegroundColor Green
    } else {
        Write-Host "ERROR: Missing GetAccountReconciliation method" -ForegroundColor Red
        $ErrorCount++
    }
} else {
    Write-Host "ERROR: Service file not found" -ForegroundColor Red
    $ErrorCount++
}

Write-Host "`n3. Validating Controller Updates" -ForegroundColor Cyan
$controllerFile = ".\controllers\cashbank_integrated_controller.go"
if (Test-Path $controllerFile) {
    Write-Host "SUCCESS: Controller file exists" -ForegroundColor Green
    $controllerContent = Get-Content $controllerFile -Raw
    
    if ($controllerContent -match "GetAccountReconciliation.*integratedService\.GetAccountReconciliation") {
        Write-Host "SUCCESS: Controller uses new reconciliation method" -ForegroundColor Green
    } else {
        Write-Host "ERROR: Controller not updated properly" -ForegroundColor Red
        $ErrorCount++
    }
} else {
    Write-Host "ERROR: Controller file not found" -ForegroundColor Red
    $ErrorCount++
}

Write-Host "`n4. Validating Frontend Types" -ForegroundColor Cyan
$typesFile = ".\frontend_components\types\cashBankIntegration.types.ts"
if (Test-Path $typesFile) {
    Write-Host "SUCCESS: Frontend types file exists" -ForegroundColor Green
    $typesContent = Get-Content $typesFile -Raw
    
    if ($typesContent -match "interface IntegratedAccountDetail") {
        Write-Host "SUCCESS: Found IntegratedAccountDetail interface" -ForegroundColor Green
    } else {
        Write-Host "ERROR: Missing IntegratedAccountDetail interface" -ForegroundColor Red
        $ErrorCount++
    }
} else {
    Write-Host "ERROR: Frontend types file not found" -ForegroundColor Red
    $ErrorCount++
}

Write-Host "`nSUMMARY" -ForegroundColor Magenta
Write-Host "=======" -ForegroundColor Magenta

if ($ErrorCount -eq 0) {
    Write-Host "ALL CHECKS PASSED! Integration fixes are ready." -ForegroundColor Green
    Write-Host "Ready for Phase 1.4 - End-to-end Testing" -ForegroundColor Green
} else {
    Write-Host "FOUND $ErrorCount ISSUES that need to be fixed." -ForegroundColor Red
}

Write-Host "`nNext Steps:" -ForegroundColor Blue
Write-Host "1. Run backend server: go run ." -ForegroundColor White
Write-Host "2. Test integration endpoints" -ForegroundColor White
Write-Host "3. Test frontend components" -ForegroundColor White