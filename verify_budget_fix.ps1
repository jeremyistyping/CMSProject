
$baseUrl = "http://localhost:8080/api/v1"

# Register/Login
$registerData = @{
    username   = "final_test_$(Get-Random)"
    email      = "final_test_$(Get-Random)@example.com"
    password   = "password123"
    first_name = "Final"
    last_name  = "Test"
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

# Create Project with problematic budget value
$projectData = @{
    project_name        = "Budget Precision Test - Frontend Simulation"
    project_description = "Testing budget precision with large numbers"
    customer            = "Test Customer"
    city                = "Test City"
    address             = "Test Address"
    project_type        = "New Build"
    budget              = 1300000000
    deadline            = (Get-Date).AddDays(30).ToString("yyyy-MM-ddTHH:mm:ssZ")
}
$projectJson = $projectData | ConvertTo-Json
Write-Host "Creating project with budget: $($projectData.budget)"

$project = Invoke-RestMethod -Uri "$baseUrl/projects" -Method Post -Body $projectJson -ContentType "application/json" -Headers $headers

Write-Host "`nCreated Project:"
Write-Host "  ID: $($project.id)"
Write-Host "  Name: $($project.project_name)"
Write-Host "  Budget: $($project.budget)"

if ($project.budget -eq 1300000000) {
    Write-Host "`n✓ SUCCESS: Budget saved correctly as 1300000000" -ForegroundColor Green
}
else {
    Write-Host "`n✗ FAILURE: Budget mismatch!" -ForegroundColor Red
    Write-Host "  Expected: 1300000000" -ForegroundColor Red
    Write-Host "  Got: $($project.budget)" -ForegroundColor Red
}

# Also test with the edit page scenario
Write-Host "`n--- Testing Project 14 Update ---"
$project14 = Invoke-RestMethod -Uri "$baseUrl/projects/14" -Method Get -Headers $headers
Write-Host "Project 14 current budget: $($project14.budget)"

if ($project14.budget -eq 1300000000) {
    Write-Host "✓ Project 14 budget is correct after manual update" -ForegroundColor Green
}
else {
    Write-Host "! Project 14 budget needs re-checking" -ForegroundColor Yellow
}
