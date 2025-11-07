# Test endpoints
Write-Host "Testing health endpoint..."
try {
    $health = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/health" -Method GET
    Write-Host "Health response:" $health.status
} catch {
    Write-Host "Health error:" $_.Exception.Message
}

Write-Host "`nTesting login endpoint..."
try {
    $loginData = @{
        email = "admin@company.com"
        password = "admin"
    } | ConvertTo-Json

    $login = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method POST -ContentType "application/json" -Body $loginData
    Write-Host "Login successful! Token length:" $login.access_token.Length
    Write-Host "User:" $login.user.username
} catch {
    Write-Host "Login error:" $_.Exception.Message
    Write-Host "Status code:" $_.Exception.Response.StatusCode
}

Write-Host "`nTesting sales integrated-payment endpoint without auth..."
try {
    $payment = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/sales/1/integrated-payment" -Method POST -ContentType "application/json" -Body "{}"
    Write-Host "Payment response:" $payment
} catch {
    Write-Host "Payment error:" $_.Exception.Message
    Write-Host "Status code:" $_.Exception.Response.StatusCode
}