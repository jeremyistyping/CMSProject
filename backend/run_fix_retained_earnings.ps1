# Fix Retained Earnings Balance Script

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "FIX RETAINED EARNINGS BALANCE" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if PostgreSQL is accessible
$pgPath = "C:\Program Files\PostgreSQL\14\bin\psql.exe"
if (-not (Test-Path $pgPath)) {
    $pgPath = "C:\Program Files\PostgreSQL\15\bin\psql.exe"
}
if (-not (Test-Path $pgPath)) {
    $pgPath = "C:\Program Files\PostgreSQL\16\bin\psql.exe"
}

if (-not (Test-Path $pgPath)) {
    Write-Host "ERROR: PostgreSQL psql not found. Please run manually:" -ForegroundColor Red
    Write-Host ""
    Write-Host "psql -U postgres -d accounting_db -f fix_retained_earnings_balance.sql" -ForegroundColor Yellow
    Write-Host ""
    exit 1
}

Write-Host "Found PostgreSQL at: $pgPath" -ForegroundColor Green
Write-Host ""

# Get database credentials from .env file
$envFile = ".env"
if (Test-Path $envFile) {
    Write-Host "Reading database config from .env..." -ForegroundColor Cyan
    $envContent = Get-Content $envFile
    
    $dbHost = ($envContent | Select-String "DB_HOST=(.*)").Matches.Groups[1].Value
    $dbPort = ($envContent | Select-String "DB_PORT=(.*)").Matches.Groups[1].Value
    $dbName = ($envContent | Select-String "DB_NAME=(.*)").Matches.Groups[1].Value
    $dbUser = ($envContent | Select-String "DB_USER=(.*)").Matches.Groups[1].Value
    
    if ([string]::IsNullOrEmpty($dbHost)) { $dbHost = "localhost" }
    if ([string]::IsNullOrEmpty($dbPort)) { $dbPort = "5432" }
    if ([string]::IsNullOrEmpty($dbName)) { $dbName = "accounting_db" }
    if ([string]::IsNullOrEmpty($dbUser)) { $dbUser = "postgres" }
    
    Write-Host "   Host: $dbHost" -ForegroundColor Gray
    Write-Host "   Port: $dbPort" -ForegroundColor Gray
    Write-Host "   Database: $dbName" -ForegroundColor Gray
    Write-Host "   User: $dbUser" -ForegroundColor Gray
    Write-Host ""
} else {
    # Default values
    $dbHost = "localhost"
    $dbPort = "5432"
    $dbName = "accounting_db"
    $dbUser = "postgres"
    
    Write-Host "WARNING: .env not found, using defaults" -ForegroundColor Yellow
    Write-Host ""
}

# Confirm before running
Write-Host "WARNING: This will fix the Retained Earnings balance in the database." -ForegroundColor Yellow
Write-Host "WARNING: Make sure you have a backup before proceeding!" -ForegroundColor Yellow
Write-Host ""
$confirm = Read-Host "Do you want to continue? (yes/no)"

if ($confirm -ne "yes") {
    Write-Host "Operation cancelled." -ForegroundColor Red
    exit 0
}

Write-Host ""
Write-Host "Running fix script..." -ForegroundColor Cyan
Write-Host ""

# Set password environment variable if needed
$env:PGPASSWORD = "postgres"  # Change this if your password is different

# Run the SQL script
& $pgPath -h $dbHost -p $dbPort -U $dbUser -d $dbName -f "fix_retained_earnings_balance.sql"

if ($LASTEXITCODE -eq 0) {
    Write-Host ""
    Write-Host "========================================" -ForegroundColor Green
    Write-Host "FIX COMPLETED SUCCESSFULLY!" -ForegroundColor Green
    Write-Host "========================================" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Cyan
    Write-Host "1. Refresh your Balance Sheet report" -ForegroundColor White
    Write-Host "2. Verify that the difference is now 0" -ForegroundColor White
    Write-Host "3. If still not balanced, check the server logs" -ForegroundColor White
    Write-Host ""
} else {
    Write-Host ""
    Write-Host "Fix script failed. Please check the error messages above." -ForegroundColor Red
    Write-Host ""
}

# Clean up
Remove-Item Env:\PGPASSWORD -ErrorAction SilentlyContinue
