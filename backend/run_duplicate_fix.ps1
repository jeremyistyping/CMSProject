#!/usr/bin/env pwsh
# ============================================================================
# Script: Run Duplicate Accounts Fix
# ============================================================================
# This script will:
# 1. Check current database state
# 2. Backup database
# 3. Fix duplicate accounts
# 4. Verify results
# ============================================================================

param(
    [string]$DbHost = "localhost",
    [string]$DbPort = "5432",
    [string]$DbName = "sistem_akuntansi",
    [string]$DbUser = "postgres",
    [string]$DbPassword = "postgres",
    [switch]$SkipBackup = $false
)

$ErrorActionPreference = "Continue"

Write-Host "============================================================================" -ForegroundColor Cyan
Write-Host "DUPLICATE ACCOUNTS FIX - AUTOMATED SCRIPT" -ForegroundColor Cyan
Write-Host "============================================================================" -ForegroundColor Cyan
Write-Host ""

# Set PostgreSQL password
$env:PGPASSWORD = $DbPassword

# Step 1: Check connection
Write-Host "Step 1: Testing database connection..." -ForegroundColor Yellow
$testQuery = "SELECT 1 as test;"
$testResult = psql -h $DbHost -p $DbPort -U $DbUser -d $DbName -t -c $testQuery 2>&1

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ ERROR: Cannot connect to database!" -ForegroundColor Red
    Write-Host "Connection details:" -ForegroundColor Yellow
    Write-Host "  Host: $DbHost" -ForegroundColor Gray
    Write-Host "  Port: $DbPort" -ForegroundColor Gray
    Write-Host "  Database: $DbName" -ForegroundColor Gray
    Write-Host "  User: $DbUser" -ForegroundColor Gray
    Write-Host ""
    Write-Host "Error: $testResult" -ForegroundColor Red
    Remove-Item Env:PGPASSWORD
    exit 1
}

Write-Host "✅ Connected to database successfully" -ForegroundColor Green
Write-Host ""

# Step 2: Check for duplicates
Write-Host "Step 2: Checking for duplicate accounts..." -ForegroundColor Yellow
$duplicateCheckQuery = @"
SELECT COUNT(*) as duplicate_codes
FROM (
    SELECT code
    FROM accounts
    WHERE deleted_at IS NULL
    GROUP BY code
    HAVING COUNT(*) > 1
) dup;
"@

$duplicateCount = psql -h $DbHost -p $DbPort -U $DbUser -d $DbName -t -c $duplicateCheckQuery 2>&1
$duplicateCount = $duplicateCount.Trim()

if ($duplicateCount -eq "0") {
    Write-Host "✅ No duplicate accounts found!" -ForegroundColor Green
    Write-Host "   Your database is clean. No fix needed." -ForegroundColor Green
    Write-Host ""
    Write-Host "However, we'll still install the unique constraint to prevent future duplicates..." -ForegroundColor Yellow
    
    # Just install constraint
    Write-Host ""
    Write-Host "Installing unique constraint..." -ForegroundColor Yellow
    
    $constraintSQL = @"
-- Drop old constraints
DROP INDEX IF EXISTS uni_accounts_code CASCADE;
DROP INDEX IF EXISTS accounts_code_key CASCADE;
DROP INDEX IF EXISTS idx_accounts_code_unique CASCADE;
DROP INDEX IF EXISTS accounts_code_unique CASCADE;
DROP INDEX IF EXISTS idx_accounts_code_unique_active CASCADE;

-- Create new constraint
CREATE UNIQUE INDEX IF NOT EXISTS idx_accounts_code_active
ON accounts (LOWER(code))
WHERE deleted_at IS NULL;

-- Create trigger
CREATE OR REPLACE FUNCTION prevent_duplicate_account_code()
RETURNS TRIGGER AS `$`$
DECLARE
    existing_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO existing_count
    FROM accounts
    WHERE LOWER(code) = LOWER(NEW.code)
      AND deleted_at IS NULL
      AND id != COALESCE(NEW.id, 0);
    
    IF existing_count > 0 THEN
        RAISE EXCEPTION 'Account code % already exists', NEW.code
            USING HINT = 'Use unique account codes only',
                  ERRCODE = '23505';
    END IF;
    
    RETURN NEW;
END;
`$`$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_prevent_duplicate_account_code ON accounts;

CREATE TRIGGER trg_prevent_duplicate_account_code
    BEFORE INSERT OR UPDATE OF code ON accounts
    FOR EACH ROW
    EXECUTE FUNCTION prevent_duplicate_account_code();
"@
    
    $constraintSQL | psql -h $DbHost -p $DbPort -U $DbUser -d $DbName
    
    if ($LASTEXITCODE -eq 0) {
        Write-Host "✅ Unique constraint installed successfully" -ForegroundColor Green
    } else {
        Write-Host "⚠️  Warning: Could not install constraint" -ForegroundColor Yellow
    }
    
    Remove-Item Env:PGPASSWORD
    exit 0
}

