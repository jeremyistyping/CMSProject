# PowerShell Script to Clean Database and Start Application
# This script will clean up problematic constraints and then start the Go backend

Write-Host "🚀 Starting Database Cleanup and Application Launch..." -ForegroundColor Green
Write-Host "=================================================" -ForegroundColor Cyan

# Get PostgreSQL connection info from environment or default values
$DB_HOST = if ($env:DB_HOST) { $env:DB_HOST } else { "localhost" }
$DB_PORT = if ($env:DB_PORT) { $env:DB_PORT } else { "5432" }
$DB_NAME = if ($env:DB_NAME) { $env:DB_NAME } else { "accounting_db" }
$DB_USER = if ($env:DB_USER) { $env:DB_USER } else { "postgres" }
$DB_PASSWORD = if ($env:DB_PASSWORD) { $env:DB_PASSWORD } else { "password" }

Write-Host "Database Configuration:" -ForegroundColor Yellow
Write-Host "  Host: $DB_HOST" -ForegroundColor White
Write-Host "  Port: $DB_PORT" -ForegroundColor White
Write-Host "  Database: $DB_NAME" -ForegroundColor White
Write-Host "  User: $DB_USER" -ForegroundColor White

# Check if psql is available
try {
    $psqlVersion = psql --version 2>$null
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ PostgreSQL client (psql) found: $psqlVersion" -ForegroundColor Green
        
        # Set PGPASSWORD environment variable for passwordless connection
        $env:PGPASSWORD = $DB_PASSWORD
        
        Write-Host "`n🧹 Running database constraint cleanup..." -ForegroundColor Yellow
        
        # Run the cleanup SQL script
        $cleanupResult = psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "cleanup_constraints.sql" 2>&1
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Database cleanup completed successfully!" -ForegroundColor Green
            Write-Host $cleanupResult -ForegroundColor Gray
        } else {
            Write-Host "⚠️  Database cleanup completed with warnings:" -ForegroundColor Yellow
            Write-Host $cleanupResult -ForegroundColor Gray
            Write-Host "Continuing with application start (warnings are usually OK)..." -ForegroundColor Yellow
        }
        
        # Clear PGPASSWORD for security
        Remove-Item Env:PGPASSWORD -ErrorAction SilentlyContinue
        
    } else {
        Write-Host "⚠️  PostgreSQL client (psql) not found in PATH" -ForegroundColor Yellow
        Write-Host "   Skipping database cleanup. You may need to run cleanup_constraints.sql manually." -ForegroundColor Yellow
    }
} catch {
    Write-Host "⚠️  Could not check for psql: $($_.Exception.Message)" -ForegroundColor Yellow
    Write-Host "   Skipping database cleanup. You may need to run cleanup_constraints.sql manually." -ForegroundColor Yellow
}

Write-Host "`n🔄 Starting Go application..." -ForegroundColor Yellow

# Check if we're in the correct directory
if (Test-Path "main.go") {
    Write-Host "✅ Found main.go in current directory" -ForegroundColor Green
} else {
    Write-Host "❌ main.go not found in current directory!" -ForegroundColor Red
    Write-Host "   Please make sure you're running this script from the backend directory" -ForegroundColor Red
    Write-Host "   Current directory: $(Get-Location)" -ForegroundColor Gray
    pause
    exit 1
}

# Set up environment variables if they don't exist
if (-not $env:DATABASE_URL) {
    $env:DATABASE_URL = "postgres://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=disable"
    Write-Host "🔧 Set DATABASE_URL: $env:DATABASE_URL" -ForegroundColor Cyan
}

if (-not $env:JWT_SECRET) {
    $env:JWT_SECRET = "your-super-secret-jwt-key-change-in-production"
    Write-Host "🔧 Set default JWT_SECRET (change this in production!)" -ForegroundColor Cyan
}

if (-not $env:PORT) {
    $env:PORT = "8080"
    Write-Host "🔧 Set PORT: $env:PORT" -ForegroundColor Cyan
}

Write-Host "`n📦 Installing/updating Go dependencies..." -ForegroundColor Yellow
go mod tidy

if ($LASTEXITCODE -eq 0) {
    Write-Host "✅ Dependencies updated successfully" -ForegroundColor Green
} else {
    Write-Host "⚠️  Dependency update failed, but continuing..." -ForegroundColor Yellow
}

Write-Host "`n🚀 Starting the application..." -ForegroundColor Green
Write-Host "   Access the API at: http://localhost:$env:PORT" -ForegroundColor Cyan
Write-Host "   Press Ctrl+C to stop the application`n" -ForegroundColor Gray

# Start the application
try {
    go run main.go
} catch {
    Write-Host "`n❌ Application failed to start: $($_.Exception.Message)" -ForegroundColor Red
    Write-Host "   Check the error messages above for troubleshooting" -ForegroundColor Yellow
    pause
    exit 1
}

Write-Host "`n👋 Application stopped." -ForegroundColor Yellow