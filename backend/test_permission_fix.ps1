#!/usr/bin/env pwsh

# Test script untuk verifikasi permission fix pada accounts untuk purchases
Write-Host "🔧 Testing Permission Fix untuk Purchase Accounts" -ForegroundColor Cyan

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

# Test 2: Login untuk mendapatkan token
Write-Host "`n2. Testing Login..." -ForegroundColor Yellow
$loginBody = @{
    username = "manager"
    password = "manager123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-WebRequest -Uri "$baseUrl/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    $loginData = $loginResponse.Content | ConvertFrom-Json
    $token = $loginData.token
    
    if ($token) {
        Write-Host "✅ Login berhasil, token diperoleh" -ForegroundColor Green
    } else {
        Write-Host "❌ Login gagal, tidak ada token" -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "❌ Login error: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "   Mencoba dengan kredensial alternatif..." -ForegroundColor Yellow
    
    # Try alternative credentials
    $altLoginBody = @{
        username = "admin"
        password = "admin123"
    } | ConvertTo-Json
    
    try {
        $loginResponse = Invoke-WebRequest -Uri "$baseUrl/auth/login" -Method POST -Body $altLoginBody -ContentType "application/json"
        $loginData = $loginResponse.Content | ConvertFrom-Json
        $token = $loginData.token
        Write-Host "✅ Login berhasil dengan admin credentials" -ForegroundColor Green
    } catch {
        Write-Host "❌ Semua login attempts gagal" -ForegroundColor Red
        exit 1
    }
}

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

# Test 3: Akses accounts catalog tanpa type (harus berhasil sekarang)
Write-Host "`n3. Testing Accounts Catalog Access (No Type)..." -ForegroundColor Yellow
try {
    $catalogResponse = Invoke-WebRequest -Uri "$baseUrl/accounts/catalog" -Method GET -Headers $headers
    if ($catalogResponse.StatusCode -eq 200) {
        $catalogData = $catalogResponse.Content | ConvertFrom-Json
        Write-Host "✅ Accounts catalog accessible - Count: $($catalogData.count)" -ForegroundColor Green
        Write-Host "   Sample accounts: $($catalogData.data[0..2] | ConvertTo-Json -Compress)" -ForegroundColor Cyan
    }
} catch {
    Write-Host "❌ Accounts catalog tidak bisa diakses: $($_.Exception.Message)" -ForegroundColor Red
    
    # Show error detail
    if ($_.Exception.Response) {
        $errorStream = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorStream)
        $errorBody = $reader.ReadToEnd()
        Write-Host "   Error detail: $errorBody" -ForegroundColor Red
    }
}

# Test 4: Akses accounts catalog dengan type=EXPENSE
Write-Host "`n4. Testing Accounts Catalog with EXPENSE Type..." -ForegroundColor Yellow
try {
    $expenseResponse = Invoke-WebRequest -Uri "$baseUrl/accounts/catalog?type=EXPENSE" -Method GET -Headers $headers
    if ($expenseResponse.StatusCode -eq 200) {
        $expenseData = $expenseResponse.Content | ConvertFrom-Json
        Write-Host "✅ EXPENSE accounts accessible - Count: $($expenseData.count)" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ EXPENSE accounts tidak bisa diakses: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 5: Akses accounts catalog dengan type=LIABILITY (untuk credit)
Write-Host "`n5. Testing Accounts Catalog with LIABILITY Type..." -ForegroundColor Yellow
try {
    $liabilityResponse = Invoke-WebRequest -Uri "$baseUrl/accounts/catalog?type=LIABILITY" -Method GET -Headers $headers
    if ($liabilityResponse.StatusCode -eq 200) {
        $liabilityData = $liabilityResponse.Content | ConvertFrom-Json
        Write-Host "✅ LIABILITY accounts accessible - Count: $($liabilityData.count)" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ LIABILITY accounts tidak bisa diakses: $($_.Exception.Message)" -ForegroundColor Red
}

# Test 6: Test endpoint credit khusus
Write-Host "`n6. Testing Credit Accounts Endpoint..." -ForegroundColor Yellow
try {
    $creditResponse = Invoke-WebRequest -Uri "$baseUrl/accounts/credit?type=LIABILITY" -Method GET -Headers $headers
    if ($creditResponse.StatusCode -eq 200) {
        $creditData = $creditResponse.Content | ConvertFrom-Json
        Write-Host "✅ Credit accounts endpoint accessible - Count: $($creditData.count)" -ForegroundColor Green
    }
} catch {
    Write-Host "⚠️ Credit accounts endpoint: $($_.Exception.Message)" -ForegroundColor Orange
}

# Test 7: Test user permissions
Write-Host "`n7. Testing User Permissions..." -ForegroundColor Yellow
try {
    $permResponse = Invoke-WebRequest -Uri "$baseUrl/permissions/me" -Method GET -Headers $headers
    if ($permResponse.StatusCode -eq 200) {
        $permData = $permResponse.Content | ConvertFrom-Json
        Write-Host "✅ User permissions retrieved" -ForegroundColor Green
        
        if ($permData.permissions.accounts) {
            $accountPerm = $permData.permissions.accounts
            Write-Host "   Accounts permissions: View=$($accountPerm.can_view), Create=$($accountPerm.can_create)" -ForegroundColor Cyan
        }
        
        if ($permData.permissions.purchases) {
            $purchasePerm = $permData.permissions.purchases
            Write-Host "   Purchase permissions: View=$($purchasePerm.can_view), Create=$($purchasePerm.can_create)" -ForegroundColor Cyan
        }
    }
} catch {
    Write-Host "⚠️ User permissions: $($_.Exception.Message)" -ForegroundColor Orange
}

# Test 8: Test purchases endpoint
Write-Host "`n8. Testing Purchases Endpoint..." -ForegroundColor Yellow
try {
    $purchaseResponse = Invoke-WebRequest -Uri "$baseUrl/purchases" -Method GET -Headers $headers
    if ($purchaseResponse.StatusCode -eq 200) {
        Write-Host "✅ Purchases endpoint accessible" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ Purchases endpoint tidak bisa diakses: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "`n🎯 Permission Fix Test Complete!" -ForegroundColor Cyan
Write-Host "`nJika test 3, 4, dan 5 berhasil, maka permission fix sudah bekerja" -ForegroundColor Green
Write-Host "Frontend purchase form seharusnya sekarang bisa load accounts dropdown" -ForegroundColor Green