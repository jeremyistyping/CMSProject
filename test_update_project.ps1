
$baseUrl = "http://localhost:8080/api/v1"
$projectId = 14

# Register/Login
$registerData = @{
    username   = "updater_$(Get-Random)"
    email      = "updater_$(Get-Random)@example.com"
    password   = "password123"
    first_name = "Update"
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

# Fetch existing project
$project = Invoke-RestMethod -Uri "$baseUrl/projects/$projectId" -Method Get -Headers $headers

# Update Budget
$project.budget = 1300000000
$updateJson = $project | ConvertTo-Json -Depth 10

Write-Host "Updating Project $projectId Budget to 1300000000..."

$updatedProject = Invoke-RestMethod -Uri "$baseUrl/projects/$projectId" -Method Put -Body $updateJson -ContentType "application/json" -Headers $headers

Write-Host "Updated Project Budget: $($updatedProject.budget)"

if ($updatedProject.budget -eq 1300000000) {
    Write-Host "SUCCESS: Budget updated correctly." -ForegroundColor Green
}
else {
    Write-Host "FAILURE: Budget mismatch. Expected 1300000000, got $($updatedProject.budget)" -ForegroundColor Red
}
