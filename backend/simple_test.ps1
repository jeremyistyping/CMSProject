Write-Host "Testing Permission Fix" -ForegroundColor Green

$baseUrl = "http://localhost:8080/api/v1"

# Test Health
Write-Host "1. Testing Health..." -ForegroundColor Yellow
$health = Invoke-WebRequest -Uri "$baseUrl/health"
Write-Host "Health: $($health.StatusCode)" -ForegroundColor Green

# Test Login
Write-Host "2. Testing Login..." -ForegroundColor Yellow
$loginBody = '{"username":"manager","password":"manager123"}'
try {
    $loginResponse = Invoke-WebRequest -Uri "$baseUrl/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    $token = ($loginResponse.Content | ConvertFrom-Json).token
    Write-Host "Login Success" -ForegroundColor Green
} catch {
    $loginBody = '{"username":"admin","password":"admin123"}'
    $loginResponse = Invoke-WebRequest -Uri "$baseUrl/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    $token = ($loginResponse.Content | ConvertFrom-Json).token
    Write-Host "Login Success with admin" -ForegroundColor Green
}

$headers = @{"Authorization" = "Bearer $token"}

# Test Accounts Catalog
Write-Host "3. Testing Accounts Catalog..." -ForegroundColor Yellow
try {
    $catalog = Invoke-WebRequest -Uri "$baseUrl/accounts/catalog" -Headers $headers
    $data = $catalog.Content | ConvertFrom-Json
    Write-Host "Accounts Catalog Success - Count: $($data.count)" -ForegroundColor Green
} catch {
    Write-Host "Accounts Catalog Failed: $($_.Exception.Message)" -ForegroundColor Red
}

# Test EXPENSE Type
Write-Host "4. Testing EXPENSE accounts..." -ForegroundColor Yellow
try {
    $expense = Invoke-WebRequest -Uri "$baseUrl/accounts/catalog?type=EXPENSE" -Headers $headers
    $data = $expense.Content | ConvertFrom-Json
    Write-Host "EXPENSE accounts Success - Count: $($data.count)" -ForegroundColor Green
} catch {
    Write-Host "EXPENSE accounts Failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "Test Complete!" -ForegroundColor Cyan