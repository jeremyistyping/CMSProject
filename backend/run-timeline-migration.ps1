# PowerShell script to run timeline_schedules migration
$ErrorActionPreference = "Stop"

Write-Host "Running Timeline Schedules Migration..." -ForegroundColor Green

# Database connection details
$server = "localhost"
$port = "5432"
$database = "CMSNew"
$username = "postgres"
$password = "Moon"

# Read SQL file
$sqlFile = "migrations\054_create_timeline_schedules_table.sql"
$sql = Get-Content $sqlFile -Raw

# Connection string
$connectionString = "Server=$server;Port=$port;Database=$database;User Id=$username;Password=$password;"

try {
    # Load Npgsql assembly (PostgreSQL .NET driver)
    Add-Type -Path "C:\Program Files\PostgreSQL\*\Npgsql.dll" -ErrorAction SilentlyContinue
    
    if (-not ([System.Management.Automation.PSTypeName]'Npgsql.NpgsqlConnection').Type) {
        Write-Host "Npgsql not found, trying alternative method..." -ForegroundColor Yellow
        
        # Alternative: Use .NET SqlClient (works for basic SQL)
        $connection = New-Object System.Data.SqlClient.SqlConnection
        $connection.ConnectionString = $connectionString
        $connection.Open()
        
        $command = $connection.CreateCommand()
        $command.CommandText = $sql
        $command.ExecuteNonQuery()
        
        $connection.Close()
        Write-Host "✅ Migration completed successfully!" -ForegroundColor Green
    }
} catch {
    Write-Host "❌ Error running migration: $_" -ForegroundColor Red
    Write-Host ""
    Write-Host "Alternative: Run migration using pgAdmin or manually:" -ForegroundColor Yellow
    Write-Host "1. Open pgAdmin" -ForegroundColor Cyan
    Write-Host "2. Connect to database 'CMSNew'" -ForegroundColor Cyan
    Write-Host "3. Execute the SQL from: $sqlFile" -ForegroundColor Cyan
    exit 1
}

