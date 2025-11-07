# Test different login credentials
$credentials = @(
    @{ email = "admin@test.com"; password = "admin123" },
    @{ email = "admin@admin.com"; password = "admin123" },
    @{ email = "admin@example.com"; password = "admin123" },
    @{ email = "admin"; password = "admin123" }
)

foreach ($cred in $credentials) {
    Write-Host "Testing: $($cred.email) / $($cred.password)" -ForegroundColor Yellow
    
    $loginBody = $cred | ConvertTo-Json
    $loginHeaders = @{ "Content-Type" = "application/json" }

    try {
        $loginResponse = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/auth/login" -Method Post -Headers $loginHeaders -Body $loginBody
        $loginData = $loginResponse.Content | ConvertFrom-Json
        Write-Host "✅ SUCCESS: $($cred.email)" -ForegroundColor Green
        Write-Host "Token: $($loginData.token.Substring(0,50))..." -ForegroundColor Green
        break
    }
    catch {
        Write-Host "❌ FAILED: $($cred.email)" -ForegroundColor Red
    }
}
