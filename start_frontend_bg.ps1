
$frontendDir = Join-Path $PSScriptRoot "frontend"
Set-Location $frontendDir
Start-Process -FilePath "npm" -ArgumentList "run", "dev" -RedirectStandardOutput "..\frontend_log.txt" -RedirectStandardError "..\frontend_error.txt" -WindowStyle Hidden
Write-Host "Frontend started in background. Logs at frontend_log.txt"
