# Test Milestone API Endpoints
Write-Host "🧪 Testing Milestone API Endpoints..." -ForegroundColor Cyan

# 1. Login
Write-Host "`n1️⃣ Logging in..." -ForegroundColor Yellow
$loginBody = @{
    email = "admin@company.com"
    password = "password123"
} | ConvertTo-Json

try {
    $loginResponse = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/auth/login" `
        -Method Post `
        -ContentType "application/json" `
        -Body $loginBody
    
    $token = $loginResponse.token
    Write-Host "✅ Login successful! Token: $($token.Substring(0,20))..." -ForegroundColor Green
} catch {
    Write-Host "❌ Login failed: $_" -ForegroundColor Red
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}

# 2. Get all projects
Write-Host "`n2️⃣ Getting all projects..." -ForegroundColor Yellow
try {
    $projects = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/projects" `
        -Method Get `
        -Headers $headers
    
    Write-Host "✅ Found $($projects.Count) projects" -ForegroundColor Green
    if ($projects.Count -gt 0) {
        $projectId = $projects[0].id
        Write-Host "   Using project ID: $projectId" -ForegroundColor Cyan
    } else {
        Write-Host "⚠️ No projects found, creating one..." -ForegroundColor Yellow
        
        # Create a test project
        $projectBody = @{
            name = "Test Project for Milestones"
            code = "TEST-MILESTONE-001"
            description = "Test project for milestone functionality"
            start_date = (Get-Date).ToString("yyyy-MM-dd")
            end_date = (Get-Date).AddMonths(3).ToString("yyyy-MM-dd")
            status = "planning"
            budget = 100000000
        } | ConvertTo-Json
        
        $newProject = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/projects" `
            -Method Post `
            -Headers $headers `
            -Body $projectBody
        
        $projectId = $newProject.id
        Write-Host "✅ Created project ID: $projectId" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ Failed to get projects: $_" -ForegroundColor Red
    Write-Host "Response: $($_.Exception.Response)" -ForegroundColor Red
    exit 1
}

# 3. Get milestones for project
Write-Host "`n3️⃣ Getting milestones for project $projectId..." -ForegroundColor Yellow
try {
    $milestones = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/projects/$projectId/milestones" `
        -Method Get `
        -Headers $headers
    
    Write-Host "✅ Found $($milestones.Count) milestones" -ForegroundColor Green
} catch {
    Write-Host "❌ Failed to get milestones: $_" -ForegroundColor Red
    Write-Host "Response: $($_.Exception.Response)" -ForegroundColor Red
}

# 4. Create a milestone
Write-Host "`n4️⃣ Creating a milestone..." -ForegroundColor Yellow
$milestoneBody = @{
    name = "Initial Planning"
    description = "Complete initial project planning phase"
    target_date = (Get-Date).AddDays(14).ToString("yyyy-MM-dd")
    status = "pending"
    order_number = 1
    weight = 15
} | ConvertTo-Json

try {
    $newMilestone = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/projects/$projectId/milestones" `
        -Method Post `
        -Headers $headers `
        -Body $milestoneBody
    
    $milestoneId = $newMilestone.id
    Write-Host "✅ Created milestone ID: $milestoneId" -ForegroundColor Green
    Write-Host "   Name: $($newMilestone.name)" -ForegroundColor Cyan
} catch {
    Write-Host "❌ Failed to create milestone: $_" -ForegroundColor Red
}

# 5. Update milestone
if ($milestoneId) {
    Write-Host "`n5️⃣ Updating milestone $milestoneId..." -ForegroundColor Yellow
    $updateBody = @{
        name = "Initial Planning (Updated)"
        description = "Complete initial project planning phase with stakeholder review"
        target_date = (Get-Date).AddDays(14).ToString("yyyy-MM-dd")
        status = "in_progress"
        order_number = 1
        weight = 20
    } | ConvertTo-Json
    
    try {
        $updated = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/projects/$projectId/milestones/$milestoneId" `
            -Method Put `
            -Headers $headers `
            -Body $updateBody
        
        Write-Host "✅ Updated milestone successfully" -ForegroundColor Green
        Write-Host "   Status: $($updated.status)" -ForegroundColor Cyan
    } catch {
        Write-Host "❌ Failed to update milestone: $_" -ForegroundColor Red
    }
}

# 6. Complete milestone
if ($milestoneId) {
    Write-Host "`n6️⃣ Completing milestone $milestoneId..." -ForegroundColor Yellow
    try {
        $completed = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/projects/$projectId/milestones/$milestoneId/complete" `
            -Method Post `
            -Headers $headers
        
        Write-Host "✅ Completed milestone successfully" -ForegroundColor Green
        Write-Host "   Status: $($completed.status)" -ForegroundColor Cyan
        Write-Host "   Completion Date: $($completed.actual_completion_date)" -ForegroundColor Cyan
    } catch {
        Write-Host "❌ Failed to complete milestone: $_" -ForegroundColor Red
    }
}

# 7. Get single milestone
if ($milestoneId) {
    Write-Host "`n7️⃣ Getting single milestone $milestoneId..." -ForegroundColor Yellow
    try {
        $milestone = Invoke-RestMethod -Uri "http://localhost:8080/api/v1/projects/$projectId/milestones/$milestoneId" `
            -Method Get `
            -Headers $headers
        
        Write-Host "✅ Retrieved milestone successfully" -ForegroundColor Green
        Write-Host "   Name: $($milestone.name)" -ForegroundColor Cyan
        Write-Host "   Status: $($milestone.status)" -ForegroundColor Cyan
    } catch {
        Write-Host "❌ Failed to get milestone: $_" -ForegroundColor Red
    }
}

# 8. Delete milestone
if ($milestoneId) {
    Write-Host "`n8️⃣ Deleting milestone $milestoneId..." -ForegroundColor Yellow
    try {
        Invoke-RestMethod -Uri "http://localhost:8080/api/v1/projects/$projectId/milestones/$milestoneId" `
            -Method Delete `
            -Headers $headers
        
        Write-Host "✅ Deleted milestone successfully" -ForegroundColor Green
    } catch {
        Write-Host "❌ Failed to delete milestone: $_" -ForegroundColor Red
    }
}

Write-Host "`n✅ Milestone API tests completed!" -ForegroundColor Green

