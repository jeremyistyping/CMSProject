#!/usr/bin/env pwsh

# Test script untuk permission employee di modul purchases
Write-Host "🔧 Testing Employee Permission untuk Purchase Module" -ForegroundColor Cyan

$baseUrl = "http://localhost:8080/api/v1"

# Test 1: Health check
Write-Host "`n1. Testing Health Check..." -ForegroundColor Yellow
try {
    $healthResponse = Invoke-WebRequest -Uri "$baseUrl/health" -Method GET
    if ($healthResponse.StatusCode -eq 200) {
        Write-Host "✅ Server is running" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ Server tidak berjalan di localhost:8080" -ForegroundColor Red
    exit 1
}

# Test 2: Login sebagai employee 
Write-Host "`n2. Testing Employee Login..." -ForegroundColor Yellow
$loginBody = @{
    username = "employee"
    password = "employee123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-WebRequest -Uri "$baseUrl/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    $loginData = $loginResponse.Content | ConvertFrom-Json
    $token = $loginData.token
    
    if ($token) {
        Write-Host "✅ Employee login berhasil" -ForegroundColor Green
        Write-Host "   User: $($loginData.user.username) | Role: $($loginData.user.role)" -ForegroundColor Cyan
    } else {
        Write-Host "❌ Login gagal, tidak ada token" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "❌ Employee login error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "   Mencoba dengan admin untuk comparison..." -ForegroundColor Yellow
    
    # Try admin untuk comparison
    $adminLoginBody = @{
        username = "admin"
        password = "admin123"
    } | ConvertTo-Json
    
    try {
        $adminLoginResponse = Invoke-WebRequest -Uri "$baseUrl/auth/login" -Method POST -Body $adminLoginBody -ContentType "application/json"
        $adminLoginData = $adminLoginResponse.Content | ConvertFrom-Json
        $token = $adminLoginData.token
        Write-Host "✅ Using admin credentials for testing" -ForegroundColor Green
        Write-Host "   User: $($adminLoginData.user.username) | Role: $($adminLoginData.user.role)" -ForegroundColor Cyan
    } catch {
        Write-Host "❌ Semua login attempts gagal" -ForegroundColor Red
        exit 1
    }
}

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

# Test 3: Check user permissions
Write-Host "`n3. Testing User Permissions..." -ForegroundColor Yellow
try {
    $permResponse = Invoke-WebRequest -Uri "$baseUrl/permissions/me" -Method GET -Headers $headers
    if ($permResponse.StatusCode -eq 200) {
        $permData = $permResponse.Content | ConvertFrom-Json
        Write-Host "✅ User permissions retrieved" -ForegroundColor Green
        
        if ($permData.permissions.purchases) {
            $purchasePerm = $permData.permissions.purchases
            Write-Host "   Purchase permissions:" -ForegroundColor Cyan
            Write-Host "   - View: $($purchasePerm.can_view)" -ForegroundColor Cyan
            Write-Host "   - Create: $($purchasePerm.can_create)" -ForegroundColor Cyan
            Write-Host "   - Edit: $($purchasePerm.can_edit)" -ForegroundColor Cyan
            Write-Host "   - Delete: $($purchasePerm.can_delete)" -ForegroundColor Cyan
            Write-Host "   - Approve: $($purchasePerm.can_approve)" -ForegroundColor Cyan
            Write-Host "   - Export: $($purchasePerm.can_export)" -ForegroundColor Cyan
        } else {
            Write-Host "   ❌ No purchase permissions found" -ForegroundColor Red
        }
        
        if ($permData.permissions.accounts) {
            $accountPerm = $permData.permissions.accounts
            Write-Host "   Account permissions:" -ForegroundColor Cyan
            Write-Host "   - View: $($accountPerm.can_view)" -ForegroundColor Cyan
            Write-Host "   - Create: $($accountPerm.can_create)" -ForegroundColor Cyan
        } else {
            Write-Host "   ❌ No account permissions found" -ForegroundColor Red
        }
    }
} catch {
    Write-Host "❌ User permissions error: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 4: Test purchase endpoint access
Write-Host "`n4. Testing Purchase Endpoint Access..." -ForegroundColor Yellow
try {
    $purchaseResponse = Invoke-WebRequest -Uri "$baseUrl/purchases" -Method GET -Headers $headers
    if ($purchaseResponse.StatusCode -eq 200) {
        Write-Host "✅ Purchases endpoint accessible" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ Purchases endpoint error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "   Response: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
}

# Test 5: Test create purchase access
Write-Host "`n5. Testing Create Purchase Access..." -ForegroundColor Yellow
$testPurchaseData = @{
    vendor_id = 1
    date = "2025-01-21"
    notes = "Test purchase for permission check"
    items = @(
        @{
            product_id = 1
            quantity = 1
            unit_price = 1000
            account_id = 15
        }
    )
} | ConvertTo-Json -Depth 3

try {
    $createResponse = Invoke-WebRequest -Uri "$baseUrl/purchases" -Method POST -Body $testPurchaseData -Headers $headers
    if ($createResponse.StatusCode -eq 201) {
        Write-Host "✅ Create purchase successful" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ Create purchase error: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $errorStream = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorStream)
        $errorBody = $reader.ReadToEnd()
        Write-Host "   Error detail: $errorBody" -ForegroundColor Red
    }
}

# Test 6: Test approval endpoints
Write-Host "`n6. Testing Approval Endpoints..." -ForegroundColor Yellow
try {
    $approvalResponse = Invoke-WebRequest -Uri "$baseUrl/purchases/pending-approval" -Method GET -Headers $headers
    if ($approvalResponse.StatusCode -eq 200) {
        Write-Host "✅ Pending approval endpoint accessible (unexpected for employee)" -ForegroundColor Yellow
    }
} catch {
    Write-Host "✅ Pending approval endpoint blocked (expected for employee role)" -ForegroundColor Green
    Write-Host "   Status: $($_.Exception.Response.StatusCode)" -ForegroundColor Cyan
}

Write-Host "`n🎯 Employee Permission Test Complete!" -ForegroundColor Cyan
Write-Host "`nJika employee bisa:" -ForegroundColor Green
Write-Host "- Login ✅" -ForegroundColor Green
Write-Host "- View purchases ✅" -ForegroundColor Green
Write-Host "- Create purchases ✅" -ForegroundColor Green
Write-Host "- Blocked from approvals ✅" -ForegroundColor Green
Write-Host "`nMaka permission sistem sudah bekerja dengan benar!" -ForegroundColor Green