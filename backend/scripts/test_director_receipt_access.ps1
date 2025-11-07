# Test Director Receipt Access
Write-Host "🧪 Testing Director Access to Receipt Creation..." -ForegroundColor Cyan

# Configuration
$baseUrl = "http://localhost:8080"
$loginUrl = "$baseUrl/api/v1/auth/login" 
$receiptUrl = "$baseUrl/api/v1/purchases/receipts"

# Director credentials (using the director we updated)
$directorCredentials = @{
    username = "director"
    password = "director123"  # Adjust password as needed
} | ConvertTo-Json -Depth 3

Write-Host "1. Logging in as director..." -ForegroundColor Yellow

try {
    # Login as director
    $headers = @{
        "Content-Type" = "application/json"
    }
    
    $loginResponse = Invoke-RestMethod -Uri $loginUrl -Method POST -Body $directorCredentials -Headers $headers
    
    if ($loginResponse.token) {
        Write-Host "✅ Login successful!" -ForegroundColor Green
        
        # Extract token
        $token = $loginResponse.token
        
        # Test receipt creation endpoint access (without creating actual receipt)
        Write-Host "2. Testing receipt creation endpoint access..." -ForegroundColor Yellow
        
        $authHeaders = @{
            "Authorization" = "Bearer $token"
            "Content-Type" = "application/json"
        }
        
        # Test with minimal invalid data to check permissions (should get validation error, not 403)
        $testReceiptData = @{
            purchase_id = 999999  # Non-existent purchase ID
            received_date = "2025-09-21"
            receipt_items = @()
        } | ConvertTo-Json -Depth 3
        
        try {
            $receiptResponse = Invoke-RestMethod -Uri $receiptUrl -Method POST -Body $testReceiptData -Headers $authHeaders
            Write-Host "✅ Permissions OK - Got past authentication!" -ForegroundColor Green
            Write-Host "   Response: $($receiptResponse | ConvertTo-Json -Depth 2)" -ForegroundColor Gray
        }
        catch {
            $statusCode = $_.Exception.Response.StatusCode.value__
            $errorMessage = $_.Exception.Message
            
            if ($statusCode -eq 403) {
                Write-Host "❌ PERMISSION DENIED (403) - Director still cannot create receipts!" -ForegroundColor Red
                Write-Host "   Error: $errorMessage" -ForegroundColor Red
            } elseif ($statusCode -eq 400 -or $statusCode -eq 404) {
                Write-Host "✅ PERMISSIONS OK - Got validation/not found error (not permission error)" -ForegroundColor Green
                Write-Host "   Status: $statusCode - $errorMessage" -ForegroundColor Gray
                Write-Host "   This means the director can access the endpoint but the test data is invalid (which is expected)" -ForegroundColor Gray
            } else {
                Write-Host "⚠️  Unexpected error (Status: $statusCode)" -ForegroundColor Orange
                Write-Host "   Error: $errorMessage" -ForegroundColor Orange
            }
        }
        
    } else {
        Write-Host "❌ Login failed - No token received" -ForegroundColor Red
    }
}
catch {
    $errorDetails = $_.Exception.Response
    if ($errorDetails) {
        $statusCode = $errorDetails.StatusCode.value__
        Write-Host "❌ Login failed with status $statusCode" -ForegroundColor Red
    } else {
        Write-Host "❌ Network error: $($_.Exception.Message)" -ForegroundColor Red
    }
    Write-Host "   Make sure the backend server is running on http://localhost:8080" -ForegroundColor Yellow
}

Write-Host "`n🎯 Test completed!" -ForegroundColor Cyan