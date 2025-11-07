# Fix activity_logs user_id constraint
# This script updates the database to allow NULL values in user_id column

Write-Host "================================" -ForegroundColor Cyan
Write-Host "Fix Activity Logs Constraint" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host ""

# Database configuration
$DB_HOST = "localhost"
$DB_PORT = "5432"
$DB_NAME = "accounting_db"
$DB_USER = "postgres"

Write-Host "Database: $DB_NAME" -ForegroundColor Yellow
Write-Host ""

# SQL script path
$SQL_FILE = "fix_activity_logs_user_id_constraint.sql"

# Check if SQL file exists
if (-not (Test-Path $SQL_FILE)) {
    Write-Host "❌ Error: SQL file not found: $SQL_FILE" -ForegroundColor Red
    exit 1
}

Write-Host "📄 SQL file: $SQL_FILE" -ForegroundColor Green
Write-Host ""

# Prompt for password
$DB_PASSWORD = Read-Host "Enter PostgreSQL password for user '$DB_USER'" -AsSecureString
$BSTR = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($DB_PASSWORD)
$PlainPassword = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto($BSTR)

# Set environment variable for password
$env:PGPASSWORD = $PlainPassword

Write-Host "🔧 Applying migration..." -ForegroundColor Yellow
Write-Host ""

# Execute SQL file
try {
    # Using docker exec to run psql inside container
    $result = docker exec -i postgres_db psql -U $DB_USER -d $DB_NAME -f "/sql/$SQL_FILE" 2>&1
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Migration applied successfully!" -ForegroundColor Green
        Write-Host ""
        Write-Host "Changes made:" -ForegroundColor Cyan
        Write-Host "  - Dropped NOT NULL constraint on activity_logs.user_id" -ForegroundColor White
        Write-Host "  - Re-created foreign key constraint allowing NULL values" -ForegroundColor White
        Write-Host ""
        Write-Host "📝 Anonymous users can now be logged without user_id" -ForegroundColor Green
    } else {
        Write-Host "❌ Migration failed!" -ForegroundColor Red
        Write-Host $result -ForegroundColor Red
        exit 1
    }
} catch {
    Write-Host "❌ Error executing migration: $_" -ForegroundColor Red
    exit 1
} finally {
    # Clear password from environment
    $env:PGPASSWORD = $null
}

Write-Host ""
Write-Host "================================" -ForegroundColor Cyan
Write-Host "Done!" -ForegroundColor Green
Write-Host "================================" -ForegroundColor Cyan
