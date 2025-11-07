# PowerShell script to add stock to products for testing sales functionality

# Login to get authentication token
$loginBody = @{
    username = "admin"
    password = "password123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" -Method POST -Body $loginBody -ContentType "application/json"
    $token = $loginResponse.access_token
    Write-Host "✅ Login successful, token obtained"
    
    # Create headers with authorization
    $headers = @{
        "Authorization" = "Bearer $token"
        "Content-Type" = "application/json"
    }
    
    # Get list of products
    Write-Host "📋 Getting products list..."
    $productsResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/products" -Method GET -Headers $headers
    $products = $productsResponse.data
    
    Write-Host "Found $($products.Length) products"
    
    # Add stock to each product using the adjust-stock endpoint
    foreach ($product in $products) {
        Write-Host "📦 Adding stock to product: $($product.name) (ID: $($product.id))"
        
        $stockAdjustment = @{
            product_id = $product.id
            adjustment_type = "IN"  # Stock increase
            quantity = 100  # Add 100 units
            reason = "Initial stock for testing sales functionality"
            reference = "INIT-$(Get-Date -Format 'yyyyMMdd')"
        } | ConvertTo-Json
        
        try {
            $adjustResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/products/adjust-stock" -Method POST -Body $stockAdjustment -Headers $headers
            Write-Host "  ✅ Added 100 units to $($product.name)"
        }
        catch {
            Write-Host "  ❌ Failed to add stock to $($product.name): $($_.Exception.Message)"
        }
    }
    
    Write-Host "`n✅ Stock adjustment completed!"
    Write-Host "📊 Updated product stock levels for testing sales functionality"
}
catch {
    Write-Host "❌ Login failed: $($_.Exception.Message)"
    Write-Host "Make sure the backend server is running on http://localhost:8080"
}