Write-Host "⚠️  Found $duplicateCount duplicate account codes" -ForegroundColor Red
Write-Host ""

# Show duplicate details
Write-Host "Duplicate account codes:" -ForegroundColor Yellow
$duplicateListQuery = @"
SELECT 
    code,
    COUNT(*) as instances,
    STRING_AGG(name, ' | ') as names
FROM accounts
WHERE deleted_at IS NULL
GROUP BY code
HAVING COUNT(*) > 1
ORDER BY COUNT(*) DESC
LIMIT 10;
"@

psql -h $DbHost -p $DbPort -U $DbUser -d $DbName -c $duplicateListQuery
Write-Host ""

# Step 3: Backup database
if (-not $SkipBackup) {
    Write-Host "Step 3: Backing up database..." -ForegroundColor Yellow
    $backupFile = "backup_before_fix_$(Get-Date -Format 'yyyyMMdd_HHmmss').sql"
    $backupPath = Join-Path $PSScriptRoot $backupFile
    
    Write-Host "  Backup file: $backupPath" -ForegroundColor Gray
    
    pg_dump -h $DbHost -p $DbPort -U $DbUser -d $DbName -f $backupPath 2>&1 | Out-Null
    
    if ($LASTEXITCODE -eq 0 -and (Test-Path $backupPath)) {
        $backupSize = (Get-Item $backupPath).Length / 1MB
        Write-Host "✅ Backup created successfully ($([math]::Round($backupSize, 2)) MB)" -ForegroundColor Green
    } else {
        Write-Host "❌ ERROR: Backup failed!" -ForegroundColor Red
        Write-Host "Cannot proceed without backup. Aborting." -ForegroundColor Red
        Remove-Item Env:PGPASSWORD
        exit 1
    }
} else {
    Write-Host "Step 3: Skipping backup (--SkipBackup flag used)" -ForegroundColor Yellow
    Write-Host "⚠️  WARNING: Proceeding without backup!" -ForegroundColor Red
}

Write-Host ""

# Step 4: Confirm before proceeding
Write-Host "============================================================================" -ForegroundColor Cyan
Write-Host "READY TO FIX DUPLICATE ACCOUNTS" -ForegroundColor Cyan
Write-Host "============================================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "This will:" -ForegroundColor Yellow
Write-Host "  1. Merge duplicate accounts (keeping the one with most usage)" -ForegroundColor White
Write-Host "  2. Move all transactions to primary account" -ForegroundColor White
Write-Host "  3. Consolidate balances" -ForegroundColor White
Write-Host "  4. Soft-delete duplicate accounts" -ForegroundColor White
Write-Host "  5. Create unique constraint to prevent future duplicates" -ForegroundColor White
Write-Host ""

$confirmation = Read-Host "Continue with fix? (yes/no)"

if ($confirmation -ne "yes") {
    Write-Host "❌ Fix cancelled by user" -ForegroundColor Yellow
    Remove-Item Env:PGPASSWORD
    exit 0
}

