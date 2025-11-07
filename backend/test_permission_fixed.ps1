Write-Host "Testing Permission Fix for Accounts Catalog" -ForegroundColor Green

$baseUrl = "http://localhost:8080/api/v1"

# Test Health
Write-Host "1. Testing Health..." -ForegroundColor Yellow
try {
    $health = Invoke-WebRequest -Uri "$baseUrl/health"
    Write-Host "✅ Health: $($health.StatusCode)" -ForegroundColor Green
} catch {
    Write-Host "❌ Health failed: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}

# Test Login dengan format yang benar
Write-Host "2. Testing Login..." -ForegroundColor Yellow
$loginBody = @{
    username = "admin"
    password = "admin123"  
} | ConvertTo-Json

try {
    $loginResponse = Invoke-WebRequest -Uri "$baseUrl/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    $loginData = $loginResponse.Content | ConvertFrom-Json
    $token = $loginData.token
    Write-Host "✅ Login Success - Token acquired" -ForegroundColor Green
} catch {
    Write-Host "❌ Login Failed: $($_.Exception.Message)" -ForegroundColor Red
    # Try with different credentials
    Write-Host "   Trying with manager credentials..." -ForegroundColor Yellow
    
    $loginBody2 = @{
        username = "manager"
        password = "manager123"  
    } | ConvertTo-Json
    
    try {
        $loginResponse = Invoke-WebRequest -Uri "$baseUrl/auth/login" -Method POST -Body $loginBody2 -ContentType "application/json"
        $loginData = $loginResponse.Content | ConvertFrom-Json
        $token = $loginData.token
        Write-Host "✅ Login Success with manager" -ForegroundColor Green
    } catch {
        Write-Host "❌ All login attempts failed" -ForegroundColor Red
        Write-Host "Response: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
        exit 1
    }
}

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

# Test Accounts Catalog (should work now without permission middleware)
Write-Host "3. Testing Accounts Catalog..." -ForegroundColor Yellow
try {
    $catalog = Invoke-WebRequest -Uri "$baseUrl/accounts/catalog" -Headers $headers
    $data = $catalog.Content | ConvertFrom-Json
    Write-Host "✅ Accounts Catalog Success - Count: $($data.count)" -ForegroundColor Green
    
    if ($data.data -and $data.data.Count -gt 0) {
        Write-Host "   Sample account: $($data.data[0].code) - $($data.data[0].name)" -ForegroundColor Cyan
    }
} catch {
    Write-Host "❌ Accounts Catalog Failed: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Response) {
        $errorResponse = $_.Exception.Response.GetResponseStream()
        $reader = New-Object System.IO.StreamReader($errorResponse)
        $errorBody = $reader.ReadToEnd()
        Write-Host "   Error detail: $errorBody" -ForegroundColor Red
    }
}

# Test EXPENSE Type specifically
Write-Host "4. Testing EXPENSE accounts..." -ForegroundColor Yellow
try {
    $expense = Invoke-WebRequest -Uri "$baseUrl/accounts/catalog?type=EXPENSE" -Headers $headers
    $data = $expense.Content | ConvertFrom-Json
    Write-Host "✅ EXPENSE accounts Success - Count: $($data.count)" -ForegroundColor Green
} catch {
    Write-Host "❌ EXPENSE accounts Failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test LIABILITY Type for credit
Write-Host "5. Testing LIABILITY accounts..." -ForegroundColor Yellow
try {
    $liability = Invoke-WebRequest -Uri "$baseUrl/accounts/catalog?type=LIABILITY" -Headers $headers
    $data = $liability.Content | ConvertFrom-Json
    Write-Host "✅ LIABILITY accounts Success - Count: $($data.count)" -ForegroundColor Green
} catch {
    Write-Host "❌ LIABILITY accounts Failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test Credit endpoint
Write-Host "6. Testing Credit endpoint..." -ForegroundColor Yellow
try {
    $credit = Invoke-WebRequest -Uri "$baseUrl/accounts/credit?type=LIABILITY" -Headers $headers
    $data = $credit.Content | ConvertFrom-Json
    Write-Host "✅ Credit endpoint Success - Count: $($data.count)" -ForegroundColor Green
} catch {
    Write-Host "❌ Credit endpoint Failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test user permissions
Write-Host "7. Testing User Permissions..." -ForegroundColor Yellow
try {
    $perms = Invoke-WebRequest -Uri "$baseUrl/permissions/me" -Headers $headers
    $permData = $perms.Content | ConvertFrom-Json
    Write-Host "✅ User permissions retrieved" -ForegroundColor Green
    Write-Host "   Role: $($permData.role)" -ForegroundColor Cyan
    
    if ($permData.permissions.accounts) {
        $accountPerm = $permData.permissions.accounts
        Write-Host "   Accounts - View: $($accountPerm.can_view), Create: $($accountPerm.can_create)" -ForegroundColor Cyan
    }
    
    if ($permData.permissions.purchases) {
        $purchasePerm = $permData.permissions.purchases  
        Write-Host "   Purchases - View: $($purchasePerm.can_view), Create: $($purchasePerm.can_create)" -ForegroundColor Cyan
    }
} catch {
    Write-Host "⚠️ User permissions: $($_.Exception.Message)" -ForegroundColor Orange
}

Write-Host "`n🎯 Permission Fix Test Complete!" -ForegroundColor Cyan
Write-Host "Jika test 3, 4, dan 5 berhasil, maka backend permission fix sudah bekerja" -ForegroundColor Green