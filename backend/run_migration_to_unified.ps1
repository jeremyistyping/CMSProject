# ================================================================
# MIGRATION TO UNIFIED JOURNALS SYSTEM - PowerShell Script
# ================================================================

param(
    [string]$DbHost = "localhost",
    [string]$DbName = "sistem_akuntansi",
    [string]$DbUser = "postgres",
    [string]$DbPassword = "postgres",
    [switch]$DryRun
)

$ErrorActionPreference = "Stop"

function Write-ColorOutput {
    param([string]$Message, [string]$Color = "White")
    Write-Host $Message -ForegroundColor $Color
}

function Write-Header {
    param([string]$Text)
    Write-Host ""
    Write-ColorOutput ("=" * 80) "Cyan"
    Write-ColorOutput $Text "Cyan"
    Write-ColorOutput ("=" * 80) "Cyan"
    Write-Host ""
}

# Set environment variable for password
$env:PGPASSWORD = $DbPassword

Write-Header "MIGRATION TO UNIFIED JOURNALS SYSTEM (SSOT)"

Write-ColorOutput "Database: $DbName@$DbHost" "Cyan"
Write-ColorOutput "User: $DbUser" "Cyan"

if ($DryRun) {
    Write-ColorOutput "Mode: DRY RUN (no changes will be committed)" "Yellow"
} else {
    Write-ColorOutput "Mode: PRODUCTION (changes will be committed)" "Red"
}

# Get script directory
$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$MigrationScript = Join-Path $ScriptDir "migrate_to_unified_journals.sql"
$BackupDir = Join-Path $ScriptDir "backups"
$BackupFile = Join-Path $BackupDir "backup_before_unified_migration_$(Get-Date -Format 'yyyyMMdd_HHmmss').sql"

# Create backup directory
if (-not (Test-Path $BackupDir)) {
    New-Item -ItemType Directory -Path $BackupDir | Out-Null
}

# Step 1: Backup Database
Write-Header "STEP 1: DATABASE BACKUP"

Write-ColorOutput "Creating backup: $BackupFile..." "Yellow"

try {
    $pgDumpCmd = "pg_dump -h $DbHost -U $DbUser -d $DbName -f `"$BackupFile`" 2>&1"
    $output = Invoke-Expression $pgDumpCmd
    
    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "✓ Backup created successfully" "Green"
        $fileSize = (Get-Item $BackupFile).Length / 1MB
        Write-ColorOutput "  Size: $([math]::Round($fileSize, 2)) MB" "Cyan"
    } else {
        Write-ColorOutput "✗ Backup failed" "Red"
        Write-Host $output
        exit 1
    }
} catch {
    Write-ColorOutput "✗ Backup error: $_" "Red"
    exit 1
}

# Step 2: Run Migration Script
Write-Header "STEP 2: RUNNING MIGRATION"

if ($DryRun) {
    Write-ColorOutput "Preparing DRY RUN mode..." "Yellow"
    # Modify script to use ROLLBACK instead of COMMIT
    $scriptContent = Get-Content $MigrationScript -Raw
    $scriptContent = $scriptContent -replace "-- ROLLBACK;", "ROLLBACK;"
    $scriptContent = $scriptContent -replace "^COMMIT;", "-- COMMIT;"
    
    $dryRunScript = Join-Path $ScriptDir "migrate_to_unified_journals_dryrun.sql"
    $scriptContent | Out-File -FilePath $dryRunScript -Encoding UTF8
    
    Write-ColorOutput "Running migration in DRY RUN mode..." "Yellow"
    psql -h $DbHost -U $DbUser -d $DbName -f $dryRunScript
    
    Remove-Item $dryRunScript -Force
} else {
    Write-ColorOutput "⚠️  WARNING: This will make permanent changes to the database!" "Yellow"
    Write-ColorOutput "   - Delete old journal_entries and journal_lines" "Yellow"
    Write-ColorOutput "   - Reset all accounts.balance to 0" "Yellow"
    Write-ColorOutput "   - Recalculate balances from unified_journal_ledger" "Yellow"
    Write-Host ""
    
    $confirmation = Read-Host "Type 'YES' to proceed with migration"
    
    if ($confirmation -ne "YES") {
        Write-ColorOutput "`n✗ Migration cancelled by user" "Yellow"
        exit 0
    }
    
    Write-ColorOutput "`nRunning migration..." "Yellow"
    psql -h $DbHost -U $DbUser -d $DbName -f $MigrationScript
}

