# Test Weekly Reports Endpoint
# This script tests the weekly reports API endpoint

Write-Host "🧪 Testing Weekly Reports API Endpoint..." -ForegroundColor Cyan
Write-Host ""

# Configuration
$baseUrl = "http://localhost:8080"
$token = "" # Will be set after login

# Function to test API endpoint
function Test-Endpoint {
    param(
        [string]$Url,
        [string]$Method = "GET",
        [hashtable]$Headers = @{},
        [string]$Body = $null
    )
    
    try {
        $params = @{
            Uri = $Url
            Method = $Method
            Headers = $Headers
            ContentType = "application/json"
        }
        
        if ($Body) {
            $params.Body = $Body
        }
        
        $response = Invoke-RestMethod @params
        return $response
    } catch {
        Write-Host "❌ Error: $($_.Exception.Message)" -ForegroundColor Red
        if ($_.Exception.Response) {
            $reader = [System.IO.StreamReader]::new($_.Exception.Response.GetResponseStream())
            $errorBody = $reader.ReadToEnd()
            Write-Host "Response: $errorBody" -ForegroundColor Yellow
        }
        return $null
    }
}

# Step 1: Login
Write-Host "1️⃣  Logging in..." -ForegroundColor Yellow
$loginBody = @{
    email = "admin@company.com"
    password = "password123"
} | ConvertTo-Json

$loginResponse = Test-Endpoint `
    -Url "$baseUrl/api/v1/auth/login" `
    -Method "POST" `
    -Body $loginBody

if ($loginResponse -and $loginResponse.access_token) {
    $token = $loginResponse.access_token
    Write-Host "✅ Login successful!" -ForegroundColor Green
    Write-Host "   Token: $($token.Substring(0, 20))..." -ForegroundColor Gray
} else {
    Write-Host "❌ Login failed!" -ForegroundColor Red
    exit 1
}

Write-Host ""

# Step 2: Get all projects
Write-Host "2️⃣  Getting all projects..." -ForegroundColor Yellow
$headers = @{
    "Authorization" = "Bearer $token"
}

$projects = Test-Endpoint `
    -Url "$baseUrl/api/v1/projects" `
    -Headers $headers

if ($projects -and $projects.data) {
    Write-Host "✅ Found $($projects.data.Count) projects" -ForegroundColor Green
    
    if ($projects.data.Count -gt 0) {
        $firstProject = $projects.data[0]
        Write-Host "   First project:" -ForegroundColor Gray
        Write-Host "   - ID: $($firstProject.id)" -ForegroundColor Gray
        Write-Host "   - Name: $($firstProject.name)" -ForegroundColor Gray
        
        $projectId = $firstProject.id
        
        Write-Host ""
        
        # Step 3: Test weekly reports endpoint
        Write-Host "3️⃣  Testing weekly reports endpoint for project ID $projectId..." -ForegroundColor Yellow
        
        $weeklyReports = Test-Endpoint `
            -Url "$baseUrl/api/v1/projects/$projectId/weekly-reports" `
            -Headers $headers
        
        if ($weeklyReports) {
            if ($weeklyReports.data) {
                Write-Host "✅ Weekly reports endpoint working!" -ForegroundColor Green
                Write-Host "   Found $($weeklyReports.data.Count) weekly reports" -ForegroundColor Gray
                
                if ($weeklyReports.data.Count -gt 0) {
                    Write-Host ""
                    Write-Host "   Sample report:" -ForegroundColor Gray
                    $sampleReport = $weeklyReports.data[0]
                    Write-Host "   - ID: $($sampleReport.id)" -ForegroundColor Gray
                    Write-Host "   - Week: $($sampleReport.week)" -ForegroundColor Gray
                    Write-Host "   - Year: $($sampleReport.year)" -ForegroundColor Gray
                    Write-Host "   - Project Manager: $($sampleReport.project_manager)" -ForegroundColor Gray
                }
            } else {
                Write-Host "✅ Endpoint accessible but no data found" -ForegroundColor Yellow
                Write-Host "   This is normal if no weekly reports have been created yet" -ForegroundColor Gray
            }
        } else {
            Write-Host "❌ Weekly reports endpoint failed!" -ForegroundColor Red
        }
    } else {
        Write-Host "⚠️  No projects found. Create a project first." -ForegroundColor Yellow
    }
} else {
    Write-Host "❌ Failed to get projects!" -ForegroundColor Red
}

Write-Host ""
Write-Host "✅ Test completed!" -ForegroundColor Cyan

