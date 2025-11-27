
$baseUrl = "http://localhost:8080/api/v1"

# Register/Login
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

# Create Project
$projectData = @{
    project_name        = "Budget Test Project"
    project_description = "Testing budget precision"
    customer            = "Test Customer"
    city                = "Test City"
    address             = "Test Address"
    project_type        = "New Build"
    budget              = 1300000000
    deadline            = (Get-Date).AddDays(30).ToString("yyyy-MM-ddTHH:mm:ssZ")
}
$projectJson = $projectData | ConvertTo-Json
Write-Host "Sending Project Data: $projectJson"

$project = Invoke-RestMethod -Uri "$baseUrl/projects" -Method Post -Body $projectJson -ContentType "application/json" -Headers $headers

Write-Host "Created Project Budget: $($project.budget)"

if ($project.budget -eq 1300000000) {
    Write-Host "SUCCESS: Budget saved correctly." -ForegroundColor Green
}
else {
    Write-Host "FAILURE: Budget mismatch. Expected 1300000000, got $($project.budget)" -ForegroundColor Red
}
