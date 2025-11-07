# Test Authentication and Get Token
Write-Host "Testing Authentication and P&L Endpoint..." -ForegroundColor Green

# Test 1: Login to get token
Write-Host "`n1. Testing Login..." -ForegroundColor Yellow
$loginBody = @{
    email = "admin@company.com"
    password = "admin123"
} | ConvertTo-Json

$loginResponse = try {
    Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
} catch {
    Write-Host "Login failed: $($_.Exception.Message)" -ForegroundColor Red
    $_.Exception.Response
}

if ($loginResponse.access_token) {
    Write-Host "Login successful! Token received." -ForegroundColor Green
    $token = $loginResponse.access_token
    Write-Host "User: $($loginResponse.user.username) ($($loginResponse.user.role))" -ForegroundColor Cyan
    
    # Test 2: Test SSOT P&L endpoint
    Write-Host "`n2. Testing SSOT P&L Endpoint..." -ForegroundColor Yellow
    
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }
    
    $plResponse = try {
        Invoke-RestMethod -Uri "http://localhost:8080/api/v1/reports/ssot-profit-loss?start_date=2025-01-01&end_date=2025-12-31&format=json" -Method GET -Headers $headers
    } catch {
        Write-Host "P&L request failed: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            Write-Host "Error details: $errorBody" -ForegroundColor Red
        }
        $null
    }
    
    if ($plResponse) {
        Write-Host "P&L endpoint accessible!" -ForegroundColor Green
        Write-Host "Status: $($plResponse.status)" -ForegroundColor Cyan
        
        if ($plResponse.data -and $plResponse.data.hasData) {
            Write-Host "P&L data found!" -ForegroundColor Green
            Write-Host "Company: $($plResponse.data.company.name)" -ForegroundColor Cyan
            Write-Host "Period: $($plResponse.data.period)" -ForegroundColor Cyan
            Write-Host "Sections: $($plResponse.data.sections.Count)" -ForegroundColor Cyan
            
            if ($plResponse.data.financialMetrics) {
                Write-Host "Financial Metrics:" -ForegroundColor Cyan
                Write-Host "  - Gross Profit: $($plResponse.data.financialMetrics.grossProfit)" -ForegroundColor White
                Write-Host "  - Net Income: $($plResponse.data.financialMetrics.netIncome)" -ForegroundColor White
            }
        } else {
            Write-Host "No P&L data available for the period" -ForegroundColor Yellow
            Write-Host "Message: $($plResponse.data.message)" -ForegroundColor Yellow
        }
    }
    
    # Test 3: Check user permissions
    Write-Host "`n3. Testing User Permissions..." -ForegroundColor Yellow
    $profileResponse = try {
        Invoke-RestMethod -Uri "http://localhost:8080/api/v1/profile" -Method GET -Headers $headers
    } catch {
        Write-Host "Profile request failed: $($_.Exception.Message)" -ForegroundColor Red
        $null
    }
    
    if ($profileResponse) {
        Write-Host "User authenticated successfully!" -ForegroundColor Green
        Write-Host "Role: $($profileResponse.role)" -ForegroundColor Cyan
        Write-Host "Username: $($profileResponse.username)" -ForegroundColor Cyan
    }
    
    # Output token for frontend testing
    Write-Host "Token for frontend testing:" -ForegroundColor Green
    Write-Host $token -ForegroundColor Yellow
    Write-Host "`nYou can use this token in browser localStorage:" -ForegroundColor Green
    Write-Host "localStorage.setItem('authToken', '$token')" -ForegroundColor Yellow
    
} else {
    Write-Host "Login failed! Please check credentials." -ForegroundColor Red
    Write-Host "Response: $($loginResponse | ConvertTo-Json)" -ForegroundColor Red
}

Write-Host "`n=== Test Complete ===" -ForegroundColor Green