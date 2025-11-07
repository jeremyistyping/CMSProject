# ================================================
# SCRIPT PERBAIKAN DOUBLE POSTING LABA DITAHAN
# ================================================
# Usage:
#   .\fix_double_posting.ps1 -Action analyze     # Analisis saja
#   .\fix_double_posting.ps1 -Action fix         # Analisis + Fix
#   .\fix_double_posting.ps1 -Action backup      # Backup database saja

param(
    [Parameter(Mandatory=$true)]
    [ValidateSet("analyze", "fix", "backup")]
    [string]$Action,
    
    [string]$DbHost = "localhost",
    [string]$DbName = "sistem_akuntansi",
    [string]$DbUser = "postgres",
    [string]$DbPassword = "postgres"
)

$ErrorActionPreference = "Stop"

# Colors for output
function Write-ColorOutput {
    param(
        [string]$Message,
        [string]$Color = "White"
    )
    Write-Host $Message -ForegroundColor $Color
}

function Write-Header {
    param([string]$Text)
    Write-Host ""
    Write-ColorOutput "================================================" "Cyan"
    Write-ColorOutput $Text "Cyan"
    Write-ColorOutput "================================================" "Cyan"
    Write-Host ""
}

# Set environment variable for password
$env:PGPASSWORD = $DbPassword

# Check if psql is available
try {
    $null = Get-Command psql -ErrorAction Stop
    Write-ColorOutput "✓ psql found" "Green"
} catch {
    Write-ColorOutput "✗ psql not found. Please install PostgreSQL client tools." "Red"
    Write-ColorOutput "  Download from: https://www.postgresql.org/download/windows/" "Yellow"
    exit 1
}

# Get script directory
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path

# Define file paths
$AnalyzeScript = Join-Path $ScriptDir "analyze_double_posting_issue.sql"
$FixScript = Join-Path $ScriptDir "fix_double_posting_issue.sql"
$BackupDir = Join-Path $ScriptDir "backups"
$BackupFile = Join-Path $BackupDir "backup_before_fix_$(Get-Date -Format 'yyyyMMdd_HHmmss').sql"

# Check if SQL scripts exist
if (-not (Test-Path $AnalyzeScript)) {
    Write-ColorOutput "✗ Analyze script not found: $AnalyzeScript" "Red"
    exit 1
}

if ($Action -eq "fix" -and -not (Test-Path $FixScript)) {
    Write-ColorOutput "✗ Fix script not found: $FixScript" "Red"
    exit 1
}

# Create backup directory if not exists
if (-not (Test-Path $BackupDir)) {
    New-Item -ItemType Directory -Path $BackupDir | Out-Null
    Write-ColorOutput "✓ Created backup directory: $BackupDir" "Green"
}

# Function to run psql command
function Invoke-Psql {
    param(
        [string]$SqlFile,
        [string]$Description
    )
    
    Write-ColorOutput "`n▶ $Description..." "Yellow"
    
    try {
        $output = psql -h $DbHost -U $DbUser -d $DbName -f $SqlFile 2>&1
        
        if ($LASTEXITCODE -eq 0) {
            Write-ColorOutput "✓ Success" "Green"
            Write-Host $output
            return $true
        } else {
            Write-ColorOutput "✗ Failed" "Red"
            Write-Host $output
            return $false
        }
    } catch {
        Write-ColorOutput "✗ Error: $_" "Red"
        return $false
    }
}

