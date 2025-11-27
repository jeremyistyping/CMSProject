
# Test Project Progress Automation
$baseUrl = "http://localhost:8080/api/v1"

# 0. Register a new user (in case admin doesn't exist or password changed)
$registerData = @{
    username   = "tester_$(Get-Random)"
    email      = "tester_$(Get-Random)@example.com"
    password   = "password123"
    first_name = "Test"
    last_name  = "User"
    role       = "admin"
}
$registerJson = $registerData | ConvertTo-Json
try {
    $registerResponse = Invoke-RestMethod -Uri "$baseUrl/auth/register" -Method Post -Body $registerJson -ContentType "application/json"
    Write-Host "Registration successful for $($registerData.email)" -ForegroundColor Green
}
catch {
    Write-Host "Registration failed (might already exist): $_" -ForegroundColor Yellow
}

# 1. Login to get token
$loginData = @{
    email    = $registerData.email
    password = $registerData.password
}
$loginJson = $loginData | ConvertTo-Json
try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method Post -Body $loginJson -ContentType "application/json"
    $token = $loginResponse.token
    Write-Host "Login successful. Token obtained." -ForegroundColor Green
}
catch {
    Write-Host "Login failed: $_" -ForegroundColor Red
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $token"
}

# 1. Create a test project
$projectData = @{
    project_name        = "Test Progress Automation Project"
    project_description = "Testing automatic progress update"
    customer            = "Test Customer"
    city                = "Test City"
    address             = "Test Address"
    project_type        = "New Build"
    budget              = 1000000
    deadline            = (Get-Date).AddDays(30).ToString("yyyy-MM-ddTHH:mm:ssZ")
    overall_progress    = 0
}
$projectJson = $projectData | ConvertTo-Json
$project = Invoke-RestMethod -Uri "$baseUrl/projects" -Method Post -Body $projectJson -ContentType "application/json" -Headers $headers
$projectId = $project.id
Write-Host "Created Project ID: $projectId with Progress: $($project.overall_progress)%"

# 4. Create a Daily Update with Category Progress
Write-Host "Creating a daily update with category progress..."
$dailyUpdateBody = @{
    date                = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")
    weather             = "Sunny"
    workers_present     = 15
    work_description    = "Foundation work completed"
    progress            = 25
    foundation_progress = 100
    utilities_progress  = 20
    interior_progress   = 5
    equipment_progress  = 0
    created_by          = "Test Script"
} | ConvertTo-Json

$dailyUpdateResponse = Invoke-RestMethod -Uri "$baseUrl/projects/$projectId/daily-updates" -Method Post -Body $dailyUpdateBody -Headers $headers -ContentType "application/json"
Write-Host "Daily Update Created: $($dailyUpdateResponse.id)"

# 5. Verify Project Progress Updated
Write-Host "Verifying project progress..."
$projectResponse = Invoke-RestMethod -Uri "$baseUrl/projects/$projectId" -Method Get -Headers $headers
if ($projectResponse.overall_progress -eq 25 -and 
    $projectResponse.foundation_progress -eq 100 -and 
    $projectResponse.utilities_progress -eq 20 -and 
    $projectResponse.interior_progress -eq 5) {
    Write-Host "SUCCESS: Project progress updated correctly!" -ForegroundColor Green
}
else {
    Write-Host "FAILURE: Project progress mismatch." -ForegroundColor Red
    Write-Host "Expected: Overall=25, Foundation=100, Utilities=20, Interior=5"
    Write-Host "Actual: Overall=$($projectResponse.overall_progress), Foundation=$($projectResponse.foundation_progress), Utilities=$($projectResponse.utilities_progress), Interior=$($projectResponse.interior_progress)"
}

# 6. Update the Daily Update (Change Progress)
Write-Host "Updating the daily update..."
$updateBody = @{
    date                = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")
    weather             = "Sunny"
    workers_present     = 15
    work_description    = "Foundation work completed (Updated)"
    created_by          = "Test Script"
    progress            = 30
    foundation_progress = 100
    utilities_progress  = 30
    interior_progress   = 10
    equipment_progress  = 5
} | ConvertTo-Json

$updatedDailyUpdate = Invoke-RestMethod -Uri "$baseUrl/projects/$projectId/daily-updates/$($dailyUpdateResponse.id)" -Method Put -Body $updateBody -Headers $headers -ContentType "application/json"
Write-Host "Daily Update Updated"

# 7. Verify Project Progress Updated Again
Write-Host "Verifying project progress after update..."
$projectResponse = Invoke-RestMethod -Uri "$baseUrl/projects/$projectId" -Method Get -Headers $headers
if ($projectResponse.overall_progress -eq 30 -and 
    $projectResponse.utilities_progress -eq 30 -and 
    $projectResponse.equipment_progress -eq 5) {
    Write-Host "SUCCESS: Project progress updated correctly after edit!" -ForegroundColor Green
}
else {
    Write-Host "FAILURE: Project progress mismatch after edit." -ForegroundColor Red
    Write-Host "Actual: Overall=$($projectResponse.overall_progress), Utilities=$($projectResponse.utilities_progress), Equipment=$($projectResponse.equipment_progress)"
}
