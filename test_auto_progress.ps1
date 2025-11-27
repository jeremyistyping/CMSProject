
$baseUrl = "http://localhost:8080/api/v1"

# Register/Login
$registerData = @{
    username   = "progress_test_$(Get-Random)"
    email      = "progress_test_$(Get-Random)@example.com"
    password   = "password123"
    first_name = "Progress"
    last_name  = "Tester"
    role       = "admin"
}
$registerJson = $registerData | ConvertTo-Json
try {
    $registerResponse = Invoke-RestMethod -Uri "$baseUrl/auth/register" -Method Post -Body $registerJson -ContentType "application/json"
}
catch {}

$loginData = @{
    email    = $registerData.email
    password = $registerData.password
}
$loginJson = $loginData | ConvertTo-Json
$loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method Post -Body $loginJson -ContentType "application/json"
$token = $loginResponse.token
$headers = @{ "Authorization" = "Bearer $token" }

# Create Test Project
$projectData = @{
    project_name        = "Auto Progress Test Project"
    project_description = "Testing auto-calculated overall progress"
    customer            = "Test Customer"
    city                = "Test City"
    address             = "Test Address"
    project_type        = "New Build"
    budget              = 500000000
    deadline            = (Get-Date).AddDays(30).ToString("yyyy-MM-ddTHH:mm:ssZ")
}
$projectJson = $projectData | ConvertTo-Json
$project = Invoke-RestMethod -Uri "$baseUrl/projects" -Method Post -Body $projectJson -ContentType "application/json" -Headers $headers

Write-Host "`n✅ Created Project: $($project.project_name) (ID: $($project.id))" -ForegroundColor Green

# Create Daily Update with Category Progress
$dailyUpdateData = @{
    date                = (Get-Date).ToString("yyyy-MM-ddTHH:mm:ssZ")
    weather             = "Sunny"
    workers_present     = 10
    foundation_progress = 40
    utilities_progress  = 20
    interior_progress   = 30
    equipment_progress  = 10
    work_description    = "Testing auto-calculated overall progress"
    materials_used      = "Test materials"
    issues              = ""
    tomorrows_plan      = "Continue testing"
    created_by          = "Auto Test"
}

$dailyUpdateJson = $dailyUpdateData | ConvertTo-Json
Write-Host "`n📝 Creating Daily Update with category progress:"
Write-Host "  Foundation: 40%"
Write-Host "  Utilities: 20%"
Write-Host "  Interior: 30%"
Write-Host "  Equipment: 10%"
Write-Host "  Expected Overall: 25% (average)"

$dailyUpdate = Invoke-RestMethod -Uri "$baseUrl/projects/$($project.id)/daily-updates" -Method Post -Body $dailyUpdateJson -ContentType "application/json" -Headers $headers

# Fetch updated project to check progress
$updatedProject = Invoke-RestMethod -Uri "$baseUrl/projects/$($project.id)" -Method Get -Headers $headers

Write-Host "`n📊 Project Progress After Daily Update:"
Write-Host "  Overall Progress: $($updatedProject.overall_progress)%" -ForegroundColor Cyan
Write-Host "  Foundation: $($updatedProject.foundation_progress)%"  
Write-Host "  Utilities: $($updatedProject.utilities_progress)%"
Write-Host "  Interior: $($updatedProject.interior_progress)%"
Write-Host "  Equipment: $($updatedProject.equipment_progress)%"

# Verify calculation
if ($updated Project.overall_progress -eq 25) {
    Write-Host "`n✅ SUCCESS: Overall progress calculated correctly as 25% (average of 40, 20, 30, 10)" -ForegroundColor Green
} else {
    Write-Host "`n❌ FAILURE: Overall progress is $($updatedProject.overall_progress)% but expected 25%" -ForegroundColor Red
}

# Test with different values
Write-Host "`n📝 Creating another Daily Update with different values:"
Write-Host "  Foundation: 80%"
Write-Host "  Utilities: 60%"
Write-Host "  Interior: 70%"
Write-Host "  Equipment: 50%"
Write-Host "  Expected Overall: 65% (average)"

$dailyUpdateData2 = @{
    date                = (Get-Date).AddDays(1).ToString("yyyy-MM-ddTHH:mm:ssZ")
    weather             = "Cloudy"
    workers_present     = 15
    foundation_progress = 80
    utilities_progress  = 60
    interior_progress   = 70
    equipment_progress  = 50
    work_description    = "Second test of auto-calculated progress"
    materials_used      = "More materials"
    issues              = ""
    tomorrows_plan      = "Final testing"
    created_by          = "Auto Test"
}

$dailyUpdateJson2 = $dailyUpdateData2 | ConvertTo-Json
$dailyUpdate2 = Invoke-RestMethod -Uri "$baseUrl/projects/$($project.id)/daily-updates" -Method Post -Body $dailyUpdateJson2 -ContentType "application/json" -Headers $headers

# Fetch project again
$updatedProject2 = Invoke-RestMethod -Uri "$baseUrl/projects/$($project.id)" -Method Get -Headers $headers

Write-Host "`n📊 Project Progress After Second Daily Update:"
Write-Host "  Overall Progress: $($updatedProject2.overall_progress)%" -ForegroundColor Cyan
Write-Host "  Foundation: $($updatedProject2.foundation_progress)%"
Write-Host "  Utilities: $($updatedProject2.utilities_progress)%"
Write-Host "  Interior: $($updatedProject2.interior_progress)%"
Write-Host "  Equipment: $($updatedProject2.equipment_progress)%"

if ($updatedProject2.overall_progress -eq 65) {
    Write-Host "`n✅ SUCCESS: Overall progress updated correctly to 65% (average of 80, 60, 70, 50)" -ForegroundColor Green
}
else {
    Write-Host "`n❌ FAILURE: Overall progress is $($updatedProject2.overall_progress)% but expected 65%" -ForegroundColor Red
}

Write-Host "`n🎉 Auto-Calculate Overall Progress Test Complete!" -ForegroundColor Yellow