# Function to backup database
function Backup-Database {
    Write-Header "DATABASE BACKUP"
    
    Write-ColorOutput "▶ Creating backup: $BackupFile..." "Yellow"
    
    try {
        # Use pg_dump
        $pgDumpCmd = "pg_dump -h $DbHost -U $DbUser -d $DbName -f `"$BackupFile`" 2>&1"
        $output = Invoke-Expression $pgDumpCmd
        
        if ($LASTEXITCODE -eq 0) {
            Write-ColorOutput "✓ Backup created successfully" "Green"
            Write-ColorOutput "  Location: $BackupFile" "Cyan"
            
            $fileSize = (Get-Item $BackupFile).Length / 1MB
            Write-ColorOutput "  Size: $([math]::Round($fileSize, 2)) MB" "Cyan"
            return $true
        } else {
            Write-ColorOutput "✗ Backup failed" "Red"
            Write-Host $output
            return $false
        }
    } catch {
        Write-ColorOutput "✗ Backup error: $_" "Red"
        return $false
    }
}

# Main execution
Write-Header "DOUBLE POSTING FIX TOOL"

Write-ColorOutput "Database: $DbName@$DbHost" "Cyan"
Write-ColorOutput "User: $DbUser" "Cyan"
Write-ColorOutput "Action: $Action" "Cyan"

switch ($Action) {
    "backup" {
        if (Backup-Database) {
            Write-Header "BACKUP COMPLETED"
            Write-ColorOutput "✓ Database backup completed successfully" "Green"
        } else {
            Write-ColorOutput "✗ Backup failed" "Red"
            exit 1
        }
    }
    
    "analyze" {
        Write-Header "ANALYZING DATABASE"
        
        if (Invoke-Psql -SqlFile $AnalyzeScript -Description "Running analysis") {
            Write-Header "ANALYSIS COMPLETED"
            Write-ColorOutput "`nReview the output above to understand the issue." "Yellow"
            Write-ColorOutput "If you want to fix the issue, run:" "Yellow"
            Write-ColorOutput "  .\fix_double_posting.ps1 -Action fix" "Cyan"
        } else {
            Write-ColorOutput "✗ Analysis failed" "Red"
            exit 1
        }
    }
    
    "fix" {
        Write-Header "FIXING DOUBLE POSTING ISSUE"
        
        # Step 1: Analyze first
        Write-ColorOutput "`n[1/3] Analysis" "Magenta"
        if (-not (Invoke-Psql -SqlFile $AnalyzeScript -Description "Pre-fix analysis")) {
            Write-ColorOutput "✗ Pre-fix analysis failed" "Red"
            exit 1
        }
        
        # Step 2: Backup
        Write-ColorOutput "`n[2/3] Backup" "Magenta"
        if (-not (Backup-Database)) {
            Write-ColorOutput "✗ Backup failed. Aborting fix." "Red"
            exit 1
        }
        
        # Step 3: Ask for confirmation
        Write-Host ""
        Write-ColorOutput "⚠️  WARNING: About to modify database!" "Yellow"
        Write-ColorOutput "   Backup location: $BackupFile" "Cyan"
        $confirmation = Read-Host "   Do you want to proceed with the fix? (yes/no)"
        
        if ($confirmation -ne "yes") {
            Write-ColorOutput "`n✗ Fix cancelled by user" "Yellow"
            exit 0
        }
        
        # Step 4: Apply fix
        Write-ColorOutput "`n[3/3] Applying Fix" "Magenta"
        if (Invoke-Psql -SqlFile $FixScript -Description "Applying database fix") {
            Write-Header "FIX COMPLETED"
            Write-ColorOutput "✓ Database fix completed successfully" "Green"
            Write-ColorOutput "`n📋 Next Steps:" "Cyan"
            Write-ColorOutput "   1. Verify the balance sheet is now balanced" "White"
            Write-ColorOutput "   2. Check Laba Ditahan balance is correct" "White"
            Write-ColorOutput "   3. Rebuild and restart the application:" "White"
            Write-ColorOutput "      go build -o app-sistem-akuntansi.exe" "Yellow"
        } else {
            Write-ColorOutput "✗ Fix failed" "Red"
            Write-ColorOutput "`n📋 Recovery Steps:" "Yellow"
            Write-ColorOutput "   1. Restore from backup:" "White"
            Write-ColorOutput "      psql -h $DbHost -U $DbUser -d $DbName -f `"$BackupFile`"" "Yellow"
            Write-ColorOutput "   2. Contact support if issue persists" "White"
            exit 1
        }
    }
}

Write-Host ""
Write-ColorOutput "Done!" "Green"
Write-Host ""

# Clean up environment variable
Remove-Item Env:\PGPASSWORD
