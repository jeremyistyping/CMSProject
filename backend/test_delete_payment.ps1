# Test Delete Payment Script
Write-Host "Testing Payment Deletion..." -ForegroundColor Yellow

# Step 1: Login to get token
$loginBody = @{
    email = "admin@company.com"
    password = "password123"
} | ConvertTo-Json

$loginHeaders = @{
    "Content-Type" = "application/json"
}

try {
    Write-Host "1. Logging in..." -ForegroundColor Green
    $loginResponse = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/auth/login" -Method Post -Headers $loginHeaders -Body $loginBody
    $loginData = $loginResponse.Content | ConvertFrom-Json
    
    Write-Host "Login response: $($loginResponse.Content)" -ForegroundColor Cyan
    
    if (-not $loginData.access_token) {
        Write-Host "❌ No token in login response" -ForegroundColor Red
        return
    }
    
    $token = $loginData.access_token
    Write-Host "✅ Login successful, got token" -ForegroundColor Green

    # Step 2: Get payments list to find a payment ID
    $authHeaders = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }

    Write-Host "2. Getting payments list..." -ForegroundColor Green
    Write-Host "Auth token: Bearer $($token.Substring(0,50))..." -ForegroundColor Cyan
    
    try {
        $paymentsResponse = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/payments?page=1&limit=10" -Method Get -Headers $authHeaders
        $paymentsData = $paymentsResponse.Content | ConvertFrom-Json
        Write-Host "✅ Payments retrieved successfully" -ForegroundColor Green
    }
    catch {
        Write-Host "❌ Failed to get payments: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            $reader.BaseStream.Position = 0
            $reader.DiscardBufferedData()
            $responseBody = $reader.ReadToEnd()
            Write-Host "Error details: $responseBody" -ForegroundColor Red
        }
        return
    }
    
    if ($paymentsData.data -and $paymentsData.data.Length -gt 0) {
        $paymentId = $paymentsData.data[0].id
        Write-Host "✅ Found payment ID: $paymentId" -ForegroundColor Green

        # Step 3: Delete the payment
        $deleteBody = @{
            reason = "Test deletion"
        } | ConvertTo-Json

        Write-Host "3. Attempting to delete payment $paymentId..." -ForegroundColor Green
        
        try {
            $deleteResponse = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/payments/$paymentId" -Method Delete -Headers $authHeaders -Body $deleteBody
            $deleteData = $deleteResponse.Content | ConvertFrom-Json
            Write-Host "✅ Payment deleted successfully: $($deleteData.message)" -ForegroundColor Green
        }
        catch {
            Write-Host "❌ Delete failed: $($_.Exception.Message)" -ForegroundColor Red
            Write-Host "Response: $($_.Exception.Response.StatusCode)" -ForegroundColor Red
            if ($_.Exception.Response) {
                $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
                $reader.BaseStream.Position = 0
                $reader.DiscardBufferedData()
                $responseBody = $reader.ReadToEnd()
                Write-Host "Error details: $responseBody" -ForegroundColor Red
            }
        }
    } else {
        Write-Host "❌ No payments found to delete" -ForegroundColor Red
    }
}
catch {
    Write-Host "❌ Login failed: $($_.Exception.Message)" -ForegroundColor Red
}

Write-Host "Test completed." -ForegroundColor Yellow
