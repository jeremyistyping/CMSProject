# Enhanced Sales Summary Test Script
# Tests all the improvements implemented for the sales summary report

Write-Host "=== Enhanced Sales Summary Report Test ===" -ForegroundColor Cyan
Write-Host "Testing improved timezone handling, logging, and error handling" -ForegroundColor Green

# Configuration
$baseUrl = "http://localhost:8080/api/reports"
$token = ""  # Get this from your authentication

# Test parameters
$testCases = @(
    @{
        name = "September 2025 (Original Problem)"
        startDate = "2025-09-01"
        endDate = "2025-09-30"
        groupBy = "month"
        expectedResult = "Should show enhanced debug info if no data"
    },
    @{
        name = "Current Month"
        startDate = (Get-Date -Day 1).ToString("yyyy-MM-dd")
        endDate = (Get-Date).ToString("yyyy-MM-dd")
        groupBy = "month"
        expectedResult = "Should show current month data or helpful debug info"
    },
    @{
        name = "Last 30 Days"
        startDate = (Get-Date).AddDays(-30).ToString("yyyy-MM-dd")
        endDate = (Get-Date).ToString("yyyy-MM-dd")
        groupBy = "day"
        expectedResult = "Should show daily breakdown for last 30 days"
    },
    @{
        name = "Invalid Date Range"
        startDate = "2025-12-31"
        endDate = "2025-01-01"
        groupBy = "month"
        expectedResult = "Should return validation error with helpful message"
    },
    @{
        name = "Future Dates"
        startDate = "2026-01-01"
        endDate = "2026-01-31"
        groupBy = "month"
        expectedResult = "Should return data or helpful debug info"
    }
)

Write-Host "`nRunning test cases..." -ForegroundColor Yellow

foreach ($testCase in $testCases) {
    Write-Host "`n--- Test Case: $($testCase.name) ---" -ForegroundColor Cyan
    Write-Host "Expected: $($testCase.expectedResult)" -ForegroundColor Gray
    
    $url = "$baseUrl/sales-summary?start_date=$($testCase.startDate)&end_date=$($testCase.endDate)&group_by=$($testCase.groupBy)&format=json"
    
    try {
        $startTime = Get-Date
        
        # Make the request
        $headers = @{}
        if ($token) {
            $headers["Authorization"] = "Bearer $token"
        }
        
        $response = Invoke-RestMethod -Uri $url -Method GET -Headers $headers -ContentType "application/json"
        
        $endTime = Get-Date
        $duration = ($endTime - $startTime).TotalMilliseconds
        
        Write-Host "✅ SUCCESS (${duration}ms)" -ForegroundColor Green
        Write-Host "Status: $($response.status)" -ForegroundColor White
        
        if ($response.data) {
            $data = $response.data
            Write-Host "Total Revenue: $($data.total_revenue)" -ForegroundColor White
            Write-Host "Total Transactions: $($data.total_transactions)" -ForegroundColor White
            Write-Host "Total Customers: $($data.total_customers)" -ForegroundColor White
            
            if ($data.processing_time) {
                Write-Host "Processing Time: $($data.processing_time)" -ForegroundColor White
            }
            
            if ($data.data_quality_score) {
                Write-Host "Data Quality Score: $($data.data_quality_score)%" -ForegroundColor White
            }
            
            # Check for enhanced debug info
            if ($data.debug_info) {
                Write-Host "📋 Debug Info Available:" -ForegroundColor Blue
                if ($data.debug_info.message) {
                    Write-Host "  Message: $($data.debug_info.message)" -ForegroundColor Gray
                }
                if ($data.debug_info.date_range_info) {
                    Write-Host "  Timezone: $($data.debug_info.date_range_info.timezone)" -ForegroundColor Gray
                    Write-Host "  Start (Jakarta): $($data.debug_info.date_range_info.start_date_jakarta)" -ForegroundColor Gray
                    Write-Host "  End (Jakarta): $($data.debug_info.date_range_info.end_date_jakarta)" -ForegroundColor Gray
                }
                if ($data.debug_info.suggestions) {
                    Write-Host "  Suggestions:" -ForegroundColor Gray
                    foreach ($suggestion in $data.debug_info.suggestions) {
                        Write-Host "    - $suggestion" -ForegroundColor Gray
                    }
                }
            }
            
            # Check metadata
            if ($response.metadata) {
                Write-Host "📊 Metadata:" -ForegroundColor Blue
                Write-Host "  Version: $($response.metadata.version)" -ForegroundColor Gray
                Write-Host "  Has Data: $($response.metadata.has_data)" -ForegroundColor Gray
                Write-Host "  Generated At: $($response.metadata.generated_at)" -ForegroundColor Gray
                Write-Host "  Timezone: $($response.metadata.timezone)" -ForegroundColor Gray
            }
        } else {
            Write-Host "⚠️ No data in response" -ForegroundColor Yellow
        }
        
    } catch {
        Write-Host "❌ ERROR" -ForegroundColor Red
        Write-Host "Message: $($_.Exception.Message)" -ForegroundColor Red
        
        # Try to parse error response
        if ($_.Exception.Response) {
            $statusCode = $_.Exception.Response.StatusCode
            Write-Host "Status Code: $statusCode" -ForegroundColor Red
            
            try {
                $errorStream = $_.Exception.Response.GetResponseStream()
                $reader = New-Object System.IO.StreamReader($errorStream)
                $errorBody = $reader.ReadToEnd() | ConvertFrom-Json
                
                if ($errorBody.debug) {
                    Write-Host "🔧 Enhanced Debug Info:" -ForegroundColor Blue
                    Write-Host "  Start Date: $($errorBody.debug.start_date)" -ForegroundColor Gray
                    Write-Host "  End Date: $($errorBody.debug.end_date)" -ForegroundColor Gray
                    Write-Host "  Parsed Start: $($errorBody.debug.parsed_start_date)" -ForegroundColor Gray
                    Write-Host "  Parsed End: $($errorBody.debug.parsed_end_date)" -ForegroundColor Gray
                    Write-Host "  Timezone: $($errorBody.debug.timezone)" -ForegroundColor Gray
                }
            } catch {
                Write-Host "Could not parse error response" -ForegroundColor Yellow
            }
        }
    }
    
    Write-Host "URL: $url" -ForegroundColor DarkGray
}

