# ==========================================
# RUN DUPLICATE ACCOUNTS CLEANUP
# ==========================================

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "CLEANING UP DUPLICATE ACCOUNTS" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Database connection
$dbHost = "localhost"
$dbPort = "5432"
$dbName = "sistem_akuntansi"
$dbUser = "postgres"
$dbPassword = "postgres"

$env:PGPASSWORD = $dbPassword

Write-Host "Connecting to database: $dbName" -ForegroundColor Yellow
Write-Host ""

# Run cleanup
Write-Host "Running cleanup script..." -ForegroundColor Green
$scriptPath = "cleanup_duplicate_accounts_now.sql"

if (Test-Path $scriptPath) {
    $result = psql -h $dbHost -p $dbPort -U $dbUser -d $dbName -f $scriptPath 2>&1
    Write-Host $result
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host ""
        Write-Host "✅ CLEANUP COMPLETED SUCCESSFULLY!" -ForegroundColor Green
        Write-Host ""
        Write-Host "Next step: Restart backend" -ForegroundColor Yellow
        Write-Host "Run: ..\restart_backend.ps1" -ForegroundColor White
    } else {
        Write-Host ""
        Write-Host "❌ CLEANUP FAILED!" -ForegroundColor Red
        Write-Host "Exit code: $LASTEXITCODE" -ForegroundColor Red
    }
} else {
    Write-Host "❌ Script file not found: $scriptPath" -ForegroundColor Red
    Write-Host "Make sure you're in the backend directory" -ForegroundColor Yellow
}

Write-Host ""

