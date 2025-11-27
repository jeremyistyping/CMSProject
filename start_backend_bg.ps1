
$backendDir = Join-Path $PSScriptRoot "backend"
Set-Location $backendDir
Start-Process -FilePath "go" -ArgumentList "run", "main.go" -RedirectStandardOutput "..\backend_log.txt" -RedirectStandardError "..\backend_error.txt" -WindowStyle Hidden
Write-Host "Backend started in background. Logs at backend_log.txt"