# Test API endpoint availability
Write-Host "`n--- API Endpoint Availability Test ---" -ForegroundColor Cyan

$endpoints = @(
    "/api/reports/sales-summary",
    "/api/reports/comprehensive/sales-summary",
    "/api/reports/available"
)

foreach ($endpoint in $endpoints) {
    $testUrl = "$baseUrl$($endpoint.Replace('/api/reports', ''))"
    Write-Host "Testing: $testUrl" -ForegroundColor Gray
    
    try {
        $response = Invoke-RestMethod -Uri $testUrl -Method GET -ContentType "application/json" -TimeoutSec 5
        Write-Host "✅ Available" -ForegroundColor Green
    } catch {
        Write-Host "❌ Not available or error: $($_.Exception.Message)" -ForegroundColor Red
    }
}

# Test timezone handling
Write-Host "`n--- Timezone Handling Test ---" -ForegroundColor Cyan
Write-Host "Current local time: $(Get-Date)" -ForegroundColor White
Write-Host "Current UTC time: $((Get-Date).ToUniversalTime())" -ForegroundColor White

$jakartaTime = try {
    [System.TimeZoneInfo]::ConvertTimeBySystemTimeZoneId((Get-Date), "SE Asia Standard Time")
} catch {
    (Get-Date).AddHours(7)  # Fallback to UTC+7
}

Write-Host "Jakarta time (WIB): $jakartaTime" -ForegroundColor White

# Summary
Write-Host "`n=== Test Summary ===" -ForegroundColor Cyan
Write-Host "✅ Enhanced Sales Summary Report Testing Completed" -ForegroundColor Green
Write-Host "📋 Check the results above for:" -ForegroundColor Blue
Write-Host "  - Timezone-aware date parsing" -ForegroundColor Gray
Write-Host "  - Enhanced error messages with debug info" -ForegroundColor Gray
Write-Host "  - Comprehensive logging (check server logs)" -ForegroundColor Gray
Write-Host "  - Data quality analysis" -ForegroundColor Gray
Write-Host "  - Processing time metrics" -ForegroundColor Gray
Write-Host "  - Helpful suggestions for troubleshooting" -ForegroundColor Gray

Write-Host "`n📝 Next Steps:" -ForegroundColor Yellow
Write-Host "1. Check server logs for detailed logging output" -ForegroundColor White
Write-Host "2. Verify database has sales data for test periods" -ForegroundColor White
Write-Host "3. Test with real authentication token if needed" -ForegroundColor White
Write-Host "4. Run the debug script to analyze database directly" -ForegroundColor White

Write-Host "`n🎯 The enhanced implementation should now provide:" -ForegroundColor Green
Write-Host "- Clear explanations when no data is found" -ForegroundColor White
Write-Host "- Timezone-aware date handling for Indonesia (WIB)" -ForegroundColor White
Write-Host "- Comprehensive debugging information" -ForegroundColor White
Write-Host "- Better error messages with actionable suggestions" -ForegroundColor White