if ($LASTEXITCODE -ne 0) {
    Write-ColorOutput "✗ Migration failed" "Red"
    Write-ColorOutput "`n📋 Recovery:" "Yellow"
    Write-ColorOutput "   Restore from backup:" "White"
    Write-ColorOutput "   psql -h $DbHost -U $DbUser -d $DbName -f `"$BackupFile`"" "Cyan"
    exit 1
}

Write-ColorOutput "✓ Migration completed successfully" "Green"

# Step 3: Verification
Write-Header "STEP 3: VERIFICATION"

Write-ColorOutput "Checking balance sheet status..." "Yellow"

$verifyQuery = @"
WITH balance_totals AS (
    SELECT 
        SUM(CASE WHEN type = 'ASSET' THEN balance ELSE 0 END) AS total_assets,
        SUM(CASE WHEN type = 'LIABILITY' THEN balance ELSE 0 END) AS total_liabilities,
        SUM(CASE WHEN type = 'EQUITY' THEN balance ELSE 0 END) AS total_equity,
        SUM(CASE WHEN type = 'REVENUE' THEN balance ELSE 0 END) AS total_revenue,
        SUM(CASE WHEN type = 'EXPENSE' THEN balance ELSE 0 END) AS total_expense
    FROM accounts
    WHERE is_active = true AND COALESCE(is_header, false) = false
)
SELECT 
    'Assets: Rp ' || ROUND(total_assets) || CHR(10) ||
    'Liabilities: Rp ' || ROUND(total_liabilities) || CHR(10) ||
    'Equity: Rp ' || ROUND(total_equity) || CHR(10) ||
    'Revenue: Rp ' || ROUND(total_revenue) || ' (not closed yet)' || CHR(10) ||
    'Expense: Rp ' || ROUND(total_expense) || ' (not closed yet)' || CHR(10) ||
    'Net Income: Rp ' || ROUND(total_revenue - total_expense) || CHR(10) ||
    'Balance Check: ' || 
    CASE 
        WHEN ABS(total_assets - (total_liabilities + total_equity + total_revenue - total_expense)) < 0.01 
        THEN 'BALANCED (with temp accounts) ✓'
        ELSE 'NOT BALANCED ✗'
    END as summary
FROM balance_totals;
"@

$verifyResult = psql -h $DbHost -U $DbUser -d $DbName -t -c $verifyQuery

Write-Host $verifyResult

# Step 4: Next Steps
Write-Header "MIGRATION COMPLETED SUCCESSFULLY!"

Write-ColorOutput "📋 NEXT STEPS:" "Cyan"
Write-ColorOutput ""
Write-ColorOutput "1. ✅ Old journal_entries have been deleted" "Green"
Write-ColorOutput "2. ✅ Account balances now reflect unified journals only" "Green"
Write-ColorOutput "3. ⚠️  Revenue & Expense accounts still have balances (need closing)" "Yellow"
Write-ColorOutput ""
Write-ColorOutput "TO CLOSE THE PERIOD:" "Cyan"
Write-ColorOutput "  Run period closing via API or application UI:" "White"
Write-ColorOutput "  - Endpoint: POST /api/period-closing/execute" "Cyan"
Write-ColorOutput "  - Date Range: From first transaction to end of fiscal year" "Cyan"
Write-ColorOutput ""
Write-ColorOutput "AFTER PERIOD CLOSING:" "Cyan"
Write-ColorOutput "  - Revenue & Expense accounts will be 0" "White"
Write-ColorOutput "  - Net Income will transfer to Retained Earnings (3201)" "White"
Write-ColorOutput "  - Balance Sheet will be fully balanced" "White"
Write-ColorOutput ""
Write-ColorOutput "BACKUP LOCATION:" "Cyan"
Write-ColorOutput "  $BackupFile" "White"
Write-ColorOutput ""

# Clean up environment variable
Remove-Item Env:\PGPASSWORD

Write-ColorOutput "Done!" "Green"
Write-Host ""
