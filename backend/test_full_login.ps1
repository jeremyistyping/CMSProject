Write-Host "Testing Login API and CORS Configuration" -ForegroundColor Green
Write-Host "================================================"

# Test 1: Direct backend API call
Write-Host "`n1. Testing direct backend API call..." -ForegroundColor Yellow
$body = @{
    email = "admin@company.com"
    password = "password123"
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" `
        -Method Post `
        -ContentType "application/json" `
        -Body $body `
        -Headers @{
            "User-Agent" = "Mozilla/5.0"
            "Origin" = "http://localhost:3000"
        }
    
    Write-Host "✅ Backend API works!" -ForegroundColor Green
    Write-Host "Response keys: $($response.PSObject.Properties.Name -join ', ')"
    
    if ($response.access_token) {
        Write-Host "✅ access_token present" -ForegroundColor Green
    }
    if ($response.token) {
        Write-Host "✅ token (compatibility) present" -ForegroundColor Green
    }
    if ($response.user) {
        Write-Host "✅ user object present" -ForegroundColor Green
    }
    if ($response.success) {
        Write-Host "✅ success flag present" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ Backend API failed:" -ForegroundColor Red
    Write-Host $_.Exception.Message
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $errorBody = $reader.ReadToEnd()
        Write-Host "Error Body: $errorBody"
    }
}

# Test 2: CORS preflight check
Write-Host "`n2. Testing CORS preflight..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/api/v1/auth/login" `
        -Method Options `
        -Headers @{
            "Origin" = "http://localhost:3000"
            "Access-Control-Request-Method" = "POST"
            "Access-Control-Request-Headers" = "Content-Type,Authorization"
        } `
        -UseBasicParsing
        
    Write-Host "✅ CORS preflight works!" -ForegroundColor Green
    Write-Host "Status: $($response.StatusCode)"
    
    $corsHeaders = $response.Headers | Where-Object { $_.Key -like "*Access-Control*" }
    foreach ($header in $corsHeaders) {
        Write-Host "CORS: $($header.Key) = $($header.Value -join ', ')"
    }
} catch {
    Write-Host "❌ CORS preflight failed:" -ForegroundColor Red
    Write-Host $_.Exception.Message
}

# Test 3: Health check endpoint
Write-Host "`n3. Testing health endpoint..." -ForegroundColor Yellow
try {
    $health = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/health" -Method Get
    Write-Host "✅ Health endpoint works!" -ForegroundColor Green
} catch {
    Write-Host "❌ Health endpoint failed:" -ForegroundColor Red
    Write-Host $_.Exception.Message
}

# Test 4: Check if frontend is running
Write-Host "`n4. Checking if frontend is running on port 3000..." -ForegroundColor Yellow
try {
    $frontendTest = Invoke-WebRequest -Uri "http://localhost:3000" -Method Get -TimeoutSec 5 -UseBasicParsing
    Write-Host "✅ Frontend is running!" -ForegroundColor Green
} catch {
    Write-Host "⚠️  Frontend might not be running on port 3000" -ForegroundColor Yellow
    Write-Host "Error: $($_.Exception.Message)"
}

Write-Host "`n================================================"
Write-Host "Tests completed!" -ForegroundColor Green

Write-Host "`nSolutions for frontend login issues:" -ForegroundColor Cyan
Write-Host "1. Ensure frontend runs on port 3000"
Write-Host "2. Clear browser cache and localStorage"
Write-Host "3. Check browser console for errors"
Write-Host "4. Verify API URL configuration"
