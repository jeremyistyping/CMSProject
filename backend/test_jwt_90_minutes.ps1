# Test Script untuk Verifikasi JWT Token 90 Menit
# Script ini akan melakukan login dan memverifikasi bahwa token yang dihasilkan memiliki masa berlaku 90 menit

Write-Host "🔍 Testing JWT Token 90-Minute Configuration" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan

# Define the API base URL
$baseUrl = "http://localhost:8080/api/v1"

# Test credentials - adjust as needed
$loginData = @{
    username = "admin"  # Adjust this to your test user
    password = "admin123"  # Adjust this to your test password
} | ConvertTo-Json

Write-Host "📝 Testing with credentials:" -ForegroundColor Yellow
Write-Host "   Username: admin" -ForegroundColor Gray
Write-Host "   Password: [HIDDEN]" -ForegroundColor Gray
Write-Host ""

try {
    # Step 1: Login to get JWT token
    Write-Host "Step 1: Attempting login..." -ForegroundColor Green
    
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method POST -Body $loginData -ContentType "application/json"
    
    if ($loginResponse.access_token) {
        Write-Host "✅ Login successful!" -ForegroundColor Green
        Write-Host "🎫 Access Token: $($loginResponse.access_token.Substring(0, 50))..." -ForegroundColor Gray
        Write-Host "🔄 Refresh Token: $($loginResponse.refresh_token.Substring(0, 50))..." -ForegroundColor Gray
        Write-Host "⏱️  Expires In: $($loginResponse.expires_in) seconds" -ForegroundColor Gray
        Write-Host "📅 Expires At: $($loginResponse.expires_at)" -ForegroundColor Gray
        Write-Host ""
        
        # Verify expires_in is 5400 seconds (90 minutes)
        if ($loginResponse.expires_in -eq 5400) {
            Write-Host "✅ VERIFIED: Token expires in 5400 seconds (90 minutes)" -ForegroundColor Green
        } else {
            Write-Host "❌ ERROR: Token expires in $($loginResponse.expires_in) seconds (expected 5400)" -ForegroundColor Red
        }
        
        # Step 2: Analyze the token using our analyzer
        Write-Host ""
        Write-Host "Step 2: Analyzing token with JWT analyzer..." -ForegroundColor Green
        
        $token = $loginResponse.access_token
        & go run analyze_jwt_token.go $token
        
        # Step 3: Test the token with a protected endpoint
        Write-Host ""
        Write-Host "Step 3: Testing token with protected endpoint..." -ForegroundColor Green
        
        $headers = @{
            "Authorization" = "Bearer $token"
            "Content-Type" = "application/json"
        }
        
        try {
            $testResponse = Invoke-RestMethod -Uri "$baseUrl/notifications/approvals?limit=5" -Method GET -Headers $headers
            Write-Host "✅ Protected endpoint access successful!" -ForegroundColor Green
            Write-Host "📋 Retrieved $($testResponse.total) notifications" -ForegroundColor Gray
        }
        catch {
            if ($_.Exception.Response.StatusCode -eq 401) {
                Write-Host "❌ Token validation failed (401 Unauthorized)" -ForegroundColor Red
                Write-Host "   This could indicate the token is already expired or invalid" -ForegroundColor Yellow
            } else {
                Write-Host "⚠️  Endpoint test failed: $($_.Exception.Message)" -ForegroundColor Yellow
            }
        }
        
    } else {
        Write-Host "❌ Login failed - no access token received" -ForegroundColor Red
    }
}
catch {
    if ($_.Exception.Message -like "*Connection*") {
        Write-Host "❌ Cannot connect to server at $baseUrl" -ForegroundColor Red
        Write-Host "   Please make sure the backend server is running" -ForegroundColor Yellow
    } else {
        Write-Host "❌ Login failed: $($_.Exception.Message)" -ForegroundColor Red
        
        # Try to show more details
        if ($_.Exception.Response) {
            $statusCode = $_.Exception.Response.StatusCode.value__
            Write-Host "   HTTP Status Code: $statusCode" -ForegroundColor Gray
        }
    }
}

Write-Host ""
Write-Host "🏁 Test completed!" -ForegroundColor Cyan
Write-Host ""
Write-Host "📋 Summary of Changes Made:" -ForegroundColor Cyan
Write-Host "  ✅ Updated middleware/jwt.go - access token expiry: 15m → 90m" -ForegroundColor Green
Write-Host "  ✅ Updated config/config.go - default JWT_ACCESS_EXPIRY: 15m → 90m" -ForegroundColor Green
Write-Host "  ✅ Updated token response ExpiresIn: 900s → 5400s" -ForegroundColor Green
Write-Host "  ✅ Updated monitoring logic to reflect 90-minute expiry" -ForegroundColor Green
Write-Host "  ✅ Updated refresh token logging to use 90-minute expiry" -ForegroundColor Green
Write-Host ""
Write-Host "💡 Notes:" -ForegroundColor Yellow
Write-Host "  • Restart the backend server to apply changes" -ForegroundColor Gray
Write-Host "  • New tokens will have 90-minute expiry" -ForegroundColor Gray
Write-Host "  • Existing tokens will keep their original expiry time" -ForegroundColor Gray