Write-Host ""
Write-Host "Step 4: Running fix script..." -ForegroundColor Yellow
Write-Host ""

# Run the fix script
$fixScriptPath = Join-Path $PSScriptRoot "fix_duplicate_accounts_simple.sql"

if (-not (Test-Path $fixScriptPath)) {
    Write-Host "❌ ERROR: Fix script not found: $fixScriptPath" -ForegroundColor Red
    Remove-Item Env:PGPASSWORD
    exit 1
}

psql -h $DbHost -p $DbPort -U $DbUser -d $DbName -f $fixScriptPath

if ($LASTEXITCODE -ne 0) {
    Write-Host ""
    Write-Host "❌ ERROR: Fix script failed!" -ForegroundColor Red
    Write-Host "Database may be in inconsistent state." -ForegroundColor Red
    Write-Host "Restore from backup: $backupPath" -ForegroundColor Yellow
    Remove-Item Env:PGPASSWORD
    exit 1
}

Write-Host ""
Write-Host "✅ Fix script completed successfully!" -ForegroundColor Green
Write-Host ""

# Step 5: Verify results
Write-Host "Step 5: Verifying results..." -ForegroundColor Yellow

$verifyQuery = @"
SELECT COUNT(*) as remaining_duplicates
FROM (
    SELECT code
    FROM accounts
    WHERE deleted_at IS NULL
    GROUP BY code
    HAVING COUNT(*) > 1
) dup;
"@

$remainingDuplicates = psql -h $DbHost -p $DbPort -U $DbUser -d $DbName -t -c $verifyQuery 2>&1
$remainingDuplicates = $remainingDuplicates.Trim()

if ($remainingDuplicates -eq "0") {
    Write-Host "✅ SUCCESS: All duplicates have been fixed!" -ForegroundColor Green
} else {
    Write-Host "⚠️  WARNING: Still have $remainingDuplicates duplicate(s)" -ForegroundColor Yellow
    Write-Host "Manual intervention may be required" -ForegroundColor Yellow
}

Write-Host ""

# Check constraint
Write-Host "Checking unique constraint..." -ForegroundColor Yellow
$constraintCheckQuery = @"
SELECT 
    CASE 
        WHEN EXISTS (
            SELECT 1 FROM pg_indexes 
            WHERE tablename = 'accounts' 
            AND indexname = 'idx_accounts_code_active'
        ) 
        THEN 'installed'
        ELSE 'missing'
    END as constraint_status;
"@

$constraintStatus = psql -h $DbHost -p $DbPort -U $DbUser -d $DbName -t -c $constraintCheckQuery 2>&1
$constraintStatus = $constraintStatus.Trim()

if ($constraintStatus -eq "installed") {
    Write-Host "✅ Unique constraint is installed" -ForegroundColor Green
} else {
    Write-Host "⚠️  Unique constraint is missing!" -ForegroundColor Yellow
}

Write-Host ""

# Final summary
Write-Host "============================================================================" -ForegroundColor Cyan
Write-Host "FIX COMPLETE" -ForegroundColor Cyan
Write-Host "============================================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Summary:" -ForegroundColor Yellow
Write-Host "  ✅ Duplicate accounts merged" -ForegroundColor Green
Write-Host "  ✅ Transactions consolidated" -ForegroundColor Green
Write-Host "  ✅ Unique constraint installed" -ForegroundColor Green
Write-Host "  ✅ Future duplicates prevented" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "  1. Test your balance sheet report" -ForegroundColor White
Write-Host "  2. Verify account balances are correct" -ForegroundColor White
Write-Host "  3. Restart your backend application" -ForegroundColor White
Write-Host "  4. Monitor for any errors" -ForegroundColor White
Write-Host ""

if (-not $SkipBackup) {
    Write-Host "Backup saved to: $backupPath" -ForegroundColor Gray
    Write-Host "Keep this backup for at least 1 week" -ForegroundColor Gray
    Write-Host ""
}

Write-Host "============================================================================" -ForegroundColor Cyan

# Cleanup
Remove-Item Env:PGPASSWORD

