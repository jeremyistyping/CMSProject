
$baseUrl = "http://localhost:8080/api/v1"

# Register a new user
$registerData = @{
    username   = "checker_$(Get-Random)"
    email      = "checker_$(Get-Random)@example.com"
    password   = "password123"
    first_name = "Check"
    last_name  = "User"
    role       = "admin"
}
$registerJson = $registerData | ConvertTo-Json
try {
    $registerResponse = Invoke-RestMethod -Uri "$baseUrl/auth/register" -Method Post -Body $registerJson -ContentType "application/json"
    Write-Host "Registration successful for $($registerData.email)"
}
catch {
    Write-Host "Registration failed: $_" -ForegroundColor Yellow
}

# Login
$loginData = @{
    email    = $registerData.email
    password = $registerData.password
}
$loginJson = $loginData | ConvertTo-Json
try {
    $loginResponse = Invoke-RestMethod -Uri "$baseUrl/auth/login" -Method Post -Body $loginJson -ContentType "application/json"
    $token = $loginResponse.token
}
catch {
    Write-Host "Login failed: $_" -ForegroundColor Red
    exit 1
}

$headers = @{
    "Authorization" = "Bearer $token"
}

# Get all projects
$projects = Invoke-RestMethod -Uri "$baseUrl/projects" -Method Get -Headers $headers

# Find the project
$targetProject = $projects | Where-Object { $_.project_name -like "*Padel Jakarta Senayan*" }

if ($targetProject) {
    Write-Host "Project Found: $($targetProject.project_name)"
    Write-Host "Budget (Raw): $($targetProject.budget)"
    $targetProject | ConvertTo-Json -Depth 5
}
else {
    Write-Host "Project not found."